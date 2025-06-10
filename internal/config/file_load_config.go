package config

import (
	"encoding/json"
	"os"
)

type Token struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}
type Config struct {
	Wallet Wallet `json:"wallet"`
}
type Wallet struct {
	Name      string  `json:"name"`
	AddressId string  `json:"addressId"`
	Token     []Token `json:"token"`
}

func LoadCofig() (*Config, error) {
	configFile, err := os.ReadFile("internal/config/config.json")
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
