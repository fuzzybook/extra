# LiveKit Server Locale (Docker)

Questo ambiente consente di eseguire un'istanza locale di **LiveKit Server (SFU WebRTC)** tramite Docker, senza dipendere dalla connessione internet o dai server cloud.

---

## 1. Avvio rapido

Dalla cartella `livekit`:

```bash
cd livekit
./start.sh
```

Oppure tramite Docker Compose:

```bash
docker compose up -d
# oppure: docker-compose up -d
```

Oppure con Docker standard:

```bash
docker run -d \
  --name local-livekit \
  --restart unless-stopped \
  -p 7880:7880 \
  -p 7881:7881 \
  -p 7882:7882/udp \
  -p 50000-50020:50000-50020/udp \
  -v $(pwd)/livekit.yaml:/etc/livekit.yaml:ro \
  livekit/livekit-server:latest --config /etc/livekit.yaml
```

---

## 2. Arresto

```bash
./stop.sh
# oppure: docker compose down
```

---

## 3. Configurazione porte

| Porta | Protocollo | Utilizzo |
|---|---|---|
| **`7880`** | TCP (HTTP/WS) | Segnalazione WebSocket e API Twirp RoomService |
| **`7881`** | TCP | Fallback WebRTC TCP (se UDP è filtrato) |
| **`7882`** | UDP | Canale primario WebRTC ICE |
| **`50000-50020`** | UDP | Range porte multimediali audio/video WebRTC |

---

## 4. Credenziali predefinite

Configurate nel file [`livekit.yaml`](./livekit.yaml):

* **API Key**: `devkey`
* **API Secret**: `secret`
* **URL Locale**: `ws://127.0.0.1:7880` (HTTP: `http://127.0.0.1:7880`)

---

## 5. Configurazione in Encore (Backend locale)

Quando l'app Encore lavora in locale senza essere collegata al cloud di Encore, non occorre usare `encore secret set`. È sufficiente modificare il file [`backend/.secrets.local.cue`](../backend/.secrets.local.cue):

```cue
PasswordPepper: "__generare_con_openssl_rand_hex_32__"
LiveKitURL: "ws://127.0.0.1:7880"
LiveKitAPIKey: "devkey"
LiveKitAPISecret: "secret"
```

Encore ricaricherà automaticamente i secret al salvataggio del file.
