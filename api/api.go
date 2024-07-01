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

/*
Пакет api содержит обработчики запросов к различным HTTP endpoint'ам REST API приложения.
Определяет модели данных для ответов. Использует встроенный HTTP сервер Go.
Таким образом формируется API для мониторинга и управления git-sync.
*/

package api

import (
	"encoding/json"
	"net/http"
)

// StatusResponse Структура ответа API
type StatusResponse struct {
	Status string `json:"status"`
}

// TODO: Запуск сервера будет выполняться в пакете http. Нужно тут оставить Handle
func SetupRoutes() {
	http.HandleFunc("/status", handlerStatus)
}

func handlerStatus(w http.ResponseWriter, r *http.Request) {

	var resp StatusResponse

	json.NewEncoder(w).Encode(resp)

}
