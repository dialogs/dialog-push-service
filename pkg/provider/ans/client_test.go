package ans

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/stretchr/testify/require"
)

// Environment for tests:
// 1. download certificate PEM
// 2. create environment variable "APPLE_PUSH_CERTIFICATE" with path to certificate
// 3. create file with devices tokens. format:
//	{
//     "android": "<token>",
//     "ios": "<token>"
//	}
// 4. create environment variable "PUSH_DEVICES" with path to file with devices tokens

func TestSendOk(t *testing.T) {

	client := getClient(t)

	payload := map[string]interface{}{
		"aps": map[string]interface{}{
			"alert": map[string]interface{}{
				"title": "test-message",
				"body":  time.Now().Format(time.RFC3339Nano),
			},
		},
	}

	jPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	req := &Request{
		Token:   getDeviceToken(t),
		Payload: jPayload,
	}

	res, err := client.Send(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, res.ID, 36) // example: CDB997A0-0C7C-8E2E-DBB5-13E89D5C756E

	require.Equal(t,
		&Response{
			ID:         res.ID,
			StatusCode: 200,
			Reason:     "",
			Timestamp:  time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		res)
}

func TestSendError(t *testing.T) {

	client := getClient(t)

	req := &Request{}
	res, err := client.Send(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, res.ID, 36) // example: CDB997A0-0C7C-8E2E-DBB5-13E89D5C756E

	require.Equal(t,
		&Response{
			ID:         res.ID,
			StatusCode: 400,
			Reason:     "MissingDeviceToken",
			Timestamp:  time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		res)
}

func TestGetTopics(t *testing.T) {

	{
		topics, err := GetTopics([]byte{
			0x30, 0x18, 0xc, 0x1, 'a',
			0x30, 0x2, 0xc, 0x2, 'b', 'c',
			0xc, 0x01, 'd',
			0x30, 0x3, 0xc, 0x3, 'e', 'f', 'g',
			0xc, 0x03, 'h', 'i', 'j'})
		require.NoError(t, err)
		require.Equal(t, []string{"a", "bc", "d", "efg", "hij"}, topics)
	}

	{
		topics, err := GetTopics([]byte{
			0x30, 0x3, 0xc, 0x1, 'a'})
		require.NoError(t, err)
		require.Equal(t, []string{"a"}, topics)
	}

	{
		topics, err := GetTopics([]byte{
			0x30, 0x8, 0xc, 0x1, 'a',
			0xc, 0x03, 'h', 'i', 'j'})
		require.NoError(t, err)
		require.Equal(t, []string{"a", "hij"}, topics)
	}

}

func getClient(t *testing.T) *Client {
	t.Helper()

	pem := getCertificatePem(t)
	client, err := NewFromPem(pem)
	require.NoError(t, err)

	return client
}

func getCertificatePem(t *testing.T) []byte {
	t.Helper()

	pem, err := test.GetIOSCertificatePem()
	require.NoError(t, err)

	return pem
}

func getDeviceToken(t *testing.T) string {
	t.Helper()

	_, token, err := test.GetPushDevices()
	require.NoError(t, err)

	return token
}
