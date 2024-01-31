package app

import (
	"context"
	"fmt"
	"git-sync/git"
	"git-sync/internal/flags"
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
func NewGitSync(flags *flags.ConsoleFlags, ctx context.Context) (*GitSync, error) {

	gitRepo, err := git.NewGitRepository(flags)
	if err != nil {
		return nil, err
	}

	gitSync := &GitSync{
		mutex:    sync.Mutex{},
		ctx:      ctx,
		interval: flags.GetDuration("sync-interval"),
		git:      gitRepo,
	}

	return gitSync, nil
}

func (gitsync *GitSync) Start() {

	fmt.Println("Начало синхронизации")

	// итоговая команда для синхронизации
	// **
	//err := gitsync.git.SyncRepository()
	//if err != nil {
	//	fmt.Println("Ошибка синхронизации:", err)
	//}
	// **

	// Создаем тикер для периодической синхронизации
	ticker := time.NewTicker(gitsync.interval)
	defer ticker.Stop()

	for {
		select {
		case <-gitsync.ctx.Done():
			// Контекст отменен, завершаем функцию
			fmt.Println("Завершение синхронизациии")
			return
		case <-ticker.C:
			// Выполняем синхронизацию репозитория Git
			// err := syncRepository(options)
			err := gitsync.git.SyncRepository()
			if err != nil {
				fmt.Println("Ошибка синхронизации:", err)
			}
		}
	}
}

func (gs *GitSync) Stop() {
	return
}
