package conjurapi

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/cyberark/conjur-api-go/conjurapi/logging"
)

var supportedAuthnTypes = []string{"authn", "ldap", "oidc"}

type Config struct {
	Account           string `yaml:"account,omitempty"`
	ApplianceURL      string `yaml:"appliance_url,omitempty"`
	NetRCPath         string `yaml:"netrc_path,omitempty"`
	SSLCert           string `yaml:"-"`
	SSLCertPath       string `yaml:"cert_file,omitempty"`
	AuthnType         string `yaml:"authn_type,omitempty"`
	ServiceID         string `yaml:"service_id,omitempty"`
	CredentialStorage string `yaml:"credential_storage,omitempty"`
}

func (c *Config) IsHttps() bool {
	return c.SSLCertPath != "" || c.SSLCert != ""
}

func (c *Config) Validate() error {
	errors := []string{}

	if c.ApplianceURL == "" {
		errors = append(errors, "Must specify an ApplianceURL")
	}

	if c.Account == "" {
		errors = append(errors, "Must specify an Account")
	}

	if c.AuthnType != "" && !contains(supportedAuthnTypes, c.AuthnType) {
		errors = append(errors, fmt.Sprintf("AuthnType must be one of %v", supportedAuthnTypes))
	}

	if (c.AuthnType == "ldap" || c.AuthnType == "oidc") && c.ServiceID == "" {
		errors = append(errors, fmt.Sprintf("Must specify a ServiceID when using %s", c.AuthnType))
	}

	if len(errors) == 0 {
		return nil
	} else if logging.ApiLog.Level == logrus.DebugLevel {
		errors = append(errors, fmt.Sprintf("config: %+v", c))
	}
	return fmt.Errorf("%s", strings.Join(errors, " -- "))
}

func (c *Config) ReadSSLCert() ([]byte, error) {
	if c.SSLCert != "" {
		return []byte(c.SSLCert), nil
	}
	return os.ReadFile(c.SSLCertPath)
}

func (c *Config) BaseURL() string {
	prefix := ""
	if !strings.HasPrefix(c.ApplianceURL, "http") {
		if c.IsHttps() {
			prefix = "https://"
		} else {
			prefix = "http://"
		}
	}
	return prefix + c.ApplianceURL
}

func mergeValue(a, b string) string {
	if len(b) != 0 {
		return b
	}
	return a
}

func (c *Config) merge(o *Config) {
	c.ApplianceURL = mergeValue(c.ApplianceURL, o.ApplianceURL)
	c.Account = mergeValue(c.Account, o.Account)
	c.SSLCert = mergeValue(c.SSLCert, o.SSLCert)
	c.SSLCertPath = mergeValue(c.SSLCertPath, o.SSLCertPath)
	c.NetRCPath = mergeValue(c.NetRCPath, o.NetRCPath)
	c.CredentialStorage = mergeValue(c.CredentialStorage, o.CredentialStorage)
	c.AuthnType = mergeValue(c.AuthnType, o.AuthnType)
	c.ServiceID = mergeValue(c.ServiceID, o.ServiceID)
}

func (c *Config) mergeYAML(filename string) error {
	// Read the YAML file
	buf, err := os.ReadFile(filename)

	if err != nil {
		logging.ApiLog.Debugf("Failed reading %s, %v\n", filename, err)
		// It is not an error if this file does not exist
		return nil
	}

	// Parse the YAML file into a new struct containing the same
	// fields as Config, plus a few extra fields for compatibility
	aux := struct {
		ConjurVersion string `yaml:"version"`
		Config        `yaml:",inline"`
		// BEGIN COMPATIBILITY WITH PYTHON CLI
		ConjurURL     string `yaml:"conjur_url"`
		ConjurAccount string `yaml:"conjur_account"`
		// END COMPATIBILITY WITH PYTHON CLI
	}{}

	if err := yaml.Unmarshal(buf, &aux); err != nil {
		logging.ApiLog.Errorf("Parsing error %s: %s\n", filename, err)
		return err
	}

	// Now merge the parsed config into the current config object
	logging.ApiLog.Debugf("Config from %s: %+v\n", filename, aux.Config)
	c.merge(&aux.Config)

	// BEGIN COMPATIBILITY WITH PYTHON CLI
	// The Python CLI uses the keys conjur_url and conjur_account
	// instead of appliance_url and account. Check if those keys
	// are present and use them if the new keys are not present.
	if c.ApplianceURL == "" && aux.ConjurURL != "" {
		c.ApplianceURL = aux.ConjurURL
	}

	if c.Account == "" && aux.ConjurAccount != "" {
		c.Account = aux.ConjurAccount
	}
	// END COMPATIBILITY WITH PYTHON CLI

	return nil
}

func (c *Config) mergeEnv() {
	env := Config{
		ApplianceURL:      os.Getenv("CONJUR_APPLIANCE_URL"),
		SSLCert:           os.Getenv("CONJUR_SSL_CERTIFICATE"),
		SSLCertPath:       os.Getenv("CONJUR_CERT_FILE"),
		Account:           os.Getenv("CONJUR_ACCOUNT"),
		NetRCPath:         os.Getenv("CONJUR_NETRC_PATH"),
		CredentialStorage: os.Getenv("CONJUR_CREDENTIAL_STORAGE"),
		AuthnType:         os.Getenv("CONJUR_AUTHN_TYPE"),
		ServiceID:         os.Getenv("CONJUR_SERVICE_ID"),
	}

	logging.ApiLog.Debugf("Config from environment: %+v\n", env)
	c.merge(&env)
}

func (c *Config) Conjurrc() []byte {
	data, _ := yaml.Marshal(&c)
	return data
}

func LoadConfig() (Config, error) {
	config := Config{}

	home, err := os.UserHomeDir()
	if err != nil {
		logging.ApiLog.Warningf("Could not detect homedir.")
	}

	// Default to using ~/.netrc, subsequent configuration can
	// override it if the home dir is set.
	if home != "" {
		config = Config{NetRCPath: path.Join(home, ".netrc")}
	}

	err = config.mergeYAML(path.Join(getSystemPath(), "conjur.conf"))
	if err != nil {
		return config, err
	}

	conjurrc := os.Getenv("CONJURRC")
	if conjurrc == "" && home != "" {
		conjurrc = path.Join(home, ".conjurrc")
	}
	if conjurrc != "" {
		config.mergeYAML(conjurrc)
	}

	config.mergeEnv()

	logging.ApiLog.Debugf("Final config: %+v\n", config)
	return config, nil
}

func getSystemPath() string {
	if runtime.GOOS == "windows" {
		//No way to use SHGetKnownFolderPath()
		//Hardcoding should be fine for now since CONJURRC is available
		return "C:\\windows"
	} else {
		return "/etc"
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
