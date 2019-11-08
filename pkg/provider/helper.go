package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

var (
	ErrInternalServerError = errors.New("remote push server: internal error")
	ErrServiceUnavailable  = errors.New("remote push server: service unavailable")
)

func SendWithRetry(maxRetries int, send func() (statusCode int, _ error)) error {

	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		hasAttempts := attempt < maxRetries-1

		statusCode, err := send()
		if err != nil {
			if hasAttempts && err == context.DeadlineExceeded {
				continue
			}
			return err

		} else if statusCode == http.StatusInternalServerError {
			if hasAttempts {
				continue
			}
			return ErrInternalServerError

		} else if statusCode == http.StatusServiceUnavailable {
			if hasAttempts {
				continue
			}
			return ErrServiceUnavailable

		}

		break
	}

	return nil
}

// DecodeJSONResponse unmarshal response in json format to the object.
// If server returns invalid json data, the method represents a response body
// as an error
func DecodeJSONResponse(r io.Reader, retval interface{}) error {

	decoder := json.NewDecoder(r)

	err := decoder.Decode(retval)
	if err == nil {
		return nil
	}

	if _, ok := err.(*json.SyntaxError); ok {
		errInfo := bytes.NewBuffer(nil)
		_, errCopy := io.CopyN(errInfo, decoder.Buffered(), 2000)
		if errCopy != nil && errCopy != io.EOF {
			return err
		}

		return errors.New(errInfo.String())
	}

	return err
}

func JSONWithoutSecrets(obj interface{}) ([]byte, error) {

	out, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return RemoveSecretsFromJSON(out), nil
}

var (
	_ReString = regexp.MustCompile(`:(\s*)"([^"\\]|\\.)+"`)
	_Secret   = []byte(`:"*"`)
)

func RemoveSecretsFromJSON(in []byte) []byte {

	if len(in) == 0 {
		return in
	}

	return _ReString.ReplaceAll(in, _Secret)
}
