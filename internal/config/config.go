package config

type Config struct {
	Prompt string
	Colors map[string]string
}

func LoadConfig() *Config {
	// TODO: Load from YAML/JSON file
	return &Config{
		Prompt: "supershell> ",
		Colors: map[string]string{},
	}
}
