package worker

type Request struct {
	Devices       []string
	CorrelationID string
	Payload       interface{}
}
