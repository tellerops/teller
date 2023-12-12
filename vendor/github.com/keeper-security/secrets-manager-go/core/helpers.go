package core

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strings"

	klog "github.com/keeper-security/secrets-manager-go/core/logger"
)

func GetServerHostname(hostname string, configStore IKeyValueStorage) string {
	hostnameToUse := defaultKeeperHostname
	if envHostname := strings.TrimSpace(os.Getenv("KSM_HOSTNAME")); envHostname != "" {
		hostnameToUse = envHostname
	} else if cfgHostname := strings.TrimSpace(configStore.Get(KEY_HOSTNAME)); cfgHostname != "" {
		hostnameToUse = cfgHostname
	} else if codedHostname := strings.TrimSpace(hostname); codedHostname != "" {
		hostnameToUse = codedHostname
	}

	// Parse URL to get only domain:
	hostnameToUse = strings.TrimSpace(hostnameToUse)
	hostnameToReturn := hostnameToUse

	if !strings.HasPrefix(strings.ToLower(hostnameToUse), "http") {
		hostnameToUse = "https://" + hostnameToUse
	}
	if u, err := url.Parse(hostnameToUse); err == nil && u.Host != "" {
		hostnameToReturn = u.Host
		if host, _, err := net.SplitHostPort(u.Host); err == nil && host != "" {
			hostnameToReturn = host
		}
	}

	klog.Debug("Keeper server hostname: " + hostnameToReturn)

	return hostnameToReturn
}

func IsJson(jsonStr string) bool {
	var js interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

func ObjToDict(obj interface{}) map[string]interface{} {
	if o, ok := obj.(map[string]interface{}); ok {
		return o
	}
	content, err := json.Marshal(obj)
	if err != nil {
		content = []byte("{}")
	}
	return JsonToDict(string(content))

}

func GetFolderByKey(folderUid string, folders []*Folder) *Folder {
	for _, f := range folders {
		if f.uid == folderUid {
			return f
		}
	}
	return nil
}
