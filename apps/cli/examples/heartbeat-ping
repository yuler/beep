#!/usr/bin/env bash
# Example: ping a cloud Heartbeat beeper after a local job succeeds.
# Official Heartbeat evaluation still runs in Core. This script is only a ping source.
#
# Job slug: heartbeat-ping
# Config: { "ping_url": "https://core.example.com/api/v1/beeper_apps/heartbeat/pings/TOKEN" }

set -euo pipefail

PING_URL="${BEEP_CONFIG_PING_URL:-}"
if [ -z "$PING_URL" ]; then
  echo "BEEP_CONFIG_PING_URL (job config ping_url) is required"
  exit 1
fi

echo "posting heartbeat ping"
curl -sS -X POST "$PING_URL"
echo "ok"
