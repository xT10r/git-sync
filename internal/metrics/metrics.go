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
	"fmt"
	"git-sync/git"
	"git-sync/internal/version"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 1. Sync Metrics
	SyncRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_requests_total",
			Help: "Total number of synchronization requests",
		},
		[]string{"type", "result", "repository"},
	)

	SyncDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sync_duration_seconds",
			Help:    "Time spent on synchronization operations",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60},
		},
		[]string{"type", "result"},
	)

	SyncErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_errors_total",
			Help: "Total number of synchronization errors",
		},
		[]string{"error_type", "repository"},
	)

	SyncStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sync_status",
			Help: "Current synchronization status",
		},
		[]string{"repository", "branch", "metric"},
	)

	SyncInProgress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_in_progress",
			Help: "Number of active synchronizations",
		},
	)

	// 2. Repository Metrics
	RepoHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repo_health",
			Help: "Repository health status",
		},
		[]string{"repository", "check"},
	)

	RepoSyncInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repo_sync_info",
			Help: "Information about the synchronized repository",
		},
		[]string{"repository", "branch", "remote_url"},
	)

	RepoCommitInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repo_commit_info",
			Help: "Information about the latest commit",
		},
		[]string{"repository", "branch", "commit_hash", "author", "message_truncated"},
	)

	RepoCommitDrift = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repo_commit_drift",
			Help: "Repository commit drift from remote",
		},
		[]string{"repository", "branch"},
	)

	RepoLastSyncAgeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repo_last_sync_age_seconds",
			Help: "Age of last synchronization in seconds",
		},
		[]string{"repository"},
	)

	// 3. Performance and Resource Metrics
	SyncFilesProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_files_processed_total",
			Help: "Total number of files processed",
		},
		[]string{"repository", "operation"},
	)

	SyncDataTransferredBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_data_transferred_bytes_total",
			Help: "Total bytes transferred during synchronization",
		},
		[]string{"repository", "direction"},
	)

	DiskUsageBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "disk_usage_bytes",
			Help: "Disk space usage",
		},
		[]string{"repository", "type"},
	)

	MemoryUsageBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Memory usage by the process",
		},
	)

	GoroutinesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Number of goroutines",
		},
	)

	// 4. HTTP Metrics
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HttpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"method", "endpoint"},
	)

	HttpRequestsInProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_progress",
			Help: "Number of active HTTP requests",
		},
		[]string{"method", "endpoint"},
	)

	// 5. System and Informational Metrics
	BuildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_info",
			Help: "Build information",
		},
		[]string{"version", "commit", "build_date", "go_version"},
	)

	ServiceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_info",
			Help: "Service information",
		},
		[]string{"name", "environment", "instance"},
	)

	ServiceUptimeSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
	)

	ServiceHealthy = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_healthy",
			Help: "Service health status (0/1)",
		},
	)

	// 6. Kubernetes Metrics
	K8sReadyStatus = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_ready_status",
			Help: "Kubernetes readiness probe status (0/1)",
		},
	)

	K8sLiveStatus = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_live_status",
			Help: "Kubernetes liveness probe status (0/1)",
		},
	)

	K8sPodRestartsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "k8s_pod_restarts_total",
			Help: "Total number of pod restarts",
		},
	)

	// 7. Webhook Metrics
	WebhookRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_requests_total",
			Help: "Total number of webhook requests",
		},
		[]string{"result"},
	)

	WebhookRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_request_duration_seconds",
			Help:    "Processing time for webhook requests",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{},
	)

	// Webhook rate limiting metrics
	WebhookRateLimitedIPs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "webhook_rate_limited_ips",
			Help: "Current rate limited IPs (1 if rate limited, 0 if not)",
		},
		[]string{"ip"},
	)

	WebhookRequestsPerIP = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "webhook_requests_per_ip",
			Help: "Current number of requests per IP in the rate limit window",
		},
		[]string{"ip"},
	)

	// Legacy metrics for backward compatibility
	SyncRepoInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "git_sync_repo_info",
			Help: "Information about the synchronized repository (deprecated)",
		},
		[]string{"repository", "branch"},
	)

	SyncCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "git_sync_sync_count",
			Help: "Total number of synchronizations with changes (deprecated)",
		},
	)

	SyncTotalCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "git_sync_sync_total_count",
			Help: "Total number of synchronizations (deprecated)",
		},
	)

	SyncTotalErrorCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "git_sync_sync_total_error_count",
			Help: "Total number of synchronization errors (deprecated)",
		},
	)

	CommitInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "git_sync_commit_info",
		Help: "Information about the latest commit (deprecated).",
	}, []string{"hash", "author", "email", "date", "message"})
)

func init() {
	// Register all metrics
	prometheus.MustRegister(SyncRequestsTotal)
	prometheus.MustRegister(SyncDurationSeconds)
	prometheus.MustRegister(SyncErrorsTotal)
	prometheus.MustRegister(SyncStatus)
	prometheus.MustRegister(SyncInProgress)
	prometheus.MustRegister(RepoHealth)
	prometheus.MustRegister(RepoSyncInfo)
	prometheus.MustRegister(RepoCommitInfo)
	prometheus.MustRegister(RepoCommitDrift)
	prometheus.MustRegister(RepoLastSyncAgeSeconds)
	prometheus.MustRegister(SyncFilesProcessedTotal)
	prometheus.MustRegister(SyncDataTransferredBytesTotal)
	prometheus.MustRegister(DiskUsageBytes)
	prometheus.MustRegister(MemoryUsageBytes)
	prometheus.MustRegister(GoroutinesCount)
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDurationSeconds)
	prometheus.MustRegister(HttpRequestsInProgress)
	prometheus.MustRegister(BuildInfo)
	prometheus.MustRegister(ServiceInfo)
	prometheus.MustRegister(ServiceUptimeSeconds)
	prometheus.MustRegister(ServiceHealthy)
	prometheus.MustRegister(K8sReadyStatus)
	prometheus.MustRegister(K8sLiveStatus)
	prometheus.MustRegister(K8sPodRestartsTotal)
	prometheus.MustRegister(WebhookRequestsTotal)
	prometheus.MustRegister(WebhookRequestDurationSeconds)
	prometheus.MustRegister(WebhookRateLimitedIPs)
	prometheus.MustRegister(WebhookRequestsPerIP)
	prometheus.MustRegister(SyncRepoInfo)
	prometheus.MustRegister(SyncCount)
	prometheus.MustRegister(SyncTotalCount)
	prometheus.MustRegister(SyncTotalErrorCount)
	prometheus.MustRegister(CommitInfo)

	// Initialize build info metrics
	InitBuildMetrics()

	// Initialize default metrics
	InitDefaultMetrics()

	// Initialize service uptime
	go func() {
		startTime := time.Now()
		for {
			ServiceUptimeSeconds.Set(time.Since(startTime).Seconds())
			time.Sleep(1 * time.Second)
		}
	}()
}

// InitBuildMetrics initializes the build information metrics
func InitBuildMetrics() {
	BuildInfo.WithLabelValues(version.Version, version.Commit, version.Date, "unknown").Set(1)
}

// InitDefaultMetrics initializes default metric values
func InitDefaultMetrics() {
	// Set default healthy state for service and Kubernetes metrics
	UpdateServiceHealth(true)
	UpdateK8sReadyStatus(true)
	UpdateK8sLiveStatus(true)
}

// DecrementSyncInProgress decrements the sync in progress counter
func DecrementSyncInProgress() {
	SyncInProgress.Add(-1)
}

// DecrementHttpRequestInProgress decrements the HTTP requests in progress counter
func DecrementHttpRequestInProgress(method, endpoint string) {
	HttpRequestsInProgress.WithLabelValues(method, endpoint).Add(-1)
}

// UpdateCommitInfo updates the commit information metrics
func UpdateCommitInfo(gci *git.CommitInfo, repositoryName string) {
	unixTimestamp := gci.Date.UnixNano() / int64(time.Millisecond)
	CommitInfo.Reset()
	CommitInfo.WithLabelValues(gci.Hash, gci.Author, gci.Email, fmt.Sprintf("%d", unixTimestamp), gci.Message).Set(1)

	// Also update the new metrics format
	messageTruncated := gci.Message
	if len(messageTruncated) > 100 {
		messageTruncated = messageTruncated[:100]
	}
	RepoCommitInfo.WithLabelValues(repositoryName, "main", gci.Hash, gci.Author, messageTruncated).Set(1)
}

// UpdateSyncRepoInfo updates the repository information metrics
func UpdateSyncRepoInfo(gro *git.GitRepositoryOptions, repositoryName string) {
	SyncRepoInfo.Reset()
	SyncRepoInfo.WithLabelValues(gro.Url(), gro.Branch()).Set(1)

	// Also update the new metrics format
	RepoSyncInfo.WithLabelValues(repositoryName, gro.Branch(), gro.Url()).Set(1)
}

// IncrementSyncRequest increments the sync requests counter
func IncrementSyncRequest(syncType, result, repository string) {
	SyncRequestsTotal.WithLabelValues(syncType, result, repository).Inc()
}

// ObserveSyncDuration records the duration of a sync operation
func ObserveSyncDuration(syncType, result string, duration float64) {
	SyncDurationSeconds.WithLabelValues(syncType, result).Observe(duration)
}

// IncrementSyncError increments the sync errors counter
func IncrementSyncError(errorType, repository string) {
	SyncErrorsTotal.WithLabelValues(errorType, repository).Inc()
	SyncTotalErrorCount.Inc() // For backward compatibility
}

// UpdateSyncStatus updates the sync status gauge
func UpdateSyncStatus(repository, branch, metric string, value float64) {
	SyncStatus.WithLabelValues(repository, branch, metric).Set(value)
}

// IncrementSyncInProgress increments the sync in progress counter
func IncrementSyncInProgress() {
	SyncInProgress.Inc()
}

// UpdateRepoHealth updates the repository health gauge
func UpdateRepoHealth(repository, check string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	RepoHealth.WithLabelValues(repository, check).Set(value)
}

// UpdateRepoCommitDrift updates the repository commit drift gauge
func UpdateRepoCommitDrift(repository, branch string, drift int) {
	RepoCommitDrift.WithLabelValues(repository, branch).Set(float64(drift))
}

// UpdateRepoLastSyncAge updates the repository last sync age gauge
func UpdateRepoLastSyncAge(repository string, ageSeconds float64) {
	RepoLastSyncAgeSeconds.WithLabelValues(repository).Set(ageSeconds)
}

// IncrementFilesProcessed increments the files processed counter
func IncrementFilesProcessed(repository, operation string) {
	SyncFilesProcessedTotal.WithLabelValues(repository, operation).Inc()
}

// IncrementDataTransferred increments the data transferred counter
func IncrementDataTransferred(repository, direction string, bytes int64) {
	SyncDataTransferredBytesTotal.WithLabelValues(repository, direction).Add(float64(bytes))
}

// UpdateDiskUsage updates the disk usage gauge
func UpdateDiskUsage(repository, usageType string, bytes int64) {
	DiskUsageBytes.WithLabelValues(repository, usageType).Set(float64(bytes))
}

// UpdateMemoryUsage updates the memory usage gauge
func UpdateMemoryUsage(bytes int64) {
	MemoryUsageBytes.Set(float64(bytes))
}

// UpdateGoroutinesCount updates the goroutines count gauge
func UpdateGoroutinesCount(count int) {
	GoroutinesCount.Set(float64(count))
}

// IncrementHttpRequest increments the HTTP requests counter
func IncrementHttpRequest(method, endpoint, statusCode string) {
	HttpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// ObserveHttpRequestDuration records the duration of an HTTP request
func ObserveHttpRequestDuration(method, endpoint string, duration float64) {
	HttpRequestDurationSeconds.WithLabelValues(method, endpoint).Observe(duration)
}

// IncrementHttpRequestInProgress increments the HTTP requests in progress counter
func IncrementHttpRequestInProgress(method, endpoint string) {
	HttpRequestsInProgress.WithLabelValues(method, endpoint).Inc()
}

// UpdateServiceInfo updates the service information gauge
func UpdateServiceInfo(name, environment, instance string) {
	ServiceInfo.WithLabelValues(name, environment, instance).Set(1)
}

// UpdateServiceHealth updates the service health gauge
func UpdateServiceHealth(healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	ServiceHealthy.Set(value)
}

// UpdateK8sReadyStatus updates the Kubernetes readiness status gauge
func UpdateK8sReadyStatus(ready bool) {
	value := 0.0
	if ready {
		value = 1.0
	}
	K8sReadyStatus.Set(value)
}

// UpdateK8sLiveStatus updates the Kubernetes liveness status gauge
func UpdateK8sLiveStatus(live bool) {
	value := 0.0
	if live {
		value = 1.0
	}
	K8sLiveStatus.Set(value)
}

// IncrementK8sPodRestarts increments the Kubernetes pod restarts counter
func IncrementK8sPodRestarts() {
	K8sPodRestartsTotal.Inc()
}

// IncrementWebhookRequest increments the webhook requests counter
func IncrementWebhookRequest(result string) {
	WebhookRequestsTotal.WithLabelValues(result).Inc()
}

// ObserveWebhookRequestDuration records the duration of a webhook request
func ObserveWebhookRequestDuration(duration float64) {
	WebhookRequestDurationSeconds.WithLabelValues().Observe(duration)
}

// SetWebhookRateLimitedIP sets the rate limited status for an IP
func SetWebhookRateLimitedIP(ip string, limited bool) {
	value := 0.0
	if limited {
		value = 1.0
	}
	WebhookRateLimitedIPs.WithLabelValues(ip).Set(value)
}

// SetWebhookRequestsPerIP sets the number of requests for an IP
func SetWebhookRequestsPerIP(ip string, count float64) {
	WebhookRequestsPerIP.WithLabelValues(ip).Set(count)
}

// RemoveWebhookRateLimitedIP removes an IP from the rate limited metrics
func RemoveWebhookRateLimitedIP(ip string) {
	WebhookRateLimitedIPs.DeleteLabelValues(ip)
	WebhookRequestsPerIP.DeleteLabelValues(ip)
}
