package providers

import (
	"fmt"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/spectralops/teller/pkg/core"
)

type ResourceFilter struct {
	Kind   string
	Search string
	Limit  int
	Offset int
}

type ConjurClient interface {
	AddSecret(variableID string, secretValue string) error
	RetrieveSecret(variableID string) ([]byte, error)
}

type CyberArkConjur struct {
	client ConjurClient
}

func NewConjurClient() (core.Provider, error) {
	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, err
	}

	conjur, err := conjurapi.NewClientFromKey(config,
		authn.LoginPair{
			Login:  os.Getenv("CONJUR_AUTHN_LOGIN"),
			APIKey: os.Getenv("CONJUR_AUTHN_API_KEY"),
		},
	)
	if err != nil {
		return nil, err
	}

	return &CyberArkConjur{client: conjur}, nil
}

func (c *CyberArkConjur) Name() string {
	return "cyberark_conjur"
}

func (c *CyberArkConjur) Put(p core.KeyPath, val string) error {
	err := c.putSecret(p, val)

	return err
}
func (c *CyberArkConjur) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement put mapping yet", c.Name())
}

func (c *CyberArkConjur) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("provider %q does not implement get mapping yet", c.Name())
}

func (c *CyberArkConjur) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := c.getSecret(p)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(string(secret))
	return &ent, nil
}

func (c *CyberArkConjur) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", c.Name())
}

func (c *CyberArkConjur) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", c.Name())
}

func (c *CyberArkConjur) getSecret(kp core.KeyPath) ([]byte, error) {
	return c.client.RetrieveSecret(kp.Path)
}

func (c *CyberArkConjur) putSecret(kp core.KeyPath, val string) error {
	return c.client.AddSecret(kp.Path, val)
}
