package service

import (
	"encoding/json"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"strings"
)

func XCTestHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bundleID := strings.TrimSpace(r.URL.Query().Get("bundleid"))
		fork := strings.TrimSpace(r.URL.Query().Get("fork"))
		udid := strings.TrimSpace(r.URL.Query().Get("udid"))
		if udid == "" {
			serverError("missing udid", http.StatusBadRequest, w)
			return
		}
		if bundleID == "" {
			serverError("missing bundleID", http.StatusBadRequest, w)
			return
		}
		if fork == "true" {
			log.Infof("running test for app %s and device %s in separate process", bundleID, udid)
			cmd := exec.Command("ios", "runtest", bundleID, "--udid="+udid)
			output, err := cmd.CombinedOutput()
			outputString := string(output)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "cmd": cmd, "output": outputString}).Errorf("error removing signature with codesign")
				jsonResponse(map[string]interface{}{"logs": outputString, "err": err.Error()}, http.StatusInternalServerError, w)
				return
			}
			log.WithFields(log.Fields{"cmd": cmd, "output": outputString}).Infof("go-ios runtest finished")
			jsonResponse(map[string]interface{}{"logs": outputString}, http.StatusOK, w)
			return
		}

		device, err := ios.GetDevice(udid)
		if err != nil {
			serverError(err.Error(), http.StatusInternalServerError, w)
			return
		}
		log.Info("server runs test")
		err = testmanagerd.RunXCUITest(bundleID, device)
		log.Info("server done running test")
		if err != nil {
			serverError(err.Error(), http.StatusInternalServerError, w)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

	}
}

//HealthHandler is a simple health check. It executes a basic codesign operation to make sure
//codesign really works and is set up correctly
func HealthHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := ios.ListDevices()
		if err != nil {
			serverError("failed getting devicelist", http.StatusInternalServerError, w)
			return
		}
		json, err := json.Marshal(
			map[string]string{
				"version": GetVersion(),
				"devices": devices.String(),
			},
		)
		if err != nil {
			serverError("failed encoding json", http.StatusInternalServerError, w)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)

	}
}

func serverError(message string, code int, w http.ResponseWriter) {
	json, err := json.Marshal(
		map[string]string{"error": message},
	)
	if err != nil {
		log.Warnf("error encoding json:%+v", err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(json)
}

func jsonResponse(jsonObject map[string]interface{}, code int, w http.ResponseWriter) {
	json, err := json.Marshal(
		jsonObject,
	)
	if err != nil {
		log.Warnf("error encoding json:%+v", err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(json)
}
