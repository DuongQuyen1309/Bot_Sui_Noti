package config

import (
	"encoding/json"
	"os"
)

type Token struct {
	Address string `json:"address"`
}
type Config struct {
	Wallet Wallet `json:"wallet"`
}
type Wallet struct {
	AddressId string  `json:"addressId"`
	Token     []Token `json:"token"`
}

func LoadCofig() (*Config, error) {
	configFile, err := os.ReadFile(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
