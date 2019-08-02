package converter

import (
	"fmt"

	"github.com/dialogs/dialog-go-lib/enum"
)

const (
	KindApi    Kind = 0 // default
	KindBinary Kind = 1
)

type Kind int

var _KindEnum = enum.New("converter kind").
	Add(KindApi, "api").
	Add(KindBinary, "binary")

func KindStringKeys() []string {
	return _KindEnum.StringKeys()
}

func KindByString(src string) Kind {
	mode, ok := _KindEnum.GetByString(src)
	if !ok {
		return KindApi
	}
	return mode.(Kind)
}

func (k Kind) String() string {
	val, ok := _KindEnum.GetByIndex(k)
	if !ok {
		return fmt.Sprintf("invalid converter kind: %d", k)
	}

	return val
}
