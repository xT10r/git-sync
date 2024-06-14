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

package gitsync_test

import (
	"context"
	"git-sync/internal/gitsync"
	"git-sync/internal/handlers"
	"git-sync/mock"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	ctx := context.Background()

	// Создаем макет флагов для использования в тесте
	mockFlags := mock.Flags()

	// Парсим флаги
	err := mockFlags.Parse(nil)
	if err != nil {
		t.Fatalf("error parsing flags: %v", err)
	}

	// Создаем фейковый контекст с отменой через 100 миллисекунд
	fakeCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Создаем экземпляр gitsync с использованием макета флагов и фейкового контекста
	gitSync, err := gitsync.NewGitSync(mockFlags, fakeCtx)
	if err != nil {
		t.Fatalf("Error initializing GitSync: %v", err)
	}

	// Создаем мок Gitter
	mockGitter := &mock.Gitter{}

	// Запуск синхронизации в отдельной горутине
	go gitSync.Start(mockGitter)

	// Отправляем сообщение в канал вебхуков для проверки второй ветки select
	go func() {
		handlers.WebhookCh <- "127.0.0.1"
	}()

	// Ждем некоторое время для проверки, что синхронизация запущена и остановлена
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что контекст завершился (первая ветка select)
	select {
	case <-gitSync.GetCtx().Done():
		// Контекст успешно завершен
	default:
		t.Error("Context was not cancelled as expected")
	}

	// Ждем тикера (третья ветка select)
	time.Sleep(1000 * time.Millisecond)

	// Проверяем, что все случаи были покрыты
	if len(handlers.WebhookCh) != 0 {
		t.Error("Webhook channel was not emptied as expected")
	}
}
