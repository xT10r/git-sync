/*
Функции Start() и Stop() можно вызывать из пакета cmd при старте и остановке приложения.
Пакет daemon использует библиотеку go-daemon для перехода процесса в фон и создания pid-файла.
*/

package daemon

import (
	"git-sync/logger"

	"github.com/sevlyar/go-daemon"
)

var pidFile = "/var/run/git-sync.pid"

// Запускает приложение как демон
func Start() {

	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		logger.GetLogger().Fatal(err)
	}
	if d != nil {
		return
	}

	// Запуск основной логики приложения

}

// Останавливает демон
func Stop() {
	// if err := daemon.Stop(pidFile); err != nil {
	//	logger.Fatal(err)
	//}
}
