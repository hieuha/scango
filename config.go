package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type config struct {
	Version       string
	Timeout       int
	Redisauth     string
	Redisdb       int
	Rediserver    string
	Concurrentmax int
	CoreLog       string
	LogLevel      int
}

func LoadConfig(path string) error {
	if _, err := toml.DecodeFile(path, &Config); err != nil {
		return fmt.Errorf("could not load config: %s", err)
	}
	return nil
}
