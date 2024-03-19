# syntax=docker/dockerfile:1
#
# STAGE 1: prepare
#
FROM golang:1.22.1 as prepare
 
WORKDIR /app
 
COPY vendor .
 
#
# STAGE 2: build
#
FROM prepare AS build
 
COPY cmd cmd
COPY internal internal
COPY vendor vendor
COPY web web
COPY go.mod .
COPY go.sum .

ARG VERSION
ARG REPO_URL
ARG SHA_COMMIT
ARG BUILD_TIME
RUN CGO_ENABLED=0 go build -mod vendor -ldflags "-s -w -X tabby-syncd/internal/buildinfo.Version=$VERSION -X tabby-syncd/internal/buildinfo.RepoUrl=$REPO_URL -X tabby-syncd/internal/buildinfo.ShaCommit=$SHA_COMMIT -X tabby-syncd/internal/buildinfo.BuildTime=$BUILD_TIME" -o /tabby-sync-server cmd/tabby-sync-server/main.go
 
#
# STAGE 3: run
#
FROM scratch as run
 
COPY --from=build /tabby-sync-server /tabby-sync-server
RUN --mount=from=busybox:uclibc,dst=/usr/ ["busybox", "mkdir", "-m777", "/data"] # Create empty /data directory in scratch container

ENV HTTP_ADDRESS=":8080"
ENV DB_FILE="/data/configstore.db"
ENV ADMIN_KEY=""
# LOG_LEVEL values: DEBUG|INFO|WARN|ERROR
ENV LOG_LEVEL="INFO"

EXPOSE 8080
ENTRYPOINT ["/tabby-sync-server"]
