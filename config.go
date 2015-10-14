package main

import (
	"os"

	"strings"

	"github.com/naoina/toml"
)

const (
	configFile = "config.toml"
)

type BuzzFeedConfig struct {
	Feeds []string
}

type TwitterConfig struct {
	PostTweet      bool   `toml:"post_tweet"`
	ConsumerToken  string `toml:"consumer_token"`
	ConsumerSecret string `toml:"consumer_secret"`
	AccessToken    string `toml:"acces_token"`
	AccessSecret   string `toml:"access_secret"`
}

type Config struct {
	BuzzFeed BuzzFeedConfig `toml:"buzzfeed"`
	Twitter  TwitterConfig `toml:"twitter"`
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

	decoder := toml.NewDecoder(file)

	if err := decoder.Decode(&out); err != nil {
		Logger.Fatalln(err)
		return nil
	}

	GlobalConfig = &out

	return &out
}

func (this TwitterConfig) IsValid() bool {
	vals := []string{this.AccessSecret, this.AccessToken, this.ConsumerSecret, this.ConsumerToken}

	for _, val := range vals {
		if strings.TrimSpace(val) == "" {
			return false
		}
	}

	return true
}
