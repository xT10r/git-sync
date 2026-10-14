// Package syncer coordinates one sequential reconciliation worker.
package syncer

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"git-sync/internal/snapshot"
	"github.com/prometheus/client_golang/prometheus"
)

type Repository interface {
	Sync(context.Context) (snapshot.Result, error)
	Current() snapshot.Result
}

type Status struct {
	SHA         string    `json:"sha"`
	Path        string    `json:"path"`
	Ready       bool      `json:"ready"`
	Degraded    bool      `json:"degraded"`
	Syncing     bool      `json:"syncing"`
	LastAttempt time.Time `json:"last_attempt"`
	LastSuccess time.Time `json:"last_success"`
	PublishedAt time.Time `json:"published_at"`
	LastError   string    `json:"last_error,omitempty"`
}

type Service struct {
	mu                                sync.RWMutex
	repo                              Repository
	interval, timeLimit               time.Duration
	logger                            *slog.Logger
	pending                           chan struct{}
	status                            Status
	started, stopping                 bool
	attempts                          *prometheus.CounterVec
	duration                          prometheus.Histogram
	lastSuccess, lastPublished, ready prometheus.Gauge
}

func New(repo Repository, interval, timeout time.Duration, logger *slog.Logger, reg prometheus.Registerer) (*Service, error) {
	if repo == nil || logger == nil || reg == nil || interval <= 0 || timeout <= 0 {
		return nil, errors.New("invalid sync service configuration")
	}
	current := repo.Current()
	s := &Service{repo: repo, interval: interval, timeLimit: timeout, logger: logger, pending: make(chan struct{}, 1), status: Status{SHA: current.SHA, Path: current.Path, Ready: current.SHA != "", PublishedAt: current.PublishedAt}}
	s.attempts = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "git_sync_attempts_total", Help: "Completed attempts by outcome."}, []string{"outcome"})
	for _, v := range []string{"published", "unchanged", "error"} {
		s.attempts.WithLabelValues(v)
	}
	s.duration = prometheus.NewHistogram(prometheus.HistogramOpts{Name: "git_sync_duration_seconds", Help: "Duration of reconciliation attempts."})
	s.lastSuccess = prometheus.NewGauge(prometheus.GaugeOpts{Name: "git_sync_last_success_timestamp_seconds", Help: "Last successful remote reconciliation."})
	s.lastPublished = prometheus.NewGauge(prometheus.GaugeOpts{Name: "git_sync_last_publication_timestamp_seconds", Help: "Current snapshot publication time."})
	s.ready = prometheus.NewGauge(prometheus.GaugeOpts{Name: "git_sync_ready", Help: "A published snapshot is available."})
	for _, collector := range []prometheus.Collector{s.attempts, s.duration, s.lastSuccess, s.lastPublished, s.ready} {
		if err := reg.Register(collector); err != nil {
			return nil, err
		}
	}
	if current.SHA != "" {
		s.ready.Set(1)
		s.lastPublished.Set(float64(current.PublishedAt.Unix()))
	}
	return s, nil
}
func (s *Service) Status() Status { s.mu.RLock(); defer s.mu.RUnlock(); return s.status }

// Trigger coalesces bursts into at most one pending reconciliation.
func (s *Service) Trigger() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.stopping {
		return false
	}
	select {
	case s.pending <- struct{}{}:
	default:
	}
	return true
}
func (s *Service) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("sync worker already started")
	}
	s.started = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.stopping = true; s.mu.Unlock() }()
	if ctx.Err() != nil {
		return nil
	}
	s.reconcile(ctx)
	timer := time.NewTimer(s.interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.pending:
		case <-timer.C:
		}
		if ctx.Err() != nil {
			return nil
		}
		s.reconcile(ctx)
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(s.interval)
	}
}
func (s *Service) reconcile(parent context.Context) {
	start := time.Now().UTC()
	s.mu.Lock()
	s.status.Syncing = true
	s.status.LastAttempt = start
	s.mu.Unlock()
	ctx, cancel := context.WithTimeout(parent, s.timeLimit)
	result, err := s.repo.Sync(ctx)
	cancel()
	end := time.Now().UTC()
	s.duration.Observe(end.Sub(start).Seconds())
	s.mu.Lock()
	s.status.Syncing = false
	if err != nil {
		s.status.Degraded = true
		// Transport errors can contain URLs/credentials: expose a bounded safe description.
		s.status.LastError = "snapshot synchronization failed"
		var operation *snapshot.OperationError
		if errors.As(err, &operation) {
			s.status.LastError += " during " + operation.Operation
		}
		if errors.Is(err, context.DeadlineExceeded) {
			s.status.LastError = "snapshot synchronization timed out"
		}
		if errors.Is(err, context.Canceled) {
			s.status.LastError = "snapshot synchronization canceled"
		}
		safeError := s.status.LastError
		s.mu.Unlock()
		s.attempts.WithLabelValues("error").Inc()
		s.logger.Warn(safeError)
		return
	}
	s.status.SHA = result.SHA
	s.status.Path = result.Path
	s.status.Ready = result.SHA != ""
	s.status.Degraded = false
	s.status.LastError = ""
	s.status.LastSuccess = end
	s.status.PublishedAt = result.PublishedAt
	s.mu.Unlock()
	s.lastSuccess.Set(float64(end.Unix()))
	if result.SHA != "" {
		s.ready.Set(1)
	}
	outcome := "unchanged"
	if result.Published {
		outcome = "published"
		s.lastPublished.Set(float64(result.PublishedAt.Unix()))
		s.logger.Info("snapshot published", "sha", result.SHA)
	}
	s.attempts.WithLabelValues(outcome).Inc()
}
