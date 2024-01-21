package core

type ConfigKey string

const (
	KEY_URL                  ConfigKey = "url" // base URL for the Secrets Manager service
	KEY_SERVER_PUBLIC_KEY_ID ConfigKey = "serverPublicKeyId"
	KEY_CLIENT_ID            ConfigKey = "clientId"
	KEY_CLIENT_KEY           ConfigKey = "clientKey"         // The key that is used to identify the client before public key
	KEY_APP_KEY              ConfigKey = "appKey"            // The application key with which all secrets are encrypted
	KEY_OWNER_PUBLIC_KEY     ConfigKey = "appOwnerPublicKey" // The application owner public key, to create records
	KEY_PRIVATE_KEY          ConfigKey = "privateKey"        // The client's private key
	KEY_PUBLIC_KEY           ConfigKey = "publicKey"         // The client's public key
	KEY_HOSTNAME             ConfigKey = "hostname"          // base hostname for the Secrets Manager service
	defaultOwnerPublicKeyId  string    = "7"
)

func GetDefaultOwnerPublicKey() string {
	if ownerKey, found := keeperServerPublicKeys[defaultOwnerPublicKeyId]; found {
		return ownerKey
	}
	return ""
}

func GetConfigKey(value string) ConfigKey {
	switch value {
	case string(KEY_URL):
		return KEY_URL
	case string(KEY_SERVER_PUBLIC_KEY_ID):
		return KEY_SERVER_PUBLIC_KEY_ID
	case string(KEY_CLIENT_ID):
		return KEY_CLIENT_ID
	case string(KEY_CLIENT_KEY):
		return KEY_CLIENT_KEY
	case string(KEY_APP_KEY):
		return KEY_APP_KEY
	case string(KEY_OWNER_PUBLIC_KEY):
		return KEY_OWNER_PUBLIC_KEY
	case string(KEY_PRIVATE_KEY):
		return KEY_PRIVATE_KEY
	case string(KEY_PUBLIC_KEY):
		return KEY_PUBLIC_KEY
	case string(KEY_HOSTNAME):
		return KEY_HOSTNAME
	default:
		return ""
	}
}

func GetConfigKeys() []ConfigKey {
	return []ConfigKey{
		KEY_URL,
		KEY_SERVER_PUBLIC_KEY_ID,
		KEY_CLIENT_ID,
		KEY_CLIENT_KEY,
		KEY_APP_KEY,
		KEY_OWNER_PUBLIC_KEY,
		KEY_PRIVATE_KEY,
		KEY_PUBLIC_KEY,
		KEY_HOSTNAME,
	}
}
