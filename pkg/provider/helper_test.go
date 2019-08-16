package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendWithRetry(t *testing.T) {

	for maxRetries := 0; maxRetries < 4; maxRetries++ {
		require.NoError(t,
			SendWithRetry(maxRetries, func() (int, error) { return 0, nil }))

		require.NoError(t,
			SendWithRetry(maxRetries, func() (int, error) { return http.StatusOK, nil }))

		require.NoError(t,
			SendWithRetry(maxRetries, func() (int, error) { return http.StatusOK, nil }))

		require.NoError(t,
			SendWithRetry(maxRetries, func() (int, error) { return http.StatusBadRequest, nil }))

		require.Equal(t,
			ErrInternalServerError,
			SendWithRetry(maxRetries, func() (int, error) { return http.StatusInternalServerError, nil }))

		require.Equal(t,
			ErrServiceUnavailable,
			SendWithRetry(maxRetries, func() (int, error) { return http.StatusServiceUnavailable, nil }))

		require.Equal(t,
			context.DeadlineExceeded,
			SendWithRetry(maxRetries, func() (int, error) { return http.StatusOK, context.DeadlineExceeded }))
	}

	{
		var counter int
		require.Equal(t,
			errors.New("test error"),
			SendWithRetry(6, func() (int, error) {
				counter++
				switch counter {
				case 1:
					return http.StatusInternalServerError, nil
				case 2:
					return http.StatusServiceUnavailable, nil
				case 3:
					return 0, context.DeadlineExceeded
				default:
					return 0, errors.New("test error")
				}
			}))
	}

	{
		var counter int
		require.NoError(t,
			SendWithRetry(5, func() (int, error) {
				counter++
				switch counter {
				case 1:
					return http.StatusInternalServerError, nil
				case 2:
					return http.StatusServiceUnavailable, nil
				case 3:
					return 0, context.DeadlineExceeded
				default:
					return http.StatusOK, nil
				}
			}))
	}

}
