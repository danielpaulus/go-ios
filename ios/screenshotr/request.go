package screenshotr

type screenShotRequest struct {
	MessageType string
}

func newScreenShotRequest() plistArray {
	request := make([]interface{}, 2)
	request[0] = "DLMessageProcessMessage"
	request[1] = screenShotRequest{"ScreenShotRequest"}
	return request
}
