package core

import (
	"fmt"
	"strings"

	klog "github.com/keeper-security/secrets-manager-go/core/logger"
)

const (
	versionMajor                 string = "16"
	version                      string = "16.6.2"
	keeperSecretsManagerClientId string = "mg16.6.2" // Golang client ID starts with "mg" + version
	defaultKeeperHostname        string = "keepersecurity.com"
	clientIdHashTag              string = "KEEPER_SECRETS_MANAGER_CLIENT_ID" // Tag for hashing the client key to client id
)

var (
	keeperServers = map[string]string{
		"US":  "keepersecurity.com",
		"EU":  "keepersecurity.eu",
		"AU":  "keepersecurity.com.au",
		"GOV": "govcloud.keepersecurity.us",
		"JP":  "keepersecurity.jp",
		"CA":  "keepersecurity.ca",
	}
)

// getClientVersion returns the version of the client
func GetClientVersion(hardcode bool) string {
	// For the client version number we use the defined major and minor and revision numbers of the module version.
	// For example, module version of 0.1.23 would create a client version would be 16.1.23.

	// Get the version of the keeper secrets manager core
	result := versionMajor + ".2.0"

	// Allow the default version to be hard coded. If not build the client version from the module version.
	if !hardcode {
		if versionParts := strings.Split(version, "."); len(versionParts) > 2 {
			versionMinor := strings.TrimSpace(versionParts[1])
			parts := strings.FieldsFunc(versionParts[2], func(r rune) bool { return '0' > r || r > '9' })
			versionRevision := strings.TrimSpace(parts[0])
			if versionMajor != "" && versionMinor != "" && versionRevision != "" {
				result = fmt.Sprintf("%s.%s.%s", versionMajor, versionMinor, versionRevision)
			} else {
				klog.Error("Unable to determine the client version - using default: " + result)
			}
		}
	}
	return result
}

var (
	// Right now the client version is being hardcoded.
	// keeperSecretsManagerSdkClientId string            = "mg" + GetClientVersion(true)
	keeperServerPublicKeys map[string]string = map[string]string{
		"7":  "BK9w6TZFxE6nFNbMfIpULCup2a8xc6w2tUTABjxny7yFmxW0dAEojwC6j6zb5nTlmb1dAx8nwo3qF7RPYGmloRM",
		"8":  "BKnhy0obglZJK-igwthNLdknoSXRrGB-mvFRzyb_L-DKKefWjYdFD2888qN1ROczz4n3keYSfKz9Koj90Z6w_tQ",
		"9":  "BAsPQdCpLIGXdWNLdAwx-3J5lNqUtKbaOMV56hUj8VzxE2USLHuHHuKDeno0ymJt-acxWV1xPlBfNUShhRTR77g",
		"10": "BNYIh_Sv03nRZUUJveE8d2mxKLIDXv654UbshaItHrCJhd6cT7pdZ_XwbdyxAOCWMkBb9AZ4t1XRCsM8-wkEBRg",
		"11": "BA6uNfeYSvqagwu4TOY6wFK4JyU5C200vJna0lH4PJ-SzGVXej8l9dElyQ58_ljfPs5Rq6zVVXpdDe8A7Y3WRhk",
		"12": "BMjTIlXfohI8TDymsHxo0DqYysCy7yZGJ80WhgOBR4QUd6LBDA6-_318a-jCGW96zxXKMm8clDTKpE8w75KG-FY",
		"13": "BJBDU1P1H21IwIdT2brKkPqbQR0Zl0TIHf7Bz_OO9jaNgIwydMkxt4GpBmkYoprZ_DHUGOrno2faB7pmTR7HhuI",
		"14": "BJFF8j-dH7pDEw_U347w2CBM6xYM8Dk5fPPAktjib-opOqzvvbsER-WDHM4ONCSBf9O_obAHzCyygxmtpktDuiE",
		"15": "BDKyWBvLbyZ-jMueORl3JwJnnEpCiZdN7yUvT0vOyjwpPBCDf6zfL4RWzvSkhAAFnwOni_1tQSl8dfXHbXqXsQ8",
		"16": "BDXyZZnrl0tc2jdC5I61JjwkjK2kr7uet9tZjt8StTiJTAQQmnVOYBgbtP08PWDbecxnHghx3kJ8QXq1XE68y8c",
		"17": "BFX68cb97m9_sweGdOVavFM3j5ot6gveg6xT4BtGahfGhKib-zdZyO9pwvv1cBda9ahkSzo1BQ4NVXp9qRyqVGU",
	}
)
