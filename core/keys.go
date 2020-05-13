package core

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// KeysConfig provides information about keys for automatic unlocking.
type KeysConfig struct {
	Keys []string `json:"keys"`
}

// FetchKeysConfig fetches keys from the JSON configuration file.
func FetchKeysConfig(configPath string) (*KeysConfig, error) {
	path := filepath.Join(configPath, "keys.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	keys := &KeysConfig{}
	err = json.Unmarshal(data, keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}
