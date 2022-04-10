package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestSecretFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockSecretFileClient(ctrl)
	path := "/run/secrets/foo_secret"
	pathmap := "/run/secrets"
	dirResult := map[string]string{
		"foo_secret":     "somesecretstring",
		"another_secret": "anothersomesecretstring",
	}
	fooResult := map[string]string{"foo_secret": "somesecretstring"}
	client.EXPECT().Read(gomock.Eq(path)).Return(fooResult, nil).AnyTimes()
	client.EXPECT().Read(gomock.Eq(pathmap)).Return(dirResult, nil).AnyTimes()
	s := SecretFile{
		client: client,
		logger: GetTestLogger(),
	}

	p := core.NewPopulate(map[string]string{"stage": "prod"})
	kpmap := p.KeyPath(core.KeyPath{Field: "", Path: "/run/secrets", Decrypt: true})
	fooPath := p.KeyPath(core.KeyPath{Field: "FOO_SECRET", Path: "/run/secrets/foo_secret", Decrypt: true})

	ent, err := s.Get(fooPath)
	assert.Nil(t, err)
	assert.Equal(t, ent.Value, "somesecretstring")

	ents, err := s.GetMapping(kpmap)
	assert.Nil(t, err)
	assert.Equal(t, len(ents), 2)

	assert.Equal(t, ents[0].Key, "foo_secret")
	assert.Equal(t, ents[0].Value, "somesecretstring")

	assert.Equal(t, ents[1].Key, "another_secret")
	assert.Equal(t, ents[1].Value, "anothersomesecretstring")
}

func TestSecretFileFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockSecretFileClient(ctrl)
	client.EXPECT().Read(gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := SecretFile{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "BAR_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
