package binary

import (
	"encoding/json"

	"github.com/dialogs/dialog-push-service/pkg/converter"
)

type Request struct {
}

func NewRequestConverter() *Request {
	return &Request{}
}

func (r *Request) Convert(in interface{}, out interface{}) error {

	body, err := converter.GetBinaryPushBody(in)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, out); err != nil {
		return err
	}

	return nil
}
