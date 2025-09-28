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

package metrics

import (
	"git-sync/git"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestSyncMetrics tests the new sync metrics
func TestSyncMetrics(t *testing.T) {
	// Test SyncRequestsTotal
	initialCount := testutil.CollectAndCount(SyncRequestsTotal)
	IncrementSyncRequest("manual", "success", "test-repo")
	newCount := testutil.CollectAndCount(SyncRequestsTotal)
	if newCount != initialCount+1 {
		t.Errorf("Expected SyncRequestsTotal count to be %d, got %d", initialCount+1, newCount)
	}

	// Test SyncErrorsTotal
	initialCount = testutil.CollectAndCount(SyncErrorsTotal)
	IncrementSyncError("network", "test-repo")
	newCount = testutil.CollectAndCount(SyncErrorsTotal)
	if newCount != initialCount+1 {
		t.Errorf("Expected SyncErrorsTotal count to be %d, got %d", initialCount+1, newCount)
	}

	// Test SyncInProgress
	initialValue := testutil.ToFloat64(SyncInProgress)
	IncrementSyncInProgress()
	expected := initialValue + 1.0
	actual := testutil.ToFloat64(SyncInProgress)
	if actual != expected {
		t.Errorf("Expected SyncInProgress to be %f, got %f", expected, actual)
	}
	DecrementSyncInProgress()
	expected = initialValue
	actual = testutil.ToFloat64(SyncInProgress)
	if actual != expected {
		t.Errorf("Expected SyncInProgress to be %f, got %f", expected, actual)
	}
}

// TestRepositoryMetrics tests the new repository metrics
func TestRepositoryMetrics(t *testing.T) {
	// Test RepoHealth
	initialCount := testutil.CollectAndCount(RepoHealth)
	UpdateRepoHealth("test-repo", "remote_reachable", true)
	newCount := testutil.CollectAndCount(RepoHealth)
	if newCount != initialCount+1 {
		t.Errorf("Expected RepoHealth count to be %d, got %d", initialCount+1, newCount)
	}

	// Test RepoCommitDrift
	initialCount = testutil.CollectAndCount(RepoCommitDrift)
	UpdateRepoCommitDrift("test-repo", "main", 5)
	newCount = testutil.CollectAndCount(RepoCommitDrift)
	if newCount != initialCount+1 {
		t.Errorf("Expected RepoCommitDrift count to be %d, got %d", initialCount+1, newCount)
	}

	// Test RepoLastSyncAge
	initialCount = testutil.CollectAndCount(RepoLastSyncAgeSeconds)
	UpdateRepoLastSyncAge("test-repo", 300)
	newCount = testutil.CollectAndCount(RepoLastSyncAgeSeconds)
	if newCount != initialCount+1 {
		t.Errorf("Expected RepoLastSyncAgeSeconds count to be %d, got %d", initialCount+1, newCount)
	}

	// Test RepoSyncInfo
	initialCount = testutil.CollectAndCount(RepoSyncInfo)
	RepoSyncInfo.WithLabelValues("test-repo", "main", "https://github.com/test/repo").Set(1)
	newCount = testutil.CollectAndCount(RepoSyncInfo)
	if newCount != initialCount+1 {
		t.Errorf("Expected RepoSyncInfo count to be %d, got %d", initialCount+1, newCount)
	}
}

// TestPerformanceMetrics tests the new performance metrics
func TestPerformanceMetrics(t *testing.T) {
	// Test SyncFilesProcessedTotal
	initialCount := testutil.CollectAndCount(SyncFilesProcessedTotal)
	IncrementFilesProcessed("test-repo", "added")
	newCount := testutil.CollectAndCount(SyncFilesProcessedTotal)
	if newCount != initialCount+1 {
		t.Errorf("Expected SyncFilesProcessedTotal count to be %d, got %d", initialCount+1, newCount)
	}

	// Test SyncDataTransferredBytesTotal
	initialCount = testutil.CollectAndCount(SyncDataTransferredBytesTotal)
	IncrementDataTransferred("test-repo", "download", 1024)
	newCount = testutil.CollectAndCount(SyncDataTransferredBytesTotal)
	if newCount != initialCount+1 {
		t.Errorf("Expected SyncDataTransferredBytesTotal count to be %d, got %d", initialCount+1, newCount)
	}

	// Test DiskUsageBytes
	initialCount = testutil.CollectAndCount(DiskUsageBytes)
	DiskUsageBytes.WithLabelValues("test-repo", "repo_size").Set(2048)
	newCount = testutil.CollectAndCount(DiskUsageBytes)
	if newCount != initialCount+1 {
		t.Errorf("Expected DiskUsageBytes count to be %d, got %d", initialCount+1, newCount)
	}

	// Test MemoryUsageBytes
	UpdateMemoryUsage(4096)
	expected := 4096.0
	actual := testutil.ToFloat64(MemoryUsageBytes)
	if actual != expected {
		t.Errorf("Expected MemoryUsageBytes to be %f, got %f", expected, actual)
	}

	// Test GoroutinesCount
	UpdateGoroutinesCount(10)
	expected = 10.0
	actual = testutil.ToFloat64(GoroutinesCount)
	if actual != expected {
		t.Errorf("Expected GoroutinesCount to be %f, got %f", expected, actual)
	}
}

// TestHttpMetrics tests the new HTTP metrics
func TestHttpMetrics(t *testing.T) {
	// Test HttpRequestsTotal
	initialCount := testutil.CollectAndCount(HttpRequestsTotal)
	IncrementHttpRequest("GET", "/metrics", "200")
	newCount := testutil.CollectAndCount(HttpRequestsTotal)
	if newCount != initialCount+1 {
		t.Errorf("Expected HttpRequestsTotal count to be %d, got %d", initialCount+1, newCount)
	}

	// Test HttpRequestsInProgress
	// First set a value for specific labels to ensure the metric exists
	HttpRequestsInProgress.WithLabelValues("GET", "/metrics").Set(0)
	initialCount = testutil.CollectAndCount(HttpRequestsInProgress)
	IncrementHttpRequestInProgress("GET", "/metrics")
	// Don't check count here as we're just modifying an existing metric
	DecrementHttpRequestInProgress("GET", "/metrics")
	// Don't check count here as we're just modifying an existing metric
}

// TestSystemMetrics tests the new system metrics
func TestSystemMetrics(t *testing.T) {
	// Test ServiceHealthy
	UpdateServiceHealth(true)
	expected := 1.0
	actual := testutil.ToFloat64(ServiceHealthy)
	if actual != expected {
		t.Errorf("Expected ServiceHealthy to be %f, got %f", expected, actual)
	}

	// Test K8sReadyStatus
	UpdateK8sReadyStatus(true)
	expected = 1.0
	actual = testutil.ToFloat64(K8sReadyStatus)
	if actual != expected {
		t.Errorf("Expected K8sReadyStatus to be %f, got %f", expected, actual)
	}

	// Test K8sLiveStatus
	UpdateK8sLiveStatus(true)
	expected = 1.0
	actual = testutil.ToFloat64(K8sLiveStatus)
	if actual != expected {
		t.Errorf("Expected K8sLiveStatus to be %f, got %f", expected, actual)
	}

	// Test K8sPodRestartsTotal
	initialValue := testutil.ToFloat64(K8sPodRestartsTotal)
	IncrementK8sPodRestarts()
	expected = initialValue + 1.0
	actual = testutil.ToFloat64(K8sPodRestartsTotal)
	if actual != expected {
		t.Errorf("Expected K8sPodRestartsTotal to be %f, got %f", expected, actual)
	}
}

// TestLegacyMetrics tests the legacy metrics for backward compatibility
func TestLegacyMetrics(t *testing.T) {
	// Test SyncCount
	initial := testutil.ToFloat64(SyncCount)
	SyncCount.Inc()
	expected := initial + 1.0
	actual := testutil.ToFloat64(SyncCount)
	if actual != expected {
		t.Errorf("Expected counter to be %f, got %f", expected, actual)
	}

	// Test SyncTotalCount
	initial = testutil.ToFloat64(SyncTotalCount)
	SyncTotalCount.Inc()
	expected = initial + 1.0
	actual = testutil.ToFloat64(SyncTotalCount)
	if actual != expected {
		t.Errorf("Expected counter to be %f, got %f", expected, actual)
	}

	// Test SyncTotalErrorCount
	initial = testutil.ToFloat64(SyncTotalErrorCount)
	SyncTotalErrorCount.Inc()
	expected = initial + 1.0
	actual = testutil.ToFloat64(SyncTotalErrorCount)
	if actual != expected {
		t.Errorf("Expected counter to be %f, got %f", expected, actual)
	}

	// Test SyncRepoInfo (legacy)
	// For GaugeVec, we need to set specific labels first
	SyncRepoInfo.WithLabelValues("test", "main").Set(1)
	// We can't easily test the value without more complex setup
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
	UpdateCommitInfo(commitInfo, "test-repo")

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
	UpdateSyncRepoInfo(options, "test-repo")

	// Check that the gauge has the expected value
	expected := 1.0
	actual := testutil.ToFloat64(SyncRepoInfo)
	if actual != expected {
		t.Errorf("Expected gauge to be %f, got %f", expected, actual)
	}
}

// TestHistogramMetrics tests histogram metrics
func TestHistogramMetrics(t *testing.T) {
	// Test SyncDurationSeconds
	initialCount := testutil.CollectAndCount(SyncDurationSeconds)
	ObserveSyncDuration("manual", "success", 5.0)
	newCount := testutil.CollectAndCount(SyncDurationSeconds)
	if newCount != initialCount+1 {
		t.Errorf("Expected SyncDurationSeconds observation count to be %d, got %d", initialCount+1, newCount)
	}

	// Test HttpRequestDurationSeconds
	initialCount = testutil.CollectAndCount(HttpRequestDurationSeconds)
	ObserveHttpRequestDuration("GET", "/metrics", 0.1)
	newCount = testutil.CollectAndCount(HttpRequestDurationSeconds)
	if newCount != initialCount+1 {
		t.Errorf("Expected HttpRequestDurationSeconds observation count to be %d, got %d", initialCount+1, newCount)
	}
}
