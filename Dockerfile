# syntax=docker/dockerfile:1
ARG GO_VERSION=1.27.2
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/git-sync ./cmd/ \
    && mkdir -p /out/root

FROM scratch
COPY LICENSE NOTICE /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/git-sync /app/git-sync
COPY --from=builder --chown=65532:65532 /out/root /git
ENV GITSYNC_LOCAL_PATH=/git \
    GITSYNC_REPOSITORY_BRANCH=main \
    GITSYNC_INTERVAL=30s \
    GITSYNC_TIMEOUT=2m \
    GITSYNC_HTTP_SERVER_ADDR=0.0.0.0:8080
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/git-sync"]
