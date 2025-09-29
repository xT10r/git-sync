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

package version

import (
	"testing"
)

func TestString(t *testing.T) {
	// Test that String() returns a properly formatted version string
	expected := "dev (none, unknown, dirty=false)"
	actual := String()

	if actual != expected {
		t.Errorf("String() = %q, want %q", actual, expected)
	}

	// Test with different values
	Version = "v1.0.0"
	Commit = "abc123"
	Date = "2024-01-01T00:00:00Z"
	Dirty = "true"

	expected = "v1.0.0 (abc123, 2024-01-01T00:00:00Z, dirty=true)"
	actual = String()

	if actual != expected {
		t.Errorf("String() = %q, want %q", actual, expected)
	}
}
