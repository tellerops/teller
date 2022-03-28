package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/hashicorp/vault/api"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestHashicorpVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockHashicorpClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"

	tests := map[string]struct {
		out api.Secret
	}{
		"data.data": {
			out: api.Secret{
				Data: map[string]interface{}{
					"data": map[string]interface{}{
						"MG_KEY":    "shazam",
						"SMTP_PASS": "mailman",
					},
				},
			},
		},
		"data": {
			out: api.Secret{
				Data: map[string]interface{}{
					"MG_KEY":    "shazam",
					"SMTP_PASS": "mailman",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client.EXPECT().Read(gomock.Eq(path)).Return(&tc.out, nil).AnyTimes()
			client.EXPECT().Read(gomock.Eq(pathmap)).Return(&tc.out, nil).AnyTimes()
			s := HashicorpVault{
				client: client,
				logger: GetTestLogger(),
			}
			AssertProvider(t, &s, true)
		})
	}
}

func TestHashicorpVaultFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockHashicorpClient(ctrl)
	client.EXPECT().Read(gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := HashicorpVault{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
