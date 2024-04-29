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

package daemon

import (
	"fmt"
	"git-sync/logger"

	"github.com/sevlyar/go-daemon"
)

// TODO: реализовать при необходимости

var pidFile = "/var/run/git-sync.pid"

// Запускает приложение как демон
func Start() {

	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("%v", err))
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
