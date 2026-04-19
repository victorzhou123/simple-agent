package config

import "simple-agent/model/claude"

type Config struct {
	Claude claude.Config
}

func (cfg *Config) LoadFromEnv() {
	cfg.Claude.LoadFromEnv()
}
