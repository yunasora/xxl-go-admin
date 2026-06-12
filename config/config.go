package config

import (
	"encoding/json"
	"os"
)

var Cfg *Config

type Config struct {
	ServerPort           string `json:"server_port"`
	DBPath               string `json:"db_path"`
	GinMode              string `json:"gin_mode"`
	AccessToken          string `json:"access_token"`
	AppName              string `json:"app_name"`
	HTTPTimeout          int    `json:"http_timeout"`
	RegistryScanInterval int    `json:"registry_scan_interval"`
	RegistryTimeout      int    `json:"registry_timeout"`
	SchedulerInterval    int    `json:"scheduler_interval"`
}

func Load(path string) error {

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	Cfg = &Config{}
	return json.Unmarshal(data, Cfg)
}
