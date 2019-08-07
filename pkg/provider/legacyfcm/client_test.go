package legacyfcm

import (
	"context"
	"testing"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/stretchr/testify/require"
)

// Environment for tests:
// 1. copy server key from https://console.firebase.google.com/project/_/settings/cloudmessaging/android:com.example.push
// 2. save server key in file. File format:
//	{
//		"key":"<server key>"
//	}
// 2. create environment variable "GOOGLE_LEGACY_APPLICATION_CREDENTIALS" with path to server key file
// 3. create file with devices tokens. format:
//	{
//     "android": "<token>",
//     "ios": "<token>"
//	}
// 4. create environment variable "PUSH_DEVICES" with path to file with devices tokens

func TestSendOk(t *testing.T) {

	token := getDeviceToken(t)

	req := &Request{
		To: token,
	}

	client := getClient(t)
	resp, err := client.Send(context.Background(), req)
	require.NoError(t, err)

	require.True(t, resp.MulticastID > 0, resp.MulticastID)
	require.NotEmpty(t, resp.Results[0].MessageID)

	require.Equal(t,
		&Response{
			MulticastID: resp.MulticastID,
			Success:     1,
			Failure:     0,
			StatusCode:  200,
			Results: []*ResponseResult{
				{
					MessageID:      resp.Results[0].MessageID,
					RegistrationID: "",
					Error:          "",
				},
			},
		},
		resp)
}

func TestSendError(t *testing.T) {

	req := &Request{
		To: "",
	}

	client := getClient(t)
	resp, err := client.Send(context.Background(), req)
	require.NoError(t, err)

	require.True(t, resp.MulticastID > 0, resp.MulticastID)

	require.Equal(t,
		&Response{
			MulticastID: resp.MulticastID,
			Success:     0,
			Failure:     1,
			StatusCode:  200,
			Results: []*ResponseResult{
				{
					MessageID:      "",
					RegistrationID: "",
					Error:          ErrorCodeMissingRegistration,
				},
			},
		},
		resp)
}

func getClient(t *testing.T) *Client {
	t.Helper()

	key := getAccountKey(t)

	client, err := New(key, 2, time.Second)
	require.NoError(t, err)

	return client
}

func getAccountKey(t *testing.T) string {
	t.Helper()

	key, err := test.GetAccountKey()
	require.NoError(t, err)

	return key
}

func getDeviceToken(t *testing.T) string {
	t.Helper()

	token, _, err := test.GetPushDevices()
	require.NoError(t, err)

	return token
}
