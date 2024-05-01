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

package handlers_test

import (
	"bytes"
	"encoding/json"
	"git-sync/internal/handlers"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
)

func TestWebhookHandlerFunc(t *testing.T) {

	defer func() {
		close(handlers.WebhookCh) // Закрываем канал после отправки данных
	}()

	// Ждем некоторое время для запуска сервера.
	time.Sleep(100 * time.Millisecond)

	// Создаем запрос для передачи в наш обработчик.
	req, err := http.NewRequest("GET", "http://localhost:8080/webhook", nil)
	req.RemoteAddr = "127.0.0.1:15200"
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Вызываем метод ServeHTTP непосредственно нашего хендлера и передаем ему наш запрос и ResponseRecorder.
	handler := http.HandlerFunc(handlers.WebhookHandlerFunc)
	handler.ServeHTTP(rr, req)

	// Проверяем, что статус код соответствует ожидаемому.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем, что тело ответа соответствует структуре WebhookResponse
	var response handlers.WebhookResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	// Проверяем, что сообщение в теле ответа соответствует ожидаемому
	expectedMessage := "Синхронизация запущена по вебхуку"
	if response.Message != expectedMessage {
		t.Errorf("handler returned unexpected message: got %v, want %v", response.Message, expectedMessage)
	}

	// Проверяем, что IP-адрес был отправлен в канал
	select {
	case ipAddress := <-handlers.WebhookCh:
		if ipAddress != req.RemoteAddr {
			t.Errorf("expected IP address %v, got %v", req.RemoteAddr, ipAddress)
		}
	case <-time.After(2 * time.Second):
		t.Error("expected IP address was not sent to the channel")
	}
}

func TestMetricsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handlers.MetricsHandler()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Используем парсер метрик Prometheus для чтения метрик из тела ответа
	parser := expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем наличие нужной метрики в коллекции
	expectedMetric := "go_info"
	if _, ok := metricFamilies[expectedMetric]; !ok {
		t.Errorf("handler response does not contain expected metric %q", expectedMetric)
	}
}
