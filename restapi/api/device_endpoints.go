package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/danielpaulus/go-ios/ios/simlocation"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Activate activates the device. Devices need to be activated and contact Apple servers before they can be used.
// Info                godoc
// @Summary      Activate the device by udid
// @Description  Returns and error if activation fails. Otherwise  {"message":"Activation successful"}
// @Tags         general_device_specific, activation
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/activate [post]
func Activate(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	err := mobileactivation.Activate(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, GenericResponse{Message: "Activation successful"})
}

func GetImages(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := imagemounter.NewImageMounter(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	signatures, err := conn.ListImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	res := make([]string, len(signatures))
	for i, sig := range signatures {
		res[i] = fmt.Sprintf("%x", sig)
	}
	c.JSON(http.StatusOK, res)

}
func InstallImage(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	auto := c.Query("auto")
	if auto == "true" {
		basedir := c.Query("basedir")
		if basedir == "" {
			basedir = "./devimages"
		}

		path, err := imagemounter.DownloadImageFor(device, basedir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		err = imagemounter.MountImage(device, path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, "ok")
		return
	}
	body := c.Request.Body
	defer body.Close()

	tempfile, err := os.CreateTemp(os.TempDir(), "go-ios")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	tempfilepath := tempfile.Name()
	defer os.Remove(tempfilepath)
	_, err = io.Copy(tempfile, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	err = tempfile.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	err = imagemounter.MountImage(device, tempfilepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "ok")
	return
}

// Info gets device info
// Info                godoc
// @Summary      Get lockdown info for a device by udid
// @Description  Returns all lockdown values and additional instruments properties for development enabled devices.
// @Tags         general_device_specific
// @Produce      json
// @Param        udid  path      string  true  "device udid"
// @Success      200  {object}  map[string]interface{}
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/info [get]
func Info(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	allValues, err := ios.GetValuesPlist(device)
	if err != nil {
		print(err)
	}
	svc, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		log.Debugf("could not open instruments, probably dev image not mounted %v", err)
	}
	if err == nil {
		info, err := svc.NetworkInformation()
		if err != nil {
			log.Debugf("error getting networkinfo from instruments %v", err)
		} else {
			allValues["instruments:networkInformation"] = info
		}
		info, err = svc.HardwareInformation()
		if err != nil {
			log.Debugf("error getting hardwareinfo from instruments %v", err)
		} else {
			allValues["instruments:hardwareInformation"] = info
		}
	}
	c.IndentedJSON(http.StatusOK, allValues)
}

// Screenshot grab screenshot from a device
// Screenshot                godoc
// @Summary      Get screenshot for device
// @Description Takes a png screenshot and returns it.
// @Tags         general_device_specific
// @Produce      png
// @Param        udid  path      string  true  "device udid"
// @Success      200  {object}  []byte
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/screenshot [get]
func Screenshot(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	screenshotService, err := instruments.NewScreenshotService(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}
	defer screenshotService.Close()

	imageBytes, err := screenshotService.TakeScreenshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "application/octet-stream", imageBytes)
}

// Change the current device location
// @Summary      Change the current device location
// @Description Change the current device location to provided latitude and longtitude
// @Tags         general_device_specific
// @Produce      json
// @Param        latitude  query      string  true  "Location latitude"
// @Param        longtitude  query      string  true  "Location longtitude"
// @Success      200  {object}  GenericResponse
// @Failure		 422  {object}  GenericResponse
// @Failure		 500  {object}  GenericResponse
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/setlocation [post]
func SetLocation(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	latitude := c.Query("latitude")
	if latitude == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "latitude query param is missing"})
		return
	}

	longtitude := c.Query("longtitude")
	if longtitude == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "longtitude query param is missing"})
		return
	}

	err := simlocation.SetLocation(device, latitude, longtitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenericResponse{Message: "Device location set to latitude=" + latitude + ", longtitude=" + longtitude})
}

// Reset to the actual device location
// @Summary      Reset the changed device location
// @Description  Reset the changed device location to the actual one
// @Tags         general_device_specific
// @Produce      json
// @Success      200
// @Failure      500  {object}  GenericResponse
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/resetlocation [post]
func ResetLocation(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	err := simlocation.ResetLocation(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenericResponse{Message: "Device location reset"})
}

// Get the list of installed profiles
// @Summary      get the list of profiles
// @Description  get the list of installed profiles from the ios device
// @Tags         general_device_specific
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  GenericResponse
// @Failure      404  {object}  GenericResponse
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/profiles [get]
func GetProfiles(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	mcinstallconn, err := mcinstall.New(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed getting device list with error", "error": err.Error()})
		return
	}

	defer mcinstallconn.Close()

	profileInfo, err := mcinstallconn.HandleList()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Failed getting profile list with error", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profileInfo)
}

//========================================
// DEVICE STATE CONDITIONS
//========================================

var (
	deviceConditionsMap   = make(map[string]deviceCondition)
	deviceConditionsMutex sync.Mutex
)

type deviceCondition struct {
	ProfileType  instruments.ProfileType
	Profile      instruments.Profile
	StateControl *instruments.DeviceStateControl
}

// Get a list of the available conditions that can be applied on the device
// @Summary      Get a list of available device conditions
// @Description  Get a list of the available conditions that can be applied on the device
// @Tags         general_device_specific
// @Produce      json
// @Success      200  {object}  []instruments.ProfileType
// @Failure      500  {object}  GenericResponse
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/conditions [get]
func GetSupportedConditions(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	control, err := instruments.NewDeviceStateControl(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	profileTypes, err := control.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, profileTypes)
}

// Enable condition on a device
// @Summary      Enable condition on a device
// @Description  Enable condition on a device by provided profileTypeID and profileID
// @Tags         general_device_specific
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Param        profileTypeID  query      string  true  "Identifier of the profile type, eg. SlowNetworkCondition"
// @Param        profileID  query      string  true  "Identifier of the sub-profile, eg. SlowNetwork100PctLoss"
// @Success      200  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/enable-condition [put]
func EnableDeviceCondition(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber

	deviceConditionsMutex.Lock()
	defer deviceConditionsMutex.Unlock()

	conditionedDevice, exists := deviceConditionsMap[udid]
	if exists {
		c.JSON(http.StatusOK, GenericResponse{Error: "Device has an active condition - profileTypeID=" + conditionedDevice.ProfileType.Identifier + ", profileID=" + conditionedDevice.Profile.Identifier})
		return
	}

	profileTypeID := c.Query("profileTypeID")
	if profileTypeID == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "profileTypeID query param is missing"})
		return
	}

	profileID := c.Query("profileID")
	if profileID == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "profileID query param is missing"})
		return
	}

	control, err := instruments.NewDeviceStateControl(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	profileTypes, err := control.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	profileType, profile, err := instruments.VerifyProfileAndType(profileTypes, profileTypeID, profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	err = control.Enable(profileType, profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	// When we apply a condition using a specific *instruments.DeviceStateControl pointer, we need that same pointer to disable it
	// Creating a new *DeviceStateControl and providing the same profileType WILL NOT disable the already active condition
	// For this reason we keep a map of `deviceConditions` that contain their original *DeviceStateControl pointers
	// which we can use in `DisableDeviceCondition()` to successfully disable the active condition
	newDeviceCondition := deviceCondition{ProfileType: profileType, Profile: profile, StateControl: control}
	deviceConditionsMap[device.Properties.SerialNumber] = newDeviceCondition

	c.JSON(http.StatusOK, GenericResponse{Message: "Enabled condition for ProfileType=" + profileTypeID + " and Profile=" + profileID})
}

// Disable the currently active condition on a device
// @Summary      Disable the currently active condition on a device
// @Description  Disable the currently active condition on a device
// @Tags         general_device_specific
// @Produce      json
// @Success      200  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Param        udid path string true "Device UDID"
// @Router       /device/{udid}/disable-condition [post]
func DisableDeviceCondition(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber

	deviceConditionsMutex.Lock()
	defer deviceConditionsMutex.Unlock()

	conditionedDevice, exists := deviceConditionsMap[udid]
	if !exists {
		c.JSON(http.StatusOK, GenericResponse{Error: "Device has no active condition"})
		return
	}

	// Disable() does not throw an error if the respective condition is not active on the device
	err := conditionedDevice.StateControl.Disable(conditionedDevice.ProfileType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	delete(deviceConditionsMap, udid)

	c.JSON(http.StatusOK, GenericResponse{Message: "Device condition disabled"})
}

// ========================================
// DEVICE PAIRING
// ========================================
// Pairs a device
// @Summary      Pair a device with/without supervision
// @Description  Pair a device with/without supervision
// @Tags         general_device_specific
// @Produce      json
// @Success      200  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Failure      422  {object}  GenericResponse
// @Param        udid path string true "Device UDID"
// @Param        supervised query string true "Set if device is supervised - true/false"
// @Param 		 p12file formData file false "Supervision *.p12 file"
// @Param 		 supervision_password formData string false "Supervision password"
// @Router       /device/{udid}/pair [post]
func PairDevice(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	supervised := c.Query("supervised")
	if supervised == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "supervised query param is missing (true/false)"})
		return
	}

	if supervised == "false" {
		err := ios.Pair(device)
		if err != nil {
			c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, GenericResponse{Message: "Device paired"})
		return
	}

	file, _, err := c.Request.FormFile("p12file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: "Could not parse p12 file from form-data or no file provided, err:" + err.Error()})
		return
	}
	p12fileBuf := new(bytes.Buffer)
	p12fileBuf.ReadFrom(file)

	supervision_password := c.Request.Header.Get("Supervision-Password")
	if supervision_password == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "you must provide non-empty `Supervision-Password` header with the request"})
		return
	}

	err = ios.PairSupervised(device, p12fileBuf.Bytes(), supervision_password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenericResponse{Message: "Device paired"})
}
