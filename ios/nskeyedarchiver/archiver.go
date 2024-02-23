package nskeyedarchiver

import (
	"fmt"
	"reflect"

	"howett.net/plist"
)

/*
type NSKeyedObject struct {
	isPrimitive bool
	primitive   interface{}
}
*/

func ArchiveXML(object interface{}) (string, error) {
	plist, err := archiveObject(object)
	if err != nil {
		return "", err
	}
	return toPlist(plist)
}

func ArchiveBin(object interface{}) ([]byte, error) {
	plist, err := archiveObject(object)
	if err != nil {
		return []byte{}, err
	}
	return toBinaryPlist(plist)
}

func archiveObject(object interface{}) (interface{}, error) {
	SetupEncoders()
	archiverSkeleton := createSkeleton(true)
	objects := make([]interface{}, 1)
	objects[0] = null
	objects, pid := archive(object, objects)
	archiverSkeleton[topKey] = map[string]interface{}{"root": pid}

	archiverSkeleton[objectsKey] = objects
	return archiverSkeleton, nil
}

func createSkeleton(withRoot bool) map[string]interface{} {
	var topDict map[string]interface{}
	if withRoot {
		topDict = map[string]interface{}{"root": plist.UID(1)}
	}

	return map[string]interface{}{
		versionKey:  versionValue,
		archiverKey: NsKeyedArchiver,
		topKey:      topDict,
	}
}

func archive(object interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	if object, ok := isPrimitiveObject(object); ok {
		index := len(objects)
		objects = append(objects, object)
		return objects, plist.UID(index)
	}

	if v, ok := object.([]interface{}); ok {
		return serializeArray(v, objects)
	}

	if v, ok := object.(map[string]interface{}); ok {
		return serializeMap(v, objects, buildClassDict("NSDictionary", "NSObject"))
	}
	typeOf := reflect.TypeOf(object)
	name := typeOf.Name()
	// seems like Name() can be empty for pointer types
	if name == "" {
		name = typeOf.String()
	}

	if encoderFunc, ok := encodableClasses[name]; ok {
		return encoderFunc(object, objects)
	}

	panic(fmt.Errorf("NSKeyedArchiver Unsupported object: '%s' of type:%s", object, typeOf))
}

func serializeArray(array []interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	arrayDict := map[string]interface{}{}
	arrayObjectIndex := len(objects)
	objects = append(objects, arrayDict)

	classDefinitionIndex := len(objects)
	objects = append(objects, buildClassDict("NSArray", "NSObject"))
	arrayDict["$class"] = plist.UID(classDefinitionIndex)
	itemRefs := make([]plist.UID, len(array))
	for index, item := range array {
		var uid plist.UID
		objects, uid = archive(item, objects)
		itemRefs[index] = uid
	}
	arrayDict["NS.objects"] = itemRefs
	return objects, plist.UID(arrayObjectIndex)
}

func serializeMutableArray(array []interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	arrayDict := map[string]interface{}{}
	arrayObjectIndex := len(objects)
	objects = append(objects, arrayDict)

	classDefinitionIndex := len(objects)
	objects = append(objects, buildClassDict("NSMutableArray", "NSArray", "NSObject"))
	arrayDict["$class"] = plist.UID(classDefinitionIndex)
	itemRefs := make([]plist.UID, len(array))
	for index, item := range array {
		var uid plist.UID
		objects, uid = archive(item, objects)
		itemRefs[index] = uid
	}
	arrayDict["NS.objects"] = itemRefs
	return objects, plist.UID(arrayObjectIndex)
}

func serializeSet(set []interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	setDict := map[string]interface{}{}
	setObjectIndex := len(objects)
	objects = append(objects, setDict)

	classDefinitionIndex := len(objects)
	objects = append(objects, buildClassDict("NSSet", "NSObject"))
	setDict["$class"] = plist.UID(classDefinitionIndex)
	itemRefs := make([]plist.UID, len(set))
	for index, item := range set {
		var uid plist.UID
		objects, uid = archive(item, objects)
		itemRefs[index] = uid
	}
	setDict["NS.objects"] = itemRefs
	return objects, plist.UID(setObjectIndex)
}

func serializeMap(mapObject map[string]interface{}, objects []interface{}, classDict map[string]interface{}) ([]interface{}, plist.UID) {
	dictDict := map[string]interface{}{}
	dictionaryRef := len(objects)
	objects = append(objects, dictDict)

	index := len(objects)
	objects = append(objects, classDict)
	dictDict["$class"] = plist.UID(index)

	keyRefs := make([]plist.UID, len(mapObject))

	i := 0
	keys := make([]string, len(mapObject))
	for k := range mapObject {
		keys[i] = k
		i++
	}

	index = 0
	for _, key := range keys {
		var uid plist.UID
		objects, uid = archive(key, objects)
		keyRefs[index] = uid
		index++
	}
	dictDict["NS.keys"] = keyRefs

	index = 0
	valueRefs := make([]plist.UID, len(mapObject))
	for _, key := range keys {
		var uid plist.UID
		objects, uid = archive(mapObject[key], objects)
		valueRefs[index] = uid
		index++
	}
	dictDict["NS.objects"] = valueRefs

	return objects, plist.UID(dictionaryRef)
}

func isArray(object interface{}) bool {
	return reflect.TypeOf(object).Kind() == reflect.Array
}

func isMap(object interface{}) bool {
	return reflect.TypeOf(object).Kind() == reflect.Map
}
