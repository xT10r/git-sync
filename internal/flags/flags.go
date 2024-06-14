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
	if err := validateFlagURL(fs, constants.FlagRepoUrl, "Repository URL"); err != nil {
		return err
	}

	// Local path
	if err := validateFlagLocalPath(fs, constants.FlagLocalPath, "Local Path"); err != nil {
		return err
	}

	// Repo Branch
	validateFlagOptional(fs, constants.FlagRepoBranch, "Repository Branch")

	// Repo user
	validateFlagOptional(fs, constants.FlagRepoAuthUser, "Repository User")

	// Repo token
	validateFlagOptional(fs, constants.FlagRepoAuthToken, "Repository Token")

	// Sync interval
	if err := validateFlagSyncInterval(fs, constants.FlagSyncInterval, "Sync Interval"); err != nil {
		return err
	}

	// HTTP Server Addr
	if err := validateFlagsHttpServer(fs); err != nil {
		return err
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

func validateFlagURL(fs *flag.FlagSet, fn string, desc string) error {

	repoUrl, isExists := getFlagValue(fs, fn)

	if !isExists {
		return fmt.Errorf("%s is not set", desc)
	}

	_, err := url.Parse(repoUrl)
	if err != nil {
		return fmt.Errorf("неверный формат URL-ссылки: %s", err)
	}
	return nil
}

func validateFlagLocalPath(fs *flag.FlagSet, fn string, desc string) error {

	localPath, isExists := getFlagValue(fs, fn)

	if !isExists {
		return fmt.Errorf("%s is not set", desc)
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("указанный путь не существует: %s", localPath)
	}
	return nil
}

func validateFlagOptional(fs *flag.FlagSet, fn string, desc string) {
	if fv, isExists := getFlagValue(fs, fn); !isExists || fv == "" {
		logger.GetLogger().Warning("%s is not set\n", desc)
	}
}

func validateFlagSyncInterval(fs *flag.FlagSet, fn string, desc string) error {

	fv, isExists := getFlagValue(fs, fn)
	if !isExists {
		return fmt.Errorf("%s is not set", desc)
	}

	if duration, err := time.ParseDuration(fv); err != nil {
		return fmt.Errorf("не удалось привести строковое значение интервала синхронизации к длительности")
	} else {
		if duration <= 0 {
			return fmt.Errorf("интервал синхронизации должен быть положительным")
		}
	}
	return nil
}

func validateFlagsHttpServer(fs *flag.FlagSet) error {

	// HTTP Server Addr
	httpServerAddr, _ := getFlagValue(fs, constants.FlagHttpServerAddr)

	if len(httpServerAddr) == 0 {
		return nil
	}

	// Разделение адреса на IP и порт
	parts := strings.Split(httpServerAddr, ":")
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

	// HTTP Server Auth username
	username, _ := getFlagValue(fs, constants.FlagHttpServerAuthUsername)
	password, _ := getFlagValue(fs, constants.FlagHttpServerAuthPassword)

	// HTTP Server Auth Baerer Token
	token, _ := getFlagValue(fs, constants.FlagHttpServerAuthToken)

	if len(username) == 0 && len(password) == 0 && len(token) == 0 {
		logger.GetLogger().Warning("HTTP-сервер: аутентификация не активна")
	}

	return nil
}

