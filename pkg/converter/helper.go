package converter

import (
	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/spf13/viper"
)

func GetKindFromConfig(src *viper.Viper) Kind {
	return KindByString(src.GetString("converter-kind"))
}

func GetAPIPushBody(in interface{}) (*api.PushBody, error) {

	body, ok := in.(*api.PushBody)
	if !ok {
		return nil, ErrInvalidIncomingDataType
	}

	return body, nil
}

func GetBinaryPushBody(in interface{}) ([]byte, error) {

	body, ok := in.([]byte)
	if !ok {
		return nil, ErrInvalidIncomingDataType
	}

	return body, nil
}

func PeerTypeProtobufToMPS(peerType api.PeerType) int {
	switch peerType {
	case api.Private:
		return 1
	case api.Group:
		return 2
	case api.SIP:
		return 4
	default:
		return 0
	}
}
