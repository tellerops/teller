package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestDotenv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockDotEnvClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	out := map[string]string{
		"MG_KEY":    "shazam",
		"SMTP_PASS": "mailman",
	}
	client.EXPECT().Read(gomock.Eq(path)).Return(out, nil).AnyTimes()
	client.EXPECT().Read(gomock.Eq(pathmap)).Return(out, nil).AnyTimes()
	client.EXPECT().Read(gomock.Eq(pathmap)).Return(out, nil).AnyTimes()
	s := Dotenv{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestDotenvFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockDotEnvClient(ctrl)
	client.EXPECT().Read(gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := Dotenv{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
