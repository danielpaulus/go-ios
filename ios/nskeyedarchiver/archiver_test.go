package nskeyedarchiver_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	archiver "github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestArchiveSlice(t *testing.T) {
	option := make(map[string]interface{})
	option["name"] = "james"
	option["age"] = 20
	children := []string{"abc", "def", "ok"}
	option["children"] = children
	data, err := archiver.ArchiveXML(option)
	if err != nil {
		t.FailNow()
	}
	intf, err := archiver.Unarchive([]byte(data))
	val := intf[0].(map[string]interface{})["children"].([]interface{})
	assert.Equal(t, "abc", val[0])
	assert.Equal(t, "def", val[1])
	assert.Equal(t, "ok", val[2])
	print(val)
}

// TODO currently only partially decoding XCTestConfig is supported, fix later
func TestXCTestconfig(t *testing.T) {
	uuid := uuid.New()
	config := nskeyedarchiver.NewXCTestConfiguration("productmodulename", uuid, "targetAppBundle", "targetAppPath", "testBundleUrl", nil, nil, false, semver.MustParse("17.0.0"))
	result, err := nskeyedarchiver.ArchiveXML(config)
	if err != nil {
		log.Error(err)
		t.Fatal()
	}
	print(result)
	log.Info(uuid.String())
	res, err := nskeyedarchiver.Unarchive([]byte(result))
	assert.NoError(t, err)
	log.Info(res)

	nskeyedBytes, err := os.ReadFile("fixtures/xctestconfiguration.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestXCTCaps(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/XCTCapabilities.bin")
	if err != nil {

		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestDTCPUClusterInfo(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/dtcpuclusterinfo.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}
	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestDTTapMessage(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/dttapmessage.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}
	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestDTSysmonTap(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/DTSysmonTapMessage.bin")
	if err != nil {

		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestNSUUID(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/nsuuid.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestXCTestIdentifier(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/xctestidentifier.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestNSValue(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/nsvalue.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}
	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestWTF(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/int64-value-in-nskeyedarchive.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestXCActivityRecord(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/XCActivityRecord.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestDTTapHeartbeatMessage(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/DTTapHeartbeatMessage.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestDTTapstatusmessage(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/dttapstatusmessage.bin")
	if err != nil {
		log.Error(err)
		t.Fatal()
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

// TODO currently uint64 dicts are decoded by converting the keys to strings, might wanna fix this later
func TestIntKeyDictionary(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/uint64-key-dictionary.bin")
	if err != nil {
		t.Fatal(err)
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestArchiverStringArray(t *testing.T) {
	arr := []interface{}{"a", "n", "c"}
	b, err := nskeyedarchiver.ArchiveXML(arr)

	if assert.NoError(t, err) {
		result, err := nskeyedarchiver.Unarchive([]byte(b))
		assert.NoError(t, err)
		assert.Equal(t, arr, result[0])
	}
}

func TestArchiverEmptyArray(t *testing.T) {
	arr := []interface{}{}
	b, err := nskeyedarchiver.ArchiveXML(arr)

	if assert.NoError(t, err) {
		result, err := nskeyedarchiver.Unarchive([]byte(b))
		assert.NoError(t, err)
		assert.Equal(t, arr, result[0])
	}
}

func TestNSDate(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/ax_statechange.bin")
	if err != nil {
		t.Fatal(err)
	}

	unarchivedObject, err := archiver.Unarchive(nskeyedBytes)
	assert.NoError(t, err)
	log.Info(unarchivedObject)
}

func TestNSNull(t *testing.T) {
	nskeyedBytes, err := os.ReadFile("fixtures/ax_focus_on_element.bin")
	if err != nil {
		t.Fatal(err)
	}

	unarchivedObject, _ := archiver.Unarchive(nskeyedBytes)
	assert.Equal(t, reflect.TypeOf(unarchivedObject[0]).Name(), "NSNull")
	log.Info(unarchivedObject[0])
	expected := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n<plist version=\"1.0\"><dict><key>$archiver</key><string>NSKeyedArchiver</string><key>$objects</key><array><string>$null</string><dict><key>$class</key><dict><key>CF$UID</key><integer>2</integer></dict></dict><dict><key>$classes</key><array><string>NSNull</string><string>NSObject</string></array><key>$classname</key><string>NSNull</string></dict></array><key>$top</key><dict><key>root</key><dict><key>CF$UID</key><integer>1</integer></dict></dict><key>$version</key><integer>100000</integer></dict></plist>"
	xml, err := archiver.ArchiveXML(unarchivedObject[0])
	if assert.NoError(t, err) {
		assert.Equal(t, expected, xml)
	}
}

func TestArchiver3(t *testing.T) {
	dat, err := os.ReadFile("fixtures/payload_dump.json")
	if err != nil {
		t.Fatal(err)
	}

	var payloads []string
	json.Unmarshal([]byte(dat), &payloads)

	plistBytes, _ := hex.DecodeString(payloads[0])
	nska, err := archiver.Unarchive(plistBytes)
	value := nska[0]
	result, err := archiver.ArchiveBin(value)
	/*if assert.NoError(t, err) {
		output := convertToJSON(nska)
		print(output)
		assert.Equal(t, plistBytes, result)
		assert.NoError(t, err)
	}*/
	nska2, err2 := archiver.Unarchive(result)
	if assert.NoError(t, err2) {
		assert.Equal(t, nska2, nska)
	}
}

// TestDecoderJson tests if real DTX nsKeyedArchived plists can be decoded without error
func TestArchiver(t *testing.T) {
	dat, err := os.ReadFile("fixtures/payload_dump.json")
	if err != nil {
		t.Fatal(err)
	}

	var payloads []string
	json.Unmarshal([]byte(dat), &payloads)
	for _, plistHex := range payloads {
		plistBytes, _ := hex.DecodeString(plistHex)
		expectedNska, _ := archiver.Unarchive(plistBytes)

		archivedNskaBin, err := archiver.ArchiveBin(expectedNska[0])
		archivedNskaXml, err2 := archiver.ArchiveXML(expectedNska[0])

		if assert.NoError(t, err) && assert.NoError(t, err2) {
			actualNskaBin, unarchiverErrBin := archiver.Unarchive(archivedNskaBin)
			actualNskaXml, unarchiverErrXml := archiver.Unarchive([]byte(archivedNskaXml))
			if assert.NoError(t, unarchiverErrBin) && assert.NoError(t, unarchiverErrXml) {
				assert.Equal(t, expectedNska, actualNskaBin)
				assert.Equal(t, expectedNska, actualNskaXml)
			}
		}
	}
}

// TestDecoderJson tests if real DTX nsKeyedArchived plists can be decoded without error
func TestDecoderJson(t *testing.T) {
	dat, err := os.ReadFile("fixtures/payload_dump.json")
	if err != nil {
		t.Fatal(err)
	}

	var payloads []string
	json.Unmarshal([]byte(dat), &payloads)
	for _, plistHex := range payloads {
		plistBytes, _ := hex.DecodeString(plistHex)
		nska, err := archiver.Unarchive(plistBytes)
		output := convertToJSON(nska)
		print(output)
		assert.NoError(t, err)
	}
}

func TestDecoder(t *testing.T) {
	testCases := map[string]struct {
		filename string
		expected string
	}{
		"test one value":       {"onevalue", "[true]"},
		"test all primitives":  {"primitives", "[1,1,1,1.5,\"YXNkZmFzZGZhZHNmYWRzZg==\",true,\"Hello, World!\",\"Hello, World!\",\"Hello, World!\",false,false,42]"},
		"test arrays and sets": {"arrays", "[[1,1,1,1.5,\"YXNkZmFzZGZhZHNmYWRzZg==\",true,\"Hello, World!\",\"Hello, World!\",\"Hello, World!\",false,false,42],[true,\"Hello, World!\",42],[true],[42,true,\"Hello, World!\"]]"},
		"test nested arrays":   {"nestedarrays", "[[[true],[42,true,\"Hello, World!\"]]]"},
		"test dictionaries":    {"dict", "[{\"array\":[true,\"Hello, World!\",42],\"int\":1,\"string\":\"string\"}]"},
	}

	for _, tc := range testCases {
		dat, err := os.ReadFile("fixtures/" + tc.filename + ".xml")
		if err != nil {
			t.Fatal(err)
		}
		objects, err := archiver.Unarchive(dat)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, convertToJSON(objects))

		dat, err = os.ReadFile("fixtures/" + tc.filename + ".bin")
		if err != nil {
			t.Fatal(err)
		}
		objects, err = archiver.Unarchive(dat)
		assert.Equal(t, tc.expected, convertToJSON(objects))
	}
}

func TestValidation(t *testing.T) {
	testCases := map[string]struct {
		filename string
	}{
		"$archiver key is missing":         {"missing_archiver"},
		"$archiver is not nskeyedarchiver": {"wrong_archiver"},
		"$top key is missing":              {"missing_top"},
		"$objects key is missing":          {"missing_objects"},
		"$version key is missing":          {"missing_version"},
		"$version is wrong":                {"wrong_version"},
		"plist is invalid":                 {"broken_plist"},
	}

	for _, tc := range testCases {
		dat, err := os.ReadFile("fixtures/" + tc.filename + ".xml")
		if err != nil {
			t.Fatal(err)
		}
		_, err = archiver.Unarchive(dat)
		assert.Error(t, err)
	}
}

func convertToJSON(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}
