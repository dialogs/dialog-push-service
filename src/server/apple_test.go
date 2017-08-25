package main

import (
	"encoding/json"
	"testing"

	apns "github.com/sideshow/apns2"

	"go.uber.org/zap"
)

func TestUserInfo(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("Unable to start logger")
	}
	encData := []byte{1, 2}
	provider := APNSDeliveryProvider{logger: logger}
	task := PushTask{body: &PushBody{Body: &PushBody_EncryptedPush{EncryptedPush: &EncryptedPush{EncryptedData: encData, Nonce: 1}}}}
	payload := provider.getPayload(task)
	if payload == nil {
		t.Fatal("Nil payload")
	}
	n := &apns.Notification{Payload: payload}
	bytes, err := n.MarshalJSON()
	if err != nil {
		t.Fatal("JSON marshalling failed")
	}
	res := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &res); err != nil {
		t.Fatalf("Unparseable JSON from APNS: %#v", err)
	}
	if res["user_info"] == nil {
		t.Fatalf("No user_info in notification")
	}
}
