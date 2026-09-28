package nskeyedarchiver

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/danielpaulus/go-ios/ios/golog"
	plist "howett.net/plist"
)

// Unarchive extracts NSKeyedArchiver Plists, either in XML or Binary format, and returns an array of the archived objects converted to usable Go Types.
// Primitives will be extracted just like regular Plist primitives (string, float64, int64, []uint8 etc.).
// NSArray, NSMutableArray, NSSet and NSMutableSet will transformed into []interface{}
// NSDictionary and NSMutableDictionary will be transformed into map[string] interface{}. I might add non string keys later.
func Unarchive(xml []byte) (result []interface{}, err error) {
	return UnarchiveReader(bytes.NewReader(xml))
}

// UnarchiveReader extracts NSKeyedArchiver Plists from an io.ReadSeeker, either in XML or Binary format, and returns an array of the archived objects converted to usable Go Types.
// Primitives will be extracted just like regular Plist primitives (string, float64, int64, []uint8 etc.).
// NSArray, NSMutableArray, NSSet and NSMutableSet will be transformed into []interface{}.
// NSDictionary and NSMutableDictionary will be transformed into map[string]interface{}. Non-string keys might be added later.
func UnarchiveReader(r io.ReadSeeker) (result []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			err = fmt.Errorf("Unarchive: %s\n%s", r, stacktrace)
		}
	}()

	SetupDecoders()
	plist, err := plistFromReader(r)
	if err != nil {
		return nil, err
	}
	nsKeyedArchiverData, ok := plist.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid NSKeyedArchiver plist: root is not a dictionary, got %T", plist)
	}

	err = verifyCorrectArchiver(nsKeyedArchiverData)
	if err != nil {
		return nil, err
	}
	top, ok := nsKeyedArchiverData[topKey].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid NSKeyedArchiver plist: '%s' is not a dictionary, got %T", topKey, nsKeyedArchiverData[topKey])
	}
	objects, ok := nsKeyedArchiverData[objectsKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid NSKeyedArchiver plist: '%s' is not an array, got %T", objectsKey, nsKeyedArchiverData[objectsKey])
	}
	return extractObjectsFromTop(top, objects)
}

// maxUnarchiveDepth bounds how deeply extractObjects will recurse into nested
// NSArray/NSSet/NSDictionary containers. A hostile archive whose NS.objects UIDs
// form a cycle would otherwise recurse forever and trigger an uncatchable Go
// "fatal error: stack overflow" that the recover() above cannot backstop. The
// bound is far above anything a well-formed archive produces, so valid archives
// decode identically.
const maxUnarchiveDepth = 1000

func extractObjectsFromTop(top map[string]interface{}, objects []interface{}) ([]interface{}, error) {
	objectCount := len(top)
	if root, ok := top["root"]; ok {
		rootUID, ok := root.(plist.UID)
		if !ok {
			return nil, fmt.Errorf("invalid NSKeyedArchiver plist: 'root' is not a UID, got %T", root)
		}
		return extractObjects([]plist.UID{rootUID}, objects, 0)
	}
	objectRefs := make([]plist.UID, objectCount)
	// convert the Dictionary with the objectReferences into a flat list of UIDs, so we can reuse the extractObjects function later
	for i := 0; i < objectCount; i++ {
		objectIndex, ok := top[fmt.Sprintf("$%d", i)].(plist.UID)
		if !ok {
			return nil, fmt.Errorf("invalid NSKeyedArchiver plist: '$top' entry $%d is not a UID, got %T", i, top[fmt.Sprintf("$%d", i)])
		}
		objectRefs[i] = objectIndex
	}
	return extractObjects(objectRefs, objects, 0)
}

// extractObjects resolves a list of object UIDs into Go values. depth tracks how
// deep we have recursed into nested containers; it is bounded by
// maxUnarchiveDepth to make a cyclic/self-referential archive fail with an error
// instead of an uncatchable stack overflow (see CRIT-2).
func extractObjects(objectRefs []plist.UID, objects []interface{}, depth int) ([]interface{}, error) {
	if depth > maxUnarchiveDepth {
		return nil, fmt.Errorf("max unarchive depth %d exceeded, aborting to avoid stack overflow (possibly a cyclic or maliciously nested archive)", maxUnarchiveDepth)
	}
	objectCount := len(objectRefs)
	returnValue := make([]interface{}, objectCount)
	for i := 0; i < objectCount; i++ {
		objectIndex := objectRefs[i]
		if int(objectIndex) < 0 || int(objectIndex) >= len(objects) {
			return nil, fmt.Errorf("object UID %d out of range for $objects of length %d", objectIndex, len(objects))
		}
		objectRef := objects[objectIndex]
		if object, ok := isPrimitiveObject(objectRef); ok {
			returnValue[i] = object
			continue
		}
		// if this crashes, I forgot a primitive type
		nonPrimitiveObjectRef, ok := objectRef.(map[string]interface{})
		if !ok {
			return []interface{}{}, fmt.Errorf("object not a dictionary: %+v", objectRef)
		}
		if object, ok := isArrayObject(nonPrimitiveObjectRef, objects); ok {
			nestedRefs, ok := object[nsObjects].([]interface{})
			if !ok {
				return nil, fmt.Errorf("NS.objects is not an array, got %T", object[nsObjects])
			}
			extractObjects, err := extractObjects(toUidList(nestedRefs), objects, depth+1)
			if err != nil {
				return nil, err
			}
			returnValue[i] = extractObjects
			continue
		}

		if object, ok := isDictionaryObject(nonPrimitiveObjectRef, objects); ok {
			dictionary, err := extractDictionary(object, objects, depth+1)
			if err != nil {
				return nil, err
			}
			returnValue[i] = dictionary
			continue
		}

		if object, ok := isNSMutableDataObject(nonPrimitiveObjectRef, objects); ok {
			returnValue[i] = object[nsDataKey]
			continue
		}

		if object, ok := isNSMutableString(nonPrimitiveObjectRef, objects); ok {
			returnValue[i] = object[nsStringKey]
			continue
		}

		obj, err := decodeNonstandardObject(nonPrimitiveObjectRef, objects)
		if err != nil {
			return nil, err
		}
		returnValue[i] = obj

	}
	return returnValue, nil
}

func decodeNonstandardObject(object map[string]interface{}, objects []interface{}) (interface{}, error) {
	className, err := resolveClass(object[class], objects)
	if err != nil {
		return nil, err
	}
	factory := decodableClasses[className]
	if factory == nil {
		return nil, fmt.Errorf("Unknown class:%s", className)
	}
	return factory(object, objects), nil
}

func isArrayObject(object map[string]interface{}, objects []interface{}) (map[string]interface{}, bool) {
	className, err := resolveClass(object[class], objects)
	if err != nil {
		return nil, false
	}
	if className == nsArray || className == nsMutableArray || className == nsSet || className == nsMutableSet {
		return object, true
	}
	return object, false
}

func isDictionaryObject(object map[string]interface{}, objects []interface{}) (map[string]interface{}, bool) {
	className, err := resolveClass(object[class], objects)
	if err != nil {
		return nil, false
	}
	if className == nsDictionary || className == nsMutableDictionary {
		return object, true
	}
	return object, false
}

func isNSMutableDataObject(object map[string]interface{}, objects []interface{}) (map[string]interface{}, bool) {
	className, err := resolveClass(object[class], objects)
	if err != nil {
		return nil, false
	}
	if className == nsMutableData {
		return object, true
	}
	return object, false
}

func isNSMutableString(object map[string]interface{}, objects []interface{}) (map[string]interface{}, bool) {
	className, err := resolveClass(object[class], objects)
	if err != nil {
		return nil, false
	}
	if className == nsMutableString {
		return object, true
	}
	return object, false
}

func extractDictionary(object map[string]interface{}, objects []interface{}, depth int) (map[string]interface{}, error) {
	if depth > maxUnarchiveDepth {
		return nil, fmt.Errorf("max unarchive depth %d exceeded, aborting to avoid stack overflow (possibly a cyclic or maliciously nested archive)", maxUnarchiveDepth)
	}
	nsKeysList, ok := object[nsKeys].([]interface{})
	if !ok {
		return nil, fmt.Errorf("NS.keys is not an array, got %T", object[nsKeys])
	}
	keyRefs := toUidList(nsKeysList)
	keys, err := extractObjects(keyRefs, objects, depth+1)
	if err != nil {
		return nil, err
	}

	nsObjectsList, ok := object[nsObjects].([]interface{})
	if !ok {
		return nil, fmt.Errorf("NS.objects is not an array, got %T", object[nsObjects])
	}
	valueRefs := toUidList(nsObjectsList)
	values, err := extractObjects(valueRefs, objects, depth+1)
	if err != nil {
		return nil, err
	}
	mapSize := len(keys)
	result := make(map[string]interface{}, mapSize)
	if mapSize == 0 {
		return result, nil
	}
	if _, ok := keys[0].(string); !ok {
		golog.Warn("non string key dict found, lazy decoding by converting keys to strings :-), fix later", "module", logModule)
		for i := 0; i < mapSize; i++ {
			key := keys[i].(uint64)
			result[fmt.Sprintf("uint64{%d}", key)] = values[i]
		}

		return result, nil
	}

	for i := 0; i < mapSize; i++ {
		result[keys[i].(string)] = values[i]
	}

	return result, nil
}

func resolveClass(classInfo interface{}, objects []interface{}) (string, error) {
	if v, ok := classInfo.(plist.UID); ok {
		classDict := objects[v].(map[string]interface{})
		return classDict[className].(string), nil
	}
	return "", fmt.Errorf("Could not find class for %s", classInfo)
}

func isPrimitiveObject(object interface{}) (interface{}, bool) {
	if v, ok := object.(int32); ok {
		return v, ok
	}
	if v, ok := object.(int); ok {
		return v, ok
	}
	if v, ok := object.(uint64); ok {
		return v, ok
	}
	if v, ok := object.(float64); ok {
		return v, ok
	}
	if v, ok := object.(bool); ok {
		return v, ok
	}
	if v, ok := object.(string); ok {
		return v, ok
	}
	if v, ok := object.([]uint8); ok {
		return v, ok
	}
	if v, ok := object.(int64); ok {
		return v, ok
	}
	return object, false
}
