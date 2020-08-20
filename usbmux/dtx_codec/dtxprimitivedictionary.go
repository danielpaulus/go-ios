package dtx

import (
	"container/list"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	archiver "github.com/danielpaulus/nskeyedarchiver"
)

// That is by far the weirdest concept I have ever seen.
// Looking at disassembled code you can see this is a custom dictionary type
// only used for DTX. In practice however, the keys are always null and the
// values are used as a simple array containing the method arguments for the
// method this message is invoking. (The payload object usually contains method names or returnvalues)
type DtxPrimitiveDictionary struct {
	keyValuePairs *list.List
	values        []interface{}
	valueTypes    []uint32
}

type DtxPrimitiveKeyValuePair struct {
	keyType   uint32
	key       interface{}
	valueType uint32
	value     interface{}
}

func (d DtxPrimitiveDictionary) String() string {
	result := "["
	for i, v := range d.valueTypes {
		var prettyString []byte
		if v == bytearray {
			bytes := d.values[i].([]byte)
			prettyString = bytes
			msg, err := archiver.Unarchive(bytes)
			if err == nil {
				prettyString, _ = json.Marshal(msg)
			}
			result += fmt.Sprintf("{t:%s, v:%s},\n", toString(v), prettyString)
			continue
		}
		if v == t_uint32 {
			result += fmt.Sprintf("{t:%s, v:%d},\n", toString(v), d.values[i])
			continue
		}
		result += fmt.Sprintf("{t:%s, v:%s},\n", toString(v), d.values[i])
	}
	result += "]"
	return result
}

func decodeAuxiliary(auxBytes []byte) DtxPrimitiveDictionary {
	result := DtxPrimitiveDictionary{}
	result.keyValuePairs = list.New()
	for {
		keyType, key, remainingBytes := readEntry(auxBytes)
		auxBytes = remainingBytes
		valueType, value, remainingBytes := readEntry(auxBytes)
		auxBytes = remainingBytes
		pair := DtxPrimitiveKeyValuePair{keyType, key, valueType, value}
		result.keyValuePairs.PushBack(pair)
		if len(auxBytes) == 0 {
			break
		}
	}

	size := result.keyValuePairs.Len()

	result.valueTypes = make([]uint32, size)
	result.values = make([]interface{}, size)

	e := result.keyValuePairs.Front()
	for i := 0; i < size; i++ {
		result.valueTypes[i] = e.Value.(DtxPrimitiveKeyValuePair).valueType
		result.values[i] = e.Value.(DtxPrimitiveKeyValuePair).value
	}

	return result
}

func readEntry(auxBytes []byte) (uint32, interface{}, []byte) {
	readType := binary.LittleEndian.Uint32(auxBytes)
	if readType == null {
		return null, nil, auxBytes[4:]
	}
	if readType == t_uint32 {
		return t_uint32, auxBytes[4:8], auxBytes[8:]
	}
	if hasLength(readType) {
		length := binary.LittleEndian.Uint32(auxBytes[4:])
		data := auxBytes[8 : 8+length]
		return readType, data, auxBytes[8+length:]
	}
	log.Fatalf("Unknown DtxPrimitiveDictionaryType: %d  rawbytes:%x", readType, auxBytes)
	return 0, nil, nil
}

const (
	null      uint32 = 0x0A
	bytearray uint32 = 0x02
	t_uint32  uint32 = 0x03
)

func toString(t uint32) string {
	switch t {
	case null:
		return "null"
	case bytearray:
		return "binary"
	case t_uint32:
		return "uint32"
	default:
		return "unknown"
	}
}

func hasLength(typeCode uint32) bool {
	return typeCode == bytearray
}
