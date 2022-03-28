package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestConsul(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockConsulClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	out := api.KVPair{
		Key:   "MG_KEY",
		Value: []byte("shazam"),
	}
	outlist := api.KVPairs{
		{
			Key:   "SMTP_PASS",
			Value: []byte("mailman"),
		},
		{
			Key:   "MG_KEY",
			Value: []byte("shazam"),
		},
	}
	client.EXPECT().Get(gomock.Eq(path), gomock.Any()).Return(&out, nil, nil).AnyTimes()
	client.EXPECT().List(gomock.Eq(pathmap), gomock.Any()).Return(outlist, nil, nil).AnyTimes()
	s := Consul{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestConsulFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockConsulClient(ctrl)
	client.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error")).AnyTimes()
	s := Consul{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
