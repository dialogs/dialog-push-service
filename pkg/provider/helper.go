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
