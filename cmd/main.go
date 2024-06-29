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

package main

import (
	"context"
	"fmt"
	"git-sync/git"
	"git-sync/internal/flags"
	"git-sync/internal/gitsync"
	"git-sync/internal/http"
	"git-sync/logger"
	"os"
	"os/signal"

	"syscall"
)

func main() {

	// Создаем контекст и функцию для отмены контекста
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flagSet := flags.NewConsoleFlags()

	// Проверка, были ли заданый обязательные флаги
	if err := flagSet.CheckRequiredFlags(); err != nil {
		logger.GetLogger().Error("%v", err)
		fmt.Printf("\n")
		flagSet.Gitsync.PrintDefaults()
		os.Exit(0)
	}

	// Проверка правильности заполнения флагов
	if err := flagSet.ValidateFlags(); err != nil {
		logger.GetLogger().Error("%v", err)
		os.Exit(0)
	}

	gitSync, err := gitsync.NewGitSync(flagSet.Gitsync, ctx)
	if err != nil {
		logger.GetLogger().Error("Error creating GitSync object: %v\n", err)
		return
	}

	gitRepo, err := git.NewGitRepository(flagSet.Gitsync)
	if err != nil {
		logger.GetLogger().Error("Error creating GitRepository object: %v\n", err)
		return
	}

	// Запускаем http-сервер
	http.StartServer(flagSet.Gitsync, ctx)

	// Запускаем периодическую синхронизацию в отдельной горутине
	go gitSync.Start(gitRepo)

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
	logger.GetLogger().Info("Received signal %s. Shutting down...\n", sig)
}
