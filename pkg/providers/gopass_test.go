package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"

	// "github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestGopass(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGopassClient(ctrl)
	path := "settings/prod/billing-svc"

	secretShazamVal := `shazam
settings / prod / billing-svc
`
	secretMailmanVal := `mailman
settings / prod / billing-svc
`
	secretShazam, _ := secparse.Parse([]byte(secretShazamVal))
	secretMailman, _ := secparse.Parse([]byte(secretMailmanVal))
	outlist := []string{
		"settings/prod/billing-svc/all/1",
		"settings/prod/billing-svc/all/2",
	}

	client.EXPECT().Get(context.TODO(), gomock.Eq(path), gomock.Any()).Return(secretShazam, nil).AnyTimes()
	client.EXPECT().Get(context.TODO(), gomock.Eq("settings/prod/billing-svc/all/1"), gomock.Any()).Return(secretShazam, nil).AnyTimes()
	client.EXPECT().Get(context.TODO(), gomock.Eq("settings/prod/billing-svc/all/2"), gomock.Any()).Return(secretMailman, nil).AnyTimes()
	client.EXPECT().List(context.TODO()).Return(outlist, nil).AnyTimes()
	s := Gopass{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestGopassFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGopassClient(ctrl)
	client.EXPECT().Get(context.TODO(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := Gopass{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	client.EXPECT().List(context.TODO()).Return([]string{"a"}, errors.New("error")).AnyTimes()
	assert.NotNil(t, err)
	_, err = s.GetMapping(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
