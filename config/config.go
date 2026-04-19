package config

import (
	"reflect"

	modelCfg "simple-agent/model/config"
	"simple-agent/utils"
)

type Config struct {
	Model modelCfg.Config `json:"model"`
}

func LoadConfig(path string, cfg *Config) error {
	if err := utils.LoadJSONFile(path, cfg); err != nil {
		return err
	}

	cfg.Model.LoadFromEnv()
	return nil
}

func (cfg *Config) SetDefaultAndValidate() error {
	return utils.SetDefaultAndValidate(reflect.ValueOf(cfg))
}
