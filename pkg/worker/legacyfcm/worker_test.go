package legacyfcm

import (
	"context"
	"errors"
	"testing"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var payload = []byte(`{"notification":{"title":"title","body":"body text"}}`)

func TestWokerNew(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	require.Equal(t, worker.KindFcmLegacy, w.Kind())
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
			Error:     worker.ErrInvalidDeviceToken,
		},
		<-chOut)

	_, ok := <-chOut
	require.False(t, ok)
}

func TestWokerSendErrInvalidIncomingDataType(t *testing.T) {

	cfg := getConfig(t)
	logger := getLogger(t)

	w, err := New(cfg, logger, metric.New())
	require.NoError(t, err)

	chOut := w.Send(context.Background(), &worker.Request{
		Devices: []string{"token1", "token2"},
	})

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: "token1",
			Error:       converter.ErrInvalidIncomingDataType,
		},
		<-chOut)

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: "token2",
			Error:       converter.ErrInvalidIncomingDataType,
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
		Payload: payload,
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
		Payload: payload,
	})

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: token,
		},
		<-chOut)

	require.Equal(t,
		&worker.Response{
			ProjectID:   w.ProjectID(),
			DeviceToken: "token2",
			Error:       errors.New("invalid registration token"),
		},
		<-chOut)

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
		"project-id":     "project-id-123",
		"key":            getLegacyFCMKey(t),
		"nop-mode":       "true",
		"workers":        "-1",
		"converter-kind": converter.KindBinary.String(),
	} {
		src.Set(k, v)
	}

	c, err := NewConfig(src)
	require.NoError(t, err)

	return c
}

func getLegacyFCMKey(t *testing.T) string {
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
