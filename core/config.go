package core

import (
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config is the configuration for the daemon.
type Config struct {
	Verbosity string            `json:"verbosity"`
	Server    *ServerConfig     `json:"server"`
	Stores    []*Store          `json:"stores"`
	Rules     []*RuleDefinition `json:"rules"`
	ETH2Dir   string            `json:"eth2dir"`
}

// ServerConfig contains configuration for the server.
type ServerConfig struct {
	Name        string `json:"name"`
	Port        int    `json:"port"`
	CertPath    string `json:"certificate_path"`
	StoragePath string `json:"storage_path"`
}

const (
	defaultPort = 12346
)

// NewConfig creates a new configuration.
// Configuration can come from the configuration file or environment variables.
func NewConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	viper.SetEnvPrefix("walletd")

	// Explicit environment variable bindings.
	if err := viper.BindEnv("server_name"); err != nil {
		return nil, errors.Wrap(err, "Failed to bind server_name")
	}
	if err := viper.BindEnv("port"); err != nil {
		return nil, errors.Wrap(err, "Failed to bind port")
	}

	c := &Config{}
	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	if c.Server == nil {
		c.Server = &ServerConfig{}
	}

	if viper.GetString("server_name") != "" {
		c.Server.Name = viper.GetString("server_name")
	}
	if viper.GetInt("port") != 0 {
		c.Server.Port = viper.GetInt("port")
	}
	if c.Server.Port == 0 {
		c.Server.Port = defaultPort
	}
	if c.Server.CertPath == "" {
		c.Server.CertPath = filepath.Join(configPath, "security")
	}
	if c.Server.StoragePath == "" {
		c.Server.StoragePath = filepath.Join(configPath, "storage")
	}

	return c, nil
}
