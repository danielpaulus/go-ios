package dtx

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	archiver "github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

// PrimitiveDictionary contains a custom dictionary type
// only used for DTX. In practice however, the keys of all dictionaries are always null and the
// values are used as a simple array containing the method arguments for the
// method this message is invoking. (The payload object usually contains method names or returnvalues)
type PrimitiveDictionary struct {
	keyValuePairs *list.List
	values        []interface{}
	valueTypes    []uint32
}

type PrimitiveKeyValuePair struct {
	keyType   uint32
	key       interface{}
	valueType uint32
	value     interface{}
}

func NewPrimitiveDictionary() PrimitiveDictionary {
	return PrimitiveDictionary{keyValuePairs: list.New()}
}

func (d *PrimitiveDictionary) AddInt32(value int) {
	d.keyValuePairs.PushBack(PrimitiveKeyValuePair{t_null, nil, t_uint32, uint32(value)})
}
func (d *PrimitiveDictionary) AddBytes(value []byte) {
	d.keyValuePairs.PushBack(PrimitiveKeyValuePair{t_null, nil, t_bytearray, value})
}

func (d PrimitiveDictionary) GetArguments() []interface{} {
	return d.values
}

//AddNsKeyedArchivedObject wraps the object in a NSKeyedArchiver envelope before saving it to the dictionary as a []byte.
//This will panic on error because NSKeyedArchiver has to support everything that is put in here during runtime.
//If not, it is a non-recoverable bug and needs to be fixed anyway.
func (d *PrimitiveDictionary) AddNsKeyedArchivedObject(object interface{}) {
	archivedObject, err := nskeyedarchiver.ArchiveBin(object)
	if err != nil {
		panic(err)
	}
	d.AddBytes(archivedObject)
}

//ToBytes serializes this PrimitiveDictionary to a byte slice
func (d PrimitiveDictionary) ToBytes() ([]byte, error) {
	size := d.keyValuePairs.Len()
	if size == 0 {
		return make([]byte, 0), nil
	}
	var buf bytes.Buffer
	writer := io.Writer(&buf)

	e := d.keyValuePairs.Front()
	for i := 0; i < size; i++ {
		valuetype := e.Value.(PrimitiveKeyValuePair).valueType
		value := e.Value.(PrimitiveKeyValuePair).value
		keytype := e.Value.(PrimitiveKeyValuePair).keyType
		if keytype != t_null {
			return make([]byte, 0), fmt.Errorf("Encoding primitive dictionary keys is not supported. Unknown type: %d", keytype)
		}
		binary.Write(writer, binary.LittleEndian, t_null)
		err := writeEntry(valuetype, value, writer)
		if err != nil {
			return make([]byte, 0), err
		}
		e = e.Next()
	}

	return buf.Bytes(), nil
}

func writeEntry(valuetype uint32, value interface{}, buf io.Writer) error {
	if valuetype == t_null {
		binary.Write(buf, binary.LittleEndian, t_null)
		return nil
	}
	if valuetype == t_uint32 {
		binary.Write(buf, binary.LittleEndian, t_uint32)
		binary.Write(buf, binary.LittleEndian, value)
		return nil
	}
	if valuetype == t_bytearray {
		data := value.([]byte)
		length := uint32(len(data))
		binary.Write(buf, binary.LittleEndian, t_bytearray)
		binary.Write(buf, binary.LittleEndian, length)
		_, err := buf.Write(data)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("Unknown DtxPrimitiveDictionaryType: %d ", valuetype)
}

func (d PrimitiveDictionary) String() string {
	result := "["
	for i, v := range d.valueTypes {
		var prettyString string
		if v == t_bytearray {
			bytes := d.values[i].([]byte)
			prettyString = fmt.Sprintf("%x", bytes)
			msg, err := archiver.Unarchive(bytes)
			if err == nil {
				jsonBytes, _ := json.Marshal(msg[0])
				prettyString = string(jsonBytes)
			} else {
				log.Warnf("failed decoding with %+v", err)

			}
			result += fmt.Sprintf("{t:%s, v:%s},", toString(v), prettyString)
			continue
		}
		if v == t_string {
			result += d.values[i].(string)
		}
		if v == t_uint32 {
			result += fmt.Sprintf("{t:%s, v:%d},", toString(v), d.values[i])
			continue
		}
		result += fmt.Sprintf("{t:%s, v:%s},", toString(v), d.values[i])
	}
	result += "]"
	return result
}

func DecodeAuxiliary(auxBytes []byte) PrimitiveDictionary {
	result := PrimitiveDictionary{}
	result.keyValuePairs = list.New()
	for {
		keyType, key, remainingBytes := readEntry(auxBytes)
		auxBytes = remainingBytes
		valueType, value, remainingBytes := readEntry(auxBytes)
		auxBytes = remainingBytes
		pair := PrimitiveKeyValuePair{keyType, key, valueType, value}
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
		result.valueTypes[i] = e.Value.(PrimitiveKeyValuePair).valueType
		result.values[i] = e.Value.(PrimitiveKeyValuePair).value
		e = e.Next()
	}

	return result
}
func isNSKeyedArchiverEncoded(datatype uint32, obj interface{}) bool {
	if datatype != t_bytearray {
		return false
	}
	data := obj.([]byte)
	return bytes.Index(data, []byte(nskeyedarchiver.NsKeyedArchiver)) != -1

}
func readEntry(auxBytes []byte) (uint32, interface{}, []byte) {
	readType := binary.LittleEndian.Uint32(auxBytes)
	if readType == t_null {
		return t_null, nil, auxBytes[4:]
	}
	if readType == t_uint32 {
		return t_uint32, binary.LittleEndian.Uint32(auxBytes[4:8]), auxBytes[8:]
	}
	if readType == t_int64 {
		return t_int64, binary.LittleEndian.Uint64(auxBytes[4:12]), auxBytes[12:]
	}

	if hasLength(readType) {
		length := binary.LittleEndian.Uint32(auxBytes[4:])
		data := auxBytes[8 : 8+length]
		if readType == t_string {
			return readType, string(data), auxBytes[8+length:]
		}
		return readType, data, auxBytes[8+length:]
	}
	panic(fmt.Sprintf("Unknown DtxPrimitiveDictionaryType: %d  rawbytes:%x", readType, auxBytes))
}

const (
	t_null      uint32 = 0x0A
	t_string    uint32 = 0x01
	t_bytearray uint32 = 0x02
	t_uint32    uint32 = 0x03
	t_int64     uint32 = 0x06
)

func toString(t uint32) string {
	switch t {
	case t_null:
		return "null"
	case t_bytearray:
		return "binary"
	case t_string:
		return "string"
	case t_uint32:
		return "uint32"
	case t_int64:
		return "int64"
	default:
		return "unknown"
	}
}

func hasLength(typeCode uint32) bool {
	return typeCode == t_bytearray || typeCode == t_string
}

type AuxiliaryEncoder struct {
	buf bytes.Buffer
}

func (a *AuxiliaryEncoder) AddNsKeyedArchivedObject(object interface{}) {
	a.writeEntry(t_null, nil)
	bytes, err := archiver.ArchiveBin(object)
	if err != nil {
		panic(err)
	}
	a.writeEntry(t_bytearray, bytes)
}

func (a *AuxiliaryEncoder) writeEntry(entryType uint32, object interface{}) {

	binary.Write(&a.buf, binary.LittleEndian, entryType)
	if entryType == t_null {
		return
	}
	if entryType == t_uint32 {
		binary.Write(&a.buf, binary.LittleEndian, object.(int32))
	}
	if entryType == t_bytearray {
		binary.Write(&a.buf, binary.LittleEndian, int32(len(object.([]byte))))
		a.buf.Write(object.([]byte))

	}
	if entryType == t_string {
		binary.Write(&a.buf, binary.LittleEndian, int32(len(object.([]byte))))
		a.buf.Write([]byte(object.(string)))

	}
	panic(fmt.Sprintf("Unknown DtxPrimitiveDictionaryType: %d", entryType))

}

func (a *AuxiliaryEncoder) GetBytes() []byte {
	return a.GetBytes()
}
