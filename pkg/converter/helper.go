package converter

import (
	"bytes"
	"errors"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/viper"
	"google.golang.org/grpc/status"
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

func ErrorByIncomingMessage(body *api.PushBody) error {

	marshaller := jsonpb.Marshaler{}
	buf := bytes.NewBuffer(nil)
	if err := marshaller.Marshal(buf, body); err != nil {
		st, _ := status.FromError(err)
		return errors.New("incoming body to json:" + st.Message())
	}

	return errors.New("invalid incoming payload data:" + buf.String())
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
