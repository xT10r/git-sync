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
	"git-sync/internal/constants"
	"git-sync/internal/handlers"
	"git-sync/internal/interfaces"
	"git-sync/internal/metrics"
	"git-sync/logger"
	"time"
)

type GitSync struct {
	ctx      context.Context
	interval time.Duration // Интервал обновления репозитория
}

// NewGitSync создает экземпляр SyncOptions с значениями по умолчанию.
func NewGitSync(f *flag.FlagSet, ctx context.Context) (*GitSync, error) {

	gitSync := &GitSync{
		ctx:      ctx,
		interval: f.Lookup(constants.FlagSyncInterval).Value.(flag.Getter).Get().(time.Duration),
	}

	return gitSync, nil
}

func (gitsync *GitSync) Start(gitRepo interfaces.Gitter) {

	logger.GetLogger().Info("Синхронизатор: начало синхронизации\n")

	// Создаем тикер для периодической синхронизации
	ticker := time.NewTicker(gitsync.interval)
	defer ticker.Stop()

	for {
		select {

		case <-gitsync.ctx.Done():
			// Контекст отменен, выходим
			logger.GetLogger().Info("Синхронизатор: завершение синхронизации\n")
			return

		case ip := <-handlers.WebhookCh:
			// Синхронизация по вебхуку
			_ = gitsync.Sync(gitRepo)
			logger.GetLogger().Info("Синхронизатор: синхронизация по вебхуку (client ip: %s)\n", ip)

		case <-ticker.C:
			// Синхронизация
			_ = gitsync.Sync(gitRepo)
		}
	}
}

func (gitsync *GitSync) Sync(gitRepo interfaces.Gitter) error {

	// Синхронизация локального репозитория
	err := gitRepo.Sync()
	if err != nil {
		logger.GetLogger().Error("Ошибка синхронизации: %v", err)
		metrics.SyncTotalErrorCount.Inc()
	}

	// Получаем текущий коммит
	commit, err := gitRepo.Commit()
	if err != nil {
		logger.GetLogger().Error("%v\n", err)
	} else {
		metrics.UpdateCommitInfo(commit)
	}

	// Увеличиваем счетчик с общим количеством синхронизаций
	metrics.SyncTotalCount.Inc()

	// Обновляем метрику с информацией о синхронизируемом репозитории
	metrics.UpdateSyncRepoInfo(gitRepo.Options())

	if gitRepo.HasChanges() {
		// Увеличиваем счетчик синхронизаций с изменениями
		metrics.SyncCount.Inc()
	}

	return nil
}

func (gitsync *GitSync) Stop() error {
	return nil
}

func (gitsync *GitSync) GetCtx() context.Context {
	return gitsync.ctx
}
