#!/usr/bin/env bash
set -euo pipefail

HOST=beep.yuler.cc
USER=root
SERVICE_NAME=core

# SSH to server and find the core container
# Match container names like <project>-<service>-<replica> (Dokploy), e.g. sides-beep-hhjmh1-core-1
CONTAINER_ID=$(
  ssh -o StrictHostKeyChecking=accept-new "$USER@$HOST" "docker ps --format '{{.ID}}\t{{.Names}}' | awk -F'\t' -v svc=\"${SERVICE_NAME}\" '\$2 ~ \"-\" svc \"-[0-9]+\$\" {print \$1; exit}'"
)

if [ -z "$CONTAINER_ID" ]; then
  echo "Error: No container found matching service '$SERVICE_NAME'"
  echo "Available containers:"
  ssh -o StrictHostKeyChecking=accept-new "$USER@$HOST" "docker ps --format '{{.Names}}'"
  exit 1
fi

echo "Found container: $CONTAINER_ID"
echo "Connecting to container..."

# Connect to container
ssh -t -o StrictHostKeyChecking=accept-new "$USER@$HOST" "docker exec -it $CONTAINER_ID ./bin/rails console"