#!/usr/bin/env bash
#
# This is a convenience script that takes a list of testnet manifests
# as arguments and runs each one of them sequentially. If a testnet
# fails, the container logs are dumped to stdout along with the testnet
# manifest, but the remaining testnets are still run.
#
# This is mostly used to run generated networks in nightly CI jobs.
#

set -euo pipefail

if [[ $# == 0 ]]; then
	echo "Usage: $0 [MANIFEST...]" >&2
	exit 1
fi

FAILED=()
EVENTS_PID=""

stop_events() {
	if [[ -n "$EVENTS_PID" ]]; then
		kill "$EVENTS_PID" 2>/dev/null || true
		wait "$EVENTS_PID" 2>/dev/null || true
		EVENTS_PID=""
	fi
}
trap stop_events EXIT

for MANIFEST in "$@"; do
	START=$SECONDS
	echo "==> Running testnet: $MANIFEST"
	echo "==> Manifest:"
	cat "$MANIFEST"
	DIAGNOSTICS="${E2E_LOG_DIR:-logs}/$(basename "$MANIFEST" .toml)-$(date -u +%Y%m%dT%H%M%SZ)-$$"
	mkdir -p "$DIAGNOSTICS"
	cp "$MANIFEST" "$DIAGNOSTICS/manifest.toml"
	{ docker version; docker compose version; } >"$DIAGNOSTICS/docker-version.txt" 2>&1 || true
	# Stream events before starting the runner: Engine only retains a limited
	# event history, and successful runs remove their own containers.
	docker events --since "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --filter label=e2e --format '{{json .}}' \
		>"$DIAGNOSTICS/docker-events.jsonl" 2>"$DIAGNOSTICS/docker-events.stderr" &
	EVENTS_PID=$!
	echo "==> Diagnostics: $DIAGNOSTICS"

	if ! ./build/runner -f "$MANIFEST" 2>&1 | tee "$DIAGNOSTICS/runner.log"; then
		echo "==> Testnet failed"
		FAILED+=("$MANIFEST")
		# Capture state before cleanup. Diagnostic failures must not prevent
		# collecting the remaining evidence or running the next manifest.
		docker ps -a --filter label=e2e --no-trunc >"$DIAGNOSTICS/containers.txt" 2>&1 || true
		IDS=$(docker ps -aq --filter label=e2e) || IDS=""
		if [[ -n "$IDS" ]]; then
			# Container IDs contain no whitespace; splitting selects each ID.
			# shellcheck disable=SC2086
			docker inspect $IDS >"$DIAGNOSTICS/inspect.json" 2>&1 || true
		fi

		echo "==> Dumping container logs for $MANIFEST..."
		./build/runner -f "$MANIFEST" logs 2>&1 | tee "$DIAGNOSTICS/nodes.log" || true

		echo "==> Cleaning up failed testnet $MANIFEST..."
		./build/runner -f "$MANIFEST" cleanup || true
	fi
	stop_events

	echo "==> Completed testnet $MANIFEST in $(( SECONDS - START ))s"
	echo ""
done

if [[ ${#FAILED[@]} -ne 0 ]]; then
	echo "${#FAILED[@]} testnets failed:"
	for MANIFEST in "${FAILED[@]}"; do
		echo "- $MANIFEST"
	done
	exit 1
else
	echo "All testnets successful"
fi
