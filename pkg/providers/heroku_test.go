package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestHeroku(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockHerokuClient(ctrl)
	// in heroku this isn't the path name, but an app name,
	// but for testing it doesn't matter
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	shazam := "shazam"
	mailman := "mailman"
	out := map[string]*string{
		"MG_KEY":    &shazam,
		"SMTP_PASS": &mailman,
	}
	client.EXPECT().ConfigVarInfoForApp(gomock.Any(), gomock.Eq(path)).Return(out, nil).AnyTimes()
	client.EXPECT().ConfigVarInfoForApp(gomock.Any(), gomock.Eq(pathmap)).Return(out, nil).AnyTimes()
	s := Heroku{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestHerokuFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockHerokuClient(ctrl)
	client.EXPECT().ConfigVarInfoForApp(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := Heroku{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
