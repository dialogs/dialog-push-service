package converter

import "errors"

var (
	ErrInvalidIncomingDataType    = errors.New("invalid incoming data type")
	ErrInvalidOutgoingDataType    = errors.New("invalid outgoing data type")
	ErrInvalidIncomingPayloadData = errors.New("invalid incoming payload data")
	ErrEmptyEncryptedPayload      = errors.New("encrypted push without encrypted data")
	ErrNotSupportedAlertPush      = errors.New("alerting pushes are not supported for FCM")
)

type IRequestConverter interface {
	Convert(in interface{}, out interface{}) error
}
