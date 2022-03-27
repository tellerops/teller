package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v43/github"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestGitHubPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGitHubActionClient(ctrl)

	keyID := "1234"
	key := "2Sg8iYjAxxmI2LvUXpJjkYrMxURPc8r+dB7TJyvv1234"
	publicKey := github.PublicKey{
		KeyID: &keyID,
		Key:   &key,
	}

	client.EXPECT().GetRepoPublicKey(gomock.Any(), "owner-name", "error").Return(&publicKey, nil, errors.New("some error")).AnyTimes()
	client.EXPECT().GetRepoPublicKey(gomock.Any(), "owner-name", "repo-name").Return(&publicKey, nil, nil).AnyTimes()
	client.EXPECT().CreateOrUpdateRepoSecret(gomock.Any(), "owner-name", "repo-name", gomock.Any()).Return(nil, nil).AnyTimes()

	client.EXPECT().GetRepoPublicKey(gomock.Any(), "owner-name", "create-error").Return(&publicKey, nil, nil).AnyTimes()
	client.EXPECT().CreateOrUpdateRepoSecret(gomock.Any(), "owner-name", "create-error", gomock.Any()).Return(nil, errors.New("some error")).AnyTimes()

	c := GitHub{
		clientActions: client,
		logger:        GetTestLogger(),
	}

	assert.NotNil(t, c.Put(core.KeyPath{Path: "owner-name", Field: "MG_KEY"}, "put-secret"), "owner or repo name should be invalid")
	assert.NotNil(t, c.Put(core.KeyPath{Path: "owner-name/error", Field: "MG_KEY"}, "put-secret"), "repo public key should return an error")
	assert.Nil(t, c.Put(core.KeyPath{Path: "owner-name/repo-name", Field: "MG_KEY"}, "put-secret"), "can put a new secret")
	assert.NotNil(t, c.Put(core.KeyPath{Path: "owner-name/create-error", Field: "MG_KEY"}, "put-secret"), "create or update secret should be fails")

}

func TestGitHubPutMapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGitHubActionClient(ctrl)

	keyID := "1234"
	key := "2Sg8iYjAxxmI2LvUXpJjkYrMxURPc8r+dB7TJyvv1234"
	publicKey := github.PublicKey{
		KeyID: &keyID,
		Key:   &key,
	}

	client.EXPECT().GetRepoPublicKey(gomock.Any(), "owner-name", "error").Return(&publicKey, nil, errors.New("some error")).AnyTimes()
	client.EXPECT().GetRepoPublicKey(gomock.Any(), "owner-name", "repo-name").Return(&publicKey, nil, nil).AnyTimes()
	client.EXPECT().CreateOrUpdateRepoSecret(gomock.Any(), "owner-name", "repo-name", gomock.Any()).Return(nil, nil).AnyTimes()

	client.EXPECT().GetRepoPublicKey(gomock.Any(), "owner-name", "create-error").Return(&publicKey, nil, nil).AnyTimes()
	client.EXPECT().CreateOrUpdateRepoSecret(gomock.Any(), "owner-name", "create-error", gomock.Any()).Return(nil, errors.New("some error")).AnyTimes()

	c := GitHub{
		clientActions: client,
		logger:        GetTestLogger(),
	}

	data := map[string]string{
		"key-1": "value-1",
	}
	assert.NotNil(t, c.PutMapping(core.KeyPath{Path: "owner-name", Field: "MG_KEY"}, data), "owner or repo name should be invalid")
	assert.NotNil(t, c.PutMapping(core.KeyPath{Path: "owner-name/error", Field: "MG_KEY"}, data), "repo public key should return an error")
	assert.Nil(t, c.PutMapping(core.KeyPath{Path: "owner-name/repo-name", Field: "MG_KEY"}, data), "can put a new secret")
	assert.NotNil(t, c.PutMapping(core.KeyPath{Path: "owner-name/create-error", Field: "MG_KEY"}, data), "create or update secret should be fails")

}

func TestGitHubDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGitHubActionClient(ctrl)

	client.EXPECT().DeleteRepoSecret(gomock.Any(), "owner-name", "repo-name", "MG_KEY").Return(nil, nil).AnyTimes()
	client.EXPECT().DeleteRepoSecret(gomock.Any(), "owner-name", "repo-name", "MG_KEY_ERROR").Return(nil, errors.New("some error")).AnyTimes()

	c := GitHub{
		clientActions: client,
		logger:        GetTestLogger(),
	}

	assert.NotNil(t, c.Delete(core.KeyPath{Path: "owner-name", Field: "MG_KEY"}), "owner or repo name should be invalid")
	assert.Nil(t, c.Delete(core.KeyPath{Path: "owner-name/repo-name", Env: "MG_KEY"}), "delete action should pass")
	assert.NotNil(t, c.Delete(core.KeyPath{Path: "owner-name/repo-name", Env: "MG_KEY_ERROR"}), "delete action should return an error")

}

func TestGitHubDeleteMaping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGitHubActionClient(ctrl)

	secretsResponse := github.Secrets{
		TotalCount: 2,
		Secrets: []*github.Secret{
			{Name: "MG_KEY-1"},
			{Name: "MG_KEY-2"},
		},
	}
	client.EXPECT().ListRepoSecrets(gomock.Any(), "owner-name", "repo-name", gomock.Any()).Return(&secretsResponse, nil, nil).AnyTimes()
	client.EXPECT().DeleteRepoSecret(gomock.Any(), "owner-name", "repo-name", "MG_KEY-1").Return(nil, nil).AnyTimes()
	client.EXPECT().DeleteRepoSecret(gomock.Any(), "owner-name", "repo-name", "MG_KEY-2").Return(nil, nil).AnyTimes()

	c := GitHub{
		clientActions: client,
		logger:        GetTestLogger(),
	}

	assert.Nil(t, c.DeleteMapping(core.KeyPath{Path: "owner-name/repo-name", Env: "MG_KEY"}), "delete action should pass")

}

func TestGitHubDeleteMapingWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGitHubActionClient(ctrl)

	secretsResponse := github.Secrets{
		TotalCount: 2,
		Secrets: []*github.Secret{
			{Name: "MG_KEY-1"},
			{Name: "MG_KEY-2"},
		},
	}
	client.EXPECT().ListRepoSecrets(gomock.Any(), "owner-name", "repo-name", gomock.Any()).Return(&secretsResponse, nil, nil).AnyTimes()
	client.EXPECT().DeleteRepoSecret(gomock.Any(), "owner-name", "repo-name", "MG_KEY-1").Return(nil, nil).AnyTimes()
	client.EXPECT().DeleteRepoSecret(gomock.Any(), "owner-name", "repo-name", "MG_KEY-2").Return(nil, errors.New("some error")).AnyTimes()

	c := GitHub{
		clientActions: client,
		logger:        GetTestLogger(),
	}

	assert.NotNil(t, c.DeleteMapping(core.KeyPath{Path: "owner-name/repo-name", Env: "MG_KEY"}), "delete action should pass")

}

func TestParsePathToOwnerAndRepo(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockGitHubActionClient(ctrl)

	c := GitHub{
		clientActions: client,
		logger:        GetTestLogger(),
	}

	owner, repo, err := c.parsePathToOwnerAndRepo(core.KeyPath{Path: "owner-name/repo-name", Env: "MG_KEY"})
	assert.Equal(t, owner, "owner-name", "unexpected owner name from path key")
	assert.Equal(t, repo, "repo-name", "unexpected repo name from path key")
	assert.Nil(t, err)

	_, _, err = c.parsePathToOwnerAndRepo(core.KeyPath{Path: "owner-name", Env: "MG_KEY"})
	assert.NotNil(t, err)
}
