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

package mock

import (
	"flag"
	"git-sync/git"
	"git-sync/internal/constants"
	"time"
)

func Flags() *flag.FlagSet {
	mockFlags := flag.NewFlagSet("test", flag.ContinueOnError)
	mockFlags.Duration(constants.FlagSyncInterval, 30*time.Second, "Interval for synchronization")
	mockFlags.String(constants.FlagRepoBranch, "master", "Branch of the repository")
	mockFlags.String(constants.FlagRepoAuthUser, "user", "Repository authentication user")
	mockFlags.String(constants.FlagRepoAuthToken, "token", "Repository authentication token")
	return mockFlags
}

type Gitter struct {
	hasChanges bool
}

func (m *Gitter) Sync() error {
	return nil
}

func (m *Gitter) Options() *git.GitRepositoryOptions {
	return git.NewGitRepositoryOptions("http://example.com", "master", "/path/to/local/repo", "user", "token", "origin")
}

func (m *Gitter) HasChanges() bool {
	return m.hasChanges
}

func (m *Gitter) Commit() (*git.CommitInfo, error) {
	return &git.CommitInfo{Hash: "mockhash", Date: time.Now()}, nil
}

func (m *Gitter) CommitHash() string {
	return "mockhash"
}
