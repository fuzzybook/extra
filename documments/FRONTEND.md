# Manuale Tecnico Frontend (Vue 3 + Quasar + TypeScript)

Questo documento costituisce la documentazione di riferimento completa per il frontend del progetto **Extra**. Il frontend è una Single Page Application (SPA) realizzata con **Vue 3** e **Quasar Framework v2**, compilata tramite **Vite** e sviluppata con forte rigore di tipizzazione in **TypeScript**.

---

## 1. Stack Tecnologico & Struttura del Progetto

- **Framework UI**: Quasar Framework v2 (basato su Vue 3)
- **Modello di Sviluppo**: Vue 3 Composition API con sintassi `<script setup lang="ts">`
- **Build Tool**: Vite (tramite `@quasar/app-vite`)
- **State Management**: Pinia (`^3.0.4`)
- **Routing**: Vue Router (`^5.0.6`) in modalità `hash`
- **Streaming WebRTC**: LiveKit Client SDK (`livekit-client` `^2.22.3`)
- **Manipolazione Immagini**: `vue-advanced-cropper` (`^2.8.9`)
- **Validazione Dati**: Zod (`^4.4.3`) con generatore automatico sincrono
- **Internazionalizzazione**: Vue I18n (`^11.4.6`) con 5 lingue supportate

### Struttura delle Cartelle (`frontend/src/`)
```
frontend/src/
├── App.vue                         # Componente radice dell'applicazione
├── assets/                         # Immagini statiche e loghi SVG
├── boot/
│   └── i18n.ts                     # Inizializzazione del plugin Vue I18n
├── components/
│   ├── AvatarUpload.vue            # Upload avatar con fotocamera e cropper
│   ├── BroadcastTrack.vue          # Aggancio DOM delle tracce audio/video LiveKit
│   ├── CropperImage.vue            # Wrapper di ritaglio con vue-advanced-cropper
│   ├── DrawerUserCard.vue          # Scheda utente integrata nel cassetto laterale
│   ├── EssentialLink.vue           # Componente link di navigazione
│   ├── ImageDialog.vue             # Modale di gestione/anteprima immagini
│   ├── ImmageDialog.vue            # Dialogo alternativo immagini
│   ├── ProfileStatusBadge.vue      # Badge colorato per lo stato utente
│   ├── PromoCalendarView.vue       # Vista calendario interattivo per le promozioni
│   ├── SelectImageDialog.vue       # Modale di selezione immagini dal catalogo
│   └── WeekSelector.vue            # Selettore multi-settimana (settimane 1–53)
├── css/
│   └── app.css                     # Stili globali CSS nativi
├── data/
│   └── countries.ts                # Elenco codici e nomi paesi ISO per form
├── encore/
│   ├── client.ts                   # Client TypeScript generato da Encore
│   ├── videochat.ts                # Client API wrapper per il servizio videochat
│   └── zod.ts                      # Schemi di validazione Zod generati
├── i18n/
│   └── index.ts                    # Traduzioni in en, it, fr, de, es
├── layouts/
│   ├── MainLayout.vue              # Layout principale (navigazione standard)
│   └── AdminLayout.vue             # Layout riservato all'area amministrativa
├── pages/
│   ├── IndexPage.vue               # Homepage dell'applicazione
│   ├── SecondPage.vue              # Pagina secondaria / demo
│   ├── LoginPage.vue               # Login con email e password
│   ├── RegisterPage.vue            # Registrazione nuovo utente
│   ├── WelcomePage.vue             # Atterraggio conferma token email benvenuto
│   ├── PasswordResetPage.vue       # Flusso richiesta e conferma reset password
│   ├── ProfilePage.vue             # Gestione profilo personale dell'utente loggato
│   ├── AdvertisingPage.vue         # Directory delle inserzioni pubblicitarie
│   ├── PromosPage.vue              # Gestione promozioni (Tabellare + Calendario)
│   ├── WideochatPage.vue           # Elenco stanze e dirette streaming
│   ├── BroadcastPage.vue           # Sala streaming WebRTC (Broadcaster / Viewer)
│   ├── ErrorNotFound.vue           # Errore 404 (Route non trovata)
│   ├── ErrorUnauthorized.vue       # Errore 401 (Non autorizzato / Login richiesto)
│   └── admin/
│       ├── DashboardPage.vue       # Cruscotto di amministrazione
│       ├── ProfilesPage.vue        # Gestione completa profili utenti
│       └── dialogs/
│           ├── CreateProfileDialog.vue   # Creazione profilo da parte di admin
│           ├── EditProfileDialog.vue     # Modifica anagrafica, dati personali, setup
│           ├── UpdateRoleDialog.vue      # Modifica ruolo (user / admin)
│           ├── UpdateStatusDialog.vue    # Modifica stato (Active, Suspended, Banned...)
│           └── UpdatePasswordDialog.vue  # Reimpostazione forzata password
├── router/
│   ├── index.ts                    # Inizializzazione Vue Router
│   └── routes.ts                   # Dichiarazione gerarchica delle rotte
├── stores/
│   ├── index.ts                    # Configurazione store Pinia
│   ├── profiles-store.ts           # Store per autenticazione e profilo utente
│   ├── admin-store.ts              # Store per le funzioni del pannello admin
│   ├── images-store.ts             # Store per il catalogo immagini
│   └── layout-store.ts             # Store per metriche reattive di layout (body size)
└── types/
    ├── images.ts                   # Tipi aggiuntivi per manipolazione immagini
    └── promo.ts                    # Metadati e opzioni per le posizioni promo
```

---

## 2. Configurazione Quasar & Content Security Policy (CSP)

Nel file `frontend/quasar.config.ts` è definita la configurazione di compilazione e sicurezza:
- **Router Mode**: `hash` (`vueRouterMode: 'hash'`).
- **Supporto LiveKit CSP**:
  - LiveKit richiede connessioni WebSocket (`ws:`, `wss:`) e HTTP(S) verso i server multimediali.
  - La configurazione ammette automaticamente `http://localhost:7880`, `ws://localhost:7880`, `https://*.livekit.cloud`, `wss://*.livekit.cloud`.
  - In caso di server self-hosted custom, è possibile specificare la variabile d'ambiente:
    ```bash
    QCLI_LIVEKIT_URL=wss://video.tuodominio.com pnpm dev
    ```
    che viene validata all'avvio e inserita nelle origini consentite della CSP.

---

## 3. Generazione Client Encore & Schemi Zod

Il frontend interagisce con il backend tramite un client tipizzato generato automaticamente:
1. `src/encore/client.ts`: Generato da Encore tramite `encore gen client`, espone tutte le interfacce dei DTO e i metodi per invocare i microservizi (`client.auth`, `client.profiles`, `client.admin`, `client.advertising`, `client.images`, `client.registration`).
2. `src/encore/videochat.ts`: Wrapper complementare per le chiamate REST al servizio `videochat` (creazione stanze, richiesta token, lista partecipanti, kick).
3. `tools/zod-sync.mjs`: Script Node.js che compila automaticamente gli schemi Zod corrispondenti in `src/encore/zod.ts`. Viene invocato tramite:
   ```bash
   pnpm zod:sync
   ```
   consentendo di mantenere sincronizzata la validazione dei form lato client con le definizioni dei microservizi Go.

---

## 4. State Management (Pinia Stores)

### 4.1 `useProfilesStore` (`stores/profiles-store.ts`)
Gestisce l'identità dell'utente autenticato e la sessione attiva.
- **State**:
  - `token`: Token di sessione UUID letto/scritto in Quasar `LocalStorage` (`profiles.token`).
  - `profile`: Oggetto `Profile` dell'utente corrente (ID, email, display name, avatar, ruolo, stato, tipo membro `member_type`).
  - `personalData`: Oggetto `PersonalData` (nome, cognome, indirizzo, cap, città, nazione).
  - `loading`: Flag booleano di caricamento per chiamate in corso.
  - `error`: Ultimo messaggio di errore riscontrato.
- **Costanti di supporto**:
  - `MEMBER_TYPES`: Costanti dei tipi membro (`broadcaster`, `visitor`, `affiliate`, `advertiser`, `admin`).
  - `MEMBER_TYPE_OPTIONS`: Opzioni per componenti Quasar (`q-select`) con label e colori badge dedicati.
  - `memberTypeInfo(type)`: Helper per il mapping rapido di colore badge e label.
- **Getters**:
  - `isAuthenticated`: `computed(() => token.value !== null)`.
- **Azioni**:
  - `login(params)`: Invoca `client.profiles.Login`, memorizza il token, chiama `me()` e popola il profilo.
  - `logout()`: Rimuove il token dal `LocalStorage` e azzera `profile` e `personalData`.
  - `me()`: Richiede il profilo corrente a `GET /auth/me`.
  - `register(params)`: Registra un nuovo utente tramite `client.profiles.Insert`.
  - `update(params)`: Aggiorna display name e avatar del proprio profilo.
  - `remove()`: Cancella il proprio profilo utente ed esegue il logout.
  - `fetchPersonalData()` / `savePersonalData(params)`: Recupera e salva i dati personali.
  - `setPassword(params)`: Aggiorna la password previa verifica della password corrente.
  - `uploadAvatar(image: Blob)`: Invia lo stream binario a `POST /profiles/avatar` e restituisce l'URL pubblico.

### 4.2 `useAdminStore` (`stores/admin-store.ts`)
Espone tutte le funzionalità di amministrazione per utenti con ruolo `admin`.
- **State**: `profiles` (elenco completo), `roles` (ruoli di sistema), `statuses` (stati account), `memberTypes` (tipi membro di sistema), `loading`, `error`.
- **Getters**: `updatableStatuses` (filtra gli stati che possono essere impostati manualmente).
- **Azioni**:
  - `listProfiles()`: Carica tutti i profili utente.
  - `fetchSystemOptions()`: Recupera ruoli, stati e tipi membro validi per popolare i menu a tendina.
  - `insertProfile(params)`: Crea un profilo amministrativamente.
  - `updateProfile(userID, params)`: Modifica display name e avatar.
  - `updateProfileRole(userID, role)`: Assegna ruolo `user` o `admin`.
  - `updateProfileMemberType(userID, memberType)`: Modifica il tipo membro (`broadcaster`, `visitor`, `affiliate`, `advertiser`, `admin`).
  - `updateProfileStatus(userID, status)`: Modifica lo stato dell'account (`Active`, `Suspended`, `Banned`...).
  - `updateProfilePassword(userID, params)`: Forzatura amministrativa della password.
  - `getPersonalData(userID)` / `upsertPersonalData(userID, params)`: Lettura e scrittura dati anagrafici.
  - `getPersonalSetup(userID)` / `upsertPersonalSetup(userID, params)`: Gestione del setup broadcast (`can_broadcast`, `room_name`).
  - `uploadAvatar(image: Blob)`: Caricamento avatar per conto di qualsiasi utente.

### 4.3 `useImagesStore` (`stores/images-store.ts`)
- **State**: `images` (array di immagini con metadati e URL pubblici).
- **Azioni**: `fetchImages()` (carica catalogo da `GET /images`), `saveImage(params)` (upload o aggiornamento immagine con payload base64).

### 4.4 `useBroadcastersStore` (`stores/broadcasters-store.ts`)
- **State**: `broadcasters` (elenco profili broadcaster), `categories`, `attributePresets` (suggerimenti rapidi attributi da SiteConfig), `availabilityOptions` (stati di disponibilità configurati in SiteConfig), `imageConfig` (configurazione proporzione 2:3 e larghezza 600px), `loading`, `error`.
- **Azioni**:
  - `fetchBroadcasters(params?)`: Carica i profili broadcaster con filtri opzionali.
  - `fetchCategories()`: Carica le categorie configurate dal backend.
  - `fetchAttributePresets()`: Carica l'elenco dei suggerimenti rapidi per gli attributi broadcaster da `SiteConfig`.
  - `fetchAvailabilityOptions()`: Carica l'elenco degli stati di disponibilità da `config.SiteConfig.Availability`.
  - `fetchImageConfig()`: Recupera la configurazione standard per le immagini dei broadcaster.
  - `getBroadcaster(id)`: Recupera i dettagli del singolo broadcaster.
  - `createBroadcaster(params)`: Registra un nuovo broadcaster.
  - `updateBroadcaster(id, params)`: Modifica i dati anagrafici e generali del broadcaster.
  - `updateAttributes(id, attributes)`: Aggiorna separatamente la lista di attributi personalizzati.
  - `updateHours(id, hours)`: Aggiorna separatamente la programmazione delle fasce orarie.
  - `updatePhotos(id, photos)`: Aggiorna separatamente la galleria fotografica e la copertina in formato 2:3.
  - `deleteBroadcaster(id)`: Rimuove definitivamente il profilo broadcaster.

### 4.5 `useLayoutStore` (`stores/layout-store.ts`)
- Monitora le dimensioni effettive del viewport e del body tramite un `ResizeObserver` nativo.
- Espone `bodyWidth` e `bodyHeight` per consentire ricalcoli reattivi di tabelle sticky ed elementi grafici.

---

## 5. Routing, Layouts & Guardie di Sicurezza

### Mappa delle Rotte (`router/routes.ts`)
```
/ (MainLayout.vue)
├── ''                           -> IndexPage.vue (Home)
├── 'second'                     -> SecondPage.vue
├── 'login'                      -> LoginPage.vue
├── 'register'                   -> RegisterPage.vue
├── 'welcome'                    -> WelcomePage.vue (?token=...)
├── 'password-reset'             -> PasswordResetPage.vue (?token=...)
├── 'wideochat'                  -> WideochatPage.vue (Elenco stanze)
├── 'wideochat/:roomName'        -> BroadcastPage.vue (Sala streaming live)
├── 'profile'                    -> ProfilePage.vue (Profilo utente)
├── 'advertising'                -> AdvertisingPage.vue (Directory inserzionisti)
├── 'promos'                     -> PromosPage.vue (Pannello promozioni e calendario)
├── 'broadcasters'               -> BroadcastersPage.vue (Gestione canali e profili)
└── '401'                        -> ErrorUnauthorized.vue

/admin (AdminLayout.vue)
├── 'dashboard'                  -> DashboardPage.vue
└── 'profiles'                   -> ProfilesPage.vue (Gestione utenti)

/*                               -> ErrorNotFound.vue (404)
```

### Layouts & Protezioni di Navigazione
1. **`MainLayout.vue`**:
   - Toolbar con pulsante toggle del cassetto (`leftDrawerOpen`), titolo applicazione, selettore multilingua (`q-select`) e scheda utente (`DrawerUserCard.vue`).
   - Se l'utente è amministratore (`isAdmin`), mostra il link di accesso a `/admin/dashboard`.
   - Nel lifecycle hook `onMounted`, se esiste un token memorizzato ma `profile` è nullo, esegue `profilesStore.me()` per ripristinare la sessione. In caso di sessione scaduta, esegue il logout automatico.
2. **`AdminLayout.vue`**:
   - Toolbar con intestazione a contrasto visivo (`bg-warning text-black`).
   - Nel lifecycle hook `onMounted`, verifica la presenza del token e che `profile.role === 'admin'`. In caso contrario, reindirizza immediatamente a `/401` (`ErrorUnauthorized.vue`).

---

## 6. Pagine & Funzionalità Dettagliate

### 6.1 Videochat & Streaming WebRTC (`WideochatPage.vue` e `BroadcastPage.vue`)
- **`WideochatPage.vue`**:
  - Mostra la tabella `q-table` delle stanze disponibili, unificando le stanze attive su LiveKit e quelle configurate nel setup personale dei broadcaster.
  - Per ciascuna stanza, presenta un pulsante:
    - **"Gestisci"** (`icon="settings"`) se il chiamante è il proprietario (`owner_id === profile.user_id`).
    - **"Guarda"** (`icon="live_tv"`) se il chiamante è uno spettatore.
  - Gli amministratori dispongono del pulsante per creare una nuova stanza con timeout configurabile.
- **`BroadcastPage.vue`**:
  - Gestisce l'intero ciclo di vita della connessione WebRTC con `livekit-client`:
    1. Effettua una chiamata a `POST /videochat/rooms/:roomName/open` per garantire l'esistenza della stanza.
    2. Richiede il token di accesso a `POST /videochat/token`.
    3. Inizializza l'istanza `new Room()` di LiveKit e si connette a `server_url`.
    4. Ascolta gli eventi `RoomEvent.TrackSubscribed`, `TrackUnsubscribed`, `Disconnected`, `Reconnecting`, `Reconnected`.
  - **Funzionalità Broadcaster / Owner**:
    - Toggle Videocamera (`live.localParticipant.setCameraEnabled(!cameraOn)`).
    - Toggle Microfono (`live.localParticipant.setMicrophoneEnabled(!micOn)`).
    - Toggle Condivisione Schermo (`live.localParticipant.setScreenShareEnabled(!screenOn)`).
    - Pannello di gestione partecipanti: lista spettatori connessi con opzione di disconnessione forzata (kick).
    - Chiusura definitiva della stanza (`closeRoom`) che disconnette tutti e rimuove la sessione su LiveKit.
  - **Funzionalità Spettatore (Viewer)**:
    - Nessuna richiesta di autorizzazione per videocamera o microfono.
    - Gestione automatica dell'autoplay audio: se le policy del browser bloccano l'audio, mostra un banner esplicito con pulsante "Attiva audio" (`enableAudio`).
  - **Componente `BroadcastTrack.vue`**:
    - Riceve la traccia LiveKit (`Track`) e la aggancia dinamicamente a elementi nativi HTML5 `<video>` o `<audio>` tramite `track.attach()`, curando il distacco pulito e il rilascio della memoria in `onBeforeUnmount` via `track.detach()`.

### 6.2 Gestione Pubblicità (`AdvertisingPage.vue`)
- **Directory Inserzioni**: Tabella Quasar nativa con ricerca reattiva, badge di stato (`active`/`inactive`), link email diretti e modale per la registrazione di nuove inserzioni.
- **Anteprima Promozioni Settimanali (`getWeekPromo`)**: Sezione dedicata per la selezione di una settimana (1–53) e dell'anno, con chiamata automatica a `GET /promo/week/:week`, badge con conteggi per le tre posizioni e visualizzatore JSON interattivo con formattazione e pulsante di copia rapida negli appunti.

### 6.3 Promozioni & Calendario (`PromosPage.vue`, `PromoCalendarView.vue`, `WeekSelector.vue`)
- **`PromosPage.vue`**:
  - Interfaccia con due radio button per la selezione della modalità: **Tabellare** e **Calendario**.
  - Schede per le tre posizioni promozionali (`PromoLeft`, `PromoRight1`, `PromoRight2`).
  - Selettore di sezioni dinamiche conformi alle regole di business del backend (`week`, `raccomended`, `discover`, `show`, `service`, `top`, `selected`).
  - Gestione multi-item per ciascuna promozione (titolo, URL target, flag nuova finestra, supporto mobile con immagine alternativa).
  - Integrazione con i dialoghi di selezione e caricamento immagini (`SelectImageDialog`, `CropperImage`).
- **`PromoCalendarView.vue`**:
  - Rappresentazione a griglia delle settimane dell'anno (1–53) con evidenziazione grafica delle promozioni attive per periodo.
- **`WeekSelector.vue`**:
  - Griglia interattiva di chip numerati da 1 a 53 per selezionare o deselezionare rapidamente le settimane di pianificazione pubblicitaria.

### 6.3 Gestione Immagini e Avatar (`AvatarUpload.vue`, `CropperImage.vue`)
- **`AvatarUpload.vue`**:
  - Permette la selezione di un file immagine dal filesystem oppure l'acquisizione in tempo reale dalla fotocamera/webcam tramite l'API `navigator.mediaDevices.getUserMedia`.
  - Mostra il flusso video in anteprima all'interno di un dialogo Quasar (`q-dialog`).
  - All'acquisizione, trasferisce il frame al componente `CropperImage.vue` configurato con stencil circolare.
  - Effettua il ritaglio a risoluzione ottimale, converte il risultato in un `Blob` binario e delega il caricamento alla funzione `uploader` passata via props (`profilesStore.uploadAvatar` o `adminStore.uploadAvatar`).

### 6.4 Gestione Utenti e Amministrazione (`ProfilesPage.vue` e Dialoghi)
- **`ProfilesPage.vue`**:
  - Tabella sticky con colonne: Avatar, Nome visualizzato, Email, Ruolo, Stato (`ProfileStatusBadge.vue`), Flag Artista, Data Creazione e Data Aggiornamento.
  - Menu contestuale su ogni riga (`q-menu`) con accesso rapido a:
    - **Modifica Profilo**: `EditProfileDialog.vue` (schede per anagrafica, dati personali e configurazione broadcasting `can_broadcast`/`room_name`).
    - **Aggiorna Ruolo**: `UpdateRoleDialog.vue` (passaggio tra `user` e `admin`).
    - **Aggiorna Stato**: `UpdateStatusDialog.vue` (selezione dello stato da `updatableStatuses`).
    - **Aggiorna Password**: `UpdatePasswordDialog.vue` (impostazione forzata nuova password).
  - Pulsante in testata per apertura `CreateProfileDialog.vue` per registrare un nuovo utente.

### 6.5 Gestione Broadcaster (`BroadcastersPage.vue` e Dialoghi Separati)
- **`BroadcastersPage.vue`**:
  - Doppia visualizzazione (Cards / Tabella) commutabile in tempo reale.
  - Filtri dinamici per ricerca testuale (`display_name`, `bio`), categoria, disponibilità (`available`, `live`, `busy`, `offline`) e flag di certificazione.
  - Ciascun broadcaster dispone di un menu azioni a tendina (`q-menu`) che permette la gestione separata di ogni aspetto del canale:
    1. **Modifica Profilo**: `EditBroadcasterDialog.vue` per anagrafica, categoria con colore badge, bio, disponibilità (selezionata dinamicamente dall'array `config.SiteConfig.Availability`), telefono, chip di servizi e lingue.
    2. **Modifica Attributi**: `EditAttributesDialog.vue` per metadati chiave-valore con suggerimenti rapidi caricati dinamicamente da `config.SiteConfig.BroadcasterAttributes` (`/broadcasters/attribute-presets`), permettendo di aggiornarli centralmente dal backend senza modifiche al frontend.
    3. **Modifica Orari**: `EditHoursDialog.vue` per programmazione delle fasce orarie settimanali strutturate come array di intervalli `{ from, to }`. Per l'aggiunta o modifica di ciascuna fascia si apre un popup dedicato con due componenti `q-time` (Dalle e Alle in formato 24h), calcolo durata, fasce rapide e validazione real-time anti-sovrapposizione (impedisce il salvataggio se `from >= to` o se gli orari si sovrappongono a fasce esistenti nel medesimo giorno).
    4. **Modifica Foto**: `EditPhotosDialog.vue` per la galleria fotografica proporzionata in formato `2:3` (larghezza 600px). Le foto possono essere caricate esclusivamente tramite ritaglio con Cropper dedicato (senza selezione da libreria/galleria) oppure eliminate, con possibilità di riordinamento e impostazione della copertina.
    5. **Elimina**: Conferma ed eliminazione del profilo con ripulitura a cascata di attributi, orari e foto.

---

## 7. Internazionalizzazione (i18n)

L'applicazione supporta nativamente 5 lingue:
- `it`: Italiano (predefinito)
- `en`: English
- `fr`: Français
- `de`: Deutsch
- `es`: Español

Tutte le etichette, messaggi di errore, conferme, dialoghi e tabelle sono localizzati nel file `src/i18n/index.ts`. La selezione della lingua avviene in tempo reale tramite la toolbar e viene conservata in `localStorage` (`app.locale`).

---

## 8. Comandi di Esecuzione e Sviluppo

Dalla directory `frontend/`:
```bash
# Installazione delle dipendenze
pnpm install

# Avvio del server di sviluppo con HMR e hot-reload
pnpm dev

# Controllo dei tipi TypeScript per l'intero progetto
pnpm typecheck

# Sincronizzazione automatica degli schemi Zod dal client Encore
pnpm zod:sync

# Compilazione per l'ambiente di produzione
pnpm build
```
