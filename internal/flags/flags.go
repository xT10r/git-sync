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

package flags

import (
	"flag"
	"fmt"
	"git-sync/internal/constants"
	"git-sync/logger"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// FlagSet представляет набор флагов командной строки.
type ConsoleFlags struct {
	Gitsync *flag.FlagSet
}

// NewConsoleFlags создает новый набор флагов командной строки.
func NewConsoleFlags() *ConsoleFlags {

	fs := ParseFlags()

	if err := validateFlags(fs); err != nil {
		logger.GetLogger().Error("%v", err)
		return nil
	}

	return &ConsoleFlags{
		Gitsync: fs,
	}
}

// ParseFlags инициализирует флаги с помощью набора флагов командной строки.
func ParseFlags() *flag.FlagSet {

	fs := flag.NewFlagSet("git-sync", flag.ExitOnError)

	fs.String(constants.FlagLocalPath, getEnv(constants.EnvLocalPath, ""), fmt.Sprintf("Путь к локальному репозиторию (%s)", constants.EnvLocalPath))

	fs.String(constants.FlagRepoUrl, getEnv(constants.EnvRepoUrl, ""), fmt.Sprintf("URL удаленного репозитория (%s)", constants.EnvRepoUrl))
	fs.String(constants.FlagRepoBranch, getEnv(constants.EnvRepoBranch, ""), fmt.Sprintf("Ветка удаленного репозитория (%s)", constants.EnvRepoBranch))
	fs.String(constants.FlagRepoAuthUser, getEnv(constants.EnvRepoAuthUser, ""), fmt.Sprintf("Учетная запись (%s)", constants.EnvRepoAuthUser))
	fs.String(constants.FlagRepoAuthToken, getEnv(constants.EnvRepoAuthToken, ""), fmt.Sprintf("Токен авторизации (%s)", constants.EnvRepoAuthToken))

	fs.Duration(constants.FlagSyncInterval, getEnvDuration(constants.EnvSyncInterval, 30*time.Second), fmt.Sprintf("Интервал обновления репозитория (%s)", constants.EnvSyncInterval))

	fs.String(constants.FlagHttpServerAddr, getEnv(constants.EnvHttpServerAddr, ""), fmt.Sprintf("Адрес http-сервера (+порт) (%s)", constants.EnvHttpServerAddr))
	fs.String(constants.FlagHttpServerAuthUsername, getEnv(constants.EnvHttpServerAuthUsername, ""), fmt.Sprintf("Имя пользователя http-сервера (%s)", constants.EnvHttpServerAuthUsername))
	fs.String(constants.FlagHttpServerAuthPassword, getEnv(constants.EnvHttpServerAuthPassword, ""), fmt.Sprintf("Пароль пользователя http-сервера (%s)", constants.EnvHttpServerAuthPassword))
	fs.String(constants.FlagHttpServerAuthToken, getEnv(constants.EnvHttpServerAuthToken, ""), fmt.Sprintf("Baerer-токен http-сервера (%s)", constants.EnvHttpServerAuthToken))

	fs.Parse(os.Args[1:])

	return fs
}

func validateFlags(fs *flag.FlagSet) error {

	// Repo URL
	if repoUrl, isExists := getFlagValue(fs, constants.FlagRepoUrl); !isExists {
		return fmt.Errorf("repository Url is not set")
	} else {
		if err := validateRemoteURL(repoUrl); err != nil {
			return err
		}
	}

	// Repo Branch
	if repoBranch, isExists := getFlagValue(fs, constants.FlagRepoBranch); !isExists || repoBranch == "" {
		logger.GetLogger().Warning("Repository branch is not set\n")
	}

	// Repo user
	if repoUser, isExists := getFlagValue(fs, constants.FlagRepoAuthUser); !isExists || repoUser == "" {
		logger.GetLogger().Warning("Repository user is not set\n")
	}

	// Repo token
	if repoToken, isExists := getFlagValue(fs, constants.FlagRepoAuthToken); !isExists || repoToken == "" {
		logger.GetLogger().Warning("Repository user token is not set\n")
	}

	// Local path
	if localPath, isExists := getFlagValue(fs, constants.FlagLocalPath); !isExists {
		return fmt.Errorf("local path is not set")
	} else {
		if err := validateLocalPath(localPath); err != nil {
			return err
		}
	}

	// Sync interval
	if syncInterval, isExists := getFlagValue(fs, constants.FlagSyncInterval); !isExists {
		return fmt.Errorf("local path is not set")
	} else {
		if err := validateSyncInterval(syncInterval); err != nil {
			return err
		}
	}

	// HTTP Server Addr
	if httpServerAddr, _ := getFlagValue(fs, constants.FlagHttpServerAddr); len(httpServerAddr) > 0 {
		if err := validateServerAddress(httpServerAddr); err != nil {
			return err
		}

		// HTTP Server Auth username
		username, userIsExists := getFlagValue(fs, constants.FlagHttpServerAuthUsername)

		if userIsExists && len(username) > 0 {

			// HTTP Server Auth password
			password, passwordisExists := getFlagValue(fs, constants.FlagHttpServerAuthPassword)
			if username == password && passwordisExists {
				logger.GetLogger().Warning("HTTP-сервер: имя пользователя и пароль совпадают\n")
			} else if len(password) == 0 {
				logger.GetLogger().Warning("HTTP-сервер: аутентификация не активна (отсутствует пароль пользователя)\n")
			}
		}

		// HTTP Server Auth Baerer Token
		token, tokenisExists := getFlagValue(fs, constants.FlagHttpServerAuthToken)

		if tokenisExists && !userIsExists && len(token) == 0 {
			logger.GetLogger().Warning("HTTP-сервер: не указан токен\n")
		}

	}

	return nil
}

func getFlagValue(fs *flag.FlagSet, flagName string) (string, bool) {
	if f := fs.Lookup(flagName); f != nil {
		value := f.Value.String()
		return value, true
	}

	return "", false
}

// getEnv возвращает значение переменной окружения или значение по умолчанию, если переменная не установлена.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvDuration возвращает значение переменной окружения в формате time.Duration или значение по умолчанию, если переменная не установлена или имеет некорректный формат.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func validateRemoteURL(remoteURL string) error {
	_, err := url.Parse(remoteURL)
	if err != nil {
		return fmt.Errorf("неверный формат URL-ссылки: %s", err)
	}
	return nil
}

func validateLocalPath(localPath string) error {
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("указанный путь не существует: %s", localPath)
	}
	return nil
}

func validateSyncInterval(syncIntervalStr string) error {

	if syncInterval, err := time.ParseDuration(syncIntervalStr); err != nil {
		return fmt.Errorf("не удалось привести строковое значение интервала синхронизации к длительности")
	} else {
		if syncInterval <= 0 {
			return fmt.Errorf("интервал синхронизации должен быть положительным")
		}
	}
	return nil
}

func validateServerAddress(addr string) error {
	// Разделение адреса на IP и порт
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("адрес должен быть в формате IP:PORT")
	}

	// Проверка корректности IP адреса
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return fmt.Errorf("некорректный IP адрес")
	}

	// Проверка корректности порта
	port, err := strconv.Atoi(parts[1])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("некорректный порт. Допустимый диапазон портов [1-65535]")
	}

	return nil
}
