package fcm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider/fcm"
	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWokerNew(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	require.Equal(t, worker.KindFcm, w.Kind())
	require.Equal(t, "project-id-123", w.ProjectID())
	require.Equal(t, true, w.NoOpMode())
}

func TestWokerSendErrInvalidDeviceToken(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	chOut := w.Send(context.Background(), &worker.Request{})
	require.Equal(t,
		&worker.Response{
			ProjectID: w.ProjectID(),
			Error:     worker.ErrEmptyToken,
		},
		<-chOut)

	_, ok := <-chOut
	require.False(t, ok)
}

func TestWokerSendNopOk(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	chOut := w.Send(context.Background(), &worker.Request{
		Devices: []string{"token1", "token2"},
		Payload: &fcm.Message{Notification: &fcm.Notification{Title: "title"}},
	})

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: "token1",
		},
		<-chOut)

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: "token2",
		},
		<-chOut)

	_, ok := <-chOut
	require.False(t, ok)
}

func TestWokerSendOk(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)
	token := getDeviceToken(t)

	cfg.NopMode = false

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	chOut := w.Send(context.Background(), &worker.Request{
		Devices: []string{token, "token2", token},
		Payload: &fcm.Message{Notification: &fcm.Notification{Title: "title"}},
	})

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: token,
		},
		<-chOut)

	fcmError := &fcm.SendError{
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
    ]`))}

	res := <-chOut
	require.Equal(t, fcmError, res.Error.(*worker.ResponseError).Err())
	require.Equal(t, worker.NewResponseErrorBadDeviceToken(fcmError), res.Error)
	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: "token2",
			Error:       worker.NewResponseErrorBadDeviceToken(fcmError),
		},
		res)

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: token,
		},
		<-chOut)

	_, ok := <-chOut
	require.False(t, ok)
}

func TestGetStringValueFromJSON(t *testing.T) {

	require.Equal(t,
		[]string{},
		getStringValueFromJSON(
			[]byte(""),
			"token"))

	require.Equal(t,
		[]string{"value2"},
		getStringValueFromJSON(
			[]byte(`{"token1": "value1", "token2": "value2"}`),
			"token2"))

	const src = `[
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
	]`

	require.Equal(t,
		[]string{"message.token"},
		getStringValueFromJSON(
			[]byte(src),
			"field"))

	require.Equal(t,
		[]string{
			"type.googleapis.com/google.firebase.fcm.v1.FcmError",
			"type.googleapis.com/google.rpc.BadRequest",
		},
		getStringValueFromJSON(
			[]byte(src),
			"@type"))
}

func getLogger(t *testing.T) *zap.Logger {
	t.Helper()

	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logCfg.Build()
	require.NoError(t, err)

	return logger
}

func getConfig(t *testing.T) *Config {
	t.Helper()

	src := viper.New()
	for k, v := range map[string]interface{}{
		"project-id":      "project-id-123",
		"service-account": getPathToServiceAccount(t),
		"nop-mode":        "true",
		"workers":         "-1",
	} {
		src.Set(k, v)
	}

	c, err := NewConfig(src)
	require.NoError(t, err)

	return c
}

func getPathToServiceAccount(t *testing.T) string {
	t.Helper()

	key, err := test.GetPathToGoogleServiceAccount()
	require.NoError(t, err)

	return key
}

func getDeviceToken(t *testing.T) string {
	t.Helper()

	token, _, err := test.GetPushDevices()
	require.NoError(t, err)

	return token
}
