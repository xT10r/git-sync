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
	"flag"
	"os"
	"testing"
	"time"
)

func TestGetFlagValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("testFlag", "testValue", "")

	value, exists := getFlagValue(fs, "testFlag")
	if !exists || value != "testValue" {
		t.Errorf("Expected 'testValue', got '%s'", value)
	}
}

func TestGetEnv(t *testing.T) {

	// Определяем структуру для тест-кейсов
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
	os.Setenv("TEST_ENV_DURATION", "1m")
	defer os.Unsetenv("TEST_ENV_DURATION")

	duration := getEnvDuration("TEST_ENV_DURATION", 30*time.Second)
	expected := 1 * time.Minute
	if duration != expected {
		t.Errorf("Expected '%s', got '%s'", expected, duration)
	}
}

func TestValidateFlagURL(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("urlFlag", "http://example.com", "")

	err := validateFlagURL(fs, "urlFlag", "URL Flag")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
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

func TestValidateFlagOptional(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("optionalFlag", "optionalValue", "")

	validateFlagOptional(fs, "optionalFlag", "Optional Flag")
	// Ensure no panic
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
