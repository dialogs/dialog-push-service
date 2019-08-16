package service

import (
	svcinfo "github.com/dialogs/dialog-push-service/pkg/info"

	"github.com/dialogs/dialog-go-lib/service/info"
)

// Info of the service
func Info() *info.Info {
	return svcinfo.New("push")
}
