package config

import "github.com/spf13/viper"

func init() {
	viper.AutomaticEnv()
	viper.MustBindEnv("bot-token", "BOT_TOKEN")
	viper.MustBindEnv("webhook-url", "WEBHOOK_URL")
}
