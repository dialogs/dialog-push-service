package service

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2ans"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2fcm"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2legacyfcm"

	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/dialogs/dialog-push-service/pkg/worker/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker/legacyfcm"
	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {

	applePem, err := test.GetPathToIOSCertificatePem()
	require.NoError(t, err)

	fcmKey, err := test.GetAccountKey()
	require.NoError(t, err)

	fcmServiceAccount, err := test.GetPathToGoogleServiceAccount()
	require.NoError(t, err)

	const file = "config.yaml"
	fileData := getConfigSrc(t, applePem, fcmKey, fcmServiceAccount)
	require.NoError(t, ioutil.WriteFile(file, []byte(fileData), os.ModePerm))
	defer func() { require.NoError(t, os.Remove(file)) }()

	v := viper.New()
	v.SetConfigFile(file)
	require.NoError(t, v.ReadInConfig())

	cfg, err := NewConfig(v)
	require.NoError(t, err)
	require.Equal(t,
		&Config{
			ApiPort:   "8010",
			AdminPort: "8011",
			Fcm: []*fcm.Config{
				{
					ServiceAccount: fcmServiceAccount,
					SendTries:      10,
					SendTimeout:    time.Second * 2,
					Config: &worker.Config{
						ProjectID:     "p-1",
						NopMode:       true,
						CountThreads:  4,
						ConverterKind: converter.KindApi,
					},
					APIConfig: &api2fcm.Config{
						AllowAlerts: true,
						Sandbox:     true,
					},
				},
			},
			LegacyFcm: []*legacyfcm.Config{
				{
					ServerKey: fcmKey,
					SendTries: 10,
					Config: &worker.Config{
						ProjectID:     "p-2",
						NopMode:       true,
						CountThreads:  3,
						ConverterKind: converter.KindApi,
					},
					APIConfig: &api2legacyfcm.Config{
						AllowAlerts: true,
						Sandbox:     true,
					},
				},
			},
			Ans: []*ans.Config{
				{
					PemFile: applePem,
					Config: &worker.Config{
						ProjectID:     "p-3",
						NopMode:       true,
						CountThreads:  2,
						ConverterKind: converter.KindApi,
					},
					APIConfig: &api2ans.Config{
						AllowAlerts: true,
						Topic:       "im.dlg.dialog-ee",
						Sound:       "dialog.wav",
					},
				},
			},
		},
		cfg)
}

func getConfigSrc(t *testing.T, applePem, fcmKey, fcmServiceAccount string) string {
	t.Helper()

	return `
grpc-port: 8010
http-port: 8011
fcm-v1:
  - project-id: p-1
    service-account: ` + fcmServiceAccount + `
    nop-mode: true
    send-tries: 10
    send-timeout: 2s
    allow-alerts: true
    sandbox: true
    workers: 4
google:
  - project-id: p-2
    key: ` + fcmKey + `
    nop-mode: true
    retries: 10
    allow-alerts: true
    sandbox: true
    workers: 3
apple:
  - project-id: p-3
    topic: im.dlg.dialog-ee
    nop-mode: true
    voip: true
    allow-alerts: true
    sandbox: true
    pem: ` + applePem + `
    sound: "dialog.wav"
    workers: 2
`
}
