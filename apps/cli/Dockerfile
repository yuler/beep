# syntax=docker/dockerfile:1
#
# Build from monorepo root:
#   docker build -f apps/cli/Dockerfile -t beep .
#

ARG GO_VERSION=1.27
FROM docker.io/library/golang:${GO_VERSION}-alpine AS build

WORKDIR /app

COPY apps/cli/go.mod ./
RUN go mod download

COPY apps/cli/ ./

ARG APP_VERSION=0.0.0
ARG GIT_REVISION=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X beep/internal/version.Version=${APP_VERSION} -X beep/internal/version.GitCommit=${GIT_REVISION}" \
    -trimpath \
    -o /app/bin/beep \
    ./main.go

FROM docker.io/library/alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata curl bind-tools

RUN addgroup -S beep && adduser -S beep -G beep

COPY --from=build /app/bin/beep /usr/local/bin/beep

ARG APP_VERSION
ARG GIT_REVISION
LABEL org.opencontainers.image.title="beep" \
      org.opencontainers.image.description="Command-line interface and self-hosted runner for Beep" \
      org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${GIT_REVISION}"

USER beep
WORKDIR /home/beep

ENTRYPOINT ["/usr/local/bin/beep"]
CMD ["runner", "up"]
