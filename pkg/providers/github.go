package providers

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

const (
	gitHubSplitPathCount = 2
)

// GitHubActionClient describe the GitHub action client
type GitHubActionClient interface {
	GetRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, *github.Response, error)
	CreateOrUpdateRepoSecret(ctx context.Context, owner, repo string, eSecret *github.EncryptedSecret) (*github.Response, error)
	DeleteRepoSecret(ctx context.Context, owner, repo, name string) (*github.Response, error)
	ListRepoSecrets(ctx context.Context, owner, repo string, opts *github.ListOptions) (*github.Secrets, *github.Response, error)
}

type GitHub struct {
	clientActions GitHubActionClient
	logger        logging.Logger
}

// NewGitHub create new GitHub provider
const GithubName = "GitHub"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Github",
		Authentication: "Requires `GITHUB_AUTH_TOKEN`",
		Name:           GithubName,
		ConfigTemplate: `
  # Configure via environment variables for integration:
  # GITHUB_AUTH_TOKEN: GitHub token

  github:
    env_sync:
       path: owner/github-repo
    env:
      script-value:
        path: owner/github-repo
`,
		Ops: core.OpMatrix{Put: true, PutMapping: true, Delete: true, DeleteMapping: true},
	}

	RegisterProvider(metaInfo, NewGitHub)
}

func NewGitHub(logger logging.Logger) (core.Provider, error) {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		return nil, errors.New("missing `GITHUB_AUTH_TOKEN`")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)
	client := github.NewClient(tc)

	return &GitHub{clientActions: client.Actions, logger: logger}, nil
}

func (g *GitHub) Put(p core.KeyPath, val string) error {

	owner, repoName, err := g.parsePathToOwnerAndRepo(p)
	if err != nil {
		return err
	}

	publicKey, _, err := g.getRepoPublicKey(owner, repoName)
	if err != nil {
		return err
	}

	encryptedSecret, err := g.encryptSecretWithPublicKey(publicKey, p.Env, val)
	if err != nil {
		return err
	}

	_, err = g.createOrUpdateRepoSecret(context.TODO(), owner, repoName, encryptedSecret)

	return err
}

func (g *GitHub) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		ap := p.WithEnv(k)
		err := g.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GitHub) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("does not supported by the %s provider", GithubName)
}

func (g *GitHub) Get(p core.KeyPath) (*core.EnvEntry, error) {
	return nil, fmt.Errorf("does not supported by the %s provider", GithubName)
}

func (g *GitHub) Delete(p core.KeyPath) error {
	owner, repoName, err := g.parsePathToOwnerAndRepo(p)
	if err != nil {
		return err
	}

	_, err = g.deleteRepoSecret(context.TODO(), owner, repoName, p.Env)
	return err
}

func (g *GitHub) DeleteMapping(p core.KeyPath) error {

	owner, repoName, err := g.parsePathToOwnerAndRepo(p)
	if err != nil {
		return err
	}

	opt := github.ListOptions{PerPage: 100}
	g.logger.WithFields(map[string]interface{}{
		"owner":           owner,
		"repository_name": repoName,
	}).Debug("get repo secrets")
	secrets, _, err := g.clientActions.ListRepoSecrets(context.TODO(), owner, repoName, &opt)
	if err != nil {
		return err
	}

	for _, secret := range secrets.Secrets {
		err := g.Delete(p.WithEnv(secret.Name))
		if err != nil {
			return err
		}

	}

	return nil
}

// parsePathToOwnerAndRepo parse the key path to the owner and repo name
func (g *GitHub) parsePathToOwnerAndRepo(p core.KeyPath) (string, string, error) { //nolint

	splitData := strings.SplitN(p.Path, "/", 2) //nolint
	if len(splitData) != gitHubSplitPathCount {
		return "", "", fmt.Errorf("invalid %s path, expected owner/repo got: %s", GithubName, p.Path)
	}
	return splitData[0], splitData[1], nil
}

// GetRepoPublicKey gets a public key that should be used for secret encryption.
func (g *GitHub) getRepoPublicKey(owner, repo string) (*github.PublicKey, *github.Response, error) {
	return g.clientActions.GetRepoPublicKey(context.TODO(), owner, repo)

}

// CreateOrUpdateRepoSecret creates or updates a repository secret with an encrypted value.
func (g *GitHub) createOrUpdateRepoSecret(ctx context.Context, owner, repo string, eSecret *github.EncryptedSecret) (*github.Response, error) {
	g.logger.WithFields(map[string]interface{}{
		"owner":           owner,
		"repository_name": repo,
		"name":            eSecret.Name,
	}).Debug("put repo secret")
	return g.clientActions.CreateOrUpdateRepoSecret(ctx, owner, repo, eSecret)
}

// DeleteRepoSecret deletes a secret in a repository using the secret name.
func (g *GitHub) deleteRepoSecret(ctx context.Context, owner, repo, name string) (*github.Response, error) {
	g.logger.WithFields(map[string]interface{}{
		"owner":           owner,
		"repository_name": repo,
		"name":            name,
	}).Debug("delete repo secret")
	return g.clientActions.DeleteRepoSecret(ctx, owner, repo, name)
}

// encryptSecretWithPublicKey secret secret name and value by the given GitHub public key
func (g *GitHub) encryptSecretWithPublicKey(publicKey *github.PublicKey, secretName, secretValue string) (*github.EncryptedSecret, error) {

	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return nil, fmt.Errorf("base64.StdEncoding.DecodeString was unable to decode public key: %v", err)
	}

	var boxKey [32]byte
	copy(boxKey[:], decodedPublicKey)
	secretBytes := []byte(secretValue)
	encryptedBytes, err := box.SealAnonymous([]byte{}, secretBytes, &boxKey, crypto_rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("box.SealAnonymous failed with error %w", err)
	}

	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)

	keyID := publicKey.GetKeyID()
	encryptedSecret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedString,
	}
	return encryptedSecret, nil
}
