// Copyright 2025 Aleksey Dobshikov
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
			defer func() {
				logger.GetLogger().SetOutput(os.Stderr) // Восстанавливаем вывод логгера
			}()

			validateFlagOptional(fs, tt.flagName, tt.desc)

			logOutput := buf.String()

			// Создаем регулярное выражение для проверки префикса и ожидаемого сообщения без даты
			re := regexp.MustCompile(fmt.Sprintf(`\[WARN\]  \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} %s`, regexp.QuoteMeta(tt.expected)))

			if tt.expected != "" && !re.MatchString(logOutput) {
				t.Errorf("Expected log message to contain '%s', but got '%s'", expectedPrefix+tt.expected, logOutput)
			} else if tt.expected == "" && logOutput != "" {
				t.Errorf("Expected no log message, but got '%s'", logOutput)
			}
		})
	}
}

// TestNewConsoleFlags tests the NewConsoleFlags function
func TestNewConsoleFlags(t *testing.T) {
	// Save original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test arguments with valid flag names
	os.Args = []string{"git-sync", "-repo-url=https://github.com/example/repo", "-repo-branch=main", "-local-path=/tmp"}

	cf := NewConsoleFlags()
	if cf == nil {
		t.Fatal("NewConsoleFlags returned nil")
	}

	if cf.Gitsync == nil {
		t.Error("Gitsync flag set is nil")
	}
}

// TestParseFlags tests the ParseFlags function
func TestParseFlags(t *testing.T) {
	// Save original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test arguments with valid flag names
	os.Args = []string{"git-sync", "-repo-url=https://github.com/example/repo", "-repo-branch=main", "-local-path=/tmp"}

	fs := ParseFlags()
	if fs == nil {
		t.Fatal("ParseFlags returned nil")
	}

	// Check that flags were parsed correctly
	repoFlag := fs.Lookup("repo-url")
	if repoFlag == nil {
		t.Error("repo-url flag not found")
	} else if repoFlag.Value.String() != "https://github.com/example/repo" {
		t.Errorf("Expected repo flag value 'https://github.com/example/repo', got '%s'", repoFlag.Value.String())
	}

	branchFlag := fs.Lookup("repo-branch")
	if branchFlag == nil {
		t.Error("repo-branch flag not found")
	} else if branchFlag.Value.String() != "main" {
		t.Errorf("Expected branch flag value 'main', got '%s'", branchFlag.Value.String())
	}

	pathFlag := fs.Lookup("local-path")
	if pathFlag == nil {
		t.Error("local-path flag not found")
	} else if pathFlag.Value.String() != "/tmp" {
		t.Errorf("Expected path flag value '/tmp', got '%s'", pathFlag.Value.String())
	}
}

// TestCheckRequiredFlags tests the CheckRequiredFlags method
func TestCheckRequiredFlags(t *testing.T) {
	// Test case 1: All required flags are set
	t.Run("AllRequiredFlagsSet", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "https://github.com/example/repo", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", "/tmp", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.CheckRequiredFlags()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test case 2: Missing required flags
	t.Run("MissingRequiredFlags", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		// Only set one of the required flags
		fs.String("repo-url", "https://github.com/example/repo", "")
		// repo-branch and local-path are missing

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.CheckRequiredFlags()
		if err == nil {
			t.Error("Expected error for missing required flags, got nil")
		} else {
			expectedError := "required flags are missing: repo-branch, local-path. Use --help or --config-help for more information"
			if err.Error() != expectedError {
				t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
			}
		}
	})

	// Test case 3: No flags set
	t.Run("NoFlagsSet", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		// No flags set

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.CheckRequiredFlags()
		if err == nil {
			t.Error("Expected error for missing required flags, got nil")
		} else {
			expectedError := "required flags are missing: repo-url, repo-branch, local-path. Use --help or --config-help for more information"
			if err.Error() != expectedError {
				t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
			}
		}
	})
}

// TestValidateFlags tests the ValidateFlags method
func TestValidateFlags(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test case 1: Valid flags
	t.Run("ValidFlags", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "https://github.com/example/repo", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", tempDir, "") // Use temp dir
		fs.String("repo-user", "user", "")
		fs.String("repo-token", "token", "")
		fs.String("sync-interval", "1m", "")
		fs.String("http-server-addr", "127.0.0.1:8080", "")
		fs.String("http-auth-username", "httpuser", "")
		fs.String("http-auth-password", "httppass", "")
		fs.String("http-auth-token", "httptoken", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.ValidateFlags()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test case 2: Invalid URL
	t.Run("InvalidURL", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "invalid-url", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", tempDir, "") // Use temp dir
		fs.String("sync-interval", "1m", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.ValidateFlags()
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})

	// Test case 3: Invalid path
	t.Run("InvalidPath", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "https://github.com/example/repo", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", "/non/existent/path", "")
		fs.String("sync-interval", "1m", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.ValidateFlags()
		if err == nil {
			t.Error("Expected error for invalid path, got nil")
		}
	})

	// Test case 4: Invalid interval
	t.Run("InvalidInterval", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "https://github.com/example/repo", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", tempDir, "") // Use temp dir
		fs.String("sync-interval", "invalid", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.ValidateFlags()
		if err == nil {
			t.Error("Expected error for invalid interval, got nil")
		}
	})

	// Test case 5: Negative interval
	t.Run("NegativeInterval", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "https://github.com/example/repo", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", tempDir, "") // Use temp dir
		fs.String("sync-interval", "-1m", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.ValidateFlags()
		if err == nil {
			t.Error("Expected error for negative interval, got nil")
		}
	})

	// Test case 6: Invalid HTTP server address
	t.Run("InvalidHTTPServerAddr", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("repo-url", "https://github.com/example/repo", "")
		fs.String("repo-branch", "main", "")
		fs.String("local-path", tempDir, "") // Use temp dir
		fs.String("sync-interval", "1m", "")
		fs.String("http-server-addr", "invalid-address", "")

		cf := &ConsoleFlags{Gitsync: fs}
		err := cf.ValidateFlags()
		if err == nil {
			t.Error("Expected error for invalid HTTP server address, got nil")
		}
	})
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

// TestGetEnvBool tests the getEnvBool function
func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "Environment variable not set",
			envKey:       "TEST_NOT_SET",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "Environment variable set to true",
			envKey:       "TEST_TRUE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "Environment variable set to false",
			envKey:       "TEST_FALSE",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "Environment variable set to 1",
			envKey:       "TEST_ONE",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "Environment variable set to 0",
			envKey:       "TEST_ZERO",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "Environment variable set to invalid value",
			envKey:       "TEST_INVALID",
			envValue:     "invalid",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			value := getEnvBool(tt.envKey, tt.defaultValue)
			if value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, value)
			}
		})
	}
}
