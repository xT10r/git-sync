// Copyright 2024 Aleksey Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git_test

import (
	"git-sync/git"
	"git-sync/internal/constants"
	"git-sync/mock"
	"os"
	"path"
	"strings"
	"testing"
)

func TestNewGitRepositoryWithValidURL(t *testing.T) {
	// Создаем макет флагов для использования в тесте
	mockFlags := mock.Flags()
	mockFlags.String(constants.FlagRepoUrl, "https://gitlab.com/DavidGriffith/minipro.git", "URL of the repository")
	mockFlags.String(constants.FlagLocalPath, path.Join(os.TempDir(), "minipro"), "Local path for the repository")

	// Парсим флаги
	err := mockFlags.Parse(nil)
	if err != nil {
		t.Fatalf("Error parsing flags: %v", err)
	}

	// Пытаемся создать новый GitRepository с правильным URL
	gitRepo, err := git.NewGitRepository(mockFlags)
	if err != nil {
		t.Fatalf("Error initializing GitRepository: %v", err)
	}

	// Проверяем, что экземпляр создан успешно
	if gitRepo == nil {
		t.Fatal("Expected gitRepo to be non-nil")
	}

	// Проверяем, что текущий хеш коммита не является пустым
	commit, err := gitRepo.Commit()
	if err != nil {
		t.Fatalf("Error getting current commit: %v", err)
	}

	if commit.Hash == "" {
		t.Error("Expected commit hash to be non-empty")
	}
}

func TestNewGitRepositoryWithInvalidURL(t *testing.T) {
	// Создаем макет флагов для использования в тесте
	mockFlags := mock.Flags()
	mockFlags.String(constants.FlagRepoUrl, "http://invalid-url.com", "URL of the repository")
	mockFlags.String(constants.FlagLocalPath, path.Join(os.TempDir(), "invalid-repo"), "Local path for the repository")

	// Парсим флаги
	err := mockFlags.Parse(nil)
	if err != nil {
		t.Fatalf("Error parsing flags: %v", err)
	}

	// Пытаемся создать новый GitRepository с неверным URL
	_, err = git.NewGitRepository(mockFlags)
	if err == nil {
		t.Error("Expected error due to invalid repository URL, but got nil")
	} else {
		// Проверяем, что ошибка связана с невозможностью клонирования репозитория
		expectedError := "failed to clone repository"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain %q, but got %q", expectedError, err.Error())
		}
	}
}
