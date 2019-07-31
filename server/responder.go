package server

import (
	"context"

	"github.com/dialogs/dialog-push-service/pkg/api"
)

type Responder interface {
	Send(projectId string, failures *api.DeviceIdList) error
}

type StreamResponder struct {
	ctx      context.Context
	response chan<- *PushResult
}

func NewStreamResponder(ctx context.Context, response chan<- *PushResult) *StreamResponder {
	return &StreamResponder{
		ctx:      ctx,
		response: response,
	}
}

func (s *StreamResponder) Send(projectId string, failures *api.DeviceIdList) error {
	if len(failures.DeviceIds) == 0 {
		return nil
	}

	select {
	case s.response <- &PushResult{ProjectId: projectId, Failures: failures}:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

type UnaryResponder struct {
	ctx      context.Context
	response chan<- *PushResult
}

func NewUnaryResponder(ctx context.Context, response chan<- *PushResult) *UnaryResponder {
	return &UnaryResponder{
		ctx:      ctx,
		response: response,
	}
}

func (u *UnaryResponder) Send(projectId string, failures *api.DeviceIdList) error {
	select {
	case u.response <- &PushResult{ProjectId: projectId, Failures: failures}:
		return nil
	case <-u.ctx.Done():
		return u.ctx.Err()
	}
}
