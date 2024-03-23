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

package metrics

import (
	"fmt"
	"git-sync/git"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	SyncCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "git_sync_sync_count",
			Help: "Total number of synchronizations with changes",
		},
	)
	SyncTotalCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "git_sync_sync_total_count",
			Help: "Total number of synchronizations",
		},
	)

	SyncTotalErrorCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "git_sync_sync_total_error_count",
			Help: "Total number of synchronization errors",
		},
	)

	SyncRepoInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "git_sync_repo_info",
			Help: "Information about the synchronized repository",
		},
		[]string{"repository", "branch"},
	)

	CommitInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "git_sync_commit_info",
		Help: "Information about the latest commit.",
	}, []string{"hash", "author", "email", "date", "message"})
)

func init() {
	prometheus.MustRegister(SyncCount)
	prometheus.MustRegister(SyncRepoInfo)
	prometheus.MustRegister(SyncTotalCount)
	prometheus.MustRegister(SyncTotalErrorCount)
	prometheus.MustRegister(CommitInfo)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func UpdateCommitInfo(gci *git.CommitInfo) {
	unixTimestamp := gci.Date.UnixNano() / int64(time.Millisecond)
	CommitInfo.Reset()
	CommitInfo.WithLabelValues(gci.Hash, gci.Author, gci.Email, fmt.Sprintf("%d", unixTimestamp), gci.Message).Set(1)
}

func UpdateSyncRepoInfo(gro *git.GitRepositoryOptions) {
	SyncRepoInfo.Reset()
	SyncRepoInfo.WithLabelValues(gro.Url(), gro.Branch()).Set(1)
}
