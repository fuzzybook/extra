#!/bin/bash
set -e

# Il pepper delle password non ha un valore di default: senza, l'avvio si ferma.
: "${PASSWORD_PEPPER:?PASSWORD_PEPPER non impostato}"

# Genera dinamicamente .secrets.local.cue partendo dalle variabili d'ambiente
cat <<EOF > /app/.secrets.local.cue
PasswordPepper: "${PASSWORD_PEPPER}"
LiveKitURL: "${LIVEKIT_URL:-ws://livekit:7880}"
LiveKitAPIKey: "${LIVEKIT_API_KEY:-devkey}"
LiveKitAPISecret: "${LIVEKIT_API_SECRET:-secret}"
EOF

echo "==> .secrets.local.cue configurato con successo."
echo "==> Avvio del backend Encore sulla porta 4000..."

exec encore run --listen 0.0.0.0:4000 --browser never --watch=false
