package providers

// TODO(XXX): remove this provider, no support/specialty

import (
	"fmt"
	"sort"

	"github.com/DopplerHQ/cli/pkg/configuration"
	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type DopplerClient interface {
	GetSecrets(host string, verifyTLS bool, apiKey string, project string, config string) ([]byte, http.Error)
}

type dopplerClient struct{}

func (dopplerClient) GetSecrets(host string, verifyTLS bool, apiKey, project, config string) ([]byte, http.Error) {
	return http.GetSecrets(host, verifyTLS, apiKey, project, config)
}

type Doppler struct {
	client DopplerClient
	logger logging.Logger
	config models.ScopedOptions
}

const DopplerName = "doppler"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description: "Doppler",
		Name:        DopplerName,
		Ops:         core.OpMatrix{Get: true},
	}

	RegisterProvider(metaInfo, NewDoppler)
}

func NewDoppler(logger logging.Logger) (core.Provider, error) {
	configuration.Setup()
	configuration.LoadConfig()

	return &Doppler{
		client: dopplerClient{},
		logger: logger,
		config: configuration.Get(configuration.Scope),
	}, nil
}

func (h *Doppler) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", DopplerName)
}
func (h *Doppler) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", DopplerName)
}

func (h *Doppler) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	s, err := h.getConfig(p.Path)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	for k, v := range s {
		entries = append(entries, p.FoundWithKey(k, v.ComputedValue))
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
		h.logger.WithField("path", p.Path).Debug("`env` attribute not configured. take `field` attribute")
		key = p.Field
	}

	v, ok := s[key]
	if !ok {
		h.logger.WithFields(map[string]interface{}{"key": key, "path": p.Path}).Debug("the given key not exists")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(v.ComputedValue)

	return &ent, nil
}

func (h *Doppler) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", DopplerName)
}

func (h *Doppler) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", DopplerName)
}

func (h *Doppler) getConfig(config string) (map[string]models.ComputedSecret, error) {
	h.logger.Debug("get secrets")
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
