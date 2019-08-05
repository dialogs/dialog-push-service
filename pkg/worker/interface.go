package worker

import (
	"context"
)

type IWorker interface {
	Kind() Kind
	ProjectID() string
	Send(context.Context, *Request) <-chan *Response
}
