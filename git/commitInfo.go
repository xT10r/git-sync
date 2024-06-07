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

import (
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

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
