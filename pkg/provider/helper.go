package provider

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrInternalServerError = errors.New("remote push server: internal error")
	ErrServiceUnavailable  = errors.New("remote push server: service unavailable")
)

func SendWithRetry(tries int, send func() (statusCode int, _ error)) error {

	if tries <= 0 {
		tries = 1
	}

	for try := 0; try < tries; try++ {
		hasTries := try < tries-1

		statusCode, err := send()
		if err != nil {
			if hasTries && err == context.DeadlineExceeded {
				continue
			}
			return err

		} else if statusCode == http.StatusInternalServerError {
			if hasTries {
				continue
			}
			return ErrInternalServerError

		} else if statusCode == http.StatusServiceUnavailable {
			if hasTries {
				continue
			}
			return ErrServiceUnavailable

		}

		break
	}

	return nil
}
