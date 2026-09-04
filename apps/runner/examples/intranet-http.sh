#!/usr/bin/env bash
# @name: Intranet HTTP Health Check
# @schedule: */5 * * * *
# @timeout: 30s
# @description: Ping internal health check endpoint
#
# Example: intranet HTTP check. Create a Runner Job with slug `intranet-http`.
# Place this file at ~/.beep-runner/jobs/intranet-http.sh (chmod +x).
#
# Optional job config on the server: { "target_url": "http://10.0.0.5/health" }
# Scripts can log with stdout (captured by the runner) or POST to $BEEP_LOG_URL.

set -euo pipefail

URL="${BEEP_CONFIG_TARGET_URL:-http://127.0.0.1/health}"

echo "checking ${URL}"

if command -v curl >/dev/null 2>&1; then
  http_code="$(curl -sS -o /tmp/beep-uptime-body -w '%{http_code}' --max-time 10 "$URL")"
  echo "status ${http_code}"
  if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 400 ]; then
    exit 0
  fi
  exit 2
fi

echo "curl is not installed"
exit 1
