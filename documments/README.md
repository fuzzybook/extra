# Documentazione del Progetto Extra: Encore & Quasar Fullstack Application

Benvenuto nella documentazione tecnica completa ed esaustiva dell'applicazione fullstack **Extra (Encore-Quasar-App)**.

Il progetto è una piattaforma modulare ad alte prestazioni composta da:
- **Backend**: Microservizi in Go sviluppati con il framework [Encore.dev](https://encore.dev), dotati di database PostgreSQL dedicati, sistema Pub/Sub ad eventi asincroni, Object Storage e integrazione WebRTC con **LiveKit**.
- **Frontend**: Single Page Application (SPA) realizzata in [Vue 3](https://vuejs.org) (Composition API, `<script setup>`) con [Quasar Framework v2](https://quasar.dev) (Vite-based), TypeScript, [Pinia](https://pinia.vuejs.org), internazionalizzazione in 5 lingue ([Vue I18n](https://vue-i18n.intlify.dev)), e LiveKit Client SDK.

---

## Mappa della Documentazione

La documentazione è suddivisa in sezioni modulari e dettagliate:

1. [**Architettura di Sistema e Flussi Globali**](file:///Users/fabio/CODE/extra/documments/ARCHITETTURA_SISTEMA.md)
   - Panoramica architetturale generale.
   - Modello di comunicazione Client-Server e sincronizzazione tipi TypeScript (`encore gen client`).
   - Diagrammi di sequenza Mermaid: Autenticazione asincrona Pub/Sub, Broadcasting LiveKit WebRTC, Upload e ritaglio immagini, Gestione Promozioni e Calendario.
   - Sicurezza: Autenticazione a token/cookie, Pepper SHA-256 + BCrypt, RBAC (Role-Based Access Control) e isolamento database.

2. [**Manuale Tecnico Backend (Go + Encore)**](file:///Users/fabio/CODE/extra/documments/BACKEND.md)
   - Architettura interna dei 10 microservizi:
     - `auth`: Autenticazione, hashing con pepper, sessioni, token Bearer e cron job di cleanup.
     - `profiles`: Gestione profili utente, stati anagrafici, dati personali, avatar storage e setup broadcast.
     - `admin`: Backoffice e middleware `RequireAdmin` per la gestione completa di utenti, ruoli e permessi.
     - `videochat`: Integrazione WebRTC con LiveKit, sincronizzazione stanze, emissione JWT token e controllo partecipanti.
     - `advertising`: Directory inserzionisti e motore promozionale/banner programmabile su base calendario (settimane 1-53) e posizioni (Left, Right1, Right2).
     - `images`: Storage e catalogo immagini con Encore Object Storage pubblico e endpoint streaming binario.
     - `registration`: Flusso registrazione, validazione MX del dominio email, token temporizzati e template email (benvenuto, password reset).
     - `mails`: Outbox locale e gestione invio email transazionali.
     - `extrahome`: Gestione configurazione dinamica del portale, caroselli e broadcaster.
     - `hello`: Endpoint di prova e healthcheck.
   - Schemi dei database e cronologia completa delle migrazioni SQL.
   - Segreti, variabili d'ambiente, testing e comandi di esecuzione.

3. [**Manuale Tecnico Frontend (Vue 3 + Quasar + TypeScript)**](file:///Users/fabio/CODE/extra/documments/FRONTEND.md)
   - Struttura del progetto, Quasar CLI, configurazione Vite e Content Security Policy (CSP) per LiveKit.
   - Integrazione Encore Client TypeScript generato e sincronizzazione schemi di validazione con Zod (`tools/zod-sync.mjs`).
   - Gestione dello stato globale con Pinia (`profiles-store`, `admin-store`, `images-store`, `layout-store`).
   - Routing, guardie di navigazione e gestione layout (`MainLayout.vue` e `AdminLayout.vue`).
   - Dettaglio implementativo di pagine e viste:
     - Autenticazione e Profilo (`LoginPage`, `RegisterPage`, `WelcomePage`, `PasswordResetPage`, `ProfilePage`).
     - Videochat e Dirette WebRTC (`WideochatPage`, `BroadcastPage`, componente `BroadcastTrack.vue`).
     - Pannello Pubblicità e Promozioni (`AdvertisingPage`, `PromosPage`, `PromoCalendarView.vue`, `WeekSelector.vue`).
     - Amministrazione (`ProfilesPage`, `DashboardPage` e dialog modali dedicati).
   - Componenti avanzati: Upload da webcam e file con ritaglio circolare (`AvatarUpload.vue`, `CropperImage.vue`), gestione dialoghi immagini.
   - Sistema di localizzazione multilingua (Italiano, Inglese, Francese, Tedesco, Spagnolo).

---

## Guida Rapida all'Avvio Locale

### Prerequisiti
- [Encore CLI](https://encore.dev/docs/install): `brew install encoredev/tap/encore`
- [Go](https://go.dev/): Versione 1.23+
- [Node.js](https://nodejs.org/): Versione 22+ o 24+
- [pnpm](https://pnpm.io/): Gestore pacchetti frontend

### Avvio del Backend
Dalla directory principale del progetto:
```bash
cd backend
encore run
```
- **API Server**: `http://localhost:4000`
- **Encore Developer Dashboard & Tracing**: `http://localhost:9400`

### Avvio del Frontend
In un terminale separato:
```bash
cd frontend
pnpm install
pnpm dev
```
L'applicazione Quasar sarà disponibile su: `http://localhost:9000` (con supporto Hot Module Replacement attivo).

### Sincronizzazione Tipi e Schemi Zod
Se vengono apportate modifiche ai modelli o agli endpoint Go nel backend:
```bash
cd backend
encore gen client --lang=typescript --output=../frontend/src/encore/client.ts

cd ../frontend
pnpm zod:sync
pnpm typecheck
```
