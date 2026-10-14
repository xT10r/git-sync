package httpserver

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git-sync/internal/config"
	"git-sync/internal/syncer"
	"github.com/prometheus/client_golang/prometheus"
)

type fakeService struct {
	ready, stopping bool
	calls           int
}

func (f *fakeService) Status() syncer.Status { return syncer.Status{Ready: f.ready} }
func (f *fakeService) Trigger() bool         { f.calls++; return !f.stopping }
func TestEndpoints(t *testing.T) {
	service := &fakeService{}
	handler := Handler(config.Config{HTTPToken: "secret"}, service, prometheus.NewRegistry())
	for _, tc := range []struct {
		path, method, token string
		want                int
	}{
		{"/livez", "GET", "", 200}, {"/readyz", "GET", "", 503},
		{"/status", "GET", "", 401}, {"/status", "GET", "secret", 200},
		{"/webhook", "GET", "secret", 405}, {"/webhook", "POST", "secret", 202},
		{"/metrics", "GET", "", 401}, {"/metrics", "GET", "secret", 200},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("%s %s: %d, want %d", tc.method, tc.path, rec.Code, tc.want)
		}
	}
	if service.calls != 1 {
		t.Fatal("unauthorized or GET request triggered sync")
	}
	service.ready = true
	service.stopping = true
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	req := httptest.NewRequest("POST", "/webhook", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatal(rec.Code)
	}
}
func TestBasicAuth(t *testing.T) {
	handler := Handler(config.Config{HTTPUsername: "u", HTTPPassword: "p"}, &fakeService{}, prometheus.NewRegistry())
	for _, password := range []string{"bad", "p"} {
		req := httptest.NewRequest("GET", "/status", nil)
		req.SetBasicAuth("u", password)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if (rec.Code == 200) != (password == "p") {
			t.Fatal(rec.Code)
		}
	}
}
func TestServeShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	}()
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP did not stop")
	}
	if _, err = client.Get("http://" + listener.Addr().String()); err == nil {
		t.Fatal("listener still serving")
	}
}
func TestNoSecretInUnauthenticatedResponse(t *testing.T) {
	h := Handler(config.Config{HTTPToken: "sensitive"}, &fakeService{}, prometheus.NewRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/status", nil))
	if strings.Contains(rec.Body.String(), "sensitive") {
		t.Fatal("credential leaked")
	}
}
