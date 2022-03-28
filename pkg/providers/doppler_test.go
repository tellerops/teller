package providers

import (
	"testing"

	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/golang/mock/gomock"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestDoppler(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockDopplerClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	out := []byte(`{
		"secrets": {
			"MG_KEY": {
				"computed": "shazam"
			},
			"SMTP_PASS": {
				"computed": "mailman"
			}
		}
	}`)

	client.EXPECT().GetSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(path)).Return(out, http.Error{}).AnyTimes()
	client.EXPECT().GetSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(pathmap)).Return(out, http.Error{}).AnyTimes()

	s := Doppler{
		client: client,
		config: models.ScopedOptions{},
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}
