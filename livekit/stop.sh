#!/bin/bash
set -e

echo "Arresto LiveKit Server locale..."

if command -v docker-compose &> /dev/null; then
    docker-compose down
elif docker compose version &> /dev/null; then
    docker compose down
else
    docker stop local-livekit 2>/dev/null || true
    docker rm local-livekit 2>/dev/null || true
fi

echo "LiveKit Server arrestato."
