package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type APIConfig struct {
	URL   string
	Token string
}

type Config struct {
	Node struct {
		StandardConsensusRPC string
		APIs                 []APIConfig
		MinimumBalance       int 
	}
	Discord struct {
		WebhookUrl  string
		AlertUserID []string
		AlertRoleID []string
	}
}

func LoadConfig(tomlpath string, envpath string) (*Config, error) {
	v1 := viper.New()
	v1.SetConfigFile(tomlpath)
	v1.SetConfigType("toml")
	if err := v1.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read the configuration file, %s", err)
	}

	var c Config
	if err := v1.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration, %s", err)
	}
	v2 := viper.New()
	v2.SetConfigFile(envpath)
	v2.SetConfigType("env")
	if err := v2.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read the configuration file, %s", err)
	}
	webhook := v2.GetString("discord_webhook")
	c.Discord.WebhookUrl = webhook
	return &c, nil
}
