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

package git

type GitRepositoryOptions struct {
	url        string // URL удаленного репозитория
	branch     string // Ветка репозитория
	path       string // Путь до локального репозитория
	user       string // Имя пользователя (для аутентификации)
	token      string // Токен (для аутентификации)
	originName string // имя удаленного репозитория
}

func NewGitRepositoryOptions(url, branch, path, user, token, originName string) *GitRepositoryOptions {
	return &GitRepositoryOptions{
		url:        url,
		branch:     branch,
		path:       path,
		user:       user,
		token:      token,
		originName: originName,
	}
}

func (gitRepoOptions *GitRepositoryOptions) Url() string {
	return gitRepoOptions.url
}

func (gitRepoOptions *GitRepositoryOptions) Branch() string {
	return gitRepoOptions.branch
}
