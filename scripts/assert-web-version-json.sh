#!/usr/bin/env bash
set -euo pipefail

version_json="${1:-apps/web/.output/public/version.json}"

if [[ ! -f "$version_json" ]]; then
	echo "missing $version_json" >&2
	exit 1
fi

if ! grep -qE '"version"[[:space:]]*:[[:space:]]*"[^"]+"' "$version_json"; then
	echo "missing version in $version_json" >&2
	exit 1
fi

if ! grep -qE '"gitHash"[[:space:]]*:[[:space:]]*"[^"]+"' "$version_json"; then
	echo "missing gitHash in $version_json" >&2
	exit 1
fi
