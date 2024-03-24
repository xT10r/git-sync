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
	"fmt"
	"net/http"
)

// Канал для вебхука
var WebhookCh = make(chan bool)

// TODO: Запуск сервера будет выполняться в пакете http. Нужно тут оставить Handle

// WebhookHandlerFunc обрабатывает запросы по вебхуку
func WebhookHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Отправляем сигнал о получении вебхука
	WebhookCh <- true
	// Возвращаем успешный статус выполнения
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Синхронизация запущена по вебхуку")
}
