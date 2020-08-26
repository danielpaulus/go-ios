package nskeyedarchiver

import (
	"fmt"
	"log"
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
	archiverSkeleton := createSkeleton(true)
	objects := make([]interface{}, 1)
	objects[0] = null
	objects, _ = archive(object, objects)

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

	log.Fatal(fmt.Errorf("Unsupported type:%s", object))
	return nil, 0
}

func serializeArray(array []interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	arrayDict := map[string]interface{}{}
	index := len(objects)
	objects = append(objects, arrayDict)

	index = len(objects)
	objects = append(objects, arrayClassDefinition())
	arrayDict["$class"] = plist.UID(index)
	itemRefs := make([]plist.UID, len(array))
	for index, item := range array {
		var uid plist.UID
		objects, uid = archive(item, objects)
		itemRefs[index] = uid
	}
	arrayDict["NS.objects"] = itemRefs
	return objects, plist.UID(index)
}

func arrayClassDefinition() map[string]interface{} {
	return map[string]interface{}{"$classes": []string{"NSArray", "NSObject"}, "$classname": "NSArray"}
}

func isArray(object interface{}) bool {
	return reflect.TypeOf(object).Kind() == reflect.Array
}

func isMap(object interface{}) bool {
	return reflect.TypeOf(object).Kind() == reflect.Map
}
