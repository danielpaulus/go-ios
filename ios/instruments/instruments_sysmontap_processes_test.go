package instruments

import (
	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

// keyedArchive writes minimal NSKeyedArchiver fixtures. The go-ios archiver
// only writes string-keyed dictionaries, but sysmontap's "Processes"
// dictionary is keyed by NSNumber pids, as on a real device.
type keyedArchive struct {
	objects []interface{}
	classes map[string]plist.UID
}

// archivedDict is an NSDictionary with keys of any type, in order.
type archivedDict struct {
	keys   []interface{}
	values []interface{}
}

// archivedNull is an NSNull value.
type archivedNull struct{}

func archiveForTest(t *testing.T, root interface{}) []byte {
	t.Helper()
	a := &keyedArchive{objects: []interface{}{"$null"}, classes: map[string]plist.UID{}}
	rootUID := a.value(root)
	b, err := plist.Marshal(map[string]interface{}{
		"$version":  100000,
		"$archiver": "NSKeyedArchiver",
		"$top":      map[string]interface{}{"root": rootUID},
		"$objects":  a.objects,
	}, plist.BinaryFormat)
	require.NoError(t, err)
	return b
}

func (a *keyedArchive) add(object interface{}) plist.UID {
	a.objects = append(a.objects, object)
	return plist.UID(len(a.objects) - 1)
}

func (a *keyedArchive) class(name string) plist.UID {
	if uid, ok := a.classes[name]; ok {
		return uid
	}
	uid := a.add(map[string]interface{}{"$classname": name, "$classes": []interface{}{name, "NSObject"}})
	a.classes[name] = uid
	return uid
}

func (a *keyedArchive) value(v interface{}) plist.UID {
	switch typed := v.(type) {
	case []interface{}:
		refs := make([]interface{}, len(typed))
		for i, item := range typed {
			refs[i] = a.value(item)
		}
		return a.add(map[string]interface{}{"NS.objects": refs, "$class": a.class("NSArray")})
	case archivedDict:
		keys := make([]interface{}, len(typed.keys))
		values := make([]interface{}, len(typed.values))
		for i := range typed.keys {
			keys[i] = a.value(typed.keys[i])
			values[i] = a.value(typed.values[i])
		}
		return a.add(map[string]interface{}{"NS.keys": keys, "NS.objects": values, "$class": a.class("NSDictionary")})
	case archivedNull:
		return a.add(map[string]interface{}{"$class": a.class("NSNull")})
	default:
		return a.add(typed)
	}
}

var testProcessAttributes = []string{"pid", "name", "cpuUsage", "physFootprint"}

// processDataMessage builds a sysmontap message the way the device sends
// process rows: a DTX data message whose payload is the raw archive of an
// NSArray holding one row with a pid-keyed "Processes" dictionary.
func processDataMessage(t *testing.T, processes archivedDict) dtx.Message {
	row := archivedDict{
		keys:   []interface{}{"Processes", "StartMachAbsTime", "EndMachAbsTime", "Type"},
		values: []interface{}{processes, uint64(100), uint64(200), uint64(5)},
	}
	return dtx.Message{Payload: []interface{}{archiveForTest(t, []interface{}{row})}}
}

func TestMapToProcessSnapshot_ArchivedDataMessage(t *testing.T) {
	msg := processDataMessage(t, archivedDict{
		keys: []interface{}{uint64(200), uint64(1)},
		values: []interface{}{
			[]interface{}{uint64(200), "SpringBoard", 2.5, uint64(81462792)},
			[]interface{}{uint64(1), "launchd", archivedNull{}, uint64(4096)},
		},
	})

	snapshot, ok, err := mapToProcessSnapshot(msg, testProcessAttributes)
	require.NoError(t, err)
	require.True(t, ok)

	assert.Equal(t, uint64(100), snapshot.StartMachAbsTime)
	assert.Equal(t, uint64(200), snapshot.EndMachAbsTime)
	require.Len(t, snapshot.Processes, 2)

	launchd := snapshot.Processes[0]
	assert.Equal(t, uint64(1), launchd.Pid)
	assert.Equal(t, "launchd", launchd.Name)
	assert.Contains(t, launchd.Attributes, "cpuUsage")
	assert.Nil(t, launchd.Attributes["cpuUsage"], "NSNull must decode to nil")

	springboard := snapshot.Processes[1]
	assert.Equal(t, uint64(200), springboard.Pid)
	assert.Equal(t, "SpringBoard", springboard.Name)
	assert.Equal(t, map[string]interface{}{
		"pid":           uint64(200),
		"name":          "SpringBoard",
		"cpuUsage":      2.5,
		"physFootprint": uint64(81462792),
	}, springboard.Attributes)
}

func TestMapToProcessSnapshot_DecodedPayload(t *testing.T) {
	msg := dtx.Message{Payload: []interface{}{[]interface{}{
		map[string]interface{}{
			"Processes": map[string]interface{}{
				"uint64{7}": []interface{}{uint64(7), "backboardd", 1.0, uint64(1024)},
			},
		},
	}}}

	snapshot, ok, err := mapToProcessSnapshot(msg, testProcessAttributes)
	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, snapshot.Processes, 1)
	assert.Equal(t, uint64(7), snapshot.Processes[0].Pid)
	assert.Equal(t, "backboardd", snapshot.Processes[0].Name)
}

func TestMapToProcessSnapshot_MessagesWithoutProcesses(t *testing.T) {
	testCases := map[string]dtx.Message{
		"system row":    buildSysmontapMsg(validResultMap()),
		"heartbeat":     {Payload: []interface{}{nskeyedarchiver.DTTapHeartbeatMessage{}}},
		"empty payload": {},
		"archived system row": {Payload: []interface{}{archiveForTest(t, []interface{}{
			archivedDict{keys: []interface{}{"Type"}, values: []interface{}{uint64(41)}},
		})}},
	}
	for name, msg := range testCases {
		t.Run(name, func(t *testing.T) {
			_, ok, err := mapToProcessSnapshot(msg, testProcessAttributes)
			require.NoError(t, err)
			assert.False(t, ok)
		})
	}
}

func TestMapToProcessSnapshot_Errors(t *testing.T) {
	testCases := map[string]dtx.Message{
		"data message is not an archive": {Payload: []interface{}{[]byte("not an archive")}},
		"value count does not match attributes": processDataMessage(t, archivedDict{
			keys:   []interface{}{uint64(1)},
			values: []interface{}{[]interface{}{uint64(1), "launchd"}},
		}),
		"pid is not numeric": processDataMessage(t, archivedDict{
			keys:   []interface{}{uint64(1)},
			values: []interface{}{[]interface{}{"one", "launchd", 1.0, uint64(1)}},
		}),
		"process values are not an array": processDataMessage(t, archivedDict{
			keys:   []interface{}{uint64(1)},
			values: []interface{}{"launchd"},
		}),
	}
	for name, msg := range testCases {
		t.Run(name, func(t *testing.T) {
			_, ok, err := mapToProcessSnapshot(msg, testProcessAttributes)
			assert.Error(t, err)
			assert.False(t, ok)
		})
	}
}

func TestAttributeNames(t *testing.T) {
	got := attributeNames([]interface{}{"pid", uint64(3), "name"})
	assert.Equal(t, []string{"pid", "", "name"}, got)
}
