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
	flags, err := parseFlags()
	if err != nil {
		logger.GetLogger().Error("%v", err)
		return nil
	}

	return &ConsoleFlags{
		Gitsync: flags,
	}
}

// ParseFlags инициализирует флаги с помощью набора флагов командной строки.
func parseFlags() (*flag.FlagSet, error) {

	fs := flag.NewFlagSet("git-sync", flag.ExitOnError)
	localPath := fs.String(constants.FlagLocalPath, getEnv(constants.EnvLocalPath, ""), fmt.Sprintf("Путь к локальному репозиторию (%s)", constants.EnvLocalPath))
	repoUrl := fs.String(constants.FlagRepoUrl, getEnv(constants.EnvRepoUrl, ""), fmt.Sprintf("URL удаленного репозитория (%s)", constants.EnvRepoUrl))
	repoBranch := fs.String(constants.FlagRepoBranch, getEnv(constants.EnvRepoBranch, ""), fmt.Sprintf("Ветка удаленного репозитория (%s)", constants.EnvRepoBranch))
	repoAuthUser := fs.String(constants.FlagRepoAuthUser, getEnv(constants.EnvRepoAuthUser, ""), fmt.Sprintf("Учетная запись (%s)", constants.EnvRepoAuthUser))
	repoAuthToken := fs.String(constants.FlagRepoAuthToken, getEnv(constants.EnvRepoAuthToken, ""), fmt.Sprintf("Токен авторизации (%s)", constants.EnvRepoAuthToken))
	syncInterval := fs.Duration(constants.FlagSyncInterval, getEnvDuration(constants.EnvSyncInterval, 30*time.Second), fmt.Sprintf("Интервал обновления репозитория (%s)", constants.EnvSyncInterval))
	httpServerAddr := fs.String(constants.FlagHttpServerAddr, getEnv(constants.EnvHttpServerAddr, ""), fmt.Sprintf("Адрес http-сервера (+порт) (%s)", constants.EnvHttpServerAddr))
	httpServerAuthUsername := fs.String(constants.FlagHttpServerAuthUsername, getEnv(constants.EnvHttpServerAuthUsername, ""), fmt.Sprintf("Имя пользователя http-сервера (%s)", constants.EnvHttpServerAuthUsername))
	httpServerAuthPassword := fs.String(constants.FlagHttpServerAuthPassword, getEnv(constants.EnvHttpServerAuthPassword, ""), fmt.Sprintf("Пароль пользователя http-сервера (%s)", constants.EnvHttpServerAuthPassword))
	httpServerAuthBaererToken := fs.String(constants.FlagHttpServerAuthToken, getEnv(constants.EnvHttpServerAuthBearerToken, ""), fmt.Sprintf("Baerer-токен http-сервера (%s)", constants.EnvHttpServerAuthBearerToken))
	fs.Parse(os.Args[1:])

	if err := validateRemoteURL(*repoUrl); err != nil {
		return nil, err
	}

	if err := validateLocalPath(*localPath); err != nil {
		return nil, err
	}

	if err := validateSyncInterval(*syncInterval); err != nil {
		return nil, err
	}

	if *repoBranch == "" {
		logger.GetLogger().Warning("Branch is not set\n")
	}

	if *repoAuthUser == "" {
		logger.GetLogger().Warning("User is not set\n")
	}

	if *repoAuthToken == "" {
		logger.GetLogger().Warning("Token is not set\n")
	}

	if *httpServerAddr != "" {
		if err := validateServerAddress(*httpServerAddr); err != nil {
			return nil, err
		}
	}

	if *httpServerAuthUsername != "" {
		// проверка имени пользователя
	}

	if *httpServerAuthPassword != "" {
		// проверка пароля пользователя
	}

	if *httpServerAuthBaererToken != "" {
		// проверка токена
	}

	return fs, nil
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

func validateSyncInterval(syncInterval time.Duration) error {
	if syncInterval <= 0 {
		return fmt.Errorf("интервал синхронизации должен быть положительным")
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
