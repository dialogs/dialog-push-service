package conversion

import (
	"testing"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/stretchr/testify/require"
)

func TestIgnoreRequest(t *testing.T) {

	strPtr := func(src string) *string { return &src }

	for _, testInfo := range []struct {
		Src       *api.PushBody
		ApnIgnore bool
		FcmIgnore bool
		GcmIgnore bool
	}{
		{
			Src: &api.PushBody{
				Body: &api.PushBody_EncryptedPush{
					EncryptedPush: &api.EncryptedPush{EncryptedData: []byte("push body")},
				},
			},
		},
		{
			Src: &api.PushBody{
				Body: &api.PushBody_VoipPush{
					VoipPush: &api.VoipPush{CallId: 123},
				},
			},
		},
		{
			Src: &api.PushBody{
				Body: &api.PushBody_AlertingPush{
					AlertingPush: &api.AlertingPush{},
				},
			},
		},
		{
			Src: &api.PushBody{
				Body: &api.PushBody_SilentPush{
					SilentPush: &api.SilentPush{},
				},
			},
			ApnIgnore: true,
		},
	} {

		var (
			res provider.IRequest
			err error
		)

		{
			res, err = RequestPbToAns(testInfo.Src, true, true, strPtr("topic-name"), strPtr("sound-name"))
			require.NoError(t, err)
			require.Equal(t, testInfo.ApnIgnore, res.Ignore())
		}

		{
			res, err = RequestPbToFcm(testInfo.Src, true)
			require.NoError(t, err)
			require.Equal(t, testInfo.FcmIgnore, res.Ignore())
		}

		{
			res, err = RequestPbToGcm(testInfo.Src, true)
			require.NoError(t, err)
			require.Equal(t, testInfo.GcmIgnore, res.Ignore())
		}
	}

}
