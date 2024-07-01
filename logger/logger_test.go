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

package logger_test

import (
	"bytes"
	"git-sync/logger"
	"strings"
	"testing"
)

// mockLoggerCreator - мок для интерфейса LoggerCreator
type mockLoggerCreator struct{}

// NewLogger имитирует создание нового экземпляра логгера
func (m *mockLoggerCreator) NewLogger() *logger.Logger {
	return logger.NewLogger()
}

// testCase - структура для описания тестовых случаев
type testCase struct {
	Name           string                       // Название теста
	Message        string                       // Сообщение
	ExpectedPrefix string                       // Ожидаемый префикс
	Func           func(*logger.Logger, string) // Функция, которая будет выполнена в тесте
}

func runTest(t *testing.T, _ func(*logger.Logger, string), c testCase) {
	mock := &mockLoggerCreator{}
	logger := mock.NewLogger()

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	c.Func(logger, c.Message)

	output := buf.String()

	// Проверка наличия сообщения
	if !strings.Contains(output, c.Message) {
		t.Errorf("[%s] Expected message '%s' not found in output: %s", c.Name, c.Message, output)
	}

	// Проверка наличия префикса
	if !strings.HasPrefix(output, c.ExpectedPrefix) {
		t.Errorf("[%s] Expected prefix: %s, got: %s", c.Name, c.ExpectedPrefix, output[:len(c.ExpectedPrefix)])
	}
}

func createTestCases() []testCase {
	return []testCase{
		{
			Name:           "TestInfo",
			Message:        "Test info message",
			ExpectedPrefix: logger.GetLogger().GetInfoPrefix(),
			Func: func(logger *logger.Logger, message string) {
				logger.Info(message)
			},
		},
		{
			Name:           "TestWarning",
			Message:        "Test warning message",
			ExpectedPrefix: logger.GetLogger().GetWarnPrefix(),
			Func: func(logger *logger.Logger, message string) {
				logger.Warning(message)
			},
		},
		{
			Name:           "TestDebug",
			Message:        "Test debug message",
			ExpectedPrefix: logger.GetLogger().GetDebugPrefix(),
			Func: func(logger *logger.Logger, message string) {
				logger.Debug(message)
			},
		},
		{
			Name:           "TestError",
			Message:        "Test error message",
			ExpectedPrefix: logger.GetLogger().GetErrPrefix(),
			Func: func(logger *logger.Logger, message string) {
				logger.Error(message)
			},
		},
	}
}

func TestLogger(t *testing.T) {
	testCases := createTestCases()
	for _, c := range testCases {
		runTest(t, c.Func, c)
	}
}
