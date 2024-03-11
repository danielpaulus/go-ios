package instruments

const deviceInfoChannel = "com.apple.instruments.server.services.deviceinfo"

const (
	xpcControlChannel            = "com.apple.instruments.server.services.device.xpccontrol"
	procControlChannel           = "com.apple.instruments.server.services.processcontrol"
	procControlPosixSpawnChannel = "com.apple.instruments.server.services.processcontrol.posixspawn"
	mobileNotificationsChannel   = "com.apple.instruments.server.services.mobilenotifications"
	mobileNetworkingChannel      = "com.apple.instruments.server.services.networking"
	Sysmontap                    = "com.apple.instruments.server.services.sysmontap" // 获取性能数据用
)

const appListingChannel = "com.apple.instruments.server.services.device.applictionListing"

const watchProcessControlChannel = "com.apple.dt.Xcode.WatchProcessControl"

const (
	assetsChannel           = "com.apple.instruments.server.services.assets"
	activityTraceTapChannel = "com.apple.instruments.server.services.activitytracetap"
)
