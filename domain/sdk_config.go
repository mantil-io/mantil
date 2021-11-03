package domain

import (
	"encoding/base64"
	"encoding/json"
)

const (
	EnvSDKConfig = "MANTIL_GO_CONFIG"
)

// this is a copy of the Config struct in mantil.go
type SDKConfig struct {
	ResourceTags    map[string]string
	WsForwarderName string
}

func (c *SDKConfig) Encode() string {
	buf, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(buf)
}
