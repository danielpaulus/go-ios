package nskeyedarchiver

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"howett.net/plist"
)

var (
	decodableClasses map[string]func(map[string]interface{}, []interface{}) interface{}
	encodableClasses map[string]func(object interface{}, objects []interface{}) ([]interface{}, plist.UID)
)

func SetupDecoders() {
	if decodableClasses == nil {
		decodableClasses = map[string]func(map[string]interface{}, []interface{}) interface{}{
			"DTActivityTraceTapMessage": NewDTActivityTraceTapMessage,
			"DTSysmonTapMessage":        NewDTActivityTraceTapMessage,
			"NSError":                   NewNSError,
			"NSNull":                    NewNSNullFromArchived,
			"NSDate":                    NewNSDate,
			"XCTestConfiguration":       NewXCTestConfigurationFromBytes,
			"DTTapHeartbeatMessage":     NewDTTapHeartbeatMessage,
			"XCTCapabilities":           NewXCTCapabilities,
			"NSUUID":                    NewNSUUIDFromBytes,
			"XCActivityRecord":          NewXCActivityRecord,
			"XCTAttachment":             NewXCTAttachment,
			"DTKTraceTapMessage":        NewDTKTraceTapMessage,
			"NSValue":                   NewNSValue,
			"NSArray":                   NewNSArray,
			"XCTTestIdentifier":         NewXCTTestIdentifier,
			"DTTapStatusMessage":        NewDTTapStatusMessage,
			"DTTapMessage":              NewDTTapMessage,
			"DTCPUClusterInfo":          NewDTCPUClusterInfo,
			"XCTIssue":                  NewXCTIssue,
			"XCTSourceCodeContext":      NewXCTSourceCodeContext,
			"XCTSourceCodeLocation":     NewXCTSourceCodeLocation,
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
			"XCTCapabilities":     archiveXCTCapabilities,
			"[]string":            archiveStringSlice,
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
	// contents["formatVersion"]= 2
	contents["gatherLocalizableStringsData"] = false
	contents["initializeForUITesting"] = true
	contents["maximumTestExecutionTimeAllowance"] = plist.UID(0)
	contents["productModuleName"] = productModuleName
	contents["randomExecutionOrderingSeed"] = plist.UID(0)
	contents["reportActivities"] = true
	contents["reportResultsToIDE"] = true
	contents["sessionIdentifier"] = NewNSUUID(sessionIdentifier)
	contents["systemAttachmentLifetime"] = 0
	// contents["targetApplicationArguments"] = []interface{}{} //TODO: triggers a bug
	contents["targetApplicationBundleID"] = targetApplicationBundleID
	// contents["targetApplicationEnvironment"] = //TODO: triggers a bug map[string]interface{}{}
	contents["targetApplicationPath"] = targetApplicationPath
	// testApplicationDependencies
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
	contents["userAttachmentLifetime"] = 0
	contents["preferredScreenCaptureFormat"] = 2
	contents["IDECapabilities"] = XCTCapabilities{CapabilitiesDictionary: map[string]interface{}{
		"expected failure test capability":         true,
		"test case run configurations":             true,
		"test timeout capability":                  true,
		"test iterations":                          true,
		"request diagnostics for specific devices": true,
		"delayed attachment transfer":              true,
		"skipped test capability":                  true,
		"daemon container sandbox extension":       true,
		"ubiquitous test identifiers":              true,
		"XCTIssue capability":                      true,
	}}
	return XCTestConfiguration{contents}
}

func archiveXcTestConfiguration(xctestconfigInterface interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	xctestconfig := xctestconfigInterface.(XCTestConfiguration)
	xcconfigRef := plist.UID(len(objects))
	objects = append(objects, xctestconfig.contents)
	classRef := plist.UID(len(objects))
	objects = append(objects, buildClassDict("XCTestConfiguration", "NSObject"))

	xctestconfig.contents["$class"] = classRef

	for _, key := range []string{
		"aggregateStatisticsBeforeCrash", "automationFrameworkPath", "productModuleName", "sessionIdentifier",
		"targetApplicationBundleID", "targetApplicationPath", "testBundleURL",
	} {
		var ref plist.UID
		objects, ref = archive(xctestconfig.contents[key], objects)
		xctestconfig.contents[key] = ref
	}

	return objects, xcconfigRef
}

type NSUUID struct {
	uuidbytes []byte
}

func (n NSUUID) String() string {
	uid, err := uuid.FromBytes(n.uuidbytes)
	if err != nil {
		return fmt.Sprintf("Failed converting %x to uuid with %+v", n.uuidbytes, err)
	}
	return uid.String()
}

type XCActivityRecord struct {
	/*
			finish":<interface {}(howett.net/plist.UID)>)
		"start":<interface {}(howett.net/plist.UID)>)

		"title":<interface {}(howett.net/plist.UID)>)

		"uuid":<interface {}(howett.net/plist.UID)>)

		"activityType":<interface {}(howett.net/plist.UID)>)

		"attachments":<interface {}(howett.net/plist.UID)>)

	*/
	Finish       NSDate
	Start        NSDate
	Title        string
	UUID         NSUUID
	ActivityType string
	Attachments  []XCTAttachment
}

func NewXCActivityRecord(object map[string]interface{}, objects []interface{}) interface{} {
	finish_ref := object["finish"].(plist.UID)
	finish := NSDate{}
	if _, ok := objects[finish_ref].(map[string]interface{}); ok {
		finish_raw := objects[finish_ref].(map[string]interface{})
		finish = NewNSDate(finish_raw, objects).(NSDate)
	}

	start_ref := object["start"].(plist.UID)
	start := NSDate{}
	if _, ok := objects[start_ref].(map[string]interface{}); ok {
		start_raw := objects[start_ref].(map[string]interface{})
		start = NewNSDate(start_raw, objects).(NSDate)
	}

	uuid_ref := object["uuid"].(plist.UID)
	uuid_raw := objects[uuid_ref].(map[string]interface{})
	uuid := NewNSUUIDFromBytes(uuid_raw, objects).(NSUUID)

	title_ref := object["title"].(plist.UID)
	title := objects[title_ref].(string)

	attachments_ref := object["attachments"].(plist.UID)
	attachments_raw := objects[attachments_ref].(map[string]interface{})

	attachments := make([]XCTAttachment, 0)
	for _, obj := range NewNSArray(attachments_raw, objects).(NSArray).Values {
		attachments = append(attachments, obj.(XCTAttachment))
	}

	activityType_ref := object["activityType"].(plist.UID)
	activityType := objects[activityType_ref].(string)

	return XCActivityRecord{Finish: finish, Start: start, UUID: uuid, Title: title, Attachments: attachments, ActivityType: activityType}
}

const (
	LifetimeKeepAlways      = uint64(0)
	LifetimeDeleteOnSuccess = uint64(1)
)

type XCTAttachment struct {
	lifetime              uint64
	UniformTypeIdentifier string
	fileNameOverride      string
	Payload               []uint8
	Timestamp             float64
	Name                  string
	userInfo              map[string]interface{}
}

func NewXCTAttachment(object map[string]interface{}, objects []interface{}) interface{} {
	lifetime := object["lifetime"].(uint64)
	uniformTypeIdentifier := objects[object["uniformTypeIdentifier"].(plist.UID)].(string)
	fileNameOverride := objects[object["fileNameOverride"].(plist.UID)].(string)
	payload := objects[object["payload"].(plist.UID)].([]uint8)
	timestamp := objects[object["timestamp"].(plist.UID)].(float64)
	name := objects[object["name"].(plist.UID)].(string)
	userInfo, _ := extractDictionary(objects[object["userInfo"].(plist.UID)].(map[string]interface{}), objects)

	return XCTAttachment{
		lifetime:              lifetime,
		UniformTypeIdentifier: uniformTypeIdentifier,
		fileNameOverride:      fileNameOverride,
		Payload:               payload,
		Timestamp:             timestamp,
		Name:                  name,
		userInfo:              userInfo,
	}
}

func NewNSUUIDFromBytes(object map[string]interface{}, objects []interface{}) interface{} {
	val := object["NS.uuidbytes"].([]byte)
	return NSUUID{uuidbytes: val}
}

func NewNSUUID(id uuid.UUID) NSUUID {
	bytes, err := id.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("Unexpected Error: %v", err))
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

func archiveXCTCapabilities(capsIface interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	caps := capsIface.(XCTCapabilities)
	object := map[string]interface{}{}

	objects, dictRef := serializeMap(caps.CapabilitiesDictionary, objects, buildClassDict("NSDictionary", "NSObject"))
	object["capabilities-dictionary"] = dictRef

	capsReference := len(objects)
	objects = append(objects, object)

	classref := capsReference + 1
	object[class] = plist.UID(classref)
	objects = append(objects, buildClassDict("XCTCapabilities", "NSObject"))
	return objects, plist.UID(capsReference)
}

type NSURL struct {
	Path string
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
	objects = append(objects, fmt.Sprintf("file://%s", nsurl.Path))

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

type DTKTraceTapMessage struct {
	DTTapMessagePlist map[string]interface{}
}

func NewDTKTraceTapMessage(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["DTTapMessagePlist"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return DTKTraceTapMessage{DTTapMessagePlist: plist}
}

type NSValue struct {
	NSSpecial uint64
	NSRectval string
}

func NewNSValue(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["NS.rectval"].(plist.UID)
	rectval, _ := objects[ref].(string)
	special := object["NS.special"].(uint64)
	return NSValue{NSRectval: rectval, NSSpecial: special}
}

type NSArray struct {
	Values []interface{}
}

func NewNSArray(object map[string]interface{}, objects []interface{}) interface{} {
	objectRefs := object["NS.objects"].([]interface{})

	uidList := toUidList(objectRefs)
	extractObjects, _ := extractObjects(uidList, objects)

	return NSArray{Values: extractObjects}
}

type XCTTestIdentifier struct {
	O uint64
	C []string
}

func (x XCTTestIdentifier) String() string {
	return fmt.Sprintf("XCTTestIdentifier{o:%d , c:%v}", x.O, x.C)
}

func NewXCTTestIdentifier(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["c"].(plist.UID)
	// plist, _ := extractObjects(objects[ref].(map[string]interface{}), objects)
	fd := objects[ref].(map[string]interface{})
	extractObjects, _ := extractObjects(toUidList(fd[nsObjects].([]interface{})), objects)
	stringarray := make([]string, len(extractObjects))
	for i, v := range extractObjects {
		stringarray[i] = v.(string)
	}
	o := object["o"].(uint64)
	return XCTTestIdentifier{
		O: o,
		C: stringarray,
	}
}

// TODO: make this nice, partially extracting objects is not really cool
type PartiallyExtractedXcTestConfig struct {
	values map[string]interface{}
}

func NewXCTestConfigurationFromBytes(object map[string]interface{}, objects []interface{}) interface{} {
	config := make(map[string]interface{}, len(object))
	for k, v := range object {
		value := v
		uid, ok := v.(plist.UID)
		if ok {
			value = objects[uid]
		}
		config[k] = value
	}

	return PartiallyExtractedXcTestConfig{config}
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

func (err NSError) Error() string {
	return fmt.Sprintf("Error code: %d, Domain: %s, User info: %v", err.ErrorCode, err.Domain, err.UserInfo)
}

// Apples Reference Date is Jan 1st 2001 00:00
const nsReferenceDate = 978307200000

type NSDate struct {
	Timestamp time.Time
}

type DTTapHeartbeatMessage struct {
	DTTapMessagePlist map[string]interface{}
}

type DTTapMessage struct {
	DTTapMessagePlist map[string]interface{}
}

type XCTCapabilities struct {
	CapabilitiesDictionary map[string]interface{}
}

func NewXCTCapabilities(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["capabilities-dictionary"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return XCTCapabilities{CapabilitiesDictionary: plist}
}

func NewDTTapHeartbeatMessage(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["DTTapMessagePlist"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return DTTapHeartbeatMessage{DTTapMessagePlist: plist}
}

func NewDTTapMessage(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["DTTapMessagePlist"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return DTTapMessage{DTTapMessagePlist: plist}
}

type DTTapStatusMessage struct {
	DTTapMessagePlist map[string]interface{}
}

func NewDTTapStatusMessage(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["DTTapMessagePlist"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return DTTapStatusMessage{DTTapMessagePlist: plist}
}

func NewNSDate(object map[string]interface{}, objects []interface{}) interface{} {
	value := object["NS.time"].(float64)
	milliesFloat := (1000*value + nsReferenceDate)
	millies := int64(milliesFloat)
	time := time.Unix(0, millies*int64(time.Millisecond))
	return NSDate{time}
}

func (n NSDate) String() string {
	return fmt.Sprintf("%s", n.Timestamp)
}

type DTCPUClusterInfo struct {
	ClusterID    uint64
	ClusterFlags uint64
}

func NewDTCPUClusterInfo(object map[string]interface{}, objects []interface{}) interface{} {
	return DTCPUClusterInfo{ClusterID: object["_clusterID"].(uint64), ClusterFlags: object["_clusterFlags"].(uint64)}
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

func archiveStringSlice(object interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	sl := object.([]string)
	return serializeArray(toInterfaceSlice(sl), objects)
}

func archiveNSMutableDictionary(object interface{}, objects []interface{}) ([]interface{}, plist.UID) {
	mut := object.(NSMutableDictionary)
	return serializeMap(mut.internalDict, objects, buildClassDict("NSMutableDictionary", "NSDictionary", "NSObject"))
}

type XCTIssue struct {
	RuntimeIssueSeverity uint64
	DetailedDescription  string
	CompactDescription   string
	SourceCodeContext    XCTSourceCodeContext
}

func NewXCTIssue(object map[string]interface{}, objects []interface{}) interface{} {
	runtimeIssueSeverity := object["runtimeIssueSeverity"].(uint64)
	detailedDescriptionRef := object["detailed-description"].(plist.UID)
	sourceCodeContextRef := object["source-code-context"].(plist.UID)
	compactDescriptionRef := object["compact-description"].(plist.UID)

	detailedDescription := objects[detailedDescriptionRef].(string)
	compactDescription := objects[compactDescriptionRef].(string)
	sourceCodeContext := NewXCTSourceCodeContext(objects[sourceCodeContextRef].(map[string]interface{}), objects).(XCTSourceCodeContext)

	return XCTIssue{RuntimeIssueSeverity: runtimeIssueSeverity, DetailedDescription: detailedDescription, CompactDescription: compactDescription, SourceCodeContext: sourceCodeContext}
}

type XCTSourceCodeContext struct {
	Location XCTSourceCodeLocation
}

func NewXCTSourceCodeContext(object map[string]interface{}, objects []interface{}) interface{} {
	locationRef := object["location"].(plist.UID)
	location := NewXCTSourceCodeLocation(objects[locationRef].(map[string]interface{}), objects).(XCTSourceCodeLocation)

	return XCTSourceCodeContext{Location: location}
}

type XCTSourceCodeLocation struct {
	FileUrl    NSURL
	LineNumber uint64
}

func NewXCTSourceCodeLocation(object map[string]interface{}, objects []interface{}) interface{} {
	fileUrlRef := object["file-url"].(plist.UID)
	relativeRef := objects[fileUrlRef].(map[string]interface{})["NS.relative"].(plist.UID)
	relativePath := objects[int(relativeRef)].(string)
	fileUrl := NewNSURL(relativePath)
	lineNumber := object["line-number"].(uint64)

	return XCTSourceCodeLocation{FileUrl: fileUrl, LineNumber: lineNumber}
}
