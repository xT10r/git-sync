# Git-Sync: поставщик снимков

Сервис получает выбранную ветку Git и публикует её файлы как неизменяемый снимок конкретного commit SHA. Один процесс обслуживает один репозиторий и владеет одним root. Требуется Go 1.26+ для сборки; CI проверяет Go 1.26.9 и 1.27.2.

## Файловый контракт

В `--local-path` находятся служебные данные:

- `.git-sync.json`: идентичность источника и версия формата;
- `.git-sync.lock`: блокировка одного писателя, освобождаемая ОС при завершении процесса;
- `cache/`: bare Git-кэш;
- `snapshots/<SHA>/`: готовые версии файлов;
- `current`: относительная ссылка на опубликованный снимок.

Потребитель читает файлы, не редактируя root. Его изменяемые данные должны находиться отдельно. Root не является Git checkout для работы разработчика.

Новая версия готовится в отдельном каталоге. После подготовки каталог получает имя SHA, затем `current` заменяется через rename ссылки в той же файловой системе. Ошибка fetch/подготовки/публикации не меняет прежний `current`. Повторный SHA не вызывает повторную выгрузку файлов. Другой SHA с тем же деревом всё равно публикуется как новая ревизия. Force-push не требует fast-forward.

**Чтение нескольких связанных файлов:** разрешите `current` один раз и читайте из конкретного `snapshots/<SHA>`. Повторное разрешение `current` между открытиями файлов может смешать две версии. В shell:

```sh
snapshot="/git/root/$(readlink /git/root/current)"
cat "$snapshot/config.json"
cat "$snapshot/rules.json"
```

Все готовые снимки пока сохраняются. Автоматического удаления по возрасту/количеству нет: размер volume необходимо контролировать. Это сохраняет безопасность долгих читателей. Очистка временных `.prepare-*` выполняется при запуске под блокировкой. После перезапуска `current` восстанавливает состояние публикации даже при недоступном remote.

Поддерживаемая среда публикации: Linux/macOS/FreeBSD с локальной файловой системой, flock, symlink и атомарной заменой ссылки через rename. Фактические интеграционные тесты выполнены на Linux; macOS/Windows только кросс-компиляция. На Windows запуск явно отклоняется до реализации проверенного механизма публикации. NFS/SMB, durability при отключении питания и автоматическое восстановление повреждённого Git-кэша не гарантируются. Проверка существования снимка при restart не является проверкой хешей всех файлов: нарушение запрета записи потребителем не поддерживается.

## Запуск

```sh
go build -o git-sync ./cmd/
./git-sync --repo-url=https://git.example.org/team/config.git \
  --repo-branch=main --local-path=/srv/git-sync/config \
  --sync-interval=30s --sync-timeout=2m \
  --http-server-addr=127.0.0.1:8080
```

Root должен отсутствовать, быть пустым либо уже принадлежать этому repo/branch. Существующий чужой каталог, обычный checkout и root другого источника отклоняются без очистки. Используйте выделенный подкаталог volume, не корень файловой системы.

| Флаг | Переменная окружения | По умолчанию |
|---|---|---|
| `--repo-url` | `GITSYNC_REPOSITORY_URL` | обязательно |
| `--repo-branch` | `GITSYNC_REPOSITORY_BRANCH` | `main` |
| `--local-path` | `GITSYNC_LOCAL_PATH` | обязательно; в образе `/git` |
| `--repo-user` | `GITSYNC_REPOSITORY_USER` | пусто; с Basic token используется `git`, если пользователь не задан |
| `--repo-token` | `GITSYNC_REPOSITORY_TOKEN` | пусто |
| `--repo-auth-mode` | `GITSYNC_REPOSITORY_AUTH_MODE` | `basic`; допустим также `bearer` |
| `--sync-interval` | `GITSYNC_INTERVAL` | `30s` |
| `--sync-timeout` | `GITSYNC_TIMEOUT` | `2m` |
| `--http-server-addr` | `GITSYNC_HTTP_SERVER_ADDR` | выключен; в образе `0.0.0.0:8080` |
| `--http-auth-username` | `GITSYNC_HTTP_AUTH_USERNAME` | пусто |
| `--http-auth-password` | `GITSYNC_HTTP_AUTH_PASSWORD` | пусто |
| `--http-auth-token` | `GITSYNC_HTTP_AUTH_TOKEN` | пусто |

Приоритет CLI > env > defaults. Некорректные интервалы отклоняются; корректный CLI override может заменить ошибочное env. Файл конфигурации не поддерживается. Секреты предпочтительно задавать runtime env из Secret; секреты из env не включаются в help. Не передавайте credentials в URL: такие URL запрещены.

Git transport: HTTP(S), абсолютный локальный путь и локальный `file:///...`. Basic user/token и Bearer — разные режимы; выбирайте по контракту Git-сервера. SSH, LFS, symlinks и submodules пока не поддерживаются. LFS pointers, symlinks и gitlinks вызывают ошибку подготовки; прежний снимок остаётся опубликованным. Режим executable обычных файлов сохраняется.

Первый sync выполняется сразу. Следующий — после интервала от завершения предыдущего либо по HTTP-триггеру. Бурсты триггеров объединяются в одну ожидающую проверку. Временный сбой повторяется на следующем интервале; exponential backoff/jitter пока не реализованы. Дедлайн ограничивает transport и проверяется при выгрузке файлов; ОС может не позволить мгновенно отменить блокирующую файловую операцию.

Ошибочная конфигурация даёт exit 2, фатальная ошибка запуска/HTTP — exit 1, help и штатный SIGTERM — exit 0. Ошибка отдельного sync не останавливает сервис: публикуется degraded status с сохранённым last-good.

## HTTP и метрики

| Endpoint | Метод | Семантика |
|---|---|---|
| `/livez` | GET | HTTP-процесс отвечает; без auth |
| `/readyz` | GET | 200 при наличии опубликованного снимка, иначе 503; без auth |
| `/status` | GET | SHA, путь, ready/degraded/syncing, время попытки/успеха/публикации, безопасное описание ошибки |
| `/webhook` | POST | 202 означает принятие/объединение сигнала, не завершённую синхронизацию; 503 после остановки worker |
| `/metrics` | GET | Prometheus |

Status/webhook/metrics защищены Basic либо Bearer, если auth настроена. Неполная Basic-пара и одновременный Basic+Bearer запрещены. Без auth эти endpoint доступны всем, кто имеет сетевой доступ к bind-адресу. Встроенного TLS нет; используйте ограниченную сеть или TLS proxy. `/webhook` — ручной trigger, не реализация подписей GitHub/GitLab.

Ready означает наличие данных, а не свежесть remote. После сетевого сбоя ready остаётся true, degraded становится true. При restart готовность восстанавливается из `current`; `last_success` остаётся нулевым до успешной проверки remote в новом процессе.

Метрики:
- `git_sync_attempts_total{outcome="published|unchanged|error"}`;
- `git_sync_duration_seconds`;
- `git_sync_last_success_timestamp_seconds`;
- `git_sync_last_publication_timestamp_seconds`;
- `git_sync_ready` и стандартные process/Go metrics.

Labels не содержат URL, commit messages, SHA или email. Детали SHA доступны через status и журнал публикации. Описание ошибки в status ограничено этапом и не раскрывает raw transport response/credentials.

## Docker

```sh
docker build -t git-sync:snapshots .
mkdir -p ./data
# Права volume должны позволять UID/GID 65532 создавать root.
sudo chown 65532:65532 ./data
docker run --rm --name git-sync --read-only \
  -v "$PWD/data:/git" -p 127.0.0.1:8080:8080 \
  -e GITSYNC_REPOSITORY_URL=https://git.example.org/team/config.git \
  -e GITSYNC_LOCAL_PATH=/git/root \
  git-sync:snapshots
```

Образ содержит статический бинарник и CA certificates, запускается как UID/GID 65532. Секретов в build ARG нет. Потребителю подключайте тот же volume только для чтения. В Kubernetes используйте отдельный root внутри volume, `runAsUser/runAsGroup/fsGroup: 65532`; не монтируйте разрешённую версию через `subPath`, ожидая переключения на новую. Полный deployment и семантику конкретного storage class проверьте отдельно.

## Миграция и проверки

Это несовместимое изменение файлового контракта относительно v1. Старый `local-path` не переиспользуется автоматически. Порядок миграции:
1. Создать новый выделенный root/volume и запустить snapshot-версию рядом со старой.
2. Проверить ready/status, нужную ветку/SHA, содержимое и права.
3. Переключить потребителя на контракт снимков с фиксацией версии на время чтения.
4. Проверить обновление и отказ Git-сервера; затем остановить прежний процесс.
5. Сохранить прежний volume/образ для ручного возврата до окончания проверки.

Переименованные метрики требуют обновления dashboards/alerts. Старые public Go-пакеты `git`/`logger` и незавершённые api/daemon удалены: это приложение, а не совместимая библиотека.

```sh
go vet ./...
go test -race -count=2 -coverprofile=coverage.out ./...
go test -run '^$' -bench BenchmarkUnchangedSnapshot -benchmem ./internal/snapshot
govulncheck ./...
```

После загрузки Go-зависимостей обычные тесты используют только локальные репозитории и HTTP-серверы. Проверяются ветки, force-push, одинаковое дерево/новый SHA, no-op, отказы публикации и auth, recovery, блокировка root, целостность закреплённого снимка, метрики и SIGTERM реального subprocess.
