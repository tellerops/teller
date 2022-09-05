package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/utils"
)

type GopassClient interface {
	List(ctx context.Context) ([]string, error)
	Get(ctx context.Context, name, revision string) (gopass.Secret, error)
	Set(ctx context.Context, name string, sec gopass.Byter) error
}

type Gopass struct {
	client GopassClient
	logger logging.Logger
}

const GoPassName = "gopass"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Gopass",
		Name:           GoPassName,
		Authentication: "Configuration is environment based, as defined by client standard. See variables [here](https://github.com/gopasspw/gopass/blob/master/docs/config.md).",
		ConfigTemplate: `
  # Override default configuration: https://github.com/gopasspw/gopass/blob/master/docs/config.md
  gopass:
    env_sync:
      path: foo
    env:
      ETC_DSN:
        path: foo/bar
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
	}

	RegisterProvider(metaInfo, NewGopass)
}
func NewGopass(logger logging.Logger) (core.Provider, error) {
	ctx := context.Background()
	gp, err := api.New(ctx)
	if err != nil {
		return nil, err
	}
	return &Gopass{client: gp, logger: logger}, nil
}

func (g *Gopass) Put(p core.KeyPath, val string) error {
	secret, err := g.getSecret(p.Path)
	if err != nil {
		return fmt.Errorf("%v cannot get value: %v", GoPassName, err)
	}

	secret.SetPassword(val)
	g.logger.WithField("path", p.Path).Debug("set secret")
	return g.client.Set(context.TODO(), p.Path, secret)
}

func (g *Gopass) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		ap := p.SwitchPath(fmt.Sprintf("%v/%v", p.Path, k))
		secret, err := g.getSecret(ap.Path)
		if err != nil {
			return fmt.Errorf("%v cannot get value: %v", GoPassName, err)
		}

		secret.SetPassword(v)
		g.logger.WithField("path", ap.Path).Debug("set secret")
		err = g.client.Set(context.TODO(), ap.Path, secret)
		if err != nil {
			return fmt.Errorf("%v cannot update value: %v", GoPassName, err)
		}

	}
	return nil
}

func (g *Gopass) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	g.logger.Debug("get all secrets")
	secretsPath, err := g.client.List(context.TODO())
	if err != nil {
		return nil, err
	}
	entries := []core.EnvEntry{}
	for _, secretPath := range secretsPath {
		if strings.HasPrefix(secretPath, p.Path) {
			secret, err := g.getSecret(secretPath)
			if err != nil {
				return nil, err
			}
			seg := utils.LastSegment(secretPath)
			entries = append(entries, p.FoundWithKey(seg, secret.Password()))
		}
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (g *Gopass) Get(p core.KeyPath) (*core.EnvEntry, error) {

	secret, err := g.getSecret(p.Path)
	if err != nil {
		return nil, fmt.Errorf("%v cannot get value: %v", GoPassName, err)
	}

	if secret == nil {
		g.logger.WithField("path", p.Path).Debug("secret is empty")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(secret.Password())
	return &ent, nil
}

func (g *Gopass) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", GoPassName)
}

func (g *Gopass) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", GoPassName)
}

func (g *Gopass) getSecret(path string) (gopass.Secret, error) {
	g.logger.WithField("path", path).Debug("get secret")
	secret, err := g.client.Get(context.TODO(), path, "")
	if err != nil {
		return nil, err
	}
	return secret, nil
}
