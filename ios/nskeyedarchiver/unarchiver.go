package nskeyedarchiver

import (
	"fmt"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

// Unarchive extracts NSKeyedArchiver Plists, either in XML or Binary format, and returns an array of the archived objects converted to usable Go Types.
// Primitives will be extracted just like regular Plist primitives (string, float64, int64, []uint8 etc.).
// NSArray, NSMutableArray, NSSet and NSMutableSet will transformed into []interface{}
// NSDictionary and NSMutableDictionary will be transformed into map[string] interface{}. I might add non string keys later.
func Unarchive(xml []byte) (result []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			err = fmt.Errorf("Unarchive: %s\n%s", r, stacktrace)
		}
	}()

	SetupDecoders()
	plist, err := plistFromBytes(xml)
	if err != nil {
		return nil, err
	}
	nsKeyedArchiverData := plist.(map[string]interface{})

	err = verifyCorrectArchiver(nsKeyedArchiverData)
	if err != nil {
		return nil, err
	}
	return extractObjectsFromTop(nsKeyedArchiverData[topKey].(map[string]interface{}), nsKeyedArchiverData[objectsKey].([]interface{}))
}

func extractObjectsFromTop(top map[string]interface{}, objects []interface{}) ([]interface{}, error) {
	objectCount := len(top)
	if root, ok := top["root"]; ok {
		return extractObjects([]plist.UID{root.(plist.UID)}, objects)
	}
	objectRefs := make([]plist.UID, objectCount)
	// convert the Dictionary with the objectReferences into a flat list of UIDs, so we can reuse the extractObjects function later
	for i := 0; i < objectCount; i++ {
		objectIndex := top[fmt.Sprintf("$%d", i)].(plist.UID)
		objectRefs[i] = objectIndex
	}
	return extractObjects(objectRefs, objects)
}

func extractObjects(objectRefs []plist.UID, objects []interface{}) ([]interface{}, error) {
	objectCount := len(objectRefs)
	returnValue := make([]interface{}, objectCount)
	for i := 0; i < objectCount; i++ {
		objectIndex := objectRefs[i]
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
			extractObjects, err := extractObjects(toUidList(object[nsObjects].([]interface{})), objects)
			if err != nil {
				return nil, err
			}
			returnValue[i] = extractObjects
			continue
		}

		if object, ok := isDictionaryObject(nonPrimitiveObjectRef, objects); ok {
			dictionary, err := extractDictionary(object, objects)
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

func extractDictionary(object map[string]interface{}, objects []interface{}) (map[string]interface{}, error) {
	keyRefs := toUidList(object[nsKeys].([]interface{}))
	keys, err := extractObjects(keyRefs, objects)
	if err != nil {
		return nil, err
	}

	valueRefs := toUidList(object[nsObjects].([]interface{}))
	values, err := extractObjects(valueRefs, objects)
	if err != nil {
		return nil, err
	}
	mapSize := len(keys)
	result := make(map[string]interface{}, mapSize)
	if mapSize == 0 {
		return result, nil
	}
	if _, ok := keys[0].(string); !ok {
		log.Warn("non string key dict found, lazy decoding by converting keys to strings :-), fix later")
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
