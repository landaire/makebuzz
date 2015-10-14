package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

const (
	configFile = "config.toml"
)

type BuzzFeedConfig struct {
	Feeds []string
}

type TwitterConfig struct {
	ConsumerToken, ConsumerSecret string
	AccessToken, AccessSecret     string
}

type Config struct {
	BuzzFeed BuzzFeedConfig
	Twitter  TwitterConfig
}

var (
	GlobalConfig *Config
)

func ParseConfig() *Config {
	var out Config

	file, err := os.Open(configFile)
	if err != nil {
		Logger.Fatalln(err)
		return nil
	}

	defer file.Close()

	if _, err := toml.DecodeReader(file, &out); err != nil {
		Logger.Fatalln(err)
		return nil
	}

	GlobalConfig = &out

	return &out
}
