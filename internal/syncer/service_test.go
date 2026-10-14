package syncer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"git-sync/internal/snapshot"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeRepo struct {
	calls  chan context.Context
	result snapshot.Result
	err    error
	block  bool
}

func (f *fakeRepo) Current() snapshot.Result { return f.result }
func (f *fakeRepo) Sync(ctx context.Context) (snapshot.Result, error) {
	if f.calls != nil {
		f.calls <- ctx
	}
	if f.block {
		<-ctx.Done()
		return snapshot.Result{}, ctx.Err()
	}
	return f.result, f.err
}
func newService(t *testing.T, f *fakeRepo, interval, timeout time.Duration) *Service {
	t.Helper()
	s, err := New(f, interval, timeout, slog.New(slog.NewTextHandler(io.Discard, nil)), prometheus.NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func TestOutcomeAccounting(t *testing.T) {
	f := &fakeRepo{result: snapshot.Result{SHA: "abc", Path: "/snapshot/abc", Published: true, PublishedAt: time.Now()}}
	s := newService(t, f, time.Second, time.Second)
	s.reconcile(context.Background())
	if got := s.Status(); !got.Ready || got.Degraded || got.SHA != "abc" {
		t.Fatalf("status %+v", got)
	}
	if testutil.ToFloat64(s.attempts.WithLabelValues("published")) != 1 {
		t.Fatal("publication not counted")
	}
	success := s.Status().LastSuccess
	f.err = errors.New("secret-token-in-transport-error")
	s.reconcile(context.Background())
	if got := s.Status(); !got.Ready || !got.Degraded || got.LastSuccess != success || got.LastError == f.err.Error() {
		t.Fatalf("failure status %+v", got)
	}
	if testutil.ToFloat64(s.attempts.WithLabelValues("error")) != 1 || testutil.ToFloat64(s.attempts.WithLabelValues("published")) != 1 {
		t.Fatal("failed attempt counted as publication")
	}
	f.err = nil
	f.result.Published = false
	s.reconcile(context.Background())
	if s.Status().Degraded || testutil.ToFloat64(s.attempts.WithLabelValues("unchanged")) != 1 {
		t.Fatal("recovery/no-op not recorded")
	}
}
func TestTriggerCoalesces(t *testing.T) {
	s := newService(t, &fakeRepo{}, time.Second, time.Second)
	for i := 0; i < 100; i++ {
		if !s.Trigger() {
			t.Fatal("rejected trigger")
		}
	}
	if len(s.pending) != 1 {
		t.Fatal("triggers not coalesced")
	}
	s.stopping = true
	if s.Trigger() {
		t.Fatal("accepted after stop")
	}
}
func TestRunInitialPeriodicAndShutdown(t *testing.T) {
	f := &fakeRepo{calls: make(chan context.Context, 10)}
	s := newService(t, f, 5*time.Millisecond, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	t.Cleanup(func() { cancel() })
	for i := 0; i < 2; i++ {
		select {
		case <-f.calls:
		case <-time.After(time.Second):
			t.Fatal("initial/periodic sync not invoked")
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	if s.Trigger() {
		t.Fatal("trigger accepted after stop")
	}
	if err := s.Run(context.Background()); err == nil {
		t.Fatal("second Run accepted")
	}
}
func TestTimeoutAndResponsiveStatus(t *testing.T) {
	f := &fakeRepo{calls: make(chan context.Context, 1), block: true}
	s := newService(t, f, time.Hour, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	select {
	case <-f.calls:
	case <-time.After(time.Second):
		t.Fatal("worker not started")
	}
	statusDone := make(chan Status, 1)
	go func() { statusDone <- s.Status() }()
	select {
	case status := <-statusDone:
		if !status.Syncing {
			t.Fatal("expected syncing")
		}
	case <-time.After(time.Second):
		t.Fatal("status blocked by Git operation")
	}
	deadline := time.After(time.Second)
	for !s.Status().Degraded {
		select {
		case <-deadline:
			t.Fatal("timeout not recorded")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if s.Status().LastError != "snapshot synchronization timed out" {
		t.Fatal(s.Status())
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown blocked")
	}
}
