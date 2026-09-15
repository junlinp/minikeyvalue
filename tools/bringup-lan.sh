#!/bin/bash
# Start 3 nginx volumes + mkv master for LAN access.
# Usage: ADVERTISE=192.168.0.107 ./tools/bringup-lan.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ADVERTISE="${ADVERTISE:-192.168.0.107}"
DATA="${MKV_DATA:-/mnt/d/minikeyvalue/data}"
cd "$ROOT"

chmod +x mkv-linux volume 2>/dev/null || true
mkdir -p "$DATA/volume1/tmp" "$DATA/volume2/tmp" "$DATA/volume3/tmp" "$DATA/indexdb"
chmod -R 777 "$DATA"

pkill -f "nginx: master process" 2>/dev/null || true
pkill -f "./mkv-linux" 2>/dev/null || true
sleep 0.5

PORT=3001 ./volume "$DATA/volume1/" &
PORT=3002 ./volume "$DATA/volume2/" &
PORT=3003 ./volume "$DATA/volume3/" &
sleep 1

exec ./mkv-linux \
  -port 3000 \
  -volumes localhost:3001,localhost:3002,localhost:3003 \
  -db "$DATA/indexdb" \
  -advertise "$ADVERTISE" \
  -replicas 3 \
  server
