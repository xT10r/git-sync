package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	GitURL       string
	GitBranch    string
	LocalDir     string
	SyncInterval int
}

func LoadConfig(configPath string) Config {

	var cfg Config
	viper.AddConfigPath(configPath)
	// viper.SetConfigName("config")
	// viper.SetConfigType("yaml")
	// viper.AddConfigPath(".")
	viper.AutomaticEnv()

	//err := viper.BindPFlags(flag.CommandLine)
	//if err != nil {
	//	log.Fatal(err)
	//}

	viper.Unmarshal(&cfg)

	// Валидация конфига

	return cfg
}
