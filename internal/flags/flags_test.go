// Copyright 2024 Aleksey Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flags

import (
	"bytes"
	"flag"
	"fmt"
	"git-sync/logger"
	"log"
	"os"
	"regexp"
	"testing"
	"time"
)

// Структура для тест-кейсов строковых переменных окружения

func TestGetFlagValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("testFlag", "testValue", "")

	value, exists := getFlagValue(fs, "testFlag")
	if !exists || value != "testValue" {
		t.Errorf("Expected 'testValue', got '%s'", value)
	}
}

func TestGetEnv(t *testing.T) {

	// Подготовим структуру для строковых переменных окружения
	type getEnvTestCase struct {
		envKey   string
		envValue string
		expected string
	}

	// Список тест-кейсов
	testCases := []getEnvTestCase{
		{"TEST_ENV_HAS_VALUE", "testValue", "testValue"},
		{"TEST_ENV_EMPTY_VALUE", "", "defaultValue"},
	}

	// Выполнение тестов
	for _, tc := range testCases {
		t.Run(tc.envKey, func(t *testing.T) {
			os.Setenv(tc.envKey, tc.envValue)
			defer os.Unsetenv(tc.envKey)
			value := getEnv(tc.envKey, tc.expected)
			if value != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, value)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {

	// Подготовим структуру для переменных окружения длительности
	type getEnvTestCase struct {
		envKey   string
		envValue string
		expected time.Duration
	}

	testCases := []getEnvTestCase{
		{"TEST_ENV_DURATION_VALUE", "30s", 30 * time.Second},
		{"TEST_ENV_DURATION__EMPTY_VALUE", "", 0},
		{"TEST_ENV_DURATION_DEFAULT_VALUE", "invalid", 5 * time.Second},
	}

	// Выполнение тестов
	for _, tc := range testCases {
		t.Run(tc.envKey, func(t *testing.T) {
			os.Setenv(tc.envKey, tc.envValue)
			defer os.Unsetenv(tc.envKey)
			value := getEnvDuration(tc.envKey, tc.expected)
			if value != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, value)
			}
		})
	}
}

func TestValidateFlagURL(t *testing.T) {

	tests := []struct {
		name      string
		flagName  string
		flagValue string
		desc      string
		expected  error
	}{
		{
			name:      "Flag exists and is set with valid URL",
			flagName:  "urlFlag",
			flagValue: "http://www.example.com",
			desc:      "URL Flag",
			expected:  nil,
		},
		{
			name:      "Flag does not exist",
			flagName:  "nonExistentFlag",
			flagValue: "",
			desc:      "Non-existent flag",
			expected:  fmt.Errorf("Non-existent flag is not set"),
		},
		{
			name:      "Flag exists but has invalid URL",
			flagName:  "invalidURLFlag",
			flagValue: "invalid-url",
			desc:      "Invalid URL Flag",
			expected:  fmt.Errorf("parse \"invalid-url\": invalid URI for request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			if tt.flagValue != "" {
				fs.String(tt.flagName, tt.flagValue, tt.desc)
			}

			err := validateFlagURL(fs, tt.flagName, tt.desc)

			if tt.expected != nil && err != nil && err.Error() != tt.expected.Error() {
				t.Errorf("Expected error '%v', but got '%v'", tt.expected, err)
			} else if tt.expected == nil && err != nil {
				t.Errorf("Expected no error, but got '%v'", err)
			} else if tt.expected != nil && err == nil {
				t.Errorf("Expected error '%v', but got no error", tt.expected)
			}
		})
	}
}

func TestValidateFlagLocalPath(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("pathFlag", "/", "")

	err := validateFlagLocalPath(fs, "pathFlag", "Path Flag")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}

// Тестовая функция validateFlagOptional проверяет, что при отсутствии флага или его пустом значении выводится предупреждающее сообщение
func TestValidateFlagOptional(t *testing.T) {
	tests := []struct {
		name       string
		flagName   string
		flagValue  string
		flagExists bool
		desc       string
		expected   string
	}{
		{
			name:       "Flag exists and is set",
			flagName:   "testFlag",
			flagValue:  "someValue",
			flagExists: true,
			desc:       "Test flag",
			expected:   "",
		},
		{
			name:       "Flag does not exist",
			flagName:   "nonExistentFlag",
			flagValue:  "",
			flagExists: false,
			desc:       "Non-existent flag",
			expected:   "Non-existent flag is not set",
		},
		{
			name:       "Flag exists but is empty",
			flagName:   "emptyFlag",
			flagValue:  "",
			flagExists: true,
			desc:       "Empty flag",
			expected:   "Empty flag is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			if tt.flagExists {
				fs.String(tt.flagName, tt.flagValue, tt.desc)
			}

			expectedPrefix := logger.GetLogger().GetWarnPrefix()

			// Перехватываем вывод логгера
			var buf bytes.Buffer
			logger.GetLogger().SetOutput(&buf)
			defer log.SetOutput(nil) // Восстанавливаем вывод логгера

			validateFlagOptional(fs, tt.flagName, tt.desc)

			logOutput := buf.String()

			// Создаем регулярное выражение для проверки префикса и ожидаемого сообщения без даты
			re := regexp.MustCompile(fmt.Sprintf(`\[WARN\] \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} %s`, tt.expected))

			if tt.expected != "" && !re.MatchString(logOutput) {
				t.Errorf("Expected log message to contain '%s', but got '%s'", expectedPrefix+tt.expected, logOutput)
			} else if tt.expected == "" && logOutput != "" {
				t.Errorf("Expected no log message, but got '%s'", logOutput)
			}
		})
	}
}

func TestValidateFlagSyncInterval(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("intervalFlag", "1m", "")

	err := validateFlagSyncInterval(fs, "intervalFlag", "Interval Flag")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}

func TestValidateFlagsHttpServer(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("httpServerAddr", "127.0.0.1:8080", "")
	fs.String("httpServerAuthUsername", "user", "")
	fs.String("httpServerAuthPassword", "pass", "")
	fs.String("httpServerAuthToken", "token", "")

	err := validateFlagsHttpServer(fs)
	if err != nil {
		t.Errorf("Expected no error, got '%s'", err)
	}
}
