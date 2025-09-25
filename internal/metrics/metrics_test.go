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
	"git-sync/git"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestSyncCount tests the SyncCount metric
func TestSyncCount(t *testing.T) {
	// Store the initial value
	initial := testutil.ToFloat64(SyncCount)

	// Increment the counter
	SyncCount.Inc()

	// Check that the counter has the expected value
	expected := initial + 1.0
	actual := testutil.ToFloat64(SyncCount)
	if actual != expected {
		t.Errorf("Expected counter to be %f, got %f", expected, actual)
	}
}

// TestSyncTotalCount tests the SyncTotalCount metric
func TestSyncTotalCount(t *testing.T) {
	// Store the initial value
	initial := testutil.ToFloat64(SyncTotalCount)

	// Increment the counter
	SyncTotalCount.Inc()

	// Check that the counter has the expected value
	expected := initial + 1.0
	actual := testutil.ToFloat64(SyncTotalCount)
	if actual != expected {
		t.Errorf("Expected counter to be %f, got %f", expected, actual)
	}
}

// TestSyncTotalErrorCount tests the SyncTotalErrorCount metric
func TestSyncTotalErrorCount(t *testing.T) {
	// Store the initial value
	initial := testutil.ToFloat64(SyncTotalErrorCount)

	// Increment the counter
	SyncTotalErrorCount.Inc()

	// Check that the counter has the expected value
	expected := initial + 1.0
	actual := testutil.ToFloat64(SyncTotalErrorCount)
	if actual != expected {
		t.Errorf("Expected counter to be %f, got %f", expected, actual)
	}
}

// TestUpdateCommitInfo tests the UpdateCommitInfo function
func TestUpdateCommitInfo(t *testing.T) {
	// Create a test commit info
	commitInfo := &git.CommitInfo{
		Hash:    "abc123",
		Author:  "Test Author",
		Email:   "test@example.com",
		Date:    time.Now(),
		Message: "Test commit message",
	}

	// Call the function
	UpdateCommitInfo(commitInfo)

	// Check that the gauge has the expected value
	expected := 1.0
	actual := testutil.ToFloat64(CommitInfo)
	if actual != expected {
		t.Errorf("Expected gauge to be %f, got %f", expected, actual)
	}
}

// TestUpdateSyncRepoInfo tests the UpdateSyncRepoInfo function
func TestUpdateSyncRepoInfo(t *testing.T) {
	// Create a test git repository options
	options := git.NewGitRepositoryOptions(
		"https://github.com/test/repo",
		"main",
		"/tmp/test",
		"testuser",
		"testtoken",
		"origin",
	)

	// Call the function
	UpdateSyncRepoInfo(options)

	// Check that the gauge has the expected value
	expected := 1.0
	actual := testutil.ToFloat64(SyncRepoInfo)
	if actual != expected {
		t.Errorf("Expected gauge to be %f, got %f", expected, actual)
	}
}
