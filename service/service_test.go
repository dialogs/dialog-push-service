package service

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/service"
	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ltime | log.Lmicroseconds)
}

func TestService(t *testing.T) {

	listener := newListener(t)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	_, apiPort, err := net.SplitHostPort(address)
	require.NoError(t, err)

	cfgPath := saveServiceConfig(t, apiPort)
	defer func() { require.NoError(t, os.Remove(cfgPath)) }()

	v := viper.New()
	v.SetConfigFile(cfgPath)
	require.NoError(t, v.ReadInConfig())

	svc, err := New(v, getLogger(t))
	require.NoError(t, err)
	require.NotNil(t, svc)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.Run())
	}()

	defer func() {
		require.NoError(t, svc.Close())
		wg.Wait()
	}()

	clientOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Second),
		grpc.WithBlock(),
	}
	require.NoError(t, service.PingGRPC(address, 2, clientOpts...))

	conn, err := grpc.Dial(address, clientOpts...)
	require.NoError(t, err)

	for _, testInfo := range []struct {
		Name string
		Func func(*testing.T)
	}{
		{
			Name: "ping",
			Func: func(*testing.T) { testPing(t, conn) },
		},
		{
			Name: "single push: invalid incoming data",
			Func: func(*testing.T) { testSinglePushInvalidIncomigData(t, conn) },
		},
		{
			Name: "single push: success",
			Func: func(*testing.T) { testSinglePushSuccess(t, conn) },
		},
		{
			Name: "push stream: invalid incoming data",
			Func: func(*testing.T) { testPushStreamInvalidIncomigData(t, conn) },
		},
		{
			Name: "push stream: success",
			Func: func(*testing.T) { testPushStreamSuccess(t, conn) },
		},
	} {

		if !t.Run(testInfo.Name, testInfo.Func) {
			return
		}
	}
}

func testPushStreamSuccess(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	android, ios, err := test.GetPushDevices()
	require.NoError(t, err)

	stream, err := client.PushStream(context.Background())
	require.NoError(t, err)

	defer func() {
		require.NoError(t, stream.CloseSend())
	}()

	for i := 0; i < 3; i++ {
		require.NoError(t, stream.Send(&api.Push{
			Destinations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"token1", android, "token2"}},
				"p-google":  &api.DeviceIdList{DeviceIds: []string{"token3", android, "token4"}},
				"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", ios, "token6"}},
				"p-unknown": &api.DeviceIdList{DeviceIds: []string{"token7", android, ios, "token8"}},
			},
			Body: &api.PushBody{
				Body: &api.PushBody_EncryptedPush{
					EncryptedPush: &api.EncryptedPush{
						EncryptedData: []byte("push body"),
					},
				},
			},
		}))

		res, err := stream.Recv()
		require.NoError(t, err)

		require.NoError(t, err)
		require.Equal(t,
			&api.Response{
				ProjectInvalidations: map[string]*api.DeviceIdList{
					"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"token1", "token2"}},
					"p-google":  &api.DeviceIdList{DeviceIds: []string{"token3", "token4"}},
					"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", "token6"}},
					"p-unknown": &api.DeviceIdList{},
				},
			},
			res)
	}
}

func testPushStreamInvalidIncomigData(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	android, ios, err := test.GetPushDevices()
	require.NoError(t, err)

	stream, err := client.PushStream(context.Background())
	require.NoError(t, err)

	defer func() {
		require.NoError(t, stream.CloseSend())
	}()

	for i := 0; i < 3; i++ {
		require.NoError(t, stream.Send(&api.Push{
			Destinations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"token1", android, "token2"}},
				"p-google":  &api.DeviceIdList{DeviceIds: []string{"token3", android, "token4"}},
				"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", ios, "token6"}},
				"p-unknown": &api.DeviceIdList{DeviceIds: []string{"", "-", android, ios, "token4"}},
			},
		}))

		res, err := stream.Recv()
		require.NoError(t, err)

		require.NoError(t, err)
		require.Equal(t,
			&api.Response{
				ProjectInvalidations: map[string]*api.DeviceIdList{
					"p-fcm":     &api.DeviceIdList{},
					"p-google":  &api.DeviceIdList{},
					"p-apple":   &api.DeviceIdList{},
					"p-unknown": &api.DeviceIdList{},
				},
			},
			res)
	}
}

func testSinglePushSuccess(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	android, ios, err := test.GetPushDevices()
	require.NoError(t, err)
	require.NotEmpty(t, ios)
	require.NotEmpty(t, android)

	res, err := client.SinglePush(context.Background(), &api.Push{
		Destinations: map[string]*api.DeviceIdList{
			"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", android, "token1"}},
			"p-google":  &api.DeviceIdList{DeviceIds: []string{"", "-", android, "token2"}},
			"p-apple":   &api.DeviceIdList{DeviceIds: []string{"", "-", ios, "token3"}},
			"p-unknown": &api.DeviceIdList{DeviceIds: []string{"", "-", android, ios, "token4"}},
		},
		Body: &api.PushBody{
			Body: &api.PushBody_EncryptedPush{
				EncryptedPush: &api.EncryptedPush{
					EncryptedData: []byte("push body"),
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t,
		&api.Response{
			ProjectInvalidations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", "token1"}},
				"p-google":  &api.DeviceIdList{DeviceIds: []string{"", "-", "token2"}},
				"p-apple":   &api.DeviceIdList{DeviceIds: []string{"", "-", "token3"}},
				"p-unknown": &api.DeviceIdList{},
			},
		},
		res)
}

func testSinglePushInvalidIncomigData(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	android, ios, err := test.GetPushDevices()
	require.NoError(t, err)

	res, err := client.SinglePush(context.Background(), &api.Push{
		Destinations: map[string]*api.DeviceIdList{
			"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"token1", android, "token2"}},
			"p-google":  &api.DeviceIdList{DeviceIds: []string{"token3", android, "token4"}},
			"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", ios, "token6"}},
			"p-unknown": &api.DeviceIdList{DeviceIds: []string{"token7", android, ios, "token8"}},
		},
	})

	require.NoError(t, err)
	require.Equal(t,
		&api.Response{
			ProjectInvalidations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{},
				"p-google":  &api.DeviceIdList{},
				"p-apple":   &api.DeviceIdList{},
				"p-unknown": &api.DeviceIdList{},
			},
		},
		res)
}

func testPing(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	res, err := client.Ping(context.Background(), &api.PingRequest{})
	require.NoError(t, err)
	require.Equal(t, &api.PongResponse{}, res)
}

func saveServiceConfig(t *testing.T, apiPort string) string {
	t.Helper()

	applePem, err := test.GetPathToIOSCertificatePem()
	require.NoError(t, err)

	fcmKey, err := test.GetAccountKey()
	require.NoError(t, err)

	fcmServiceAccount, err := test.GetPathToGoogleServiceAccount()
	require.NoError(t, err)

	apiPortInt, err := strconv.Atoi(apiPort)
	require.NoError(t, err)

	adminPort := strconv.Itoa(apiPortInt + 1)

	fileData := `
grpc-port: ` + apiPort + `
http-port: ` + adminPort + `
fcm-v1:
  - project-id: p-fcm
    service-account: ` + fcmServiceAccount + `
    send-tries: 10
    send-timeout: 2s
    allow-alerts: true
google:
  - project-id: p-google
    key: ` + fcmKey + `
    retries: 10
    allow-alerts: true
apple:
  - project-id: p-apple
    allow-alerts: true
    pem: ` + applePem + `
    sound: "dialog.wav"
`

	const file = "config.yaml"
	require.NoError(t, ioutil.WriteFile(file, []byte(fileData), os.ModePerm))

	return file
}

func newListener(t *testing.T) net.Listener {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	return l
}

func getLogger(t *testing.T) *zap.Logger {
	t.Helper()

	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logCfg.Build()
	require.NoError(t, err)

	return logger
}
