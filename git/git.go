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

package git

import (
	"flag"
	"fmt"
	"git-sync/internal/constants"
	"git-sync/logger"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	dateFormat string = "2006-01-02T15:04:05"
)

type GitRepository struct {
	mutex         sync.Mutex
	Options       *GitRepositoryOptions
	repository    *git.Repository
	currentCommit *CommitInfo
	hasChanges    bool
}

type GitRepositoryOptions struct {
	url        string // URL удаленного репозитория
	branch     string // Ветка репозитория
	path       string // Путь до локального репозитория
	user       string // Имя пользователя (для аутентификации)
	token      string // Токен (для аутентификации)
	originName string
}

type CommitInfo struct {
	Hash    string
	Date    time.Time
	Message string
	Author  string
	Email   string
	Reason  string
	commit  *object.Commit
	Changes []ChangeInfo
}

type ChangeInfo struct {
	ChangeType string
	FileName   string
	FromHash   string
	ToHash     string
}

// NewSyncOptions создает экземпляр SyncOptions с значениями по умолчанию.
func NewGitRepository(f *flag.FlagSet) (*GitRepository, error) {
	options := &GitRepositoryOptions{
		url:        f.Lookup(constants.FlagRepoUrl).Value.(flag.Getter).Get().(string),
		branch:     f.Lookup(constants.FlagRepoBranch).Value.(flag.Getter).Get().(string),
		path:       f.Lookup(constants.FlagLocalPath).Value.(flag.Getter).Get().(string),
		user:       f.Lookup(constants.FlagRepoAuthUser).Value.(flag.Getter).Get().(string),
		token:      f.Lookup(constants.FlagRepoAuthToken).Value.(flag.Getter).Get().(string),
		originName: "origin",
	}

	gitRepository := &GitRepository{
		mutex:         sync.Mutex{},
		Options:       options,
		repository:    nil,
		currentCommit: nil,
	}

	// Получаем репозиторий
	err := gitRepository.cloneOpenRepo()
	if err != nil {
		return nil, err
	}

	// Записываем текущий коммит
	err = gitRepository.storeCurrentCommit("init")
	if err != nil {
		return nil, err
	}

	return gitRepository, nil
}

// NewCommitInfo создает новый объект CommitInfo на основе git.Commit.
func NewCommitInfo(commit *object.Commit) *CommitInfo {

	return &CommitInfo{
		Hash:    commit.Hash.String(),
		Date:    commit.Committer.When,
		Message: strings.TrimSpace(commit.Message),
		Author:  commit.Author.Name,
		Email:   commit.Author.Email,
		commit:  commit,
		Changes: []ChangeInfo{},
	}
}

// AddChange добавляет информацию об изменении файла в CommitInfo.
func (ci *CommitInfo) AddChange(changeType, fileName, fromHash, toHash string) {
	change := ChangeInfo{
		ChangeType: changeType,
		FileName:   fileName,
		FromHash:   fromHash,
		ToHash:     toHash,
	}
	ci.Changes = append(ci.Changes, change)
}

// Синхронизация локального и удаленного репозитория
func (gitRepo *GitRepository) SyncRepository() error {

	var err error

	gitRepo.resetChangesFlag()

	// Открываем либо клонируем удаленный репозиторий
	err = gitRepo.cloneOpenRepo() // тут не фиксируются изменения
	if err != nil {
		return err
	}

	// Принимаем изменения из удаленного репозитория
	err = gitRepo.fetchRepo()
	if err != nil {
		return err
	}

	// Проверяем изменения между удаленным и локальным репозиториями
	err = gitRepo.compareCommitTrees()
	if err != nil {
		return err
	}

	// Проверяем наличие изменений в структуре локального репозитория
	err = gitRepo.compareFiles()
	if err != nil {
		return err
	}

	return nil
}

// cloneRepo клонирует репозиторий
func (gitRepo *GitRepository) cloneRepo() error {

	repoDir := gitRepo.Options.path
	gitDir := filepath.Join(gitRepo.Options.path, ".git")

	needToClone := false

	// Проверка существования каталога repoDir
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		needToClone = true
	}

	// Проверка существования каталога gitDir
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		needToClone = true
	}

	if !needToClone {
		return nil
	}

	repository, err := git.PlainClone(gitRepo.Options.path, false, &git.CloneOptions{
		URL: gitRepo.Options.url, // URL удаленного репозитория
		Auth: &http.TokenAuth{
			Token: gitRepo.Options.token, // Токен для аутентификации
		},
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	gitRepo.repository = repository

	gitRepo.setChangesFlag(true)
	gitRepo.storeCurrentCommit("local")
	err = gitRepo.showCommitMessage()
	if err != nil {
		return err
	}

	return nil
}

// openRepo открывает репозиторий
func (gitRepo *GitRepository) openRepo() error {
	// Открываем репозиторий
	repository, err := git.PlainOpen(gitRepo.Options.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %v", err)
	}
	gitRepo.repository = repository
	return nil
}

// cloneOpenRepo клонирует или открывает репозиторий
func (gitRepo *GitRepository) cloneOpenRepo() error {

	if err := gitRepo.cloneRepo(); err != nil {
		return err
	}

	if err := gitRepo.openRepo(); err != nil {
		return err
	}

	return nil

}

// fetchRepo обновляет локальный репозиторий данными из удаленного,
// используя принудительный fetch. Возвращает ошибку в случае
// возникновения проблем при выполнении операции fetch. Если репозиторий
// уже актуален и не требует обновления, возвращает nil без ошибки.
func (gitRepo *GitRepository) fetchRepo() error {

	remote, err := gitRepo.repository.Remote(gitRepo.Options.originName)
	if err != nil {
		return fmt.Errorf("failed to get remote: %v", err)
	}

	// Выполняем fetch для получения обновлений из удаленного репозитория
	err = remote.Fetch(&git.FetchOptions{
		Auth: &http.TokenAuth{
			Token: gitRepo.Options.token,
		},
		Force: true,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil // Репозиторий уже актуален, не возвращаем ошибку
	}

	if err != nil {
		return fmt.Errorf("failed to fetch remote: %v", err)
	}

	return nil
}

// pullRepo выполняет операцию Pull для текущего репозитория с использованием указанных параметров.
// Если force установлен в true, операция Pull будет выполнена с флагом Force для принудительного объединения изменений.
// В случае ошибки при выполнении операции Pull, функция возвращает ошибку.
func (gitRepo *GitRepository) pullRepo(force bool) error {

	// Получаем объект Worktree из текущего репозитория
	wt, err := gitRepo.getRepoWorktree()
	if err != nil {
		return err
	}

	// Выполняем операцию Pull с указанными параметрами
	err = wt.Pull(&git.PullOptions{
		RemoteURL:  gitRepo.Options.url,
		RemoteName: gitRepo.Options.originName,
		Force:      force,
	})

	// Обрабатываем случаи ошибок
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull changes: %v", err)
	}

	return nil
}

// resetRepo сбрасывает все изменения в локальном репозитории.
// Эта функция выполняет жесткий сброс (hard reset), удаляя все
// неотслеживаемые файлы и отменяя все изменения.
// Возвращает ошибку в случае возникновения проблем при сбросе.
func (gitRepo *GitRepository) resetRepo() error {

	// Получаем объект Worktree из текущего репозитория
	wt, err := gitRepo.getRepoWorktree()
	if err != nil {
		return err
	}

	// Выполняем жесткий сброс
	err = wt.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset changes: %v", err)
	}
	return nil
}

// getRepoWorktree возвращает указатель на объект Worktree для текущего репозитория.
// Если произошла ошибка при получении Worktree, функция возвращает nil и ошибку.
func (gitRepo *GitRepository) getRepoWorktree() (*git.Worktree, error) {
	// Получаем объект Worktree из репозитория
	wt, err := gitRepo.repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %v", err)
	}
	return wt, nil
}

// storeCurrentCommit сохраняет информацию о текущем коммите внутри GitRepository.
// Параметр "reason" представляет собой причину "сохранения" коммита.
// Функция выполняет блокировку мьютекса GitRepository для безопасной работы с данными.
func (gitRepo *GitRepository) storeCurrentCommit(reason string) error {

	var err error

	// Получаем текущий локальный коммит
	commit, err := gitRepo.getCommit(false)
	if err != nil {
		return err
	}

	// Блокируем мьютекс для безопасной работы с данными GitRepository
	gitRepo.mutex.Lock()

	// Создаем новый объект CommitInfo на основе текущего коммита
	gitRepo.currentCommit = NewCommitInfo(commit)
	gitRepo.currentCommit.Reason = reason

	// Снимаем блокировку мьютекса
	gitRepo.mutex.Unlock()

	return nil
}

// getLastCommits получает последний локальный или удаленный коммит репозитория
// в зависимости от указанного сокращенного названия ветки и флага isRemote.
func (gitRepo *GitRepository) getCommit(isRemote bool) (*object.Commit, error) {

	var ref plumbing.ReferenceName

	// Если требуется получить удаленный коммит
	if isRemote {
		remote, err := gitRepo.repository.Remote(gitRepo.Options.originName)
		if err != nil {
			return nil, fmt.Errorf("failed to get remote: %v", err)
		}

		// Формируем путь к удаленной ветке на основе указанного имени
		ref = plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remote.Config().Name, gitRepo.Options.branch))
	} else {
		// Получаем последний коммит на локальной ветке
		localRef, err := gitRepo.repository.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD reference: %v", err)
		}
		ref = plumbing.ReferenceName(localRef.Name())
	}

	branchRef, err := gitRepo.repository.Reference(ref, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference: %v", err)
	}

	commit, err := gitRepo.repository.CommitObject(branchRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %v", err)
	}

	return commit, nil
}

// Проверяем наличие изменений между локальным и удаленным коммитами
func (gitRepo *GitRepository) compareCommitTrees() error {

	// Получаем последний коммит локального репозитория
	localCommit, err := gitRepo.getCommit(false)
	if err != nil {
		return err
	}

	// Получаем последний коммит удаленного репозитория
	remoteCommit, err := gitRepo.getCommit(true)
	if err != nil {
		return err
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

	// Изменения найдены
	if diff.Len() > 0 {

		gitRepo.setChangesFlag(true)

		// принимаем изменения из удаленного репозитория (git pull --force)
		err = gitRepo.pullRepo(true)
		if err != nil {
			return err
		}

		gitRepo.storeCurrentCommit("remote")

		err = gitRepo.showCommitMessage()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// Проверяем наличие изменений в файлах лольного репозитория
func (gitRepo *GitRepository) compareFiles() error {

	wt, err := gitRepo.getRepoWorktree()
	if err != nil {
		return err
	}

	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %v", err)
	}

	if !status.IsClean() {

		gitRepo.setChangesFlag(true)

		// fmt.Println("Найдены изменения в локальном репозитории:")
		// changedFiles := strings.Split(status.String(), "\n")

		err := gitRepo.resetRepo()
		if err != nil {
			return err
		}

		gitRepo.storeCurrentCommit("local")

		// выводим сообщение в лог
		err = gitRepo.showCommitMessage()
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (gitRepo *GitRepository) resetChangesFlag() {
	gitRepo.mutex.Lock()
	defer gitRepo.mutex.Unlock()
	gitRepo.hasChanges = false
}

func (gitRepo *GitRepository) setChangesFlag(hasChanges bool) {
	if hasChanges {
		gitRepo.mutex.Lock()
		defer gitRepo.mutex.Unlock()
		gitRepo.hasChanges = hasChanges
	}
}

func (gitRepo *GitRepository) GetChangesFlag() bool {
	gitRepo.mutex.Lock()
	defer gitRepo.mutex.Unlock()
	return gitRepo.hasChanges
}

// Получение текущего коммита
func (gr *GitRepository) GetCurrentCommit() (*CommitInfo, error) {

	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	if gr.currentCommit == nil {
		return nil, fmt.Errorf("отсутствуют текущиий коммит")
	}

	return gr.currentCommit, nil
}

func (gitRepo *GitRepository) GetCurrentCommitHash() string {
	return gitRepo.currentCommit.Hash
}

// showCommitMessage выводит информацию о последнем сохраненном коммите.
// Если коммит отсутствует, функция возвращает ошибку.
func (gitRepo *GitRepository) showCommitMessage() error {

	// Блокировка мьютекса для безопасного доступа к lastCommit
	gitRepo.mutex.Lock()
	defer gitRepo.mutex.Unlock()

	// Проверка наличия сохраненного коммита
	if gitRepo.currentCommit == nil {
		return fmt.Errorf("не удалось получить сохраненный коммит")
	}

	// Получение информации о последнем коммите
	reason := gitRepo.currentCommit.Reason
	commitHash := gitRepo.currentCommit.Hash
	commitDate := gitRepo.currentCommit.Date.Format(dateFormat)
	commitMessage := strings.TrimSpace(gitRepo.currentCommit.Message)
	authorName := gitRepo.currentCommit.Author
	authorEmail := gitRepo.currentCommit.Email

	// Вывод информации о коммите в лог
	logMessage := fmt.Sprintf("%s %s %s (%s) %s %s\n", reason, commitHash, authorName, authorEmail, commitDate, commitMessage)
	logger.GetLogger().Info(logMessage)

	return nil

}

func (gitRepoOptions *GitRepositoryOptions) Url() string {
	return gitRepoOptions.url
}

func (gitRepoOptions *GitRepositoryOptions) Branch() string {
	return gitRepoOptions.branch
}
