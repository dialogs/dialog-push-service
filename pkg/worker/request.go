package worker

import "github.com/dialogs/dialog-push-service/pkg/provider"

type Request struct {
	Devices       []string
	CorrelationID string
	Payload       provider.IRequest
}
