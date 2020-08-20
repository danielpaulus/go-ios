package nskeyedarchiver

import (
	"fmt"

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
	if obj, ok := isPrimitiveObject(object); ok {
		archiverSkeleton := createSkeleton(2, true)
		archiverSkeleton[objectsKey].([]interface{})[1] = obj
		return archiverSkeleton, nil
	}

	return nil, fmt.Errorf("Unsupported type")
}

func createSkeleton(size int, withRoot bool) map[string]interface{} {
	var topDict map[string]interface{}
	if withRoot {
		topDict = map[string]interface{}{"root": plist.UID(1)}
	}
	objectsArray := make([]interface{}, size)
	objectsArray[0] = null
	return map[string]interface{}{
		versionKey:  versionValue,
		objectsKey:  objectsArray,
		archiverKey: nsKeyedArchiver,
		topKey:      topDict,
	}
}
