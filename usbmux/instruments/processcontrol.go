package instruments

const channelName = "com.apple.instruments.server.services.processcontrol"

func startProcess(path string, bundleID string, envVars map[string]string, arguments []string, options map[string]interface{}) {
	const objcMethodName = "launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:"

}
