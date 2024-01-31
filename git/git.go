/*
Основная логика работы с репозиторием будет в методах этой структуры Repo.
Пакет git экспортирует интерфейс для работы с Git, скрывая внутреннюю реализацию.
*/

package git

import (
	"context"
	"fmt"
	"git-sync/internal/flags"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitCommits struct {
	local  *object.Commit
	remote *object.Commit
}

type GitRepository struct {
	options     *GitRepositoryOptions
	repository  *git.Repository
	lastCommits *GitCommits
}

type GitRepositoryOptions struct {
	url        string // URL удаленного репозитория
	branch     string // Ветка репозитория
	path       string // Путь до локального репозитория
	user       string // Имя пользователя (для аутентификации)
	token      string // Токен (для аутентификации)
	originName string
}

// NewSyncOptions создает экземпляр SyncOptions с значениями по умолчанию.
func NewGitRepository(flags *flags.ConsoleFlags) (*GitRepository, error) {
	options := &GitRepositoryOptions{
		url:        flags.GetString("remote-url"),
		branch:     flags.GetString("remote-branch"),
		path:       flags.GetString("local-path"),
		user:       flags.GetString("git-user"),
		token:      flags.GetString("git-token"),
		originName: "origin",
	}

	gitRepository := &GitRepository{
		options:    options,
		repository: nil,
		lastCommits: &GitCommits{
			local:  nil,
			remote: nil,
		},
	}

	return gitRepository, nil
}

// SyncOptions содержит опции для синхронизации репозитория Git.
type SyncOptions struct {
	RemoteURL    string        // URL удаленного репозитория
	LocalPath    string        // Путь к локальной папке
	Username     string        // Имя пользователя (для аутентификации)
	Token        string        // Токен (для аутентификации)
	RemoteBranch string        // Ветка репозитория
	Interval     time.Duration // Интервал обновления репозитория
}

// NewSyncOptions создает экземпляр SyncOptions с значениями по умолчанию.
func NewSyncOptions(flags *flags.ConsoleFlags) (*SyncOptions, error) {
	options := &SyncOptions{
		Interval:     flags.GetDuration("sync-interval"),
		RemoteURL:    flags.GetString("remote-url"),
		RemoteBranch: flags.GetString("remote-branch"),
		LocalPath:    flags.GetString("local-path"),
		Username:     flags.GetString("git-user"),
		Token:        flags.GetString("git-token"),
	}

	return options, nil
}

// StartSync запускает периодическую синхронизацию репозитория Git с заданными параметрами.
// Функция будет выполняться до тех пор, пока контекст не будет отменен.
func StartSync(ctx context.Context, options *SyncOptions) {

	fmt.Println("Начало синхронизации")

	// Создаем тикер для периодической синхронизации
	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Контекст отменен, завершаем функцию
			fmt.Println("Завершение синхронизациии")
			return
		case <-ticker.C:
			// Выполняем синхронизацию репозитория Git
			err := syncRepository(options)
			if err != nil {
				fmt.Println("Ошибка синхронизации:", err)
			}
		}
	}
}

func (gr *GitRepository) SyncRepository() error {

	var err error

	// Открываем либо клонируем удаленный репозиторий
	err = gr.cloneOpenRepo()
	if err != nil {
		return err
	}

	// Принимаем изменения из удаленного репозитория
	err = gr.fetchRepo()
	if err != nil {
		return err
	}

	// Получаем последние коммиты (локальный и удаленный)
	err = gr.getLastCommits()
	if err != nil {
		return err
	}

	var commitsDiff bool
	commitsDiff, err = gr.checkChangesBetweenCommits()
	if err != nil {
		return err
	}

	var filesDiff bool
	filesDiff, err = gr.checkChangesBetweenFiles()
	if err != nil {
		return err
	}

	if commitsDiff || filesDiff {
		fmt.Printf("asd")
	}

	return nil
}

// Это была рабочая функция. Решил её разбить на более мелкие... (см. выше)
// SyncRepository выполняет синхронизацию репозитория Git с заданными параметрами.
func syncRepository(options *SyncOptions) error {

	remoteName := "origin"
	changesFound := false

	// Открываем или клонируем репозиторий
	repo, err := cloneRepoOld(options)
	if err != nil {
		return err
	}

	// Получаем ссылку HEAD локального репозитория
	localRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	// Получаем ссылку с именем текущей ветки из HEAD
	localBranchRef, err := repo.Reference(localRef.Name(), true)
	if err != nil {
		return fmt.Errorf("failed to get branch reference: %v", err)
	}

	// Получаем удаленный репозиторий (origin)
	remote, err := repo.Remote(remoteName)
	if err != nil {
		return fmt.Errorf("failed to get remote: %v", err)
	}

	// Выполняем fetch для получения обновлений из удаленного репозитория
	err = remote.Fetch(&git.FetchOptions{
		Auth: &http.TokenAuth{
			Token: options.Token,
		},
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch remote: %v", err)
	}

	// Формируем путь к удаленной ветке на основе имени текущей ветки локального репозитория
	remoteBranchPath := fmt.Sprintf("refs/remotes/%s/%s", remote.Config().Name, localBranchRef.Name().Short())

	// Получаем последний коммит на удаленном репозитории
	remoteRef, err := repo.Reference(plumbing.ReferenceName(remoteBranchPath), true)
	if err != nil {
		return fmt.Errorf("failed to get remote reference: %v", err)
	}

	localCommit, err := repo.CommitObject(localRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit object: %v", err)
	}

	remoteCommit, err := repo.CommitObject(remoteRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get remote commit object: %v", err)
	}

	// Получаем деревья для сравнения
	localTree, err := localCommit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get local tree: %v", err)
	}
	remoteTree, err := remoteCommit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get remote tree: %v", err)
	}

	// Сравниваем локальный и удаленный коммиты
	diff, err := localTree.Diff(remoteTree)
	if err != nil {
		return fmt.Errorf("failed to get diff: %v", err)
	}

	if diff.Len() > 0 {
		fmt.Println("Найдены изменения в удаленном репозитории:")
		for i := 0; i < diff.Len(); i++ {
			delta := diff[i]
			from, to, err := delta.Files()
			if err != nil {
				return fmt.Errorf("failed to get diff files: %v", err)
			}
			fmt.Printf("%s -> %s (%s -> %s)\n", from.Name, to.Name, from.Hash.String(), to.Hash.String())
		}
		changesFound = true
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %v", err)
	}

	changesFound, err = resetRepoOld(wt, status)
	if err != nil {
		return err
	}

	// При наличии изменений в репо выводим информацию о последнем коммите
	if changesFound {

		dateFormat := "2006-01-02 15:04:05"

		commitHash := localCommit.Hash
		commitDate := localCommit.Committer.When.Format(dateFormat)
		commitMessage := localCommit.Message

		// Pull с флагом Force для принудительного объединения изменений
		if diff.Len() > 0 {
			err = wt.Pull(&git.PullOptions{
				RemoteURL:  options.RemoteURL,
				RemoteName: remote.Config().Name,
				Force:      true,
			})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return fmt.Errorf("failed to pull changes: %v", err)
			}
			commitHash = remoteCommit.Hash
			commitDate = remoteCommit.Committer.When.Format(dateFormat)
			commitMessage = remoteCommit.Message
		}

		currentTime := time.Now().Format(dateFormat)
		fmt.Printf("[%s]\nКоммит: %s\nДата: %s\nСообщение: %s\n", currentTime, commitHash, commitDate, commitMessage)
	}

	return nil
}

// Открытие или клонирование репозитория
func (gr *GitRepository) cloneOpenRepo() error {

	var err error

	// Открываем или клонируем репозиторий
	gr.repository, err = git.PlainClone(gr.options.path, false, &git.CloneOptions{
		URL: gr.options.url, // URL удаленного репозитория
		Auth: &http.TokenAuth{
			Token: gr.options.token, // Токен для аутентификации
		},
	})
	if err != nil {
		// Если репозиторий уже существует, открываем его
		if err == git.ErrRepositoryAlreadyExists {
			gr.repository, err = git.PlainOpen(gr.options.path)
			if err != nil {
				return fmt.Errorf("failed to open repository: %v", err)
			}
		} else {
			return fmt.Errorf("failed to clone repository: %v", err)
		}
	}

	return nil
}

// Принимаем последние изменения из удаленного репозитория
func (gr *GitRepository) fetchRepo() error {

	remote, err := gr.repository.Remote(gr.options.originName)
	if err != nil {
		return fmt.Errorf("failed to get remote: %v", err)
	}

	// Выполняем fetch для получения обновлений из удаленного репозитория
	err = remote.Fetch(&git.FetchOptions{
		Auth: &http.TokenAuth{
			Token: gr.options.token,
		},
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch remote: %v", err)
	}
	return nil
}

// Получаем послдение локальный и удаленный коммиты
func (gr *GitRepository) getLastCommits() error {

	var err error
	gr.lastCommits.local, err = gr.getLastCommit(false)
	if err != nil {
		return err
	}
	gr.lastCommits.remote, err = gr.getLastCommit(true)
	if err != nil {
		return err
	}
	return nil
}

// Проверяет наличие изменений между локальным и удаленным коммитами
func (gr *GitRepository) checkChangesBetweenCommits() (bool, error) {

	// Получаем деревья для сравнения
	localTree, err := gr.lastCommits.local.Tree()
	if err != nil {
		return false, fmt.Errorf("failed to get local tree: %v", err)
	}
	remoteTree, err := gr.lastCommits.remote.Tree()
	if err != nil {
		return false, fmt.Errorf("failed to get remote tree: %v", err)
	}

	// Сравниваем локальный и удаленный коммиты
	diff, err := localTree.Diff(remoteTree)
	if err != nil {
		return false, fmt.Errorf("failed to get diff: %v", err)
	}

	if diff.Len() > 0 {
		fmt.Println("Найдены изменения в удаленном репозитории:")
		for i := 0; i < diff.Len(); i++ {
			delta := diff[i]
			from, to, err := delta.Files()
			if err != nil {
				return false, fmt.Errorf("failed to get diff files: %v", err)
			}
			fmt.Printf("%s -> %s (%s -> %s)\n", from.Name, to.Name, from.Hash.String(), to.Hash.String())
		}
		return true, nil
	}

	return false, nil
}

func (gr *GitRepository) checkChangesBetweenFiles() (bool, error) {

	wt, err := gr.repository.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %v", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %v", err)
	}

	if status.IsClean() {
		return false, nil
	}

	// Сброс изменений в локальном репозитории
	err = wt.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		return false, fmt.Errorf("failed to reset changes: %v", err)
	}

	return true, nil
}

// getLastCommits получает последний локальный или удаленный коммит репозитория
// в зависимости от указанного сокращенного названия ветки и флага isRemote.
func (gr *GitRepository) getLastCommit(isRemote bool) (*object.Commit, error) {

	var ref plumbing.ReferenceName

	// Если требуется получить удаленный коммит
	if isRemote {
		remote, err := gr.repository.Remote(gr.options.originName)
		if err != nil {
			return nil, fmt.Errorf("failed to get remote: %v", err)
		}

		// Формируем путь к удаленной ветке на основе указанного имени
		ref = plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remote.Config().Name, gr.options.branch))
	} else {
		// Получаем последний коммит на локальной ветке
		ref = plumbing.ReferenceName(gr.options.branch)
	}

	branchRef, err := gr.repository.Reference(ref, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference: %v", err)
	}

	commit, err := gr.repository.CommitObject(branchRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %v", err)
	}

	return commit, nil
}

func cloneRepoOld(options *SyncOptions) (*git.Repository, error) {

	// Открываем или клонируем репозиторий
	repo, err := git.PlainClone(options.LocalPath, false, &git.CloneOptions{
		URL: options.RemoteURL, // URL удаленного репозитория
		Auth: &http.TokenAuth{
			Token: options.Token, // Токен для аутентификации
		},
	})
	if err != nil {
		// Если репозиторий уже существует, открываем его
		if err == git.ErrRepositoryAlreadyExists {
			repo, err = git.PlainOpen(options.LocalPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open repository: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to clone repository: %v", err)
		}
	}
	return repo, nil
}

func resetRepoOld(wt *git.Worktree, status git.Status) (bool, error) {

	if status.IsClean() {
		return false, nil
	}

	// Настройки для сброса изменений в локальном репозитории
	resetOptions := &git.ResetOptions{
		Mode: git.HardReset,
	}

	// Сброс изменений в локальном репозитории
	err := wt.Reset(resetOptions)
	if err != nil {
		return false, fmt.Errorf("failed to reset changes: %v", err)
	}

	return true, nil
}
