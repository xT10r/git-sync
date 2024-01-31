/*
Здесь мы инициализируем логгер из стандартной библиотеки log, настраиваем вывод в консоль и формат со временем.
Экспортируем функцию инициализации Init() и GetLogger() для получения логгера.
*/

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
		logger = log.New(os.Stdout, "", log.LstdFlags)
	case "warn":
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)
	case "error":
		logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Ltime)
	}

	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
}

// Возвращает глобальный логгер
func GetLogger() *log.Logger {
	return logger
}

func Print(message string) {
	logger.Println(message)
}
