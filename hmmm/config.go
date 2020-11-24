package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// getConfig ...
func getConfig() (*viper.Viper, error) {
	keys := []string{"ConsumerKey", "ConsumerSecret", "AccessToken", "AccessSecret", "Intro", "Body"}
	cfg := viper.New()
	cfg.AddConfigPath(".")
	cfg.AddConfigPath("$HOME/.hmmm")
	cfg.SetConfigName("config")
	cfg.SetEnvPrefix("Hmmm")

	for _, key := range keys {
		cfg.SetDefault(key, nil)
	}

	cfg.AutomaticEnv()
	if err := cfg.ReadInConfig(); err != nil {
		fmt.Printf("Unable to read config file, %v.", err)
	}

	for _, key := range keys {
		if cfg.Get(key) == nil {
			return nil, fmt.Errorf("[%s] is required.", key)
		}
	}

	return cfg, nil
}
