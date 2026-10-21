#!/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$DIR"

echo "Avvio LiveKit Server locale..."

if command -v docker-compose &> /dev/null; then
    docker-compose up -d
elif docker compose version &> /dev/null; then
    docker compose up -d
else
    echo "docker-compose non rilevato. Avvio con docker run..."
    docker run -d \
      --name local-livekit \
      --restart unless-stopped \
      -p 7880:7880 \
      -p 7881:7881 \
      -p 7882:7882/udp \
      -p 50000-50020:50000-50020/udp \
      -v "$DIR/livekit.yaml:/etc/livekit.yaml:ro" \
      livekit/livekit-server:latest --config /etc/livekit.yaml
fi

echo "LiveKit Server attivo su ws://127.0.0.1:7880 (API Key: devkey, Secret: secret)"
