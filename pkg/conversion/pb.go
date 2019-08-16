package conversion

import (
	"bytes"
	"errors"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidIncomingDataType    = errors.New("invalid incoming data type")
	ErrInvalidOutgoingDataType    = errors.New("invalid outgoing data type")
	ErrInvalidIncomingPayloadData = errors.New("invalid incoming payload data")
	ErrEmptyEncryptedPayload      = errors.New("encrypted push without encrypted data")
	ErrNotSupportedAlertPush      = errors.New("alerting pushes are not supported for FCM")
)

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
