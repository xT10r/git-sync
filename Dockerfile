# Copyright 2024 Aleksey Dobshikov
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     https://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Stage 1: Build the binary
FROM golang:1.21 as builder
WORKDIR /src
COPY . /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o git-sync ./cmd/

# Stage 2: Create the final image
FROM alpine:3.20

ARG GITSYNC_REPOSITORY_URL="" \
    GITSYNC_REPOSITORY_BRANCH="main" \
    GITSYNC_LOCAL_PATH="/git" \
    GITSYNC_INTERVAL="30s" \
    GITSYNC_HTTP_SERVER_ADDR="0.0.0.0:8080" \
    GITSYNC_HTTP_AUTH_USERNAME="" \
    GITSYNC_HTTP_AUTH_PASSWORD="" \
    GITSYNC_HTTP_AUTH_TOKEN="" \
    GITSYNC_REPOSITORY_USER="" \
    GITSYNC_REPOSITORY_TOKEN=""

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /src/git-sync /app

ENV GITSYNC_REPOSITORY_URL=${GITSYNC_REPOSITORY_URL} \
    GITSYNC_REPOSITORY_BRANCH=${GITSYNC_REPOSITORY_BRANCH} \
    GITSYNC_LOCAL_PATH=${GITSYNC_LOCAL_PATH} \
    GITSYNC_INTERVAL=${GITSYNC_INTERVAL} \
    GITSYNC_HTTP_SERVER_ADDR=${GITSYNC_HTTP_SERVER_ADDR} \
    GITSYNC_HTTP_AUTH_USERNAME=${GITSYNC_HTTP_AUTH_USERNAME} \
    GITSYNC_HTTP_AUTH_PASSWORD=${GITSYNC_HTTP_AUTH_PASSWORD} \
    GITSYNC_HTTP_AUTH_TOKEN=${GITSYNC_HTTP_AUTH_TOKEN} \
    GITSYNC_REPOSITORY_USER=${GITSYNC_REPOSITORY_USER} \
    GITSYNC_REPOSITORY_TOKEN=${GITSYNC_REPOSITORY_TOKEN}

VOLUME ["${GITSYNC_LOCAL_PATH}"]

ENTRYPOINT ["/app/git-sync"]
