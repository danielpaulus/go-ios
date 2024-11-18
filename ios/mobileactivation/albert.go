package mobileactivation

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const activationUserAgent = "iOS Device Activator (MobileActivation-592.103.2)"

const (
	activationServerURL = "https://albert.apple.com/deviceservices/deviceActivation"
	drmHandshakeURL     = "https://albert.apple.com/deviceservices/drmHandshake"
)

var netClient = &http.Client{
	Timeout:   time.Second * 5,
	Transport: http.DefaultTransport,
}

func sendHandshakeRequest(body io.Reader) (http.Header, io.ReadCloser, error) {
	requestURL := drmHandshakeURL
	req, err := http.NewRequest("POST", requestURL, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-apple-plist")
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("User-Agent", activationUserAgent)
	response, err := netClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode > 299 {
		return response.Header, response.Body, fmt.Errorf("error activating %d, %v", response.StatusCode, response)
	}
	return response.Header, response.Body, nil
}

func sendActivationRequest(body io.Reader) (http.Header, io.ReadCloser, error) {
	requestURL := activationServerURL
	req, err := http.NewRequest("POST", requestURL, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", activationUserAgent)
	response, err := netClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode > 299 {
		return response.Header, response.Body, fmt.Errorf("error activating %d, %v", response.StatusCode, response)
	}
	return response.Header, response.Body, nil
}
