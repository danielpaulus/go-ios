package nskeyedarchiver

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"howett.net/plist"
)

var decodableClasses map[string]func(map[string]interface{}, []interface{}) interface{}
var encodableClasses map[string]func(object interface{}, objects []interface{}) ([]interface{}, plist.UID)

func SetupDecoders() {
	if decodableClasses == nil {
		decodableClasses = map[string]func(map[string]interface{}, []interface{}) interface{}{
			"DTActivityTraceTapMessage": NewDTActivityTraceTapMessage,
			"NSError":                   NewNSError,
			"NSNull":                    NewNSNullFromArchived,
			"NSDate":                    NewNSDate,
		}
	}
}

func SetupEncoders() {
	if encodableClasses == nil {
		encodableClasses = map[string]func(object interface{}, objects []interface{}) ([]interface{}, plist.UID){
			"XCTestConfiguration": archiveXcTestConfiguration,
			"NSUUID":              archiveNSUUID,
			"NSURL":               archiveNSURL,
			"NSNull":              archiveNSNull,
			"NSMutableDictionary": archiveNSMutableDictionary,
		}
	}
}

type XCTestConfiguration struct {
	contents map[string]interface{}
}

func NewXCTestConfiguration(
	productModuleName string,
	sessionIdentifier uuid.UUID,
	targetApplicationBundleID string,
	targetApplicationPath string,
	testBundleURL string,
) XCTestConfiguration {
	contents := map[string]interface{}{}

	contents["aggregateStatisticsBeforeCrash"] = map[string]interface{}{"XCSuiteRecordsKey": map[string]interface{}{}}
	contents["automationFrameworkPath"] = "/Developer/Library/PrivateFrameworks/XCTAutomationSupport.framework"
	contents["baselineFileRelativePath"] = plist.UID(0)
	contents["baselineFileURL"] = plist.UID(0)
	contents["defaultTestExecutionTimeAllowance"] = plist.UID(0)
	contents["disablePerformanceMetrics"] = false
	contents["emitOSLogs"] = false
	//contents["formatVersion"]= 2
	contents["gatherLocalizableStringsData"] = false
	contents["initializeForUITesting"] = true
	contents["maximumTestExecutionTimeAllowance"] = plist.UID(0)
	contents["productModuleName"] = productModuleName
	contents["randomExecutionOrderingSeed"] = plist.UID(0)
	contents["reportActivities"] = true
	contents["reportResultsToIDE"] = true
	contents["sessionIdentifier"] = NewNSUUID(sessionIdentifier)
	contents["systemAttachmentLifetime"] = 2
	//contents["targetApplicationArguments"] = []interface{}{} //TODO: triggers a bug
	contents["targetApplicationBundleID"] = targetApplicationBundleID
	//contents["targetApplicationEnvironment"] = //TODO: triggers a bug map[string]interface{}{}
	contents["targetApplicationPath"] = targetApplicationPath
	//testApplicationDependencies
	contents["testApplicationUserOverrides"] = plist.UID(0)
	contents["testBundleRelativePath"] = plist.UID(0)
	contents["testBundleURL"] = NewNSURL(testBundleURL)
	contents["testExecutionOrdering"] = 0
	contents["testsDrivenByIDE"] = false
	contents["testsMustRunOnMainThread"] = true
	contents["testsToRun"] = plist.UID(0)
	contents["testsToSkip"] = plist.UID(0)
	contents["testTimeoutsEnabled"] = false
	contents["treatMissingBaselinesAsFailures"] = false
	contents["userAttachmentLifetime"] = 1
	return XCTestConfiguration{contents}
}

func archiveXcTestConfiguration(xctestconfigInterface interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	xctestconfig := xctestconfigInterface.(XCTestConfiguration)
	xcconfigRef := plist.UID(len(objects))
	objects = append(objects, xctestconfig.contents)
	classRef := plist.UID(len(objects))
	objects = append(objects, buildClassDict("XCTestConfiguration", "NSObject"))

	xctestconfig.contents["$class"] = classRef

	for _, key := range []string{"aggregateStatisticsBeforeCrash", "automationFrameworkPath", "productModuleName", "sessionIdentifier",
		"targetApplicationBundleID", "targetApplicationPath", "testBundleURL"} {
		var ref plist.UID
		objects, ref = archive(xctestconfig.contents[key], objects)
		xctestconfig.contents[key] = ref
	}

	return objects, xcconfigRef
}

type NSUUID struct {
	uuidbytes []byte
}

func NewNSUUID(id uuid.UUID) NSUUID {
	bytes, err := id.MarshalBinary()
	if err != nil {
		log.Fatal("Unexpected Error", err)
	}
	return NSUUID{bytes}
}

func archiveNSUUID(uid interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	nsuuid := uid.(NSUUID)
	object := map[string]interface{}{}

	object["NS.uuidbytes"] = nsuuid.uuidbytes
	uuidReference := len(objects)
	objects = append(objects, object)

	classref := uuidReference + 1
	object[class] = plist.UID(classref)
	objects = append(objects, buildClassDict("NSUUID", "NSObject"))

	return objects, plist.UID(uuidReference)
}

type NSURL struct {
	path string
}

func NewNSURL(path string) NSURL {
	return NSURL{path}
}

func archiveNSURL(nsurlInterface interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	nsurl := nsurlInterface.(NSURL)
	object := map[string]interface{}{}

	object["NS.base"] = plist.UID(0)

	urlReference := len(objects)
	objects = append(objects, object)

	classref := urlReference + 1
	object[class] = plist.UID(classref)
	objects = append(objects, buildClassDict("NSURL", "NSObject"))

	pathRef := classref + 1
	object["NS.relative"] = plist.UID(pathRef)
	objects = append(objects, fmt.Sprintf("file://%s", nsurl.path))

	return objects, plist.UID(urlReference)
}

type DTActivityTraceTapMessage struct {
	DTTapMessagePlist map[string]interface{}
}

func NewDTActivityTraceTapMessage(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["DTTapMessagePlist"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return DTActivityTraceTapMessage{DTTapMessagePlist: plist}
}

type NSError struct {
	ErrorCode uint64
	Domain    string
	UserInfo  map[string]interface{}
}

func NewNSError(object map[string]interface{}, objects []interface{}) interface{} {
	errorCode := object["NSCode"].(uint64)
	userInfo_ref := object["NSUserInfo"].(plist.UID)
	domain_ref := object["NSDomain"].(plist.UID)
	domain := objects[domain_ref].(string)
	userinfo, _ := extractDictionary(objects[userInfo_ref].(map[string]interface{}), objects)

	return NSError{ErrorCode: errorCode, Domain: domain, UserInfo: userinfo}
}

//Apples Reference Date is Jan 1st 2001 00:00
const nsReferenceDate = 978307200000

type NSDate struct {
	timestamp time.Time
}

func NewNSDate(object map[string]interface{}, objects []interface{}) interface{} {
	value := object["NS.time"].(float64)
	milliesFloat := (1000*value + nsReferenceDate)
	millies := int64(milliesFloat)
	time := time.Unix(0, millies*int64(time.Millisecond))
	return NSDate{time}
}
func (n NSDate) String() string {
	return fmt.Sprintf("%s", n.timestamp)
}

type NSNull struct {
	class string
}

func NewNSNullFromArchived(object map[string]interface{}, objects []interface{}) interface{} {
	return NewNSNull()
}
func NewNSNull() interface{} {
	return NSNull{"NSNull"}
}

func archiveNSNull(object interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	nsnull := map[string]interface{}{}
	nsnullReference := len(objects)
	objects = append(objects, nsnull)
	objects = append(objects, buildClassDict("NSNull", "NSObject"))
	nsnull[class] = plist.UID(nsnullReference + 1)
	return objects, plist.UID(nsnullReference)
}

type NSMutableDictionary struct {
	internalDict map[string]interface{}
}

func NewNSMutableDictionary(internalDict map[string]interface{}) interface{} {
	return NSMutableDictionary{internalDict}
}

func archiveNSMutableDictionary(object interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	mut := object.(NSMutableDictionary)
	return serializeMap(mut.internalDict, objects, buildClassDict("NSMutableDictionary", "NSNull", "NSObject"))
}
