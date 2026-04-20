package config

import (
	"reflect"

	modelCfg "simple-agent/model/config"
	toolsCfg "simple-agent/tools"
	uiCfg "simple-agent/ui/config"
	"simple-agent/utils"
)

type Config struct {
	Model modelCfg.Config `json:"model"`
	Tools toolsCfg.Config `json:"tools"`
	UI    uiCfg.Config    `json:"ui"`
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
