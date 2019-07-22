package server

import (
	"context"
	"testing"
	"time"

	. "github.com/franela/goblin"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	sent1 := false
	onSend1 := func(task PushTask) { sent1 = true }

	sent2 := false
	onSend2 := func(task PushTask) { sent2 = true }

	noopCfg1 := noopConfig{ProjectID: "test1", OnSend: onSend1}
	noopCfg1.Workers = 1

	noopCfg2 := noopConfig{ProjectID: "test2", OnSend: onSend2}
	noopCfg2.Workers = 2

	testConfig := &serverConfig{Noop: []noopConfig{noopCfg1, noopCfg2}, ReadQueueSize: 2, WriteQueueSize: 2}
	server := newPushingServer(testConfig)
	g := Goblin(t)
	g.Describe("Server", func() {
		g.It("Should send single pushes", func(done Done) {
			push := &Push{
				Destinations: map[string]*DeviceIdList{
					"test1": {DeviceIds: []string{"a", "b", "c", "d", "e"}},
					"test2": {DeviceIds: []string{"f", "g"}},
				},
				CorrelationId: "test",
				Body:          &PushBody{CollapseKey: "ckey", Seq: 1},
			}

			resp, err := server.SinglePush(context.Background(), push)
			require.NoError(t, err)
			require.NotNil(t, resp)

			go func() {
				for {
					if sent1 && sent2 {
						done()
					}
					<-time.After(50 * time.Millisecond)
				}
			}()
		})
	})
}
