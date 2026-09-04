# syntax=docker/dockerfile:1
#
# Build from monorepo root:
#   docker build -f apps/runner/Dockerfile -t beep-runner .
#

ARG GO_VERSION=1.24
FROM docker.io/library/golang:${GO_VERSION}-alpine AS build

WORKDIR /app

COPY apps/runner/go.mod ./
RUN go mod download

COPY apps/runner/ ./

ARG APP_VERSION=0.0.0
ARG GIT_REVISION=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X beep-runner/internal/version.Version=${APP_VERSION} -X beep-runner/internal/version.GitCommit=${GIT_REVISION}" \
    -trimpath \
    -o /app/bin/beep-runner \
    ./main.go

FROM docker.io/library/alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata curl bind-tools

RUN addgroup -S beep && adduser -S beep -G beep

COPY --from=build /app/bin/beep-runner /usr/local/bin/beep-runner

ARG APP_VERSION
ARG GIT_REVISION
LABEL org.opencontainers.image.title="beep-runner" \
      org.opencontainers.image.description="Self-hosted workspace runner for Beep" \
      org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${GIT_REVISION}"

USER beep
WORKDIR /home/beep

ENTRYPOINT ["/usr/local/bin/beep-runner"]
CMD ["run"]
