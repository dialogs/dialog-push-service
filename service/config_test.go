package service

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/conversion"
	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/dialogs/dialog-push-service/pkg/worker/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker/gcm"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {

	applePem, err := test.GetPathToIOSCertificatePem()
	require.NoError(t, err)

	gcmKey, err := test.GetAccountKey()
	require.NoError(t, err)

	fcmServiceAccount, err := test.GetPathToGoogleServiceAccount()
	require.NoError(t, err)

	const file = "config.yaml"
	fileData := getConfigSrc(t, applePem, string(gcmKey), fcmServiceAccount)
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
					Retries:        10,
					Timeout:        time.Second * 2,
					Config: &worker.Config{
						ProjectID:    "p-1",
						NopMode:      true,
						CountThreads: 4,
						Sandbox:      true,
						Config: &conversion.Config{
							AllowAlerts: true,
						},
					},
				},
			},
			Gcm: []*gcm.Config{
				{
					ServerKey: string(gcmKey),
					Retries:   10,
					Timeout:   2 * time.Second,
					Config: &worker.Config{
						ProjectID:    "p-2",
						NopMode:      true,
						CountThreads: 3,
						Sandbox:      true,
						Config: &conversion.Config{
							AllowAlerts: true,
						},
					},
				},
			},
			Ans: []*ans.Config{
				{
					PemFile: applePem,
					Retries: 10,
					Timeout: 2 * time.Second,
					Config: &worker.Config{
						ProjectID:    "p-3",
						NopMode:      true,
						CountThreads: 2,
						Sandbox:      true,
						Config: &conversion.Config{
							AllowAlerts: true,
							Topic:       "im.dlg.dialog-ee",
							Sound:       "dialog.wav",
						},
					},
				},
			},
		},
		cfg)
}

func getConfigSrc(t *testing.T, applePem, gcmKey, fcmServiceAccount string) string {
	t.Helper()

	return `
grpc-port: 8010
http-port: 8011
fcm:
  - project-id: p-1
    service-account: ` + fcmServiceAccount + `
    nop-mode: true
    retries: 10
    timeout: 2s
    allow-alerts: true
    sandbox: true
    workers: 4
google:
  - project-id: p-2
    key: ` + gcmKey + `
    nop-mode: true
    retries: 10
    timeout: 2s
    allow-alerts: true
    sandbox: true
    workers: 3
apple:
  - project-id: p-3
    topic: im.dlg.dialog-ee
    retries: 10
    timeout: 2s
    nop-mode: true
    voip: true
    allow-alerts: true
    sandbox: true
    pem: ` + applePem + `
    sound: "dialog.wav"
    workers: 2
`
}
