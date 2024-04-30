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

package http

import (
	"context"
	"flag"
	"fmt"
	"git-sync/internal/constants"
	"git-sync/internal/metrics"
	"git-sync/logger"
	"git-sync/webhook"
	"net/http"
	"sort"

	"github.com/justinas/alice"
)

// Словарь для хранения всех путей хэндлеров
var registeredPaths = make(map[string]bool)

// Функция для регистрации хэндлеров
func registerHandler(path string, handler http.Handler, handlerFunc http.HandlerFunc) {
	if handler != nil {
		http.Handle(path, handler)
		registeredPaths[path] = true
	} else if handlerFunc != nil {
		http.HandleFunc(path, handlerFunc)
		registeredPaths[path] = true
	} else {
		panic("registerHandler: не передан хэндлер или функция хэндлера")
	}

	if path == "/" {
		delete(registeredPaths, "/")
	}
}

func basicAuthMiddleware(username, password string) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerAuthMiddleware(token string) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if authHeader != "Bearer "+token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func rootHandlerFunc(w http.ResponseWriter, r *http.Request) {

	var paths []string
	for path := range registeredPaths {
		paths = append(paths, path)
	}

	// Сортируем пути хендлеров
	sort.Strings(paths)

	// Выводим список хендлеров на основной странице
	fmt.Fprintf(w, "<h1>Список доступных хэндлеров:</h1>\n<ul>\n")
	for _, path := range paths {
		fmt.Fprintf(w, "<li><a href=\"%s\">%s</a></li>\n", path, path)
	}
	fmt.Fprintf(w, "</ul>\n")
}

func StartServer(f *flag.FlagSet, ctx context.Context) {

	addr := f.Lookup(constants.FlagHttpServerAddr).Value.(flag.Getter).Get().(string)
	basicUsername := f.Lookup(constants.FlagHttpServerAuthUsername).Value.(flag.Getter).Get().(string)
	basicPassword := f.Lookup(constants.FlagHttpServerAuthPassword).Value.(flag.Getter).Get().(string)
	bearerToken := f.Lookup(constants.FlagHttpServerAuthToken).Value.(flag.Getter).Get().(string)

	useBasicAuth := basicUsername != "" && basicPassword != ""
	useBaererToken := len(bearerToken) > 0

	if len(addr) == 0 {
		logger.GetLogger().Info("HTTP-сервер: не запущен\n")
		return
	} else {
		logger.GetLogger().Info("HTTP-сервер: http://%s", addr)
	}

	chain := alice.New()

	if useBasicAuth {
		chain = chain.Append(basicAuthMiddleware(basicUsername, basicPassword))
		logger.GetLogger().Info("HTTP-сервер: используется базовая аутентификация\n")
	}

	if useBaererToken {
		chain = chain.Append(bearerAuthMiddleware(bearerToken))
		logger.GetLogger().Info("HTTP-сервер: используется аутентификация по токену\n")
	}

	registerHandler("/metrics", chain.Then(metrics.MetricsHandler()), nil)
	registerHandler("/webhook", http.HandlerFunc(webhook.WebhookHandlerFunc), nil)
	registerHandler("/", nil, rootHandlerFunc)

	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()
}
