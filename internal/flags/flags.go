package flags

import (
	"flag"
	"fmt"
	"git-sync/logger"
	"net/url"
	"os"
	"time"
)

// FlagSet представляет набор флагов командной строки.
type ConsoleFlags struct {
	FlagSet *flag.FlagSet
}

// NewConsoleFlags создает новый набор флагов командной строки.
func NewConsoleFlags(name string) *ConsoleFlags {
	flags, err := ParseFlags(name)
	if err != nil {
		logger.GetLogger().Fatal(err)
		return nil
	}
	return flags
}

// GetString возвращает значение флага строкового типа.
func (f *ConsoleFlags) GetString(name string) string {
	return f.FlagSet.Lookup(name).Value.(flag.Getter).Get().(string)
}

// GetDuration возвращает значение флага типа Duration.
func (f *ConsoleFlags) GetDuration(name string) time.Duration {
	return f.FlagSet.Lookup(name).Value.(flag.Getter).Get().(time.Duration)
}

// ParseFlags инициализирует флаги с помощью набора флагов командной строки.
func ParseFlags(name string) (*ConsoleFlags, error) {

	flagSet := flag.NewFlagSet(name, flag.ExitOnError)
	remoteURL := flagSet.String("remote-url", getEnv("GITSYNC_REMOTE_URL", ""), "URL удаленного репозитория")
	remoteBranch := flagSet.String("remote-branch", getEnv("GITSYNC_BRANCH", ""), "Ветка удаленного репозитория")
	localPath := flagSet.String("local-path", getEnv("GITSYNC_LOCAL_PATH", ""), "Путь к локальной папке")
	gitUser := flagSet.String("git-user", getEnv("GITSYNC_USER", ""), "Учетная запись")
	gitToken := flagSet.String("git-token", getEnv("GITSYNC_TOKEN", ""), "Токен авторизации")
	syncInterval := flagSet.Duration("sync-interval", getEnvDuration("GITSYNC_INTERVAL", 30*time.Second), "Интервал обновления репозитория")

	flagSet.Parse(os.Args[1:])

	if err := validateRemoteURL(*remoteURL); err != nil {
		return nil, err
	}

	if err := validateLocalPath(*localPath); err != nil {
		return nil, err
	}

	if err := validateSyncInterval(*syncInterval); err != nil {
		return nil, err
	}

	if *remoteBranch == "" {
		fmt.Printf("Branch is not set\n")
	}

	if *gitUser == "" {
		fmt.Printf("User is not set\n")
	}

	if *gitToken == "" {
		fmt.Printf("Token is not set\n")
	}

	return &ConsoleFlags{FlagSet: flagSet}, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию, если переменная не установлена.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvDuration возвращает значение переменной окружения в формате time.Duration или значение по умолчанию, если переменная не установлена или имеет некорректный формат.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func validateRemoteURL(remoteURL string) error {
	_, err := url.Parse(remoteURL)
	if err != nil {
		return fmt.Errorf("неверный формат URL-ссылки: %s", err)
	}
	return nil
}

func validateLocalPath(localPath string) error {
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("указанный путь не существует: %s", localPath)
	}
	return nil
}

func validateSyncInterval(syncInterval time.Duration) error {
	if syncInterval <= 0 {
		return fmt.Errorf("интервал синхронизации должен быть положительным")
	}
	return nil
}
