/*
Это main пакет - точка входа приложения. Он импортирует другие пакеты, собирает конфиг, запускает логгер и основную логику приложения.
Также в проекте могут быть пакеты docs, deploy, scripts для документации, деплоя и вспомогательных скриптов соответственно.
*/

package main

import (
	"context"
	"fmt"
	"git-sync/git"
	"git-sync/internal/flags"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Создаем контекст и функцию для отмены контекста
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flagSet := flags.NewConsoleFlags("git-sync")
	syncOptions, err := git.NewSyncOptions(flagSet)
	if err != nil {
		fmt.Println("Ошибка при создании объекта SyncOptions:", err)
		return
	}

	// Запускаем периодическую синхронизацию в отдельной горутине
	go git.StartSync(ctx, syncOptions)

	// Ждем сигналов SIGINT или SIGTERM для завершения программы
	waitForSignals(cancel)

	// Отменяем контекст и ждем завершения горутин
	cancel()
}

// waitForSignals ожидает сигналы SIGINT или SIGTERM и вызывает функцию cancel для завершения программы.
func waitForSignals(cancel context.CancelFunc) {

	// Вызываем функцию cancel для завершения программы
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигналы
	sig := <-signalChan
	fmt.Printf("Получен сигнал %s. Завершение программы...\n", sig)
}
