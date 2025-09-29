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

// Package version provides version information for the application.
package version

var (
	// Version is the semantic version of the application (e.g., v1.0.0).
	Version = "dev"

	// Commit is the git commit hash of the application build.
	Commit = "none"

	// Date is the build date in UTC RFC3339 format.
	Date = "unknown"

	// Dirty indicates if there were uncommitted changes at build time.
	Dirty = "false"
)

// String returns a formatted string containing all version information.
func String() string {
	return Version + " (" + Commit + ", " + Date + ", dirty=" + Dirty + ")"
}
