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

package main

import (
	"git-sync/logger"
)

func main() {
	// Example of using global logger functions
	logger.Info("Application started")

	logger.Warning("This is a warning message")

	// This error will be logged but won't stop the application
	logger.Error("This is an error message")

	// Debug messages won't be shown by default
	logger.Debug("This debug message won't be shown")

	// Enable debug mode
	logger.SetDebug(true)

	// Now debug messages will be shown
	logger.Debug("This debug message will be shown")

	// You can also check if debug mode is enabled
	if logger.IsDebug() {
		logger.Info("Debug mode is enabled")
	}

	logger.Info("Application finished")
}
