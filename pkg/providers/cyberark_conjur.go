package providers

import (
	"fmt"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/spectralops/teller/pkg/core"
)

type ConjurClient interface {
	RetrieveSecret(variableId string) ([]byte, error)
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
	return fmt.Errorf("%v does not implement write yet", c.Name())
}
func (c *CyberArkConjur) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("%v does not implement write yet", c.Name())
}

func (c *CyberArkConjur) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("%v does not implement get mapping yet", c.Name())
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

func (c *CyberArkConjur) getSecret(kp core.KeyPath) ([]byte, error) {
	return c.client.RetrieveSecret(kp.Path)
}
