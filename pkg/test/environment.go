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

	path, err := GetPathToIOSCertificatePem()
	if err != nil {
		return nil, err
	}

	pem, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return pem, nil
}

func GetPathToIOSCertificatePem() (string, error) {

	path := os.Getenv("APPLE_PUSH_CERTIFICATE")
	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func GetGoogleServiceAccount() ([]byte, error) {

	path, err := GetPathToGoogleServiceAccount()
	if err != nil {
		return nil, err
	}

	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func GetPathToGoogleServiceAccount() (string, error) {

	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func GetGoogleCloudMessageSettings() ([]byte, error) {

	path := os.Getenv("GOOGLE_LEGACY_APPLICATION_CREDENTIALS")
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func GetAccountKey() (string, error) {

	data, err := GetGoogleCloudMessageSettings()
	if err != nil {
		return "", err
	}

	settings := &struct {
		Key string `json:"key"`
	}{}

	r := bytes.NewReader(data)
	if err := json.NewDecoder(r).Decode(settings); err != nil {
		return "", err
	}

	return settings.Key, nil
}
