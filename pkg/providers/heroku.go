package providers

import (
	"context"
	"fmt"
	"os"
	"sort"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/spectralops/teller/pkg/core"
)

type HerokuClient interface {
	ConfigVarInfoForApp(ctx context.Context, appIdentity string) (heroku.ConfigVarInfoForAppResult, error)
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
	return fmt.Errorf("%v does not implement write yet", h.Name())
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
		entries = append(entries, core.EnvEntry{Key: k, Value: val, Provider: h.Name(), ResolvedPath: p.Path})
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
		return nil, fmt.Errorf("field at '%s' does not exist", p.Path)
	}

	return &core.EnvEntry{
		Key:          p.Env,
		Value:        *k,
		ResolvedPath: p.Path,
		Provider:     h.Name(),
	}, nil
}

func (h *Heroku) getSecret(kp core.KeyPath) (heroku.ConfigVarInfoForAppResult, error) {
	return h.client.ConfigVarInfoForApp(context.TODO(), kp.Path)
}
