// Copyright 2024 Aleksey Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
