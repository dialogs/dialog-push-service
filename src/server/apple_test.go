package server

import (
	"testing"
	"crypto/tls"
	"net/http/httptest"
	"net/http"
	"time"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
)

func getApnsProvider(config apnsConfig) APNSDeliveryProvider {
	tasks := make(chan PushTask)
	return APNSDeliveryProvider{tasks: tasks, cert: tls.Certificate{}, config: config}
}

func mockApnsSend(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	defer r.Body.Close()
	grpclog.Printf("Bytes = %s", string(b))
}

func TestAppleSend(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(mockApnsSend))
	defer server.Close()

	config := apnsConfig{host: server.URL, ProjectID: "test"}
	provider := getApnsProvider(config)
	resp := make(chan []string)
	go provider.spawnWorker("test")
	for i := 0; i < 10; i++ {
		provider.tasks <- PushTask{deviceIds: []string{"1"}, body: nil, resp: resp}
	}
	time.Sleep(10 * time.Second)
}
