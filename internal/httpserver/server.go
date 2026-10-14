// Package httpserver exposes an instance-local control and monitoring API.
package httpserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"git-sync/internal/config"
	"git-sync/internal/syncer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Service interface {
	Status() syncer.Status
	Trigger() bool
}

func Handler(cfg config.Config, s Service, gatherer prometheus.Gatherer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			method(w, "GET")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.Status())
	})
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			method(w, "POST")
			return
		}
		if !s.Trigger() {
			http.Error(w, "worker stopping", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, `{"status":"accepted"}`)
	})
	mux.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := true
		if cfg.HTTPUsername != "" {
			u, p, ok := r.BasicAuth()
			allowed = ok && equal(u, cfg.HTTPUsername) && equal(p, cfg.HTTPPassword)
			if !allowed {
				w.Header().Set("WWW-Authenticate", `Basic realm="git-sync"`)
			}
		} else if cfg.HTTPToken != "" {
			allowed = equal(r.Header.Get("Authorization"), "Bearer "+cfg.HTTPToken)
		}
		if !allowed {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	})
	public := http.NewServeMux()
	for _, path := range []string{"/livez", "/readyz"} {
		path := path
		public.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				method(w, "GET")
				return
			}
			if path == "/readyz" && !s.Status().Ready {
				http.Error(w, "no published snapshot", http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "ok")
		})
	}
	public.Handle("/", protected)
	return public
}
func equal(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func method(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// Serve returns only after the listener has stopped; its caller owns worker cancellation.
func Serve(ctx context.Context, listener net.Listener, handler http.Handler) error {
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 16}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(shutdown)
		if err != nil {
			server.Close()
		}
		serveErr := <-done
		if err != nil {
			return err
		}
		if serveErr != http.ErrServerClosed {
			return serveErr
		}
		return nil
	}
}
