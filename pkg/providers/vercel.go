package providers

import (
	"fmt"
	"os"
	"sort"

	"github.com/dghubble/sling"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type VercelClient interface {
	GetProject(path string) (map[string]*string, error)
}
type VercelAPI struct {
	http *sling.Sling
}

func NewVercelAPI(token string) *VercelAPI {
	bearer := "Bearer " + token
	httpClient := sling.New().Base(VercelAPIBase).Set("Authorization", bearer)
	return &VercelAPI{http: httpClient}
}

func (v *VercelAPI) GetProject(path string) (map[string]*string, error) {
	projectsPath := "/v1" + ProjectEndPoint + "/" + path
	project := new(VercelProject)
	_, err := v.http.Get(projectsPath).ReceiveSuccess(project)
	return project.envMap(), err
}

type Vercel struct {
	client VercelClient
	logger logging.Logger
}

type VercelProject struct {
	Env []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"env"`
}

func (vp *VercelProject) envMap() map[string]*string {
	val := make(map[string]*string)
	for i := 0; i < len(vp.Env); i++ {
		// pick only plain type variables (ignore secrets)
		cur := vp.Env[i]
		if cur.Type == "plain" {
			val[cur.Key] = &cur.Value
		}
	}
	return val
}

/*
https://vercel.com/docs/api#endpoints/secrets
*/

const VercelAPIBase = "https://api.vercel.com/"

const ProjectEndPoint = "/projects"

const VercelName = "vercel"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Name:           "vercel",
		Description:    "Vercel",
		Authentication: "Requires an API key populated in your environment in: `VERCEL_TOKEN`.",
		ConfigTemplate: `
  # requires an API token in: VERCEL_TOKEN 
  vercel:
	# sync a complete environment
    env_sync:
      path: drakula-demo

	# # pick and choose variables
	# env:
	#	  JVM_OPTS:
	#      path: drakula-demo
`,
		Ops: core.OpMatrix{
			Get:        true,
			GetMapping: true,
		},
	}

	RegisterProvider(metaInfo, NewVercel)
}

func NewVercel(logger logging.Logger) (core.Provider, error) {
	vercelToken := os.Getenv("VERCEL_TOKEN")
	if vercelToken == "" {
		return nil, fmt.Errorf("please set VERCEL_TOKEN")
	}
	return &Vercel{client: NewVercelAPI(vercelToken), logger: logger}, nil
}

func (ve *Vercel) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	secret, err := ve.getSecret(p)
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

func (ve *Vercel) Get(p core.KeyPath) (*core.EnvEntry, error) { //nolint:dupl
	secret, err := ve.getSecret(p)
	if err != nil {
		return nil, err
	}

	data := secret
	k := data[p.Env]
	if p.Field != "" {
		ve.logger.WithField("path", p.Path).Debug("`env` attribute not found in returned data. take `field` attribute")
		k = data[p.Field]
	}

	if k == nil {
		ve.logger.WithField("path", p.Path).Debug("requested entry not found")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*k)
	return &ent, nil
}

func (ve *Vercel) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", VercelName)
}
func (ve *Vercel) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", VercelName)
}

func (ve *Vercel) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", VercelName)
}

func (ve *Vercel) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", VercelName)
}

func (ve *Vercel) getSecret(kp core.KeyPath) (map[string]*string, error) {
	/* https://vercel.com/docs/api#endpoints/projects/get-a-single-project */
	ve.logger.WithField("path", kp.Path).Debug("get secret")
	project, err := ve.client.GetProject(kp.Path)
	if err != nil {
		return nil, err
	}

	return project, nil
}
