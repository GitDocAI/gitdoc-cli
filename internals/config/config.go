package config

import "os"

type Config struct {
	ProjectPath string
}

func NewConfig() Config {
	return Config{
		ProjectPath: os.Getenv("NEXTRAPATH"),
	}
}
