package providers

import (
	"fmt"
	"sort"

	"github.com/DopplerHQ/cli/pkg/configuration"
	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/spectralops/teller/pkg/core"
)

type DopplerClient interface {
	GetSecrets(host string, verifyTLS bool, apiKey string, project string, config string) ([]byte, http.Error)
}

type dopplerClient struct{}

func (dopplerClient) GetSecrets(host string, verifyTLS bool, apiKey string, project string, config string) ([]byte, http.Error) {
	return http.GetSecrets(host, verifyTLS, apiKey, project, config)
}

type Doppler struct {
	client DopplerClient
	config models.ScopedOptions
}

func NewDoppler() (core.Provider, error) {
	configuration.Setup()
	configuration.LoadConfig()

	return &Doppler{
		client: dopplerClient{},
		config: configuration.Get(configuration.Scope),
	}, nil
}

func (h *Doppler) Name() string {
	return "doppler"
}

func (h *Doppler) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	s, err := h.getConfig(p.Path)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	for k, v := range s {
		entries = append(entries, core.EnvEntry{
			Key:          k,
			Value:        v.ComputedValue,
			Provider:     h.Name(),
			ResolvedPath: p.Path,
		})
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (h *Doppler) Get(p core.KeyPath) (*core.EnvEntry, error) {
	s, err := h.getConfig(p.Path)
	if err != nil {
		return nil, err
	}

	key := p.Env
	if p.Field != "" {
		key = p.Field
	}

	v, ok := s[key]
	if !ok {
		return nil, fmt.Errorf("field at '%s' does not exist", p.Path)
	}

	return &core.EnvEntry{
		Key:          p.Env,
		Value:        v.ComputedValue,
		ResolvedPath: p.Path,
		Provider:     h.Name(),
	}, nil
}

func (h *Doppler) getConfig(config string) (map[string]models.ComputedSecret, error) {
	r, herr := h.client.GetSecrets(
		h.config.APIHost.Value,
		utils.GetBool(h.config.VerifyTLS.Value, true),
		h.config.Token.Value,
		h.config.EnclaveProject.Value,
		config,
	)
	if !herr.IsNil() {
		return nil, herr.Err
	}

	return models.ParseSecrets(r)
}
