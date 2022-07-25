package providers

import (
	"fmt"
	"os"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
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
	logger logging.Logger
}

const ConjurName = "cyberark_conjur"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "CyberArk Conjure",
		Name:           ConjurName,
		Authentication: "Requires a username and API key populated in your environment:\n* `CONJUR_AUTHN_LOGIN`\n* `CONJUR_AUTHN_API_KEY`",
		ConfigTemplate: `
  # https://conjur.org
  # set CONJUR_AUTHN_LOGIN and CONJUR_AUTHN_API_KEY env vars
  # set .conjurrc file in user's home directory
  cyberark_conjur:
    env:
      FOO_BAR:
        path: /secrets/foo/bar
`,
		Ops: core.OpMatrix{Get: true, Put: true},
	}
	RegisterProvider(metaInfo, NewConjurClient)
}

func NewConjurClient(logger logging.Logger) (core.Provider, error) {
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

	return &CyberArkConjur{client: conjur, logger: logger}, nil
}

func (c *CyberArkConjur) Put(p core.KeyPath, val string) error {
	err := c.putSecret(p, val)

	return err
}
func (c *CyberArkConjur) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement put mapping yet", ConjurName)
}

func (c *CyberArkConjur) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("provider %q does not implement get mapping yet", ConjurName)
}

func (c *CyberArkConjur) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := c.getSecret(p)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		c.logger.WithField("path", p.Path).Debug("secret is empty")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(string(secret))
	return &ent, nil
}

func (c *CyberArkConjur) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", ConjurName)
}

func (c *CyberArkConjur) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", ConjurName)
}

func (c *CyberArkConjur) getSecret(kp core.KeyPath) ([]byte, error) {
	c.logger.WithField("path", kp.Path).Debug("get a secret from the path")
	return c.client.RetrieveSecret(kp.Path)
}

func (c *CyberArkConjur) putSecret(kp core.KeyPath, val string) error {
	c.logger.WithField("path", kp.Path).Debug("create secret")
	return c.client.AddSecret(kp.Path, val)
}
