package providers

import (
	"context"
	"os"
	"sort"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/spectralops/teller/pkg/core"
)

type HerokuClient interface {
	ConfigVarInfoForApp(ctx context.Context, appIdentity string) (heroku.ConfigVarInfoForAppResult, error)
	ConfigVarUpdate(ctx context.Context, appIdentity string, o map[string]*string) (heroku.ConfigVarUpdateResult, error)
}
type Heroku struct {
	client HerokuClient
}

func NewHeroku() (core.Provider, error) {
	heroku.DefaultTransport.BearerToken = os.Getenv("HEROKU_API_KEY")

	svc := heroku.NewService(heroku.DefaultClient)
	return &Heroku{client: svc}, nil
}

func (h *Heroku) Name() string {
	return "heroku"
}

func (h *Heroku) Put(p core.KeyPath, val string) error {
	k := p.EffectiveKey()
	_, err := h.client.ConfigVarUpdate(context.TODO(), p.Path, map[string]*string{k: &val})
	return err
}
func (h *Heroku) PutMapping(p core.KeyPath, m map[string]string) error {
	vars := map[string]*string{}
	for k := range m {
		v := m[k]
		vars[k] = &v
	}
	_, err := h.client.ConfigVarUpdate(context.TODO(), p.Path, vars)
	return err
}

// LINTFIX: Extract this commonly somewhere
// nolint: dupl
func (h *Heroku) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
	}

	k := secret

	entries := []core.EnvEntry{}
	for k, v := range k {
		val := ""
		if v != nil {
			val = *v
		}
		entries = append(entries, p.FoundWithKey(k, val))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (h *Heroku) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
	}

	data := secret
	k := data[p.Env]
	if p.Field != "" {
		k = data[p.Field]
	}

	if k == nil {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*k)
	return &ent, nil
}

func (h *Heroku) getSecret(kp core.KeyPath) (heroku.ConfigVarInfoForAppResult, error) {
	return h.client.ConfigVarInfoForApp(context.TODO(), kp.Path)
}
