package ios

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGetRsdPorts(t *testing.T) {
	rsd, err := NewRsdPortProvider(strings.NewReader(rsdOutput))
	assert.NoError(t, err)

	t.Run("exact match", func(t *testing.T) {
		testmanagerd := rsd.GetPort("com.apple.dt.testmanagerd.remote")
		assert.Equal(t, 50340, testmanagerd)
	})
	t.Run("fall back to shim", func(t *testing.T) {
		syslog := rsd.GetPort("com.apple.syslog_relay")
		assert.Equal(t, 50343, syslog)
	})
}

const rsdOutput = `
{
  "MessageType": "Handshake",
  "MessagingProtocolVersion": 3,
  "Properties": {
    "AppleInternal": false,
    "BoardId": 26,
    "BootSessionUUID": "aeaf223a-9447-4771-bc4c-1572f330b68c",
    "BuildVersion": "21A360",
    "CPUArchitecture": "arm64e",
    "CertificateProductionStatus": true,
    "CertificateSecurityMode": true,
    "ChipID": 32800,
    "DeviceClass": "iPhone",
    "DeviceColor": "1",
    "DeviceEnclosureColor": "1",
    "DeviceSupportsLockdown": true,
    "EffectiveProductionStatusAp": true,
    "EffectiveProductionStatusSEP": true,
    "EffectiveSecurityModeAp": true,
    "EffectiveSecurityModeSEP": true,
    "EthernetMacAddress": "90:e1:7b:21:7b:84",
    "HWModel": "D331pAP",
    "HardwarePlatform": "t8020",
    "HasSEP": true,
    "HumanReadableProductVersionString": "17.0.3",
    "Image4CryptoHashMethod": "sha2-384",
    "Image4Supported": true,
    "IsUIBuild": true,
    "IsVirtualDevice": false,
    "MobileDeviceMinimumVersion": "1600",
    "ModelNumber": "MT502",
    "OSInstallEnvironment": false,
    "OSVersion": "17.0.3",
    "ProductName": "iPhone OS",
    "ProductType": "iPhone11,6",
    "RegionCode": "ZD",
    "RegionInfo": "ZD/A",
    "RemoteXPCVersionFlags": 72057594037927940,
    "RestoreLongVersion": "21.1.360.0.0,0",
    "SecurityDomain": 1,
    "SensitivePropertiesVisible": true,
    "SerialNumber": "FFMXNFFJKPH1",
    "SigningFuse": true,
    "StoreDemoMode": false,
    "SupplementalBuildVersion": "21A360",
    "ThinningProductType": "iPhone11,6",
    "UniqueChipID": 7125711553429550,
    "UniqueDeviceID": "00008020-001950CC01EA002E"
  },
  "Services": {
    "com.apple.GPUTools.MobileService.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50358"
    },
    "com.apple.PurpleReverseProxy.Conn.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50335"
    },
    "com.apple.PurpleReverseProxy.Ctrl.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50357"
    },
    "com.apple.RestoreRemoteServices.restoreserviced": {
      "Entitlement": "com.apple.private.RestoreRemoteServices.restoreservice.remote",
      "Port": "50332",
      "Properties": {
        "ServiceVersion": 2,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.accessibility.axAuditDaemon.remoteserver.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50373"
    },
    "com.apple.afc.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50351"
    },
    "com.apple.amfi.lockdown.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50374"
    },
    "com.apple.atc.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50371"
    },
    "com.apple.atc2.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50364"
    },
    "com.apple.backgroundassets.lockdownservice.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50365"
    },
    "com.apple.bluetooth.BTPacketLogger.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50356"
    },
    "com.apple.carkit.remote-iap.service": {
      "Entitlement": "AppleInternal",
      "Port": "50370",
      "Properties": {
        "UsesRemoteXPC": true
      }
    },
    "com.apple.carkit.service.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50317"
    },
    "com.apple.commcenter.mobile-helper-cbupdateservice.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50325"
    },
    "com.apple.companion_proxy.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50320"
    },
    "com.apple.corecaptured.remoteservice": {
      "Entitlement": "com.apple.corecaptured.remoteservice-access",
      "Port": "50363",
      "Properties": {
        "UsesRemoteXPC": true
      }
    },
    "com.apple.coredevice.appservice": {
      "Entitlement": "com.apple.private.CoreDevice.canInstallCustomerContent",
      "Port": "50353",
      "Properties": {
        "Features": [
          "com.apple.coredevice.feature.launchapplication",
          "com.apple.coredevice.feature.spawnexecutable",
          "com.apple.coredevice.feature.monitorprocesstermination",
          "com.apple.coredevice.feature.installapp",
          "com.apple.coredevice.feature.uninstallapp",
          "com.apple.coredevice.feature.listroots",
          "com.apple.coredevice.feature.installroot",
          "com.apple.coredevice.feature.uninstallroot",
          "com.apple.coredevice.feature.sendsignaltoprocess",
          "com.apple.coredevice.feature.sendmemorywarningtoprocess",
          "com.apple.coredevice.feature.listprocesses",
          "com.apple.coredevice.feature.rebootdevice",
          "com.apple.coredevice.feature.listapps",
          "com.apple.coredevice.feature.fetchappicons"
        ],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.coredevice.deviceinfo": {
      "Entitlement": "com.apple.private.CoreDevice.canRetrieveDeviceInfo",
      "Port": "50333",
      "Properties": {
        "Features": [
          "com.apple.coredevice.feature.getdisplayinfo",
          "com.apple.coredevice.feature.getdeviceinfo",
          "com.apple.coredevice.feature.querymobilegestalt",
          "com.apple.coredevice.feature.getlockstate"
        ],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.coredevice.diagnosticsservice": {
      "Entitlement": "com.apple.private.CoreDevice.canObtainDiagnostics",
      "Port": "50378",
      "Properties": {
        "Features": [
          "com.apple.coredevice.feature.capturesysdiagnose"
        ],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.coredevice.fileservice.control": {
      "Entitlement": "com.apple.private.CoreDevice.canTransferFilesToDevice",
      "Port": "50347",
      "Properties": {
        "Features": [
          "com.apple.coredevice.feature.transferFiles"
        ],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.coredevice.fileservice.data": {
      "Entitlement": "com.apple.private.CoreDevice.canTransferFilesToDevice",
      "Port": "50337",
      "Properties": {
        "Features": [],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.coredevice.openstdiosocket": {
      "Entitlement": "com.apple.private.CoreDevice.canInstallCustomerContent",
      "Port": "50380",
      "Properties": {
        "Features": [],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.crashreportcopymobile.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50368"
    },
    "com.apple.crashreportmover.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50375"
    },
    "com.apple.dt.ViewHierarchyAgent.remote": {
      "Entitlement": "com.apple.private.dt.ViewHierarchyAgent.client",
      "Port": "50345",
      "Properties": {
        "UsesRemoteXPC": true
      }
    },
    "com.apple.dt.remoteFetchSymbols": {
      "Entitlement": "com.apple.private.dt.remoteFetchSymbols.client",
      "Port": "50334",
      "Properties": {
        "Features": [
          "com.apple.dt.remoteFetchSymbols.dyldSharedCacheFiles"
        ],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.dt.remotepairingdeviced.lockdown.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50328"
    },
    "com.apple.dt.testmanagerd.remote": {
      "Entitlement": "com.apple.private.dt.testmanagerd.client",
      "Port": "50340",
      "Properties": {
        "UsesRemoteXPC": false
      }
    },
    "com.apple.dt.testmanagerd.remote.automation": {
      "Entitlement": "AppleInternal",
      "Port": "50344",
      "Properties": {
        "UsesRemoteXPC": false
      }
    },
    "com.apple.fusion.remote.service": {
      "Entitlement": "com.apple.fusion.remote.service",
      "Port": "50339",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.gputools.remote.agent": {
      "Entitlement": "com.apple.private.gputoolstransportd",
      "Port": "50367",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.idamd.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50319"
    },
    "com.apple.instruments.dtservicehub": {
      "Entitlement": "com.apple.private.dt.instruments.dtservicehub.client",
      "Port": "50338",
      "Properties": {
        "Features": [
          "com.apple.dt.profile"
        ],
        "version": 1
      }
    },
    "com.apple.internal.devicecompute.CoreDeviceProxy": {
      "Entitlement": "AppleInternal",
      "Port": "50349",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": false
      }
    },
    "com.apple.internal.dt.coredevice.untrusted.tunnelservice": {
      "Entitlement": "com.apple.dt.coredevice.tunnelservice.client",
      "Port": "50321",
      "Properties": {
        "ServiceVersion": 2,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.internal.dt.remote.debugproxy": {
      "Entitlement": "com.apple.private.CoreDevice.canDebugApplicationsOnDevice",
      "Port": "50369",
      "Properties": {
        "Features": [
          "com.apple.coredevice.feature.debugserverproxy"
        ],
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.iosdiagnostics.relay.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50346"
    },
    "com.apple.misagent.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50382"
    },
    "com.apple.mobile.MCInstall.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50326"
    },
    "com.apple.mobile.assertion_agent.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50323"
    },
    "com.apple.mobile.diagnostics_relay.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50377"
    },
    "com.apple.mobile.file_relay.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50327"
    },
    "com.apple.mobile.heartbeat.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50366"
    },
    "com.apple.mobile.house_arrest.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50342"
    },
    "com.apple.mobile.insecure_notification_proxy.remote": {
      "Entitlement": "com.apple.mobile.insecure_notification_proxy.remote",
      "Port": "50318",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.mobile.insecure_notification_proxy.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.untrusted",
      "Port": "50348"
    },
    "com.apple.mobile.installation_proxy.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50336"
    },
    "com.apple.mobile.lockdown.remote.trusted": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50372",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.mobile.lockdown.remote.untrusted": {
      "Entitlement": "com.apple.mobile.lockdown.remote.untrusted",
      "Port": "50329",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.mobile.mobile_image_mounter.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50322"
    },
    "com.apple.mobile.notification_proxy.remote": {
      "Entitlement": "com.apple.mobile.notification_proxy.remote",
      "Port": "50341",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.mobile.notification_proxy.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50354"
    },
    "com.apple.mobile.storage_mounter_proxy.bridge": {
      "Entitlement": "com.apple.private.mobile_storage.remote.allowedSPI",
      "Port": "50331",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.mobileactivationd.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50355"
    },
    "com.apple.mobilebackup2.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50376"
    },
    "com.apple.mobilesync.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50330"
    },
    "com.apple.os_trace_relay.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50316"
    },
    "com.apple.osanalytics.logTransfer": {
      "Entitlement": "com.apple.ReportCrash.antenna-access",
      "Port": "50360",
      "Properties": {
        "UsesRemoteXPC": true
      }
    },
    "com.apple.pcapd.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50379"
    },
    "com.apple.preboardservice.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50359"
    },
    "com.apple.preboardservice_v2.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50350"
    },
    "com.apple.remote.installcoordination_proxy": {
      "Entitlement": "com.apple.private.InstallCoordinationRemote",
      "Port": "50361",
      "Properties": {
        "ServiceVersion": 1,
        "UsesRemoteXPC": true
      }
    },
    "com.apple.springboardservices.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50381"
    },
    "com.apple.streaming_zip_conduit.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50362"
    },
    "com.apple.sysdiagnose.remote": {
      "Entitlement": "com.apple.private.sysdiagnose.remote",
      "Port": "50352",
      "Properties": {
        "UsesRemoteXPC": true
      }
    },
    "com.apple.syslog_relay.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50343"
    },
    "com.apple.webinspector.shim.remote": {
      "Entitlement": "com.apple.mobile.lockdown.remote.trusted",
      "Port": "50324"
    }
  },
  "UUID": "d09e3fd1-62aa-47bc-b01a-23df97017d34"
}
`
