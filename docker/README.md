# Guida al Deploy con Docker Compose (Extra App - Caddy Edition)

Questo pacchetto Docker fornisce l'intera infrastruttura necessaria per compilare ed eseguire in produzione l'applicazione fullstack **Extra** (Frontend Quasar + Caddy Web Server + Backend Encore + PostgreSQL + LiveKit).

---

## 1. Architettura dei Container

La configurazione orchestrata da [docker-compose.yml](file:///Users/fabio/CODE/extra/docker/docker-compose.yml) include 4 servizi:

1. **`frontend` (extra-frontend)**:
   - Build multi-stage Node 22 + Quasar v2 SPA.
   - Web Server moderno **Caddy 2** con:
     - Gestione automatica del routing SPA (`try_files {path} /index.html`).
     - Compressione avanzata Gzip e Zstandard (`encode gzip zstd`).
     - Cache a lungo termine per gli asset statici (`/assets/*`).
     - **Reverse proxy trasparente verso il backend** con gestione automatica di WebSocket per streaming e chat.
     - Supporto nativo HTTP/2 e HTTP/3 (QUIC).
     - Predisposizione per **HTTPS automatico** (Let's Encrypt / ZeroSSL) con volumi dedicati `caddy_data` e `caddy_config`.
   - Porte esposte: `http://localhost:8080` (e `https://localhost:8443`).

2. **`backend` (extra-backend)**:
   - Container basato su Debian con **Go 1.24** e **CLI Encore**.
   - Avvia i microservizi su porta `4000` con cruscotto diagnostico su porta `9400`.
   - Entrypoint dinamico che genera i segreti (`.secrets.local.cue`) a partire dalle variabili d'ambiente.

3. **`postgres` (extra-postgres)**:
   - Immagine ufficiale **PostgreSQL 16 Alpine**.
   - Volume persistente `extra_postgres_data`.
   - Script di inizializzazione automatico per la creazione dei database dei vari microservizi (`auth`, `profiles`, `advertising`, `images`, `registration`, `extrahome`, `broadcasters`, `videochat`).
   - Porta esposta: `5432`.

4. **`livekit` (extra-livekit)**:
   - Media Server WebRTC / SFU ufficiale LiveKit configurato con [livekit.yaml](file:///Users/fabio/CODE/extra/docker/livekit.yaml).
   - Porte: `7880` (HTTP/WS), `7881` (TCP RTC), `7882` (UDP RTC).

---

## 2. Configurazione (`.env`)

Il file [.env](file:///Users/fabio/CODE/extra/docker/.env) (copiato da [.env.example](file:///Users/fabio/CODE/extra/docker/.env.example)) permette di personalizzare porte e segreti:

```env
# Porte
FRONTEND_PORT=8080
FRONTEND_SSL_PORT=8443
BACKEND_PORT=4000
BACKEND_DASHBOARD_PORT=9400
POSTGRES_PORT=5432
LIVEKIT_HTTP_PORT=7880

# Database
POSTGRES_USER=extra
POSTGRES_PASSWORD=extrasecret
POSTGRES_DB=extra

# Segreti
PASSWORD_PEPPER=__generare_con_openssl_rand_hex_32__
LIVEKIT_URL=ws://livekit:7880
LIVEKIT_API_KEY=devkey
LIVEKIT_API_SECRET=secret
```

### Abilitazione HTTPS Automatico in Produzione
Per abilitare certificati SSL/TLS validi su un server di produzione:
1. Apri [Caddyfile](file:///Users/fabio/CODE/extra/docker/Caddyfile).
2. Rimuovi `auto_https off`.
3. Sostituisci `:80` con il tuo dominio pubblico (es. `extra.tuodominio.com`).
4. Mappa le porte standard `80:80` e `443:443` in `docker-compose.yml`. Caddy richiederà e rinnoverà i certificati in piena autonomia.

---

## 3. Comandi di Gestione

Dalla cartella `docker/`:

### Avvio dell'intero stack
```bash
docker compose up -d --build
```

### Visualizzazione dei log in tempo reale
```bash
docker compose logs -f
```

### Controllo stato dei container
```bash
docker compose ps
```

### Arresto dello stack
```bash
docker compose down
```

### Arresto con rimozione dei volumi di database e certificati
```bash
docker compose down -v
```

---

## 4. Endpoint Principali

| Servizio | URL Locale | Descrizione |
|---|---|---|
| **Frontend SPA (Caddy)** | `http://localhost:8080` | Interfaccia utente web Quasar |
| **Backend API Gateway** | `http://localhost:4000` | Gateway REST microservizi Encore |
| **Encore Tracing & Dashboard** | `http://localhost:9400` | Tracing e monitoraggio chiamate API |
| **LiveKit Media Server** | `http://localhost:7880` | Endpoint WebRTC SFU |
| **PostgreSQL** | `localhost:5432` | Database relazionale |
