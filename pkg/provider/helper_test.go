package provider

import (
	"bytes"
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

func TestDecodeJSONResponse(t *testing.T) {

	{
		// test: invalid incoming data
		const In = "JSON_PARSING_ERROR: Unexpected character (a) at position 0."

		target := make(map[string]interface{})
		require.EqualError(t,
			DecodeJSONResponse(bytes.NewReader([]byte(In)), &target),
			In)
	}

	{
		// test: success incoming data
		const In = `{"k1":"val1","k2":"val2 \" val3"}`

		target := make(map[string]interface{})
		require.NoError(t,
			DecodeJSONResponse(bytes.NewReader([]byte(In)), &target))

		require.Equal(t,
			map[string]interface{}{
				"k1": "val1",
				"k2": `val2 " val3`,
			},
			target)
	}
}

func TestRemoveSecretsFromJSON(t *testing.T) {

	for k, v := range map[string]string{
		``:                                 ``,
		`"`:                                `"`,
		`""`:                               `""`,
		`"k1":"`:                           `"k1":"`,
		`"k1":""`:                          `"k1":""`,
		`"k1":"v1 \" v2"`:                  `"k1":"*"`,
		`{"k1":"val1","k2":"val2"}`:        `{"k1":"*","k2":"*"}`,
		`{"k1":"val1","k2":{"k3":"val3"}}`: `{"k1":"*","k2":{"k3":"*"}}`,
		`{"k1":"val1","k2":"val2 \" val3 \" val4"}`: `{"k1":"*","k2":"*"}`,
	} {

		require.Equal(t,
			v,
			string(RemoveSecretsFromJSON([]byte(k))), k)
	}
}

func TestJSONWithoutSecrets(t *testing.T) {

	{
		out, err := JSONWithoutSecrets(nil)
		require.NoError(t, err)
		require.Equal(t, "null", string(out))
	}

	{
		out, err := JSONWithoutSecrets(map[string]interface{}{
			"k1": "val1",
			"k2": `val2 " val3`,
		})
		require.NoError(t, err)
		require.Equal(t, `{"k1":"*","k2":"*"}`, string(out))
	}
}
