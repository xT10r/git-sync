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

package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Канал для вебхука
var WebhookCh = make(chan string)

// WebhookHandlerFunc обрабатывает запросы по вебхуку
func WebhookHandlerFunc(w http.ResponseWriter, r *http.Request) {

	// Получаем IP-адрес клиента из запроса
	ipAddress := r.RemoteAddr

	// Отправляем сигнал о получении вебхука и IP-адрес клиента в канал
	WebhookCh <- ipAddress

	// Формируем JSON-структуру с сообщением и временем
	response := struct {
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}{
		Message: "Синхронизация запущена по вебхуку",
		Time:    time.Now(),
	}

	// Кодируем JSON-структуру в ответ и отправляем
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Возвращаем успешный статус выполнения
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		// В случае ошибки кодируем сообщение об ошибке в текстовом формате
		http.Error(w, fmt.Sprintf("Ошибка кодирования JSON: %v", err), http.StatusInternalServerError)
		return
	}
}
