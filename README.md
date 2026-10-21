# extra

| Cartella | Contenuto |
|---|---|
| `backend/` | API Go su [Encore](https://encore.dev) (servizi auth, profiles, admin, registration, broadcasters, videochat, images, advertising, …) |
| `frontend/` | Pannello Quasar/Vue 3 |
| `publicsite/` | Sito pubblico Quasar (SSG) |
| `docker/` | Deploy con Docker Compose e Caddy |
| `livekit/` | Server LiveKit per lo sviluppo locale |
| `documments/` | Documentazione di architettura |
| `mockup/` | Dati e immagini di esempio |
| `supervisor/` | Report di revisione del codice (2026-10-21) |

## Segreti

Nessun segreto è nel repository. Il backend legge i suoi da `backend/.secrets.local.cue`
(ignorato da git); Docker legge `docker/.env`, da creare partendo da `docker/.env.example`
(`PASSWORD_PEPPER` è obbligatorio, generarlo con `openssl rand -hex 32`).
