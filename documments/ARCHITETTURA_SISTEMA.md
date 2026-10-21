# Architettura di Sistema e Flussi Globali

Questo documento fornisce una visione d'insieme dell'architettura dell'intero sistema **Extra**, evidenziando i pattern di comunicazione tra Frontend (Quasar SPA) e Backend (Encore Microservices), le scelte architetturali di sicurezza, i flussi asincroni Pub/Sub e l'integrazione di servizi terzi come LiveKit.

---

## 1. Topologia Architetturale

L'applicazione segue un paradigma a **microservizi modulari** lato backend e un'interfaccia a **Single Page Application (SPA)** reattiva lato frontend.

```mermaid
graph TB
    subgraph Frontend["Frontend (Quasar Framework v2 / Vue 3 / TS)"]
        UI["Interfaccia Utente & Pagine Quasar"]
        Pinia["Pinia Stores (Profiles, Admin, Images, Layout)"]
        ClientTS["Encore Generated Client (HTTP/REST)"]
        LKClient["LiveKit Client SDK (WebRTC)"]
        UI --> Pinia
        Pinia --> ClientTS
        UI --> LKClient
    end

    subgraph Backend["Backend Encore (Go Microservices)"]
        Gateway["Encore API Gateway (Port 4000)"]
        AuthHandler["AuthHandler (//encore:authhandler)"]

        subgraph Services["Microservizi"]
            AuthSvc["auth (Credenziali, Sessioni, Hashing)"]
            ProfilesSvc["profiles (Utenti, Profili, Personal Setup)"]
            AdminSvc["admin (Backoffice, Middleware RequireAdmin)"]
            VideochatSvc["videochat (LiveKit Room & Token Manager)"]
            AdvSvc["advertising (Directory Inserzionisti & Promo)"]
            ImagesSvc["images (Catalogo & Binary Storage)"]
            RegSvc["registration (Orchestrazione Registrazione & Token)"]
            MailsSvc["mails (Outbox & Transactional Delivery)"]
            HomeSvc["extrahome (Configurazione Sito, Caroselli)"]
            HelloSvc["hello (Healthcheck & Test)"]
        end

        subgraph EventBus["Pub/Sub Event Broker"]
            LoginReq["Topic: login-requests"]
            LoginRes["Topic: login-results"]
        end

        subgraph Storage["Storage & Database"]
            PGAuth[("DB: auth")]
            PGProf[("DB: profiles")]
            PGAdv[("DB: advertising")]
            PGImg[("DB: images")]
            PGReg[("DB: registration")]
            PGHome[("DB: extrahome")]
            BucketAvatars["Bucket: avatars (Pubblico)"]
            BucketImages["Bucket: images (Pubblico)"]
        end
    end

    subgraph External["Servizi Esterni"]
        LiveKitServer["LiveKit Media Server (WebRTC / SFU)"]
    end

    ClientTS -->|HTTP / Bearer Token| Gateway
    Gateway --> AuthHandler
    AuthHandler -->|Verifica Sessione| PGAuth
    AuthHandler -->|Carica Ruolo| PGProf

    Gateway --> Services
    ProfilesSvc -->|Pubblica LoginRequested| LoginReq
    LoginReq --> AuthSvc
    AuthSvc -->|Pubblica LoginCompleted| LoginRes
    LoginRes --> ProfilesSvc

    AuthSvc --> PGAuth
    ProfilesSvc --> PGProf
    AdminSvc --> PGProf
    AdvSvc --> PGAdv
    ImagesSvc --> PGImg
    RegSvc --> PGReg
    HomeSvc --> PGHome

    ProfilesSvc --> BucketAvatars
    ImagesSvc --> BucketImages

    VideochatSvc -->|Twirp / REST API| LiveKitServer
    LKClient <==>|PeerConnection WebRTC (Audio/Video)| LiveKitServer
```

---

## 2. Modello di Comunicazione & Generazione Client

1. **Protocollo HTTP/JSON**: Tutte le chiamate tra Frontend e Backend avvengono su protocollo REST HTTP standard tramite endpoint esposti da Encore.
2. **Type Safety End-to-End**:
   - I tipi di richiesta e risposta definiti in Go nei package del backend vengono convertiti automaticamente in interfacce e classi TypeScript mediante il comando:
     ```bash
     encore gen client --lang=typescript --output=../frontend/src/encore/client.ts
     ```
   - In questo modo, qualsiasi modifica a un campo in Go (es. aggiunta di un campo al profilo o cambio tipo) produce errori a tempo di compilazione nel frontend (`pnpm typecheck` con `vue-tsc`).
3. **Validazione dei Form con Zod**:
   - Il progetto include lo script `frontend/tools/zod-sync.mjs` che legge i tipi generati in `client.ts` e compila schemi di validazione Zod in `src/encore/zod.ts`.
   - Questi schemi alimentano le validazioni reattive nei componenti e form Quasar (`q-input`, `q-select`).

---

## 3. Flussi di Sequenza Principali

### 3.1 Flusso di Autenticazione Asincrona (Pub/Sub)
Un pattern architetturale distintivo del backend è la separazione tra la gestione anagrafica (`profiles`) e la sicurezza delle credenziali (`auth`), orchestrata in modo asincrono tramite Pub/Sub con polling controllato.

```mermaid
sequenceDiagram
    autonumber
    actor Utente as Client (Browser / Quasar)
    participant Profiles as Servizio profiles
    participant DBProfiles as DB profiles (user_profiles)
    participant PubSub as Encore Pub/Sub
    participant Auth as Servizio auth
    participant DBAuth as DB auth (credentials / sessions / login_results)

    Utente->>Profiles: POST /auth/login { user_email, password }
    Profiles->>DBProfiles: SELECT user_id FROM user_profiles WHERE email = $1
    DBProfiles-->>Profiles: user_id (UUID)
    Profiles->>PubSub: Publish(login-requests, { request_id, user_id, password })
    
    par Esecuzione Consumer Auth
        PubSub->>Auth: handleLoginRequested({ request_id, user_id, password })
        Auth->>DBAuth: SELECT password_hash FROM credentials WHERE user_id = $1
        DBAuth-->>Auth: password_hash (BCrypt + Pepper)
        Auth->>Auth: Hashing password fornita con Pepper SHA-256 e verifica BCrypt
        alt Credenziali Valide
            Auth->>DBAuth: INSERT INTO sessions (token, user_id, expires_at)
            Auth->>DBAuth: INSERT INTO login_results (request_id, ok=true, token, expires_at)
            Auth->>PubSub: Publish(login-results, { request_id, ok=true, token, expires_at })
        else Credenziali Errate
            Auth->>DBAuth: INSERT INTO login_results (request_id, ok=false)
            Auth->>PubSub: Publish(login-results, { request_id, ok=false })
        end
    and Polling Risultato Profiles
        loop Polling ogni 50ms (timeout 10s)
            Profiles->>DBAuth: SELECT ok, token, expires_at FROM login_results WHERE request_id = $1
            DBAuth-->>Profiles: Risultato trovato
        end
    end

    Profiles-->>Utente: 200 OK { token, expires_at }
    Utente->>Utente: Salvataggio token in LocalStorage (Bearer)
    Utente->>Profiles: GET /auth/me (con header Authorization: Bearer <token>)
    Profiles-->>Utente: Dati completi utente e ruolo
```

---

### 3.2 Flusso Videochat & Streaming LiveKit WebRTC

Il sistema gestisce dirette streaming 1-a-molti (1 Broadcaster/Owner e N Viewers). La sicurezza e l'assegnazione dei permessi di pubblicazione audio/video sono controllate interamente dal server.

```mermaid
sequenceDiagram
    autonumber
    actor Broadcaster as Proprietario / Broadcaster
    actor Viewer as Spettatore (Viewer)
    participant Frontend as Quasar Frontend
    participant Backend as videochat & profiles Svc
    participant DB as Postgres (profiles & personal_setup)
    participant LiveKit as LiveKit Media Server

    Note over Broadcaster,DB: Setup Iniziale (Eseguito da Admin)
    Backend->>DB: Admin imposta personal_setup: { can_broadcast: true, room_name: "stanza-1" }

    Note over Broadcaster,LiveKit: Apertura della Stanza
    Broadcaster->>Frontend: Apre /#/wideochat/stanza-1
    Frontend->>Backend: POST /videochat/rooms/stanza-1/open
    Backend->>DB: Acquisizione pg_advisory_xact_lock (concurrency safe)
    Backend->>DB: Verifica univocità personal_setup per la stanza
    Backend->>LiveKit: CreateRoom(name="stanza-1", metadata={ mode: "broadcast", owner_id: userID })
    LiveKit-->>Backend: Room creata o riutilizzata
    Backend-->>Frontend: Dettagli stanza (OwnerID, DisplayName, Active: true)

    Note over Broadcaster,LiveKit: Richiesta Token Broadcaster
    Frontend->>Backend: POST /videochat/token { room_name: "stanza-1" }
    Backend->>Backend: Verifica se caller == owner_id -> CanPublish = TRUE
    Backend->>Backend: Firma JWT Token LiveKit (Validità 10 min)
    Backend-->>Frontend: { participant_token, role: "owner", server_url }
    Frontend->>LiveKit: livekit.Room.connect(server_url, participant_token)
    Broadcaster->>Frontend: Attiva Videocamera / Microfono / Schermo
    Frontend->>LiveKit: publishTrack(videoTrack, audioTrack)

    Note over Viewer,LiveKit: Ingresso Spettatore
    Viewer->>Frontend: Entra in /#/wideochat/stanza-1
    Frontend->>Backend: POST /videochat/token { room_name: "stanza-1" }
    Backend->>Backend: caller != owner_id -> CanPublish = FALSE (CanSubscribe = TRUE)
    Backend-->>Frontend: { participant_token, role: "viewer", server_url }
    Frontend->>LiveKit: livekit.Room.connect(server_url, participant_token)
    LiveKit-->>Frontend: TrackSubscribed Event (flussi video/audio del broadcaster)
    Frontend->>Frontend: Montaggio traccia su <video> tramite BroadcastTrack.vue
```

---

### 3.3 Flusso Upload Immagini e Ritaglio Avatar

```mermaid
sequenceDiagram
    autonumber
    actor Utente as Utente (Browser)
    participant AvatarComp as AvatarUpload.vue / CropperImage.vue
    participant ProfilesSvc as profiles.UploadAvatar (Encore)
    participant Bucket as Encore Object Bucket (avatars)
    participant DB as Postgres (user_profiles)

    Utente->>AvatarComp: Seleziona file da disco OPPURE scatta da webcam (getUserMedia)
    AvatarComp->>AvatarComp: Apre CropperImage.vue (stencil circolare per avatar)
    Utente->>AvatarComp: Regola ritaglio e conferma
    AvatarComp->>AvatarComp: Canvas HTML5 -> blob (image/png)
    AvatarComp->>ProfilesSvc: POST /profiles/avatar (Raw binary body stream + Bearer Token)
    ProfilesSvc->>ProfilesSvc: Genera UUID v4 casuale come chiave
    ProfilesSvc->>Bucket: Upload stream con Content-Type
    Bucket-->>ProfilesSvc: Upload completato
    ProfilesSvc-->>AvatarComp: 200 OK { url: "http://.../avatars/<uuid>" }
    AvatarComp->>ProfilesSvc: PUT /profiles { display_name: "...", avatar_url: "..." }
    ProfilesSvc->>DB: UPDATE user_profiles SET avatar_url = $1 WHERE user_id = $2
    DB-->>ProfilesSvc: Record aggiornato
    ProfilesSvc-->>AvatarComp: Profilo aggiornato con nuovo avatar
```

---

## 4. Modello di Sicurezza e Autorizzazione

1. **Password Hashing con Pepper a 2 Livelli**:
   - Livello 1: Pre-hashing crittografico mediante **SHA-256** combinando la password in chiaro dell'utente con il segreto globale dell'applicazione (`secrets.PasswordPepper`). Questo protegge dalla limitazione dei 72 byte di BCrypt e garantisce che un eventuale leak del database di hash sia inservibile senza il pepper.
   - Livello 2: Hashing con **BCrypt** (`bcrypt.GenerateFromPassword(peppered, bcrypt.DefaultCost)`) dotato di salt univoco per record.
   - Protezione da riutilizzo: in fase di cambio password, il sistema impedisce di riutilizzare la password attuale (`isPasswordReuse`).
2. **Sessioni e Autenticatore Globale**:
   - Le sessioni sono persistite nella tabella `sessions` con UUID v4 e TTL di 7 giorni.
   - L'handler di autenticazione (`//encore:authhandler AuthHandler`) legge il token sia dal cookie `session` che dall'header standard `Authorization: Bearer <token>`.
   - Se il token è valido e non scaduto, recupera contestualmente il ruolo utente (`user` o `admin`) dalla tabella `user_profiles` e lo rende disponibile a tutti gli endpoint protetti tramite `auth.Data().(*AuthData)`.
3. **Controllo Accessi Basato sui Ruoli (RBAC)**:
   - **`admin`**: Accesso a tutti i servizi amministrativi, modifica ruoli, reset forzato password, abilitazione stanze broadcast, configurazione promozionale. Protetto a livello di middleware in `backend/admin/middleware.go` (`RequireAdmin`).
   - **`user`**: Accesso alle proprie risorse (modifica del solo profilo utente proprietario tramite `requireSelf`, lettura e inserimento dati personali, visualizzazione stanze).
   - **Public**: Accesso alle registrazioni, verifica disponibilità email e login.
4. **Isolamento dei Database (Database-per-Service)**:
   - Ogni microservizio (`auth`, `profiles`, `advertising`, `images`, `registration`, `extrahome`) dispone del proprio database PostgreSQL isolato e del proprio set di migrazioni SQL.
   - Le query cross-database sono confinate strettamente ai riferimenti controllati (es. `profilesDB = sqldb.Named("profiles")` in `auth` e `admin`) per garantire modularità e tracciabilità delle dipendenze.
