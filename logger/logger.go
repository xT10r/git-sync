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

package logger

import (
	"log"
	"os"
)

var logger *log.Logger

// Инициализирует глобальный логгер
func Init(level string) {

	switch level {
	case "debug":
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	case "info":
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Ltime)
	case "warn":
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Llongfile|log.Ltime)
	case "error":
		logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Ltime)
	default:
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Ltime)
	}
}

// Возвращает глобальный логгер
func GetLogger() *log.Logger {
	Init("info") // TODO: реализовать изменение типа логгирования
	return logger
}

func Print(message string) {
	logger.Println(message)
}
