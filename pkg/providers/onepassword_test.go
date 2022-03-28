package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestOnePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockOnePasswordClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"

	out := onepassword.Item{
		Fields: []*onepassword.ItemField{
			{Label: "MG_KEY", Value: "shazam"},
		},
	}
	outlist := onepassword.Item{
		Fields: []*onepassword.ItemField{
			{Label: "MG_KEY_1", Value: "mailman"},
			{Label: "MG_KEY", Value: "shazam"},
		},
	}
	client.EXPECT().GetItemByTitle(gomock.Eq(path), gomock.Any()).Return(&out, nil).AnyTimes()
	client.EXPECT().GetItemByTitle(gomock.Eq(pathmap), gomock.Any()).Return(&outlist, nil).AnyTimes()
	s := OnePassword{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestOnePasswordFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockOnePasswordClient(ctrl)
	client.EXPECT().GetItemByTitle(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := OnePassword{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
	_, err = s.GetMapping(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
