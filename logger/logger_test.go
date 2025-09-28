// Copyright 2025 Aleksey Dobshikov
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

package logger

import (
	"bytes"
	"testing"
)

func TestGlobalLoggerFunctions(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	GetLogger().SetOutput(&buf)

	// Test Info function
	Info("Test info message")
	if buf.String() == "" {
		t.Error("Info function did not write to log")
	}

	// Clear buffer
	buf.Reset()

	// Test Warning function
	Warning("Test warning message")
	if buf.String() == "" {
		t.Error("Warning function did not write to log")
	}

	// Clear buffer
	buf.Reset()

	// Test Error function
	Error("Test error message")
	if buf.String() == "" {
		t.Error("Error function did not write to log")
	}

	// Clear buffer
	buf.Reset()

	// Test Debug function when debug is disabled
	SetDebug(false)
	Debug("Test debug message")
	if buf.String() != "" {
		t.Error("Debug function wrote to log when debug mode was disabled")
	}

	// Test Debug function when debug is enabled
	SetDebug(true)
	Debug("Test debug message")
	if buf.String() == "" {
		t.Error("Debug function did not write to log when debug mode was enabled")
	}
}

func TestIsDebug(t *testing.T) {
	// Reset debug mode to false at the beginning of the test
	SetDebug(false)

	// Test debug mode is initially false
	t.Logf("Initial debug mode value: %v", IsDebug())
	if IsDebug() {
		t.Error("Debug mode should be false by default")
	}

	// Enable debug mode
	SetDebug(true)
	t.Logf("Debug mode after SetDebug(true): %v", IsDebug())
	if !IsDebug() {
		t.Error("Debug mode should be true after calling SetDebug(true)")
	}

	// Disable debug mode
	SetDebug(false)
	t.Logf("Debug mode after SetDebug(false): %v", IsDebug())
	if IsDebug() {
		t.Error("Debug mode should be false after calling SetDebug(false)")
	}
}
