package test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

func GetPushDevices() (android, ios string, _ error) {

	path := os.Getenv("PUSH_DEVICES")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	deviceses := &struct {
		Android string `json:"android"`
		IOS     string `json:"ios"`
	}{}

	r := bytes.NewReader(data)
	if err := json.NewDecoder(r).Decode(deviceses); err != nil {
		return "", "", err
	}

	return deviceses.Android, deviceses.IOS, nil
}

func GetIOSCertificatePem() ([]byte, error) {

	path := os.Getenv("APPLE_PUSH_CERTIFICATE")
	pem, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return pem, nil
}

func GetGoogleServiceAccount() ([]byte, error) {

	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
