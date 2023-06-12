package providers

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestAnsibleVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockAnsibleVaultClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	out := map[string]string{
		"MG_KEY":    "shazam",
		"SMTP_PASS": "mailman",
	}
	client.EXPECT().Read(gomock.Eq(path)).Return(out, nil).AnyTimes()
	client.EXPECT().Read(gomock.Eq(pathmap)).Return(out, nil).AnyTimes()

	s := AnsibleVault{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}
