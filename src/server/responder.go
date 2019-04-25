package main

import "context"

type Responder interface {
	Send(projectId string, failures *DeviceIdList)
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

func (s *StreamResponder) Send(projectId string, failures *DeviceIdList) {
	if len(failures.DeviceIds) == 0 {
		return
	}

	select {
	case s.response <- &PushResult{ProjectId: projectId, Failures: failures}:
	case <-s.ctx.Done():
	}
}

type UnaryResponder struct {
	response chan<- *PushResult
}

func NewUnaryResponder(response chan<- *PushResult) *UnaryResponder {
	return &UnaryResponder{
		response: response,
	}
}

func (u *UnaryResponder) Send(projectId string, failures *DeviceIdList) {
	u.response <- &PushResult{ProjectId: projectId, Failures: failures}
}
