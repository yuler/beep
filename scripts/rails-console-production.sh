#!/usr/bin/env bash
# Open a Rails console inside the production core container.
#
# Connects to the Dokploy server over SSH, finds the running core container by
# its <project>-<service> names, and runs `./bin/rails console` inside it.
set -euo pipefail

HOST=beep.yuler.cc
USER=root
PROJECT_NAME=beep
SERVICE_NAME=core

# SSH to server and find the core container
# Match Dokploy names like <stack>-<project>-<service>-<replica>, e.g. sides-beep-hhjmh1-core-1
CONTAINER_ID=$(
  ssh -o StrictHostKeyChecking=accept-new "$USER@$HOST" "docker ps --format '{{.ID}}\t{{.Names}}' | awk -F'\t' -v proj=\"${PROJECT_NAME}\" -v svc=\"${SERVICE_NAME}\" '\$2 ~ proj \"-.*-\" svc \"-[0-9]+\$\" {print \$1; exit}'"
)

if [ -z "$CONTAINER_ID" ]; then
  echo "Error: No container found matching '${PROJECT_NAME}-*-${SERVICE_NAME}'"
  echo "Available containers:"
  ssh -o StrictHostKeyChecking=accept-new "$USER@$HOST" "docker ps --format '{{.Names}}'"
  exit 1
fi

echo "Found container: $CONTAINER_ID"
echo "Connecting to container..."

# Connect to container
ssh -t -o StrictHostKeyChecking=accept-new "$USER@$HOST" "docker exec -it $CONTAINER_ID ./bin/rails console"