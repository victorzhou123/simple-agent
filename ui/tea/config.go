package tea

type Config struct {
	Tea       TeaConfig       `json:"tea"`
	Formatter FormatterConfig `json:"formatter"`
}

type TeaConfig struct {
	ModelName string `json:"model_name"`
	Endpoint  string `json:"endpoint"`
}

type FormatterConfig struct {
	Icons map[string]string `json:"icons"`
}
