package gcm

import (
	"context"
	"errors"
	"testing"

	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider/gcm"
	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var notification = []byte(`{"title":"title","body":"body text"}`)

func TestWokerNew(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	require.Equal(t, worker.KindGcm, w.Kind())
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
		Payload: &gcm.Request{},
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
		Payload: &gcm.Request{Notification: notification},
	})

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: token,
		},
		<-chOut)

	fcmError := errors.New(gcm.ErrorCodeInvalidRegistration)

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
		"project-id": "project-id-123",
		"key":        getGcmKey(t),
		"nop-mode":   "true",
		"workers":    "-1",
	} {
		src.Set(k, v)
	}

	c, err := NewConfig(src)
	require.NoError(t, err)

	return c
}

func getGcmKey(t *testing.T) []byte {
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
