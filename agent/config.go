package agent

type Config struct {
	StreamTimeOut int `json:"stream_time_out"`
}

func (cfg *Config) SetDefault() {
	if cfg.StreamTimeOut == 0 {
		cfg.StreamTimeOut = 60 * 3
	}
}
