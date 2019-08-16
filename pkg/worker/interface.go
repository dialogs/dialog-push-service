package worker

import (
	"context"

	"github.com/dialogs/dialog-push-service/pkg/conversion"
)

type IWorker interface {
	Kind() Kind
	ProjectID() string
	Send(context.Context, *Request) <-chan *Response
	ConversionConfig() *conversion.Config
	ExistVoIP() bool
}
