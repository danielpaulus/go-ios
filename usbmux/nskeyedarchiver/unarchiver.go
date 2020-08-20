package nskeyedarchiver

import (
	"fmt"
	"reflect"

	plist "howett.net/plist"
)

//Unarchive extracts NSKeyedArchiver Plists, either in XML or Binary format, and returns an array of the archived objects converted to usable Go Types.
// Primitives will be extracted just like regular Plist primitives (string, float64, int64, []uint8 etc.).
// NSArray, NSMutableArray, NSSet and NSMutableSet will transformed into []interface{}
// NSDictionary and NSMutableDictionary will be transformed into map[string] interface{}. I might add non string keys later.
func Unarchive(xml []byte) ([]interface{}, error) {
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
	//convert the Dictionary with the objectReferences into a flat list of UIDs, so we can reuse the extractObjects function later
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
		if object, ok := isArrayObject(objectRef.(map[string]interface{}), objects); ok {
			extractObjects, err := extractObjects(toUidList(object[nsObjects].([]interface{})), objects)
			if err != nil {
				return nil, err
			}
			returnValue[i] = extractObjects
			continue
		}

		if object, ok := isDictionaryObject(objectRef.(map[string]interface{}), objects); ok {
			dictionary, err := extractDictionary(object, objects)
			if err != nil {
				return nil, err
			}
			returnValue[i] = dictionary
			continue
		}

		objectType := reflect.TypeOf(objectRef).String()
		return nil, fmt.Errorf("Unknown object type:%s", objectType)

	}
	return returnValue, nil
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
	return object, false
}
