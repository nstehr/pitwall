package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/openziti/sdk-golang/ziti"
)

type Config struct {
	AuthEndpoint string `json:"authEndpoint"`
	ApiEndpoint  string `json:"apiEndpoint"`
	OAuth        OAuth  `json:"oAuth"`
	ZitiConfig   *ziti.Config
}

type OAuth struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	Expiry       time.Time `json:"expiry"`
}

func ConfigExists(configPath string) bool {
	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// instead of returning the file handle, lets return the parsed contents as the Config struct
func GetOrCreateConfig(cfgPath string) (*Config, error) {

	_, err := os.Stat(cfgPath)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	jsonBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (config *Config) WriteToFile(cfgPath string) error {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.Write(jsonBytes)
	if err != nil {
		return err
	}
	return f.Close()
}
