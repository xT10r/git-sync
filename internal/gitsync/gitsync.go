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

package gitsync

import (
	"context"
	"flag"
	"git-sync/git"
	"git-sync/internal/constants"
	"git-sync/internal/handlers"
	"git-sync/internal/metrics"
	"git-sync/logger"
	"sync"
	"time"
)

type GitSync struct {
	mutex    sync.Mutex
	ctx      context.Context
	interval time.Duration      // Интервал обновления репозитория
	git      *git.GitRepository // Данные синхронизируемого репозитория
}

// NewGitSync создает экземпляр SyncOptions с значениями по умолчанию.
func NewGitSync(f *flag.FlagSet, ctx context.Context) (*GitSync, error) {

	gitRepo, err := git.NewGitRepository(f)
	if err != nil {
		return nil, err
	}

	gitSync := &GitSync{
		mutex:    sync.Mutex{},
		ctx:      ctx,
		interval: f.Lookup(constants.FlagSyncInterval).Value.(flag.Getter).Get().(time.Duration),
		git:      gitRepo,
	}

	return gitSync, nil
}

func (gitsync *GitSync) Start() {

	logger.GetLogger().Info("Начало синхронизации\n")

	// Создаем тикер для периодической синхронизации
	ticker := time.NewTicker(gitsync.interval)
	defer ticker.Stop()

	for {
		select {

		case <-gitsync.ctx.Done():
			// Контекст отменен, выходим
			logger.GetLogger().Info("Завершение синхронизации\n")
			return

		case ip := <-handlers.WebhookCh:
			// Синхронизация по вебхуку
			_ = gitsync.sync()
			logger.GetLogger().Info("Синхронизации по вебхуку (client ip: %s)\n", ip)

		case <-ticker.C:
			// Синхронизация
			_ = gitsync.sync()
		}
	}
}

func (gitsync *GitSync) sync() error {
	// Синхронизация локального репозитория
	err := gitsync.git.SyncRepository()
	if err != nil {
		logger.GetLogger().Error("Ошибка синхронизации: %v", err)
		metrics.SyncTotalErrorCount.Inc()
	}

	// Получаем текущий коммит
	commit, err := gitsync.git.GetCurrentCommit()
	if err != nil {
		logger.GetLogger().Error("%v\n", err)
	} else {
		metrics.UpdateCommitInfo(commit)
	}

	// Увеличиваем счетчик с общим количеством синхронизаций
	metrics.SyncTotalCount.Inc()

	// Обновляем метрику с информацией о синхронизируемом репозитории
	metrics.UpdateSyncRepoInfo(gitsync.git.Options)

	if gitsync.git.GetChangesFlag() {
		// Увеличиваем счетчик синхронизаций с изменениями
		metrics.SyncCount.Inc()
	}

	return nil
}

func (gitsync *GitSync) Stop() error {
	return nil
}
