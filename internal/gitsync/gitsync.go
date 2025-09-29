// Copyright 2025 Aleksey Dobshikov
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
	"fmt"
	"os"
	"time"

	"git-sync/internal/config"
	"git-sync/internal/handlers"
	"git-sync/internal/interfaces"
	"git-sync/internal/metrics"
	"git-sync/logger"
)

type GitSync struct {
	ctx            context.Context
	interval       time.Duration // Интервал обновления репозитория
	repositoryName string        // Name of the repository being synced
}

// NewGitSync creates a new GitSync instance with values from the provided config.
func NewGitSync(f *flag.FlagSet, cfg *config.Config, ctx context.Context, repositoryName string) (*GitSync, error) {
	// Since we now support multiple repositories, we need to get the first repository
	// for backward compatibility or create a way to specify which repository to use
	var firstRepoConfig config.RepositoryConfig
	var found bool

	// Get the first repository from the config
	for _, repoConfig := range cfg.Repositories {
		firstRepoConfig = repoConfig
		found = true
		break
	}

	if !found {
		return nil, fmt.Errorf("no repositories configured")
	}

	// Use the interval from the unified configuration
	interval := time.Duration(firstRepoConfig.Sync.Interval) * time.Second

	gitSync := &GitSync{
		ctx:            ctx,
		interval:       interval,
		repositoryName: repositoryName,
	}

	return gitSync, nil
}

func (gitsync *GitSync) Start(gitRepo interfaces.Gitter) {
	logger.Info("Sync: start synchronization\n")
	logger.Debug("Starting synchronization loop for repository: %s", gitsync.repositoryName)

	// Создаем тикер для периодической синхронизации
	ticker := time.NewTicker(gitsync.interval)
	defer ticker.Stop()

	for {
		select {
		case <-gitsync.ctx.Done():
			// Контекст отменен, выходим
			logger.Info("Sync: stop synchronization\n")
			logger.Debug("Stopping synchronization for repository: %s", gitsync.repositoryName)
			return

		case webhookData := <-handlers.WebhookCh:
			// Синхронизация по вебхуку
			logger.Debug("Webhook triggered synchronization for repository: %s (client IP: %s, User-Agent: %s)",
				gitsync.repositoryName, webhookData.ClientIP, webhookData.UserAgent)
			startTime := time.Now()
			err := gitsync.Sync(gitRepo)
			duration := time.Since(startTime).Seconds()
			metrics.ObserveSyncDuration("webhook", getResultFromError(err), duration)
			if err != nil {
				// Log the error and continue
				_, _ = fmt.Fprintf(os.Stderr, "Webhook sync error for repository %s: %v", gitsync.repositoryName, err)
				metrics.IncrementSyncError("webhook", gitsync.repositoryName)
			}
			logger.Info("Sync: webhook synchronization (client IP: %s)\n", webhookData.ClientIP)

		case <-ticker.C:
			// Синхронизация
			logger.Debug("Periodic synchronization for repository: %s", gitsync.repositoryName)
			startTime := time.Now()
			err := gitsync.Sync(gitRepo)
			duration := time.Since(startTime).Seconds()
			metrics.ObserveSyncDuration("periodic", getResultFromError(err), duration)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Periodic sync error for repository %s: %v", gitsync.repositoryName, err)
				metrics.IncrementSyncError("periodic", gitsync.repositoryName)
			}
		}
	}
}

func (gitsync *GitSync) Sync(gitRepo interfaces.Gitter) error {
	logger.Debug("Starting sync operation for repository: %s", gitsync.repositoryName)

	// Синхронизация локального репозитория
	err := gitRepo.Sync()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Sync error: %v", err)
		metrics.SyncTotalErrorCount.Inc()
		return err
	}

	// Получаем текущий коммит
	commit, err := gitRepo.Commit()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	} else {
		logger.Debug("Updating commit info for repository: %s, commit: %s", gitsync.repositoryName, commit.Hash)
		metrics.UpdateCommitInfo(commit, gitsync.repositoryName)
	}

	// Увеличиваем счетчик с общим количеством синхронизаций
	metrics.SyncTotalCount.Inc()

	// Обновляем метрику с информацией о синхронизируемом репозитории
	logger.Debug("Updating repository info metrics for: %s", gitsync.repositoryName)
	metrics.UpdateSyncRepoInfo(gitRepo.Options(), gitsync.repositoryName)

	if gitRepo.HasChanges() {
		// Увеличиваем счетчик синхронизаций с изменениями
		logger.Debug("Repository %s has changes, incrementing sync count", gitsync.repositoryName)
		metrics.SyncCount.Inc()
	} else {
		logger.Debug("Repository %s has no changes", gitsync.repositoryName)
	}

	logger.Debug("Completed sync operation for repository: %s", gitsync.repositoryName)
	return nil
}

func (gitsync *GitSync) Stop() error {
	return nil
}

func (gitsync *GitSync) GetCtx() context.Context {
	return gitsync.ctx
}

// getResultFromError converts an error to a result string for metrics
func getResultFromError(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
