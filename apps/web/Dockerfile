# syntax=docker/dockerfile:1
# check=error=true
#
# Build from monorepo root:
#   docker build -f apps/web/Dockerfile -t beep-web .
#
# Mode B (Compose): VITE_CORE_URL= empty; CORE_INTERNAL_URL=http://core:80
# Mode A (split):   --build-arg VITE_CORE_URL=https://core.example.com

ARG NODE_VERSION=24.17.0
FROM docker.io/library/node:${NODE_VERSION}-alpine AS build

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.14.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json ./apps/web/

RUN pnpm install --filter web... --frozen-lockfile

# Tailwind v4 uses .gitignore to decide what not to scan. Without it, the SSR
# Vite environment picks up extra candidates and the hashed CSS URL 404s.
COPY .gitignore ./
COPY VERSION ./
COPY apps/web ./apps/web

# Empty string → same-origin `/api/v1` (Mode B). Do not omit the ARG or Vite
# falls back to the local Mode A default in src/config.ts.
ARG VITE_CORE_URL=
ENV VITE_CORE_URL=$VITE_CORE_URL

# Rails origin for Nitro /api proxy + SSR fetches (Compose service name).
ARG CORE_INTERNAL_URL=http://core:80
ENV CORE_INTERNAL_URL=$CORE_INTERNAL_URL

# Inlined into the bundle as build info for `logBuildInfo()`.
ARG APP_VERSION
ARG GIT_REVISION
ENV APP_VERSION=$APP_VERSION
ENV GIT_REVISION=$GIT_REVISION

RUN pnpm --filter web build

FROM docker.io/library/node:${NODE_VERSION}-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
# Runtime SSR / proxy target (override in Compose if needed).
ENV CORE_INTERNAL_URL=http://core:80

COPY --from=build /app/apps/web/.output ./.output

ARG APP_VERSION
ARG GIT_REVISION
LABEL org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${GIT_REVISION}"

EXPOSE 3000
USER node
CMD ["node", ".output/server/index.mjs"]
