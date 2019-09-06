package service

import (
	"context"
	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"io"
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
	// SAST: exception 'gets dinamic data'
	if len(apiPort) > 50 {
		t.Fatal("invalid data size:" + apiPort)
	}

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
			Name: "single push: alerting push success",
			Func: func(*testing.T) { testSingleAlertingPushSuccess(t, conn) },
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

	destinations := map[string]*api.DeviceIdList{
		"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"token1", android, "token2"}},
		"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"token3", android, "token4"}},
		"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", ios, "token6"}},
		"p-unknown": &api.DeviceIdList{DeviceIds: []string{"token7", android, ios, "token8"}},
	}

	const CountSend = 3

	for i := 0; i < CountSend; i++ {
		require.NoError(t, stream.Send(&api.Push{
			Destinations: destinations,
			Body: &api.PushBody{
				Body: &api.PushBody_EncryptedPush{
					EncryptedPush: &api.EncryptedPush{
						EncryptedData: []byte("push body"),
					},
				},
			},
		}))
	}

	countDestinations := len(destinations) - 1 // -1 - exclude results with unknown project

	for i := 0; i < CountSend*countDestinations; i++ {
		res, err := stream.Recv()
		require.NoError(t, err)

		for projectID := range res.ProjectInvalidations {
			switch projectID {
			case "p-fcm":
				require.Equal(t,
					&api.Response{
						ProjectInvalidations: map[string]*api.DeviceIdList{
							"p-fcm": &api.DeviceIdList{DeviceIds: []string{"token1", "token2"}},
						},
					},
					res)

			case "p-gcm":
				require.Equal(t,
					&api.Response{
						ProjectInvalidations: map[string]*api.DeviceIdList{
							"p-gcm": &api.DeviceIdList{DeviceIds: []string{"token3", "token4"}},
						},
					},
					res)
			case "p-apple":
				require.Equal(t,
					&api.Response{
						ProjectInvalidations: map[string]*api.DeviceIdList{
							"p-apple": &api.DeviceIdList{DeviceIds: []string{"token5", "token6"}},
						},
					},
					res)
			default:
				t.Fatal("invalid project" + projectID)
			}
		}
	}

	checkStreamEnd(t, stream)
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
		push := &api.Push{
			Destinations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"token1", android, "token2"}},
				"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"token3", android, "token4"}},
				"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", ios, "token6"}},
				"p-unknown": &api.DeviceIdList{DeviceIds: []string{"", "-", android, ios, "token4"}},
			},
		}

		require.NoError(t, stream.Send(push))
	}

	checkStreamEnd(t, stream)
}

func testSinglePushSuccess(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	android, ios, err := test.GetPushDevices()
	require.NoError(t, err)
	require.NotEmpty(t, ios)
	require.NotEmpty(t, android)

	logrus.Info("===test11====")
	res, err := client.SinglePush(context.Background(), &api.Push{
		Destinations: map[string]*api.DeviceIdList{
			"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", android, "token1"}},
			"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", android, "token2"}},
			"p-apple":   &api.DeviceIdList{DeviceIds: []string{"", "-", ios, "token3"}},
			"p-unknown": &api.DeviceIdList{DeviceIds: []string{"", "-", android, ios, "token4"}},
		},
		Body: &api.PushBody{
			Body: &api.PushBody_EncryptedPush{
				EncryptedPush: &api.EncryptedPush{
					PublicAlertingPush: &api.AlertingPush{
						AlertBody:  nil,
						AlertTitle: nil,
						Badge:      0,
						Peer:       nil,
						Mid:        &types.StringValue{Value:"testEncryptedValue"},
						Category:   nil,
					},
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
				"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", "token2"}},
				"p-apple":   &api.DeviceIdList{DeviceIds: []string{"", "-", "token3"}},
				"p-unknown": &api.DeviceIdList{},
			},
		},
		res)
}

func testSingleAlertingPushSuccess(t *testing.T, conn *grpc.ClientConn) {

	client := api.NewPushingClient(conn)

	android, ios, err := test.GetPushDevices()
	require.NoError(t, err)
	require.NotEmpty(t, ios)
	require.NotEmpty(t, android)

	logrus.Info("===test12====")
	res, err := client.SinglePush(context.Background(), &api.Push{
		Destinations: map[string]*api.DeviceIdList{
			"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", android, "token1"}},
			"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", android, "token2"}},
			"p-apple":   &api.DeviceIdList{DeviceIds: []string{"", "-", ios, "token3"}},
			"p-unknown": &api.DeviceIdList{DeviceIds: []string{"", "-", android, ios, "token4"}},
		},
		Body: &api.PushBody{
			Body: &api.PushBody_AlertingPush{
				AlertingPush: &api.AlertingPush{
					AlertBody:  nil,
					AlertTitle: nil,
					Badge:      0,
					Peer:       nil,
					Mid:        &types.StringValue{Value:"testMidMessage"},
					Category:   nil,
				},
			},
				//EncryptedPush: &api.EncryptedPush{
				//	EncryptedData: []byte("push body"),
				//},
			//},
		},
	})

	require.NoError(t, err)
	require.Equal(t,
		&api.Response{
			ProjectInvalidations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", "token1"}},
				"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"", "-", "token2"}},
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
			"p-gcm":     &api.DeviceIdList{DeviceIds: []string{"token3", android, "token4"}},
			"p-apple":   &api.DeviceIdList{DeviceIds: []string{"token5", ios, "token6"}},
			"p-unknown": &api.DeviceIdList{DeviceIds: []string{"token7", android, ios, "token8"}},
		},
	})

	require.NoError(t, err)
	require.Equal(t,
		&api.Response{
			ProjectInvalidations: map[string]*api.DeviceIdList{
				"p-fcm":     &api.DeviceIdList{},
				"p-gcm":     &api.DeviceIdList{},
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

	gcmKey, err := test.GetAccountKey()
	require.NoError(t, err)

	fcmServiceAccount, err := test.GetPathToGoogleServiceAccount()
	require.NoError(t, err)

	apiPortInt, err := strconv.Atoi(apiPort)
	require.NoError(t, err)

	adminPort := strconv.Itoa(apiPortInt + 1)

	fileData := `
grpc-port: ` + apiPort + `
http-port: ` + adminPort + `
fcm:
  - project-id: p-fcm
    service-account: ` + fcmServiceAccount + `
    send-tries: 10
    send-timeout: 2s
    allow-alerts: true
google:
  - project-id: p-gcm
    key: ` + string(gcmKey) + `
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

func checkStreamEnd(t *testing.T, stream api.Pushing_PushStreamClient) {
	t.Helper()

	ch := make(chan error)

	select {
	case <-time.After(time.Second):
		// test: ok
		require.NoError(t, stream.CloseSend())

	case <-func() <-chan error {
		go func() {
			res, err := stream.Recv()
			require.Nil(t, res)
			ch <- err
		}()
		return ch
	}():

		t.Fatal("can't returns value")
	}

	require.Equal(t, io.EOF, <-ch)
}
