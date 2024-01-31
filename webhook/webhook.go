/*
Основная логика - запуск HTTP сервера и обработка входящих запросов вебхуков в handler.
При получении валидного запроса вызываем метод pull из пакета git для синхронизации репозитория.
Пакет webhook инкапсулирует всю работу с вебхуками.
*/

package webhook

import (
	"net/http"
	// "github.com/x/sync-service/git"
)

func Start() {

	http.HandleFunc("/hook", handler)

	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	// Проверка и валидация запроса

	// Парсинг данных вебхука

	// Вызов метода pull из пакета git

	w.WriteHeader(200)
}
