package api

import (
	"encoding/json"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"net/http"
)

//ListenHandler returns a DeviceList
func ListenHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := ios.ListDevices()
		if err != nil {
			serverError("failed getting devicelist", http.StatusInternalServerError, w)
			return
		}
		json, err := json.Marshal(devices)
		if err != nil {
			serverError("failed encoding json", http.StatusInternalServerError, w)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(json)
		log.WithError(err).Errorf("failed writing devicelist response")

	}
}
