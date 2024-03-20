package config

import "github.com/spf13/viper"

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("admin-username", "vpakh")

	viper.MustBindEnv("port", "PORT")
	viper.MustBindEnv("bot-token", "BOT_TOKEN")
	viper.MustBindEnv("webhook-url", "WEBHOOK_URL")
	viper.MustBindEnv("project-id", "PROJECT_ID")
	viper.MustBindEnv("project-region", "PROJECT_REGION")
	viper.MustBindEnv("k-service", "K_SERVICE")
}
