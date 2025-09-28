# Git-Sync Service

## Общее описание

`Git-Sync Service` обеспечивает синхронизацию удалённого репозитория с локальным.
Любые изменения в локальном репозитории инициируют синхронизацию с удалённым репозиторием.
Необходимость синхронизации определяется путём сравнения хешей файлов (для выявления изменений) и сравнением деревьев файлов удалённого и локального репозиториев.
Сервис предоставляет метрики в формате Prometheus и поддерживает вебхуки для ручного запуска синхронизации.

## Возможности сервиса

- Синхронизация удалённого репозитория с локальным.
- Определение необходимости синхронизации по хешам файлов и сравнению деревьев каталогов локального и удалённого репозиториев.
- Обработка вебхуков для ручной синхронизации.
- Доступ к метрикам через Prometheus.
- Поддержка отладочного логирования.
- Предоставление информации о версии через CLI-флаг и HTTP-эндпоинт.

## Настройка окружения

Для работы с функционалом репозитория необходимо корректно настроить окружение. Ниже приведены инструкции для различных ОС.

### Установка Go

Go — основной язык программирования проекта. Для локальной сборки требуется Go 1.23 или новее.

#### Linux

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL/Fedora
sudo yum install golang
# или на новых версиях
sudo dnf install golang
```

#### macOS

```bash
# Homebrew
brew install go
```

#### Windows

Скачайте и установите Go с официального сайта:
https://golang.org/dl/

### Установка Make

#### Linux

На большинстве дистрибутивов Linux `make` доступен в стандартном пакетном менеджере:

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install make

# CentOS/RHEL/Fedora
sudo yum install make
# или на новых версиях
sudo dnf install make
```

#### macOS

На macOS установите Xcode Command Line Tools (включает `make`):

```bash
xcode-select --install
```

#### Windows

Для Windows 10/11 есть несколько вариантов:

1. **Chocolatey** (рекомендуется):
   ```powershell
   choco install make
   ```

2. **Winget**:
   ```powershell
   winget install ezwinports.make
   ```

3. **WSL** (Windows Subsystem for Linux):
   Установите WSL и следуйте инструкциям для Linux.

### Установка Docker

Docker необходим для сборки и запуска сервиса.

#### Linux

Следуйте официальной инструкции для вашего дистрибутива:
https://docs.docker.com/engine/install/

Также установите Docker Compose:
https://docs.docker.com/compose/install/

#### macOS

Скачайте Docker Desktop for Mac:
https://docs.docker.com/docker-for-mac/install/

#### Windows

Скачайте Docker Desktop for Windows:
https://docs.docker.com/docker-for-windows/install/

Примечание: на Windows рекомендуется использовать WSL 2 backend для лучшей производительности.

### Сборка проекта

После установки `make` и Docker вы можете собрать проект так:

```bash
# Сборка с использованием Docker (рекомендуется)
make build

# Локальная сборка (для разработки)
make build-local

# Сборка исполняемого файла Windows (.exe)
make build-windows

# Сборка для нескольких платформ
make build-all

# Запуск тестов
make test

# Запуск тестов с отчётом покрытия
make cover

# Очистка артефактов сборки
make clean
```

Для пользователей Windows без `make` предусмотрен PowerShell-скрипт:

```powershell
# Сборка исполняемого файла Windows
.\scripts\build-windows.ps1
```

Дополнительные опции сборки:

```bash
make help
```

## Конфигурация и параметры

Параметры сервиса задаются флагами командной строки, переменными окружения и конфигурационным файлом.

### Просмотр параметров конфигурации

Для вывода всех доступных опций с описаниями, переменными окружения и флагами используйте:

```bash
git-sync config-help
```

### Генерация конфигурационного файла

Для генерации примерного конфигурационного файла выполните:

```bash
git-sync gen-config
```

По умолчанию будет создан `config.yaml` в текущей директории. Чтобы указать иной путь, используйте флаг `-output`:

```bash
git-sync gen-config -output /path/to/config.yaml
```

### Использование конфигурационного файла

Сервис автоматически ищет конфигурационный файл в следующих местах:

1. Текущая директория (`./config.yaml`)
2. Директория конфигурации (`./config/config.yaml`)
3. Домашняя директория пользователя (`~/.gitsync/config.yaml`)

Пример формата (YAML):

```yaml
debug: false

# Настройки по умолчанию
defaults:
  gitlab:
    repoauth:
      user: gitlab-user
      token: your-default-gitlab-token  # Токен по умолчанию для всех репозиториев
      token_file: ""                     # Путь к файлу токена по умолчанию (опционально)
    repobranch: main
  sync:
    interval: 30

http_server:
  addr: "0.0.0.0:8080"
  auth:
    username: "admin"
    password: "your-password"

repositories:
  # Основной репозиторий — использует большинство настроек по умолчанию
  main-repo:
    gitlab:
      repourl: "https://gitlab.example.com/username/repository1.git"
      # repoauth не указан — будет взят из defaults
    sync:
      local_path: "./repo1"

  # Репозиторий с изменённым интервалом и веткой
  secondary-repo:
    gitlab:
      repoauth:
        token: "your-gitlab-token-2"     # Собственный токен
      repobranch: "develop"              # Переопределяем ветку
      repourl: "https://gitlab.example.com/username/repository2.git"
    sync:
      interval: 60                       # Переопределяем интервал
      local_path: "./repo2"

  # Репозиторий с другим пользователем, но токеном по умолчанию
  external-repo:
    gitlab:
      repoauth:
        user: "different-user"           # Переопределяем только пользователя
        # token будет взят из defaults
      repourl: "https://gitlab.example.com/anotheruser/repository.git"
    sync:
      local_path: "./external-repo"

  # Репозиторий с полностью кастомными настройками
  custom-repo:
    gitlab:
      repoauth:
        token: "your-gitlab-token-4"
        user: "custom-user"
      repobranch: "feature-branch"
      repourl: "https://gitlab.example.com/custom/repo.git"
    sync:
      interval: 300                      # 5 минут
      local_path: "./custom-repo"
```

### Как работают настройки по умолчанию

Секция `defaults` задаёт общие параметры для всех репозиториев. Если в конфигурации конкретного репозитория параметр не указан, будет использовано значение из `defaults`.

Например:
- Если не указан `repoauth.token`, используется `defaults.gitlab.repoauth.token`.
- Если не указан `repoauth.user`, используется `defaults.gitlab.repoauth.user`.
- Если не указан `repoauth.token_file`, используется `defaults.gitlab.repoauth.token_file`.
- Если не указан `repobranch`, используется `defaults.gitlab.repobranch`.
- Если не указан `sync.interval`, используется `defaults.sync.interval`.

Если в `defaults` указан `token_file`, он будет использован для всех репозиториев без собственного `token_file`. Токен считывается из файла и используется для аутентификации.

### Безопасная работа с чувствительными данными

По соображениям безопасности чувствительные данные (токены, пароли) не рекомендуется хранить непосредственно в конфигурации или передавать через переменные окружения в продакшене. Рекомендуемые подходы:

#### Среда разработки

Для удобства допускается прямое задание значений:

```bash
GITSYNC_REPOSITORY_TOKEN=your-token GITSYNC_HTTP_AUTH_USERNAME=admin GITSYNC_HTTP_AUTH_PASSWORD=your-password git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local-path=./repo
```

Либо в конфигурационном файле (только для разработки):
```yaml
gitlab:
  repourl: "https://gitlab.example.com/username/repository.git"
  repobranch: "main"
  repoauth:
    user: "gitlab-user"
    token: "your-gitlab-token"

http_server:
  addr: "0.0.0.0:8080"
  auth:
    username: "admin"
    password: "your-password"
```

#### Производственная среда

1. **Docker Secrets**:
   ```bash
   echo "your-gitlab-token" > gitlab-token
   echo "your-http-password" > http-password
   docker run -v $(pwd)/gitlab-token:/run/secrets/gitlab-token:ro \
              -v $(pwd)/http-password:/run/secrets/http-password:ro \
              -e GITSYNC_REPOSITORY_TOKEN_FILE=/run/secrets/gitlab-token \
              -e GITSYNC_HTTP_AUTH_PASSWORD_FILE=/run/secrets/http-password \
              git-sync
   ```

2. **Kubernetes Secrets**:
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: git-sync-secrets
   type: Opaque
   data:
     gitlab-token: eW91ci1naXRsYWItdG9rZW4=    # base64
     http-password: eW91ci1odHRwLXBhc3N3b3Jk   # base64
   ---
   apiVersion: v1
   kind: Pod
   metadata:
     name: git-sync
   spec:
     containers:
     - name: git-sync
       image: git-sync
       env:
       - name: GITSYNC_REPOSITORY_TOKEN_FILE
         value: /run/secrets/gitlab-token
       - name: GITSYNC_HTTP_AUTH_PASSWORD_FILE
         value: /run/secrets/http-password
       volumeMounts:
       - name: secrets
         mountPath: /run/secrets
         readOnly: true
     volumes:
     - name: secrets
       secret:
         secretName: git-sync-secrets
   ```

3. **Файл переменных окружения**:
   ```bash
   echo "GITSYNC_REPOSITORY_TOKEN=your-token" > .env.secure
   echo "GITSYNC_HTTP_AUTH_PASSWORD=your-password" >> .env.secure
   chmod 600 .env.secure
   source .env.secure
   git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local-path=./repo
   ```

### Логирование

В проекте принят унифицированный подход к логированию. Вместо вызовов вида `logger.GetLogger().Info()` и `logger.GetLogger().Error()` используется единообразный API глобальных функций `logger.Info()`, `logger.Error()` и т.п. для лучшей поддерживаемости.

#### Отладочное логирование

Отладочный режим включается одним из способов:
1. Флагом командной строки: `--debug`
2. Переменной окружения: `GITSYNC_DEBUG=true`
3. Опцией конфигурации: `debug: true`

При включённом отладочном режиме выводятся дополнительные сообщения, упрощающие диагностику.

Примеры:
```bash
git-sync --debug --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local_path=./repo
```
```bash
GITSYNC_DEBUG=true git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local_path=./repo
```
```yaml
debug: true
gitlab:
  repourl: "https://gitlab.example.com/username/repository.git"
  repobranch: "main"
  repoauth:
    user: "gitlab-user"
    token: "your-gitlab-token"
sync:
  local_path: "./repo"
  interval: 30
http_server:
  addr: "0.0.0.0:8080"
```

### Информация о версии

Доступ к информации о версии предоставляется несколькими способами:

1. **CLI-флаг**: `--version`
   ```bash
   git-sync --version
   ```
2. **HTTP-эндпоинт**: `/version` — подробная информация в формате JSON
   ```bash
   curl http://localhost:8080/version
   ```
3. **Prometheus-метрики**: метрика `git_sync_build_info` содержит сведения о сборке (метки: версия, коммит, дата, dirty-статус).

Состав информации:
- **Version** — семантическая версия (например, v1.0.0)
- **Commit** — хеш git-коммита во время сборки
- **Date** — дата сборки (UTC)
- **Dirty** — наличие несохранённых изменений на момент сборки

### Параметры командной строки и переменные окружения

Каждая переменная окружения сопоставлена соответствующему флагу.

| Аргумент командной строки (флаг) | Переменная окружения                 | Описание                                                          |
|---|---|---|
| `--local-path`                   | `GITSYNC_LOCAL_PATH`                 | Путь к локальному репозиторию.                                    |
| `--repo-url`                     | `GITSYNC_REPOSITORY_URL`             | URL удалённого репозитория.                                       |
| `--repo-branch`                  | `GITSYNC_REPOSITORY_BRANCH`          | Ветка удалённого репозитория.                                     |
| `--repo-user`                    | `GITSYNC_REPOSITORY_USER`            | Пользователь для аутентификации в репозитории.                    |
| `--repo-token`                   | `GITSYNC_REPOSITORY_TOKEN`           | Токен для аутентификации в репозитории.                           |
| `--repo-token-file`              | `GITSYNC_REPOSITORY_TOKEN_FILE`      | Путь к файлу с токеном для аутентификации в репозитории.          |
| `--sync-interval`                | `GITSYNC_INTERVAL`                   | Интервал синхронизации репозитория.                               |
| `--http-server-addr`             | `GITSYNC_HTTP_SERVER_ADDR`           | Адрес и порт HTTP-сервера.                                        |
| `--http-auth-username`           | `GITSYNC_HTTP_AUTH_USERNAME`         | Имя пользователя для аутентификации HTTP-сервера.                 |
| `--http-auth-password`           | `GITSYNC_HTTP_AUTH_PASSWORD`         | Пароль для аутентификации HTTP-сервера.                           |
| `--http-auth-token`              | `GITSYNC_HTTP_AUTH_TOKEN`            | Токен для аутентификации HTTP-сервера.                            |
| `--http-auth-token-file`         | `GITSYNC_HTTP_AUTH_TOKEN_FILE`       | Путь к файлу с токеном для аутентификации HTTP-сервера.           |
| `--debug`                        | `GITSYNC_DEBUG`                      | Включить отладочное логирование.                                  |
| `--version`                      |                                      | Показать информацию о версии.                                     |

### Приоритет конфигурации

Порядок приоритета источников значений:
1. Флаги командной строки
2. Переменные окружения
3. Конфигурационный файл
4. Значения по умолчанию

### Поддержка заголовков лицензии

Проект использует лицензию Apache License 2.0 для всех исходных файлов. Каждый Go-файл должен содержать заголовок лицензии с уведомлением об авторском праве.

#### Инструменты

Мы используем расширение VSCode licenser (`ymotongpoo.licenser`) для автоматического добавления и обновления заголовков лицензии в исходных файлах.

#### Добавление заголовков лицензии

Для добавления заголовка лицензии в файл:

1. Откройте файл в VSCode
2. Откройте палитру команд (`Ctrl+Shift+P` или `Cmd+Shift+P`)
3. Введите "licenser: Insert license header"
4. Нажмите Enter

Licenser автоматически вставит соответствующий заголовок лицензии в верхнюю часть файла.

Заголовки лицензии будут автоматически обновлены с текущим годом. Для получения дополнительной информации см. [Поддержка заголовков лицензии](./LICENSE-HEADER-MAINTENANCE.md).

### Метрики Prometheus

Имена метрик унифицированы для лучшей читаемости:

| Наименование              | Описание                                                                 |
|---|---|
| `git_sync_changes_count`  | Общее количество синхронизаций с изменениями.                            |
| `git_sync_total_count`    | Общее количество синхронизаций.                                          |
| `git_sync_error_total`    | Общее количество ошибок синхронизации.                                   |
| `git_sync_repo_info`      | Информация о синхронизируемом репозитории (метки: имя репозитория, ветка). |
| `git_sync_commit_info`    | Информация о последнем коммите (метки: хеш, автор, email, дата, сообщение). |
| `git_sync_build_info`     | Информация о сборке (метки: версия, коммит, дата, dirty).                |

Примечание: Имена метрик были улучшены для избежания повторения и повышения читаемости. Например, `git_sync_sync_count` был переименован в `git_sync_changes_count` для устранения избыточного слова "sync".

### Веб-интерфейс

Сервис предоставляет расширенную веб-панель на корневом эндпоинте (`/`) с удобным мониторингом и управлением. Панель включает:

- **Общий статус системы**: визуальный индикатор здоровья.
- **Статус синхронизации репозитория**: сведения о последней синхронизации.
- **Кнопка ручной синхронизации**: быстрый запуск синка.
- **Карточка Health**: состояние сервиса + ссылка на подробности.
- **Карточка Metrics**: ключевые метрики + ссылка на полный список Prometheus.
- **Карточка Version**: версия + ссылка на подробный эндпоинт.
- **Список эндпоинтов**: полный перечень доступных API.

Дизайн адаптивный, с понятными индикаторами состояний (OK/Warning/Error) и соответствует лучшим практикам для веб-интерфейсов мониторинга.

### API-эндпоинты

| Эндпоинт    | Метод | Описание                                                                 |
|---|---|---|
| `/status`   | GET   | Базовая информация о состоянии сервиса.                                  |
| `/health`   | GET   | Подробная информация о состоянии (включая Go runtime).                   |
| `/ready`    | GET   | Статус готовности (для Kubernetes probes).                               |
| `/version`  | GET   | Информация о версии сервиса.                                             |
| `/metrics`  | GET   | Метрики Prometheus.                                                      |
| `/webhook`  | POST  | Ручной запуск синхронизации.                                             |
| `/`         | GET   | Расширенная веб-панель мониторинга (dashboard).                          |

Все эндпоинты могут быть защищены аутентификацией:
- Basic (username/password)
- Bearer Token

При включённой аутентификации все эндпоинты требуют валидных учётных данных.

Сервис использует модульную архитектуру с компонентами:
1. **Server Package** — жизненный цикл HTTP-сервера, регистрация маршрутов.
2. **Router Package** — централизованная маршрутизация и middleware.
3. **API Package** — эндпоинты мониторинга и управления.
4. **Handlers Package** — вебхуки, метрики и пр. специализированная логика.

### Примеры использования

**Конфигурационные файлы приложений**: единый источник правды для часто меняющихся конфигураций, требующих синхронизации между инстансами.

**Скрипты развёртывания**: автоматическое обновление и синхронизация скриптов между серверами/средами, чтобы все узлы использовали одинаковые версии.

**Конфигурации серверов**: синхронизация конфигов веб-серверов/БД и др. для единообразной корректной настройки.

**Документация и инструкции**: синхронизация документации и инструкций, чтобы разработчики имели актуальные материалы.