package fcm

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/stretchr/testify/require"
)

// Environment for tests:
// 1. download service-account.json from https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
// 2. create environment variable "GOOGLE_APPLICATION_CREDENTIALS" with path to service-account.json
// 3. create file with devices tokens. format:
//	{
//     "android": "<token>",
//     "ios": "<token>"
//	}
// 4. create environment variable "PUSH_DEVICES" with path to file with devices tokens

func init() {
	log.SetFlags(log.Llongfile | log.Ltime | log.Lmicroseconds)
}

func TestGetEndpoint(t *testing.T) {

	require.Equal(t,
		"https://fcm.googleapis.com/v1/projects/project-id/messages:send",
		getEndpoint("project-id"))

	require.Equal(t,
		"https://fcm.googleapis.com/v1/projects/project%20%2F%5C/messages:send",
		getEndpoint(`project /\`))
}

func TestSendOk(t *testing.T) {

	token := getDeviceToken(t)

	msg := &Message{
		Token: token,
		Notification: &Notification{
			Title: "test-title",
			Body:  time.Now().Format(time.RFC3339Nano),
		},
	}

	client := getClient(t)

	for _, sandbox := range []bool{false, true} {
		client.sandbox = sandbox

		// some operations for check reusing token
		for i := 0; i < 3; i++ {
			resp, err := client.Send(context.Background(), msg)
			require.NoError(t, err)
			require.True(t, resp.Ok())
			require.True(t, strings.HasPrefix(resp.Name, "projects/"), resp.Name)
		}
	}
}

func TestSendError(t *testing.T) {

	msg := &Message{
		Token: "-",
	}

	client := getClient(t)
	resp, err := client.Send(context.Background(), msg)
	require.NoError(t, err)
	require.Falsef(t, resp.Ok(), "%#v", resp)

	require.Equal(t,
		&Response{
			StatusCode: 400,
			Error: &SendError{
				Code:    400,
				Message: `The registration token is not a valid FCM registration token`,
				Status:  "INVALID_ARGUMENT",
				Details: json.RawMessage([]byte(`[
      {
        "@type": "type.googleapis.com/google.firebase.fcm.v1.FcmError",
        "errorCode": "INVALID_ARGUMENT"
      },
      {
        "@type": "type.googleapis.com/google.rpc.BadRequest",
        "fieldViolations": [
          {
            "field": "message.token",
            "description": "The registration token is not a valid FCM registration token"
          }
        ]
      }
    ]`)),
			},
		},
		resp)
}

func getClient(t *testing.T) *Client {
	t.Helper()

	svcAccount := getServiceAccount(t)

	client, err := New(svcAccount, false, 2, time.Second)
	require.NoError(t, err)

	return client
}

func getServiceAccount(t *testing.T) []byte {
	t.Helper()

	data, err := test.GetGoogleServiceAccount()
	require.NoError(t, err)

	return data
}

func getDeviceToken(t *testing.T) string {
	t.Helper()

	token, _, err := test.GetPushDevices()
	require.NoError(t, err)

	return token
}
