package main
//
//import (
//	"crypto/tls"
//	"fmt"
//	"github.com/edganiukov/fcm"
//	"github.com/sideshow/apns2"
//	"google.golang.org/grpc/grpclog"
//	"gopkg.in/h2non/gock.v1"
//	"net/http"
//	"testing"
//	"time"
//)
//
//func getConfiguredServer(providers map[string]DeliveryProvider) PushingServerImpl {
//	return PushingServerImpl{providers: providers}
//}
//
//func getConfiguredFCMServer(workersCount int) PushingServerImpl {
//	googCfg := googleConfig{Key: "A", host: "lolo"}
//	googCfg.IsSandbox = true
//	provider := GoogleDeliveryProvider{tasks: make(chan PushTask), config: googCfg}
//	for i := 0; i < workersCount; i++ {
//		go provider.spawnWorker(fmt.Sprintf("gcm.%d", i))
//	}
//	return getConfiguredServer(map[string]DeliveryProvider{
//		"FCM": provider,
//	})
//}
//
//func getConfiguredApnsServer() PushingServerImpl {
//	apnsCfg := apnsConfig{}
//	provider := APNSDeliveryProvider{tasks: make(chan PushTask), cert: tls.Certificate{}, config: apnsCfg}
//	for i := 0; i < 3; i++ {
//		go provider.spawnWorker(fmt.Sprintf("apns.%d", i))
//	}
//	return getConfiguredServer(map[string]DeliveryProvider{
//		"APNS": provider,
//	})
//}
//
//func mkPush(ds map[string][]string) *Push {
//	dests := make(map[string]*DeviceIdList)
//	for p, d := range ds {
//		dests[p] = &DeviceIdList{DeviceIds: d}
//	}
//	return &Push{Destinations: dests, Body: &PushBody{Seq: 1, Body: &PushBody_SilentPush{SilentPush: &SilentPush{}}}}
//}
//
//func setupFmcGock(callsHint int) chan int {
//	calls := make(chan int, callsHint)
//	result := fcm.Result{MessageID: "1", RegistrationID: "1"}
//	fcmOkResponse := fcm.Response{Results: []fcm.Result{result}, Success: 1}
//	//fcmNotOkResponse := `
//	//	{
//	//	"failure":1,"success":0,"multicast_id":1,"canonical_ids":1,
//	//	"results":[
//	//		{"message_id":"1","registration_id":"A","error":"NotRegistered"},
//	//		{"message_id":"1","registration_id":"B","error":""}
//	//	]}
//	//	`
//	gock.New("https://fcm.googleapis.com").
//		Post("/fcm/send").
//		Persist().
//		Reply(200).
//		SetHeader("Content-Type", "application/json").
//	//BodyString(fcmNotOkResponse).
//		JSON(fcmOkResponse).
//		Map(func(r *http.Response) *http.Response {
//		calls <- 1
//		return r
//	})
//	return calls
//}
//
//func apnsMatcher(r *http.Request, g *gock.Request) (bool, error) {
//	grpclog.Printf("URL = %s", r.URL.Path)
//	return true, nil
//	//strings.HasPrefix(r.URL.Path, "/3/device")
//}
//
//func setupApnsGock(callsHint int) chan int {
//	calls := make(chan int, callsHint)
//	result := apns2.Response{StatusCode: 200}
//	gock.New("https://api.push.apple.com").
//		AddMatcher(apnsMatcher).
//		Persist().
//		Reply(200).
//		JSON(result).
//		Map(func(r *http.Response) *http.Response {
//		calls <- 1
//		return r
//	})
//	return calls
//}
//
//func expectCallsCount(t testing.TB, calls chan int, expected int, timeout time.Duration) {
//	var cnt = 0
//	for {
//		select {
//		case r := <-calls:
//			cnt += r
//
//			if cnt == expected {
//				return
//			}
//		case <-time.After(timeout):
//			t.Fatal("Timed out")
//			return
//		}
//	}
//}
//
//func drain(resps chan *Response) {
//	for range resps {
//	}
//}
//
//func TestFCMStreaming(t *testing.T) {
//	//defer gock.Off()
//	var pushes = 5
//	//fcmCalls := setupFmcGock(pushes)
//	server := getConfiguredFCMServer(3)
//	reqs := make(chan *Push)
//	resps := make(chan *Response)
//	go drain(resps)
//	go server.startStream(reqs, resps)
//	go func() {
//		for i := 0; i < pushes; i++ {
//			reqs <- mkPush(map[string][]string{"FCM": {"A", "B"}})
//		}
//	}()
//	time.Sleep(10 * time.Second)
//	//expectCallsCount(t, fcmCalls, pushes, time.Duration(pushes) * time.Millisecond)
//}
//
//func BenchmarkFCMStream(b *testing.B) {
//	defer gock.Off()
//	b.ReportAllocs()
//	fcmCalls := setupFmcGock(b.N)
//	server := getConfiguredFCMServer(1)
//	reqs := make(chan *Push)
//	resps := make(chan *Response)
//	go drain(resps)
//	go server.startStream(reqs, resps)
//	go func() {
//		pusharr := make([]*Push, b.N)
//		var push *Push
//		for i := 0; i < b.N; i++ {
//			push = mkPush(map[string][]string{"FCM": {"A", "B"}})
//			push.GetBody().Seq = int32(i)
//			pusharr = append(pusharr, push)
//		}
//		for _, push := range pusharr {
//			reqs <- push
//		}
//	}()
//	expectCallsCount(b, fcmCalls, b.N, 10 * time.Second)
//}
//
////func TestApnsStreaming(t *testing.T) {
////	defer gock.Off()
////	var pushes = 5
////	apnsCalls := setupApnsGock(pushes)
////	server := getConfiguredApnsServer()
////	reqs := make(chan *Push)
////	resps := make(chan *Response)
////	go server.startStream(reqs, resps)
////	go func() {
////		for i := 0; i < pushes; i++ {
////			reqs <- mkPush(map[string][]string{"APNS": []string{"A", "B"}})
////		}
////	}()
////	expectCallsCount(t, apnsCalls, pushes, time.Duration(pushes) * time.Millisecond)
////
////}
