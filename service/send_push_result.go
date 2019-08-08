package service

type sendPushResult struct {
	ProjectID           string
	InvalidationDevices []string
}

func newSendPushResult(projectID string) *sendPushResult {
	return &sendPushResult{
		ProjectID:           projectID,
		InvalidationDevices: make([]string, 0),
	}
}
