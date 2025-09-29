# syntax=docker/dockerfile:1.7
# Copyright 2025 Aleksey Dobshikov
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Global ARG declarations (available in all stages)
ARG RUNTIME_FAMILY=alpine
ARG RUNTIME_IMAGE=alpine:3.21

# ---------- Builder ----------
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates gcc git musl-dev

WORKDIR /src
COPY . .

# Версия для ldflags (прокидываем из CI)
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG DIRTY=false

# Multi-arch: берём целевые значения от buildx
ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false \
      -ldflags "\
        -s -w \
        -X 'git-sync/internal/version.Version=${VERSION}' \
        -X 'git-sync/internal/version.Commit=${COMMIT}' \
        -X 'git-sync/internal/version.Date=${DATE}' \
        -X 'git-sync/internal/version.Dirty=${DIRTY}' \
      " \
      -o /out/git-sync ./cmd

# отдадим рантайму готовый bundle корневых сертификатов, чтобы не ставить пакеты
RUN cp /etc/ssl/certs/ca-certificates.crt /out/ca-certificates.crt

# ---------- Runtime ----------
FROM ${RUNTIME_IMAGE} AS runtime

# OCI labels. See: https://github.com/opencontainers/image-spec/blob/main/annotations.md
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG RUNTIME_IMAGE
ARG RUNTIME_FAMILY

LABEL org.opencontainers.image.title="git-sync" \
      org.opencontainers.image.description="Git repository synchronization service with HTTP API/metrics." \
      org.opencontainers.image.url="https://github.com/xt10r/git-sync" \
      org.opencontainers.image.source="https://github.com/xt10r/git-sync" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${DATE}" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="xt10r" \
      org.opencontainers.image.base.name="${RUNTIME_IMAGE}"

# Копируем bundle сертификатов из билдера (без пакетного менеджера)
COPY --from=builder /out/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app
COPY --from=builder --chown=1001:1001 /out/git-sync /app/git-sync

USER 1001:1001

ARG GITSYNC_REPOSITORY_URL="" \
    GITSYNC_REPOSITORY_BRANCH="main" \
    GITSYNC_LOCAL_PATH="/git" \
    GITSYNC_INTERVAL="30s" \
    GITSYNC_HTTP_SERVER_ADDR="0.0.0.0:8080" \
    GITSYNC_REPOSITORY_USER="" \
    GITSYNC_DEBUG="false"
ENV GITSYNC_REPOSITORY_URL=${GITSYNC_REPOSITORY_URL} \
    GITSYNC_REPOSITORY_BRANCH=${GITSYNC_REPOSITORY_BRANCH} \
    GITSYNC_LOCAL_PATH=${GITSYNC_LOCAL_PATH} \
    GITSYNC_INTERVAL=${GITSYNC_INTERVAL} \
    GITSYNC_HTTP_SERVER_ADDR=${GITSYNC_HTTP_SERVER_ADDR} \
    GITSYNC_REPOSITORY_USER=${GITSYNC_REPOSITORY_USER} \
    GITSYNC_DEBUG=${GITSYNC_DEBUG}

VOLUME ["${GITSYNC_LOCAL_PATH}"]
 
ENTRYPOINT ["/app/git-sync"]