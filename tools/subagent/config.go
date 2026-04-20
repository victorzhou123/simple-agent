package subagent

type Config struct {
	Prompt    string `json:"prompt"`
	MaxRounds int    `json:"max_rounds"`
}

func (cfg *Config) SetDefault() {
	if cfg.MaxRounds == 0 {
		cfg.MaxRounds = 30
	}
}
