package info

import "github.com/dialogs/dialog-go-lib/service/info"

var (
	// Version of the service
	Version = "<todo>"
	// Commit in git in short format
	Commit = "<todo>"
	// GoVersion info on build moment
	GoVersion = "<todo>"
	// BuildDate is date and time in format +%Y-%m-%d_%H:%M:%S
	BuildDate = "<todo>"
)

// New returns service info
func New(name string) *info.Info {
	return &info.Info{
		Name:      name,
		Version:   Version,
		Commit:    Commit,
		GoVersion: GoVersion,
		BuildDate: BuildDate,
	}
}
