package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestCloudflareWorkersKV(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockCloudflareClient(ctrl)

	// In CloudflareWorkersKV this isn't the path name, but a Workers KV namespace ID.
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	shazam := []byte("shazam")
	mailman := []byte("mailman")

	listOut := cloudflare.ListStorageKeysResponse{ //nolint
		cloudflare.Response{},
		[]cloudflare.StorageKey{{Name: "MG_KEY"}, {Name: "SMTP_PASS"}},
		cloudflare.ResultInfo{},
	}

	client.EXPECT().ReadWorkersKV(gomock.Any(), path, gomock.Eq("MG_KEY")).Return(shazam, nil).AnyTimes()
	client.EXPECT().ReadWorkersKV(gomock.Any(), pathmap, gomock.Eq("MG_KEY")).Return(shazam, nil).AnyTimes()
	client.EXPECT().ReadWorkersKV(gomock.Any(), pathmap, gomock.Eq("SMTP_PASS")).Return(mailman, nil).AnyTimes()
	client.EXPECT().ListWorkersKVs(gomock.Any(), pathmap).Return(listOut, nil).AnyTimes()

	s := Cloudflare{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestCloudflareReadWorkersKVFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockCloudflareClient(ctrl)
	client.EXPECT().ReadWorkersKV(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := Cloudflare{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	_, missingLookupKeyError := s.Get(core.KeyPath{Field: "", Env: "", Path: "settings/{{stage}}/billing-svc"})

	assert.NotNil(t, err)
	assert.Equal(t, missingLookupKeyError.Error(), "Key required for fetching secrets. Received \"\"")
}

func TestCloudflareListWorkersKVsFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockCloudflareClient(ctrl)
	client.EXPECT().ListWorkersKVs(gomock.Any(), gomock.Any()).Return(cloudflare.ListStorageKeysResponse{}, errors.New("error")).AnyTimes()
	s := Cloudflare{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.GetMapping(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
