# Manuale Tecnico Backend (Go + Encore.dev)

Questo documento costituisce la documentazione di riferimento completa per il backend del progetto **Extra**. Il backend è implementato in **Go** e strutturato mediante il framework per microservizi cloud-native **Encore.dev**.

---

## 1. Architettura Generale e Principi Guida

- **Linguaggio**: Go (1.23+)
- **Framework**: Encore.dev (runtime cloud e locale con orchestrazione automatica di database, pub/sub, storage e cron jobs)
- **Persistenza**: PostgreSQL gestito in modalità database-per-service tramite `encore.dev/storage/sqldb`
- **Object Storage**: Bucket S3-compatibili nativi di Encore tramite `encore.dev/storage/objects`
- **Event-Driven Messaging**: Pub/Sub nativo Encore (`encore.dev/pubsub`) con semantica At-Least-Once
- **Sicurezza crittografica**: BCrypt con pepper applicativo SHA-256 pre-hashing
- **WebRTC & Live Streaming**: LiveKit Server SDK (`github.com/livekit/server-sdk-go/v2`)

---

## 2. Dettaglio dei Microservizi

### 2.1 Servizio `auth`
Il servizio `auth` è il custode delle credenziali utente e delle sessioni attive. Isola le password e i token dal resto dell'applicazione.

#### Configurazione e Segreti
- `secrets.PasswordPepper` (stringa): Pepper crittografico applicativo unito alla password prima del processo di hashing.
- `sessionTTL`: 7 giorni (`7 * 24 * time.Hour`).
- `minPasswordLength`: 8 caratteri minimi.

#### Tabelle Database (`auth`)
1. **`credentials`**:
   - `user_id UUID PRIMARY KEY`: Riferimento all'utente in `profiles`.
   - `password_hash TEXT NOT NULL`: Hash BCrypt della password con pepper.
   - `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`.
2. **`sessions`**:
   - `token UUID PRIMARY KEY`: Identificativo univoco della sessione (inviato come Bearer token o session cookie).
   - `user_id UUID NOT NULL`: Utente proprietario.
   - `created_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ` (indicizzato per cleanup rapido).
3. **`login_results`**:
   - Tabella di correlazione per la risoluzione del flusso asincrono Pub/Sub (`request_id UUID PRIMARY KEY`, `ok BOOLEAN`, `token UUID`, `expires_at TIMESTAMPTZ`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`).

#### Handler di Autenticazione Globale
```go
//encore:authhandler
func AuthHandler(ctx context.Context, p *AuthParams) (auth.UID, *AuthData, error)
```
- Ispeziona il cookie `session` o l'header `Authorization: Bearer <token>`.
- Verifica la presenza e validità della sessione in `sessions` (`expires_at > NOW()`).
- Interroga il database `profiles` per estrarre il ruolo dell'utente (`role`: `user` o `admin`).
- Restituisce `auth.UID` (ID utente) e `AuthData` (`SessionToken`, `Role`).

#### Endpoints API
| Metodo | Path | Visibilità | Descrizione |
|---|---|---|---|
| `POST` | `/auth/internal/login` | Privato (Service-to-Service) | Verifica password e crea un nuovo token di sessione. |
| `POST` | `/auth/logout` | Autenticato (`auth`) | Revoca la sessione corrente cancellando il token da `sessions`. |
| `POST` | `/auth/credentials` | Privato | Registra l'hash iniziale della password per un nuovo utente (chiamato da `profiles.Insert`). |
| `PUT` | `/auth/password` | Autenticato (`auth`) | Modifica password per l'utente loggato (richiede e valida la password corrente). |
| `PUT` | `/auth/internal/users/:userID/password` | Privato | Forzatura password da flussi fidati interni (es. admin o reset password). |

#### Pub/Sub & Cron Jobs
- **Topic `login-requests`**: Riceve messaggi `LoginRequested` da `profiles.Login`.
- **Subscription `authenticate`**: Consumer che valida le credenziali, genera la sessione e invia l'esito a `login-results`.
- **Cron Job `cleanup-expired-sessions`**: Esecuzione ogni 1 ora per eliminare sessioni scadute (`DELETE FROM sessions WHERE expires_at <= NOW()`).
- **Cron Job `cleanup-older-login-results`**: Esecuzione ogni 6 ore per eliminare righe di correlazione login più vecchie di 24 ore.

---

### 2.2 Servizio `profiles`
Gestisce le anagrafiche utente, dati personali, avatar e le impostazioni broadcast per lo streaming.

#### Tabelle Database (`profiles`)
1. **`user_profiles`**:
   - `user_id UUID PRIMARY KEY`
   - `email TEXT NOT NULL UNIQUE` (normalizzata in lowercase)
   - `display_name TEXT NOT NULL`
   - `avatar_url TEXT NOT NULL DEFAULT ''`
   - `role TEXT NOT NULL DEFAULT 'user'` (`user` o `admin`)
   - `status SMALLINT NOT NULL DEFAULT 0` (Mappato all'enum `Status`)
   - `member_type TEXT NOT NULL DEFAULT 'visitor'` (Valori supportati: `broadcaster`, `visitor`, `affiliate`, `advertiser`, `admin`)
   - `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
2. **`personal_data`**:
   - `user_id UUID PRIMARY KEY REFERENCES user_profiles(user_id) ON DELETE CASCADE`
   - `first_name`, `last_name`, `address`, `cap`, `city`, `country` (codice ISO 2 lettere), `created_at`, `updated_at`
3. **`personal_setup`**:
   - `user_id UUID PRIMARY KEY REFERENCES user_profiles(user_id) ON DELETE CASCADE`
   - `can_broadcast BOOLEAN NOT NULL DEFAULT FALSE`
   - `room_name TEXT NOT NULL DEFAULT ''`
   - Vincolo SQL: se `can_broadcast = false`, `room_name` deve essere vuoto; se `true`, deve essere compreso tra 1 e 128 byte non vuoti.

#### Enum degli Stati Utente (`Status`)
- `0`: `StatusActive` (Attivo - modificabile da admin)
- `1`: `StatusInactive` (Inattivo - modificabile)
- `2`: `StatusSuspended` (Sospeso - modificabile)
- `3`: `StatusDeleted` (Eliminato - non modificabile)
- `4`: `StatusPending` (In attesa - modificabile)
- `5`: `StatusWaitingDeletion` (In attesa di cancellazione - non modificabile)
- `6`: `StatusBanned` (Bannato - modificabile)

#### Object Storage Bucket
- **`avatars`** (`objects.NewBucket("avatars", Public: true)`): Bucket pubblico per la memorizzazione permanente delle immagini profilo.

#### Endpoints API
| Metodo | Path | Visibilità | Descrizione |
|---|---|---|---|
| `POST` | `/auth/login` | Pubblico | Login utente. Cerca l'utente per email, pubblica su `login-requests` e attende la risposta con polling (timeout 10s). |
| `GET` | `/auth/me` | Pubblico | Restituisce il profilo dell'utente autenticato chiamante o `null` se anonimo. |
| `POST` | `/profiles` | Pubblico | Registrazione diretta profilo utente con password (delega credenziali ad `auth.Register`). |
| `GET` | `/profiles/profile` | Pubblico | Alias per il recupero del profilo chiamante. |
| `PUT` | `/profiles` | Autenticato | Aggiorna `display_name` e `avatar_url` del proprio profilo utente. |
| `DELETE` | `/profiles` | Autenticato | Elimina il proprio account e cancella il record da `user_profiles`. |
| `GET` | `/profiles/personal-data` | Autenticato | Recupera i dati anagrafici personali dell'utente autenticato. |
| `PUT` | `/profiles/personal-data` | Autenticato | Inserisce o aggiorna (Upsert) i dati personali con validazione CAP e paese. |
| `GET` | `/profiles/personal-setup` | Autenticato | Restituisce la configurazione broadcast dell'utente (`can_broadcast`, `room_name`). |
| `GET` | `/profiles/personal-setup/:userID` | Autenticato (Admin) | Lettura della configurazione broadcast di un utente specifico. |
| `PUT` | `/profiles/personal-setup/:userID` | Autenticato (Admin) | Assegna o revoca il permesso broadcast e il nome stanza di un utente. |
| `POST` | `/profiles/avatar` | Autenticato (Raw HTTP) | Riceve lo stream binario dell'immagine, la salva nel bucket `avatars` e restituisce il relativo URL pubblico. |

---

### 2.3 Servizio `admin`
Fornisce il pannello di controllo backoffice per amministratori per gestire utenti, ruoli, permessi e visualizzare lo stato di sistema.

#### Middleware di Sicurezza
- `RequireAdmin` (`encore:middleware target=all`): Intercetta ogni endpoint del servizio `admin`. Se il chiamante non è autenticato o non possiede il ruolo `admin`, rigetta la richiesta con codice di errore `PermissionDenied` (403).

#### Endpoints API
| Metodo | Path | Descrizione |
|---|---|---|
| `GET` | `/admin/profiles` | Elenco di tutti i profili utente registrati, ordinati per ID. |
| `GET` | `/admin/profiles/:userID` | Dettaglio del profilo utente per ID. |
| `POST` | `/admin/profiles` | Creazione amministrativa di un nuovo profilo utente con credenziali. |
| `PUT` | `/admin/profiles/:userID` | Modifica anagrafica di base (`display_name`, `avatar_url`). |
| `PUT` | `/admin/profiles/:userID/status` | Modifica dello stato dell'account (`Active`, `Suspended`, `Banned`, ecc.). |
| `PUT` | `/admin/profiles/:userID/role` | Modifica del ruolo utente (`user` <-> `admin`). |
| `PUT` | `/admin/profiles/:userID/member-type` | Modifica del tipo membro utente (`member_type`). |
| `PUT` | `/admin/profiles/:userID/password` | Reimpostazione amministrativa forzata della password. |
| `DELETE` | `/admin/profiles/:userID` | Eliminazione definitiva dell'utente da parte dell'amministratore. |
| `GET` | `/admin/profiles/:userID/personal-data` | Lettura dei dati personali di qualsiasi utente. |
| `PUT` | `/admin/profiles/:userID/personal-data` | Salvataggio o aggiornamento forzato dei dati personali dell'utente. |
| `GET` | `/admin/system/options` | Elenco delle opzioni configurabili (ruoli disponibili, stati con flag `updatable`, tipi membro). |

---

### 2.4 Servizio `videochat`
Gestisce l'orchestrazione delle stanze live e l'interazione con il media server WebRTC **LiveKit**.

#### Configurazione e Segreti
- `secrets.LiveKitURL`: URL del server LiveKit (supporta schemi `http`, `https`, `ws`, `wss`).
- `secrets.LiveKitAPIKey`: Chiave API LiveKit per la firma JWT.
- `secrets.LiveKitAPISecret`: Segreto API LiveKit.
- `tokenTTL`: 10 minuti di validità per i JWT generati.

#### Modello di Concorrenza e Locking
Per evitare conflitti nella creazione simultanea della medesima stanza da parte di istanze concorrenti, `OpenRoom` acquisisce un lock advisory Postgres transazionale:
```go
SELECT pg_advisory_xact_lock(hashtextextended($1, 0))
```
usando la chiave `"videochat/open/" + roomName`.

#### Token JWT e Permessi LiveKit
La funzione `signJoinToken` impone i privilegi:
- **Broadcaster / Room Owner**: `CanPublish = true`, `CanSubscribe = true`.
- **Viewer / Spettatore**: `CanPublish = false`, `CanSubscribe = true`.
- Nessun utente può pubblicare dati arbitrari (`CanPublishData = false`) né aggiornare i propri metadata (`CanUpdateOwnMetadata = false`).

#### Endpoints API
| Metodo | Path | Visibilità | Descrizione |
|---|---|---|---|
| `GET` | `/videochat/status` | Pubblico | Healthcheck del servizio videochat. |
| `POST` | `/videochat/rooms` | Autenticato (Admin) | Creazione esplicita di una stanza LiveKit con metadata di broadcast. |
| `GET` | `/videochat/rooms` | Autenticato | Restituisce la lista unificata delle stanze attive su LiveKit e di quelle configurate nei setup personali (con `active: false` se spente). |
| `POST` | `/videochat/rooms/:roomName/open` | Autenticato | Apre o riutilizza una stanza: verifica che esista un setup personale abilitato univoco per quel nome stanza, la crea su LiveKit se necessario e assegna la proprietà. |
| `POST` | `/videochat/token` | Autenticato | Rilascia un token JWT LiveKit firmato per l'accesso alla stanza. Assegna il ruolo `owner` o `viewer` verificando il metadata della stanza. |
| `DELETE` | `/videochat/rooms/:roomName` | Autenticato (Owner) | Chiude la stanza su LiveKit e disconnette forzatamente tutti i partecipanti. |
| `GET` | `/videochat/rooms/:roomName/participants` | Autenticato (Owner) | Elenca tutti i partecipanti connessi alla stanza. |
| `DELETE` | `/videochat/rooms/:roomName/participants/:identity` | Autenticato (Owner) | Disconnette forzatamente uno spettatore specifico dalla diretta. |

---

### 2.5 Servizio `advertising`
Gestisce la directory delle inserzioni commerciali e la programmazione avanzata di banner/promozioni suddivise per posizione, sezione e settimane dell'anno.

#### Tabelle Database (`advertising`)
1. **`advertising`**:
   - `id UUID PRIMARY KEY`, `nome TEXT NOT NULL`, `email TEXT NOT NULL`, `status TEXT CHECK (status IN ('active', 'inactive'))`, `created_at TIMESTAMPTZ`.
2. **`promo`**:
   - `id BIGSERIAL PRIMARY KEY`
   - `position INT NOT NULL CHECK (position IN (0, 1, 2))`
   - `section VARCHAR(50) NOT NULL DEFAULT ''`
   - `item JSONB NOT NULL DEFAULT '[]'` (Array di oggetti promo contenenti `width`, `height`, `image_url`, `url`, `mobile_only`, `mobile_image_url`, `new_window`)
   - `year INT NOT NULL DEFAULT EXTRACT(YEAR FROM NOW())`
   - `weeks JSONB NOT NULL DEFAULT '[]'` (Array di interi 1–53 corrispondenti alle settimane lavorative)
   - `active_from TIMESTAMPTZ`, `created_at TIMESTAMPTZ`

#### Matrice di Validazione Posizioni & Sezioni
- **`PromoLeft` (0)**: Dimensioni target 230x0 (proporzione 2:3). Sezioni ammesse: `'week'`, `'raccomended'`, `'discover'`.
- **`PromoRight1` (1)**: Dimensioni target 230x0 (proporzione 2:3). Sezioni ammesse: `'show'`, `'service'`.
- **`PromoRight2` (2)**: Dimensioni target 230x345 (fissa). Sezioni ammesse: `'top'`, `'selected'`.

#### Endpoints API
| Metodo | Path | Descrizione |
|---|---|---|
| `GET` | `/advertising` | Elenco delle pubblicità registrate ordinate per data decrescente. |
| `POST` | `/advertising` | Registrazione di un nuovo inserzionista con validazione email formale. |
| `GET` | `/promo` | Elenco completo delle promozioni/banner configurati. |
| `GET` | `/promotions` | Alias per il recupero delle promozioni arricchite con settimane attive. |
| `GET` | `/promo/week/:week` | Payload JSON con i tre tipi di promo (`PromoLeft`, `PromoRight1`, `PromoRight2`) attivi nella settimana specificata, ordinati per sezione secondo le costanti (`week`, `raccomended`, `discover` / `show`, `service` / `top`, `selected`). |
| `GET` | `/promo/week` | Versione query-based dell'endpoint settimanale (`?week=W&year=Y`, default settimana corrente). |
| `POST` | `/promo` | Creazione di una nuova promozione associata a posizione, sezione e settimane. |
| `PUT` | `/promo/:id` | Aggiornamento dei contenuti e della programmazione di una promozione esistente. |

---

### 2.6 Servizio `images`
Gestisce lo storage persistente e il catalogo delle immagini promozionali, banner e asset generici.

#### Tabelle Database (`images`)
- **`images`**: `uuid UUID PRIMARY KEY`, `description TEXT NOT NULL`, `created_at TIMESTAMPTZ`.

#### Object Storage Bucket
- **`images`** (`objects.NewBucket("images", Public: true)`): Bucket per la conservazione fisica dei binari immagine.

#### Endpoints API
| Metodo | Path | Descrizione |
|---|---|---|
| `GET` | `/images` | Elenco di tutte le immagini registrate con URL pubblico diretto nel bucket. |
| `GET` | `/image/:uuid` | Endpoint raw HTTP che effettua lo streaming binario dell'immagine memorizzata. |
| `POST` | `/image` | Riceve una stringa Base64 (con data-URL o raw), la decodifica, ne estrae il Content-Type, la carica nel bucket e crea/aggiorna il record nel database. |

---

### 2.7 Servizio `broadcasters`
Gestisce le anagrafiche, la catalogazione, i metadati pubblici e i profili operativi avanzati dei broadcaster (artisti/trasmettitori). Fornisce funzionalità di ricerca e filtraggio per categoria, disponibilità e certificazione, oltre a consentire la gestione di servizi offerti, lingue parlate, orari e galleria fotografica.

#### Tabelle Database (`broadcasters`)
1. **`broadcasters`**:
   - `id BIGSERIAL PRIMARY KEY`: Identificativo numerico univoco del profilo broadcaster.
   - `user_id UUID NOT NULL UNIQUE`: Riferimento univoco all'utente (profilo in `profiles.user_profiles`).
   - `display_name TEXT NOT NULL`: Nome d'arte o di trasmissione.
   - `category TEXT NOT NULL`: Categoria principale (es. musica, arte, intrattenimento).
   - `category_color TEXT NOT NULL DEFAULT ''`: Colore identificativo badge della categoria (es. `#3b82f6` o `primary`).
   - `bio TEXT NOT NULL DEFAULT ''`: Descrizione o biografia dell'artista.
   - `certified BOOLEAN NOT NULL DEFAULT FALSE`: Badge certificato verificato.
   - `availability TEXT NOT NULL DEFAULT 'available'`: Stato di disponibilità (`available`, `busy`, `offline`, ecc.).
   - `phone TEXT NOT NULL DEFAULT ''`: Recapito telefonico di contatto.
   - `services JSONB NOT NULL DEFAULT '[]'::jsonb`: Array di stringhe dei servizi offerti.
   - `languages JSONB NOT NULL DEFAULT '[]'::jsonb`: Array di stringhe delle lingue supportate.
   - `attributes JSONB NOT NULL DEFAULT '[]'::jsonb`: Coppie chiave-valore di attributi (`label`, `value`).
   - `hours JSONB NOT NULL DEFAULT '[]'::jsonb`: Programmazione orari per giorno (`day`, `hours: [{ from, to }]`) con validazione backend rigorosa su ordine (`from < to`) e anti-sovrapposizione intervalli (`fromA < toB && fromB < toA`).
   - `photos JSONB NOT NULL DEFAULT '[]'::jsonb`: Galleria immagini del broadcaster (`image_url`, `alt`).
   - `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`.
   - Indici: `idx_broadcasters_user_id`, `idx_broadcasters_category`, `idx_broadcasters_availability`.

#### Controllo Accessi & Autorizzazioni
- La creazione (`POST /broadcasters`) imposta come proprietario l'utente chiamante autenticato. Gli utenti con ruolo `admin` possono specificare un `user_id` arbitrario per conto terzi.
- L'aggiornamento (`PUT /broadcasters/id/:id`) e la cancellazione (`DELETE /broadcasters/id/:id`) sono consentiti esclusivamente al proprietario del profilo (`user_id == callerID`) o agli amministratori di sistema.

#### Endpoints API
| Metodo | Path | Visibilità | Descrizione |
|---|---|---|---|
| `GET` | `/broadcasters/status` | Pubblico | Health check rapido per la disponibilità del servizio. |
| `GET` | `/broadcasters/categories` | Pubblico | Restituisce l'elenco delle categorie ufficiali dei broadcaster e i relativi colori configurati in `config.SiteConfig`. |
| `GET` | `/broadcasters/attribute-presets` | Pubblico | Restituisce i suggerimenti rapidi degli attributi broadcaster configurati in `config.SiteConfig.BroadcasterAttributes`. |
| `GET` | `/broadcasters/availability-options` | Pubblico | Restituisce l'elenco degli stati di disponibilità configurati in `config.SiteConfig.Availability`. |
| `GET` | `/broadcasters/image-config` | Pubblico | Restituisce dimensioni e proporzioni raccomandate per le immagini dei broadcaster (`width: 600`, `proportion: 2:3`). |
| `GET` | `/broadcasters` | Pubblico | Elenco dei broadcaster con ordinamento per certificazione e nome. Filtri opzionali via query: `category`, `certified`, `availability`, `search` (ricerca testuale su `display_name` e `bio`). |
| `GET` | `/broadcasters/id/:id` | Pubblico | Dettaglio del profilo broadcaster dato il suo ID numerico. |
| `GET` | `/broadcasters/user/:userID` | Pubblico | Dettaglio del profilo broadcaster dato l'UUID dell'utente. |
| `POST` | `/broadcasters` | Autenticato | Creazione del profilo broadcaster per l'utente loggato o per terzi (se admin). |
| `PUT` | `/broadcasters/id/:id` | Autenticato | Aggiornamento anagrafica principale del broadcaster (preserva attributi, ore e foto se omessi). |
| `GET` | `/broadcasters/id/:id/attributes` | Pubblico | Lettura della lista degli attributi personalizzati del broadcaster. |
| `PUT` | `/broadcasters/id/:id/attributes` | Autenticato | Aggiornamento selettivo degli attributi personalizzati del broadcaster (proprietario o admin). |
| `GET` | `/broadcasters/id/:id/hours` | Pubblico | Lettura delle fasce orarie programmate del broadcaster. |
| `PUT` | `/broadcasters/id/:id/hours` | Autenticato | Aggiornamento selettivo delle fasce orarie programmate (proprietario o admin). |
| `GET` | `/broadcasters/id/:id/photos` | Pubblico | Lettura della galleria fotografica del broadcaster. |
| `PUT` | `/broadcasters/id/:id/photos` | Autenticato | Aggiornamento selettivo della galleria fotografica e copertina (proprietario o admin). |
| `DELETE` | `/broadcasters/id/:id` | Autenticato | Rimozione definitiva del profilo broadcaster (proprietario o admin). |

---

### 2.8 Servizio `registration`
Orchestra il flusso di registrazione dei nuovi utenti, il controllo dell'affidabilità del dominio email tramite DNS, la conferma via email di benvenuto e il recupero password.

#### Tabelle Database (`registration`)
1. **`welcome_tokens`**:
   - `token UUID PRIMARY KEY`, `user_id UUID NOT NULL`, `email TEXT NOT NULL`, `created_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ` (TTL: 24 ore), `used_at TIMESTAMPTZ`.
2. **`password_reset_tokens`**:
   - `token UUID PRIMARY KEY`, `user_id UUID NOT NULL`, `email TEXT NOT NULL`, `created_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ` (TTL: 1 ora), `used_at TIMESTAMPTZ`.

#### Template Email Incorporati (`embed.FS`)
- `mails/welcome_subject.tmpl`, `welcome_text.tmpl`, `welcome_html.tmpl`
- `mails/password_reset_subject.tmpl`, `password_reset_text.tmpl`, `password_reset_html.tmpl`

#### Endpoints API
| Metodo | Path | Descrizione |
|---|---|---|
| `GET` | `/registration/email/:email` | Verifica che l'email non sia già occupata e controlla l'esistenza di record MX validi via DNS (`net.LookupMX`). |
| `POST` | `/registration` | Registra l'utente, genera un token di benvenuto a 24h e invia la mail transazionale di benvenuto. |
| `POST` | `/registration/welcome/:token` | Valida il token di benvenuto ricevuto via email e ne imposta l'utilizzo. |
| `POST` | `/registration/password-reset` | Richiede il reset password per un'email: se l'account esiste, genera un token a 1h e invia la mail con il link. Risponde sempre con esito positivo per evitare user enumeration. |
| `POST` | `/registration/password-reset/:token` | Consuma il token di reset password valido e imposta la nuova password chiamando `auth.SetUserPassword`. |

---

### 2.9 Servizio `mails`
Gestisce la consegna delle email transazionali generate da `registration`.

- **Outbox locale**: Durante lo sviluppo locale, le email vengono scritte su disco nella cartella `backend/mails/outbox/` in formato `.json` (metadati e testo) e `.html` (corpo formattato per ispezione visiva immediata).
- **Endpoint**: `POST /mails/transactional` (privato, consumato dal servizio `registration`).

---

### 2.10 Servizio `extrahome`
Gestisce la configurazione globale del portale, layout della homepage, metatag SEO (`head`), caroselli e anagrafica broadcaster.

- **Tabelle**: `siteconfig` (`id BIGSERIAL PRIMARY KEY`, `extrsite JSONB NOT NULL`, `start_at TIMESTAMPTZ`, `description TEXT`).
- **Endpoint**: `GET /extrahome/status`.

---

### 2.11 Servizio `hello`
Servizio dimostrativo e di test di base del framework Encore.
- **Endpoint**: `GET /hello/:name` (restituisce `{"message": "Hello, <name>!"}`).

---

## 3. Gestione Database & Migrazioni SQL

Ciascun database ha la propria cartella `migrations/` contenente file SQL con numerazione progressiva:
```
backend/
├── advertising/migrations/
│   ├── 1_create_advertising.up.sql
│   ├── 2_create_promo.up.sql
│   ├── 3_add_section_to_promo.up.sql
│   └── 4_add_weeks_to_promo.up.sql
├── auth/migrations/
│   ├── 1_create_auth_tables.up.sql
│   ├── 2_create_login_results.up.sql
│   └── 3_add_expired_sessions_idx.up.sql
├── broadcasters/migrations/
│   └── 1_create_broadcasters.up.sql
├── extrahome/migrations/
│   └── 1_create_siteconfig.up.sql
├── images/migrations/
│   └── 1_create_images.up.sql
├── profiles/migrations/
│   ├── 1_create_user_profiles.up.sql
│   ├── 2_user_id_uuid.up.sql
│   ├── 3_add_auth.up.sql
│   ├── 4_drop_auth_tables.up.sql
│   ├── 5_add_email.up.sql
│   ├── 6_add_role.up.sql
│   ├── 7_add_status.up.sql
│   ├── 8_artist_profile.up.sql
│   ├── 9_personal_data.up.sql
│   ├── 10_add_personal_data_cap.up.sql
│   ├── 11_personal_setup.up.sql
│   └── 12_member_type.up.sql
└── registration/migrations/
    ├── 1_create_welcome_tokens.up.sql
    └── 2_create_password_reset_tokens.up.sql
```
All'avvio (`encore run`), Encore applica automaticamente tutte le migrazioni mancanti in ordine sequenziale all'interno di transazioni isolate.

---

## 4. Ambiente di Sviluppo, Segreti e Test

### File dei Segreti Locali (`.secrets.local.cue`)
In ambiente di sviluppo locale, i segreti vengono forniti tramite il file CUE nella root del backend:
```cue
PasswordPepper: "local-dev-pepper-secret-string"
LiveKitURL: "http://localhost:7880"
LiveKitAPIKey: "devkey"
LiveKitAPISecret: "secret"
```

### Esecuzione dei Test Automatizzati
Il backend include test unitari e di integrazione completi:
```bash
cd backend
go test -v ./...
```
I test coprono:
- Validazione password e policy di complessità (`auth/password_validation_test.go`).
- Ciclo di vita dei profili e login asincrono (`profiles/login_test.go`, `profiles/validation_test.go`).
- Persistenza e controlli di sicurezza dei setup broadcast personali (`profiles/personal_setup_test.go`).
- Creazione e validazione vincoli delle promozioni e inserzioni (`advertising/promo_test.go`, `advertising/advertising_test.go`).
- Gestione stanze, concurrency lock e token LiveKit (`videochat/open_room_test.go`, `videochat/rooms_test.go`).
- Validazione dati e parametri dei profili broadcaster (`broadcasters/broadcasters_test.go`).
