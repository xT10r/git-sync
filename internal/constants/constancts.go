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

package constants

const (

	// Имена флагов
	FlagRepoUrl                string = "repo-url"
	FlagRepoBranch             string = "repo-branch"
	FlagRepoAuthUser           string = "repo-user"
	FlagRepoAuthToken          string = "repo-token"
	FlagLocalPath              string = "local-path"
	FlagSyncInterval           string = "sync-interval"    // 30 секунд
	FlagHttpServerAddr         string = "http-server-addr" // "0.0.0.0:8080"
	FlagHttpServerAuthUsername string = "http-auth-username"
	FlagHttpServerAuthPassword string = "http-auth-password"
	FlagHttpServerAuthToken    string = "http-auth-token"

	// Имена переменных окружения
	EnvRepoUrl                   string = "GITSYNC_REPOSITORY_URL"
	EnvRepoBranch                string = "GITSYNC_REPOSITORY_BRANCH"
	EnvRepoAuthUser              string = "GITSYNC_REPOSITORY_USER"
	EnvRepoAuthToken             string = "GITSYNC_REPOSITORY_TOKEN"
	EnvLocalPath                 string = "GITSYNC_LOCAL_PATH"
	EnvSyncInterval              string = "GITSYNC_INTERVAL"
	EnvHttpServerAddr            string = "GITSYNC_HTTP_SERVER_ADDR"
	EnvHttpServerAuthUsername    string = "GITSYNC_HTTP_AUTH_USERNAME"
	EnvHttpServerAuthPassword    string = "GITSYNC_HTTP_AUTH_PASSWORD"
	EnvHttpServerAuthBearerToken string = "GITSYNC_HTTP_AUTH_TOKEN"
)
