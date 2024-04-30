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

type Logger struct {
	logger *log.Logger
}

const (
	infoPrefix    string = "[INFO] "
	debugPrefix   string = "[DEBUG] "
	warningPrefix string = "[WARN] "
	errorPrefix   string = "[ERROR] "
)

var (
	logger *Logger
)

var (
	defaultFlags = log.LstdFlags | log.Lshortfile | log.Ltime
)

func NewLogger() *Logger {

	return &Logger{
		logger: log.New(os.Stdout, "", defaultFlags),
	}
}

func GetLogger() *Logger {
	if logger == nil {
		logger = NewLogger()
	}
	return logger
}

func (l *Logger) Info(format string, v ...interface{}) {
	curPrefix := l.logger.Prefix()
	l.logger.SetPrefix(infoPrefix)
	defer l.logger.SetPrefix(curPrefix)
	l.logger.Printf(format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	curPrefix := l.logger.Prefix()
	l.logger.SetPrefix(debugPrefix)
	defer l.logger.SetPrefix(curPrefix)
	l.logger.Printf(format, v...)
}

func (l *Logger) Warning(format string, v ...interface{}) {
	curPrefix := l.logger.Prefix()
	l.logger.SetPrefix(warningPrefix)
	defer l.logger.SetPrefix(curPrefix)
	l.logger.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	curPrefix := l.logger.Prefix()
	l.logger.SetPrefix(errorPrefix)
	defer l.logger.SetPrefix(curPrefix)
	l.logger.Printf(format, v...)
}
