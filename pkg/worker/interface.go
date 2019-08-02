package worker

import (
	"context"
)

type IWorker interface {
	Kind() Kind
	ProviderID() string
	Send(context.Context, *Request) <-chan *Response
}
