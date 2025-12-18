#!/usr/bin/env bash
# Restore `prometheus.yml` from `prometheus-datasource.yml` (makes a copy)
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT_DIR/monitoring/grafana/provisioning/datasources/prometheus-datasource.yml"
#DST="$ROOT_DIR/monitoring/grafana/provisioning/datasources/prometheus.yml"

if [ ! -f "$SRC" ]; then
  echo "Source file not found: $SRC"
  exit 1
fi

cp -v "$SRC" "$DST"
echo "Restored $DST from $SRC"
