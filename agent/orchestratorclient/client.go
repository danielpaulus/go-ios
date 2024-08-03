package orchestratorclient

import (
	"bytes"

	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/goccy/go-json"
)

const POOL_HEADER = "X-POOL-ID"
const AUTH_HEADER = "X-DEVICEAPI-AUTH-HEADER"

var netClient = &http.Client{
	Timeout: time.Second * 50,
}

func GetDevices() ([]models.DeviceInfo, error) {
	_, devices, err := getRequest("api/v1/devices", []models.DeviceInfo{})
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func PushSDPAnswers(sdps []models.SDP) error {
	_, _, err := putRequest(fmt.Sprintf("api/v1/pools/%s/sdp", models.GetPoolId().String()), sdps)
	return err
}

func DownloadSDPs() ([]models.SDP, error) {
	var sdps []models.SDP
	_, sdps, err := getRequest(fmt.Sprintf("api/v1/pools/%s/sdp", models.GetPoolId().String()), sdps)
	return sdps, err
}

func UpdateState(list models.DevicePool) error {
	_, _, err := putRequest("api/v1/pools", list)
	return err
}

func GetCloudconfig() (models.Config, error) {
	config := models.Config{}
	_, cloudConfig, err := getRequest("api/v1/config", config)

	return cloudConfig, err
}

func putRequest(path string, body interface{}) (http.Header, io.ReadCloser, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	//log.Info("sending" + string(data))
	return request(path, http.MethodPut, bytes.NewReader(data))
}

func postRequest(path string, body interface{}) (http.Header, io.ReadCloser, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	return request(path, http.MethodPost, bytes.NewReader(data))
}

type T interface {
}

func getRequest[t T](path string, payload t) (http.Header, t, error) {
	header, body, err := request(path, http.MethodGet, nil)
	if err != nil {
		return nil, payload, err
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, payload, err
	}

	err = json.Unmarshal(data, &payload)
	return header, payload, err
}

func request(path string, method string, body io.Reader) (http.Header, io.ReadCloser, error) {
	orchestratorURL := os.Getenv("ORCHESTRATOR_URL")
	requestURL := fmt.Sprintf("%s/%s", orchestratorURL, path)
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set(POOL_HEADER, models.GetPoolId().String())
	req.Header.Set(AUTH_HEADER, os.Getenv("API_KEY"))
	response, err := netClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode > 299 {
		return nil, nil, fmt.Errorf("orchestrator %s failed with status code: %d", path, response.StatusCode)
	}
	return response.Header, response.Body, nil
}
