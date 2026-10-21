# Report di supervisione del codice – progetto `extra`

**Data:** 21 ottobre 2026
**Perimetro:** `backend/` (Go + Encore, 11 servizi), `frontend/` (Quasar/Vue 3), `publicsite/` (Quasar SSG), `docker/`, `livekit/`, `mockup/`, file di configurazione nella radice.
**Modalità:** analisi statica in sola lettura. **Nessun file del progetto è stato modificato.** Le correzioni sono descritte qui e nel file allegato `2026-10-21_proposte-modifiche.md`.

---

## 1. Sintesi

Il progetto è ben avviato: i servizi Encore sono separati in modo leggibile, le query SQL usano sempre parametri (non ho trovato SQL injection), le password usano bcrypt con un pepper, ci sono test Go in 7 pacchetti e il flusso di reset password non rivela se un'email esiste.

Però **oggi non è pronto per andare online**. Ci sono tre blocchi di problemi:

1. **Endpoint di scrittura pubblici senza autenticazione.** Promo, immagini, inserzionisti e chatlog accettano scritture da chiunque. Messi insieme ai punti 2 e 3, permettono di **eseguire JavaScript nel sito pubblico e nel pannello admin**.
2. **Il segreto del pepper è nel repository git** e ricompare come valore di default in 5 file Docker/env.
3. **Il deploy Docker non può funzionare così com'è.** Il backend parte in modalità sviluppo (`encore run`), il frontend chiama `http://localhost:4000`, e il proxy di Caddy non copre metà delle rotte.

| Area | Valutazione | Commento breve |
|---|---|---|
| Sicurezza | 🔴 Critica | 6 problemi critici, 9 alti |
| Correttezza funzionale | 🟠 Da sistemare | Bug reali in chat, alzata di mano, login, conferma email |
| Architettura | 🟡 Migliorabile | Servizi che leggono i DB degli altri, due sistemi di chat paralleli |
| Infrastruttura / deploy | 🔴 Non funzionante in produzione | Vedi §5 |
| Qualità / test | 🟡 Discreta | Test backend presenti; frontend senza test né linter |

---

## 2. Elenco priorizzato dei problemi

Legenda: 🔴 Critico · 🟠 Alto · 🟡 Medio · ⚪ Basso. I riferimenti sono `file:riga` relativi alla cartella `extra/`.

### 🔴 Critici (da risolvere prima di qualsiasi messa online)

| ID | Problema | Dove | Impatto |
|---|---|---|---|
| C1 | **Promo creabili e modificabili da chiunque** (`public`, nessun controllo sul ruolo) | `backend/advertising/promo.go:266`, `:309` | Chiunque può inserire banner e link nella home pubblica |
| C2 | **XSS sul sito pubblico tramite link promo**: `:href="c.url"` senza validazione. Vue non blocca `javascript:` | `publicsite/src/pages/IndexPage.vue:12`, `:59`, `:80` | Con C1 → JavaScript eseguito nel browser di ogni visitatore |
| C3 | **Upload/sovrascrittura immagini anonimo**: `SaveImage` è `public`. Accetta qualsiasi `content_type` (anche `text/html`, `image/svg+xml`) e sovrascrive qualsiasi immagine esistente dato il suo UUID, ricavabile da `GET /images`. `owner` non viene mai valorizzato | `backend/images/images.go:278`, `:299`, `:314` | Defacement, hosting di contenuti malevoli sul bucket pubblico |
| C4 | **XSS memorizzato nel pannello**: la didascalia foto (`photo.alt`), scritta dal broadcaster, viene mostrata con `v-html` senza sanitizzazione | `frontend/src/components/broadcasters/EditPhotosDialog.vue:127`, `:175` | Un broadcaster può eseguire codice nella sessione dell'admin. Il token è in `localStorage` (C4 + F1) e può essere rubato |
| C5 | **Pepper delle password versionato in git**: il file `backend/.secrets.local copy.cue` è tracciato (verificato sull'indice git). `.gitignore` ignora solo `.secrets.local.cue`, scritto due volte. Lo stesso valore è default in `docker/entrypoint-backend.sh`, `docker/docker-compose.yml`, `docker/.env.example` e `docker/.env` | vedi file citati | Il pepper non protegge più gli hash in caso di furto del DB |
| C6 | **Un utente qualsiasi può auto-abilitarsi e auto-certificarsi come broadcaster**: `CreateBroadcaster`/`UpdateBroadcaster` accettano `enabled`, `used` e `certified` anche da non-admin | `backend/broadcasters/broadcasters.go:556-563`, `:640-647`, `:666-669` | Aggira il flusso di approvazione admin (`admin.UpdateProfileMemberType`) |

### 🟠 Alti

| ID | Problema | Dove |
|---|---|---|
| H1 | **Lo stato del profilo non viene mai controllato**: utenti `Banned`, `Suspended` o `Deleted` possono fare login e usare la sessione. Lo `status` non compare in nessuna query di autorizzazione | `backend/profiles/login.go:40`, `backend/auth/auth.go:57-77` |
| H2 | **Chatlog pubblico in scrittura e lettura** su qualsiasi stanza, senza limite di dimensione, incluso il tipo `CMD` | `backend/broadcasters/chatlog.go:35`, `:77` |
| H3 | **Identità falsificabile in chat**: il client invia `senderId`, `senderName`, `isOwner` e `type: 'ADMIN_ALERT'` dentro il payload e i riceventi ci credono (ignorano `participant.identity`). Il token LiveKit dei viewer ha `CanPublishData=true` | `backend/videochat/livekit.go:158`, `frontend/src/components/videochat/BroadcastChat.vue:312-340`, `BroadcastStageRight.vue:490-505`, `frontend/src/stores/chat-store.ts:86-101` |
| H4 | **Messaggi `kind: "system"` inviabili da chiunque** | `backend/videochat/messages.go:145` |
| H5 | **`SendCommand` senza autorizzazione**: qualsiasi utente loggato pubblica `command.*` (mute/clear/alert) su qualsiasi stanza. Oggi l'handler è vuoto, ma il rischio esplode appena verrà implementato | `backend/videochat/messages.go:259` |
| H6 | **Elenco inserzionisti pubblico con email** e creazione anonima | `backend/advertising/advertising.go:93`, `:126` |
| H7 | **Dati personali esposti**: `GET /broadcasters` (pubblico) restituisce telefono, `user_id` e anche i profili disabilitati (`enabled=false`) | `backend/broadcasters/broadcasters.go:396-464` |
| H8 | **Upload avatar senza limiti**: nessun limite di dimensione, `Content-Type` scelto dal client, bucket pubblico | `backend/profiles/avatars.go:35-47` |
| H9 | **Logout solo lato client**: la sessione resta valida lato server per 7 giorni. Il token sta in `localStorage` | `frontend/src/stores/profiles-store.ts:54-57`, `:101-103` |

### 🟡 Medi

| ID | Problema | Dove |
|---|---|---|
| M1 | **Transazione lasciata aperta (con riga bloccata `FOR UPDATE`)**: in `setPasswordWithCurrentCheck` il ramo "password riutilizzata" ritorna con `err == nil`, quindi la `defer` non fa rollback. Stesso schema in `replacePassword` quando `RowsAffected()==0` | `backend/auth/credentials.go:103-106`, `:122-124`, `:151-154`, `:167-169` |
| M2 | **Bug alzata di mano**: dopo 5 minuti una richiesta `pending` scaduta resta `pending`. L'indice unico parziale `WHERE status='pending'` fa fallire ogni nuova richiesta con errore 500, e l'utente non può più chiedere la parola in quella stanza. Il commento parla di `ON CONFLICT` ma la query non lo contiene | `backend/videochat/speak_requests.go:88-125`, migrazione `videochat/2_…sql` |
| M3 | **Lease dello speaker non applicato**: dopo 30 minuti lo slot risulta libero ma il permesso LiveKit di pubblicare non viene revocato, quindi possono esserci 2 speaker contemporanei | `backend/videochat/speak_requests.go:18`, `:327-331` |
| M4 | **`ListMessages` restituisce i 50 messaggi più vecchi** (`ORDER BY created_at ASC LIMIT`) invece degli ultimi | `backend/videochat/messages.go:226-232` |
| M5 | **Login sensibile a maiuscole**: la registrazione normalizza l'email in minuscolo, il login no. Anche `CheckEmail` non normalizza | `backend/profiles/login.go:40`, `backend/registration/registration.go:107-128` |
| M6 | **Reset password non atomico**: il token viene "consumato" dopo il cambio password, quindi due richieste concorrenti con lo stesso token passano entrambe. Le sessioni esistenti non vengono revocate | `backend/registration/registration.go:276-301` |
| M7 | **Conferma email senza effetto**: `ConfirmWelcome` segna il token ma non cambia lo stato del profilo, che nasce già `Active` (0). Inoltre `POST /profiles` è un secondo endpoint pubblico di registrazione che salta email e token | `backend/registration/registration.go:182-211`, `backend/profiles/profiles.go:146` |
| M8 | **Registrazione non atomica**: se `auth.Register` fallisce, il profilo resta senza credenziali e l'email risulta "già registrata". Stessa cosa in `admin.InsertProfile`, che in più non valida né normalizza i dati | `backend/profiles/profiles.go:156-170`, `backend/admin/admin.go:98-119` |
| M9 | **Cancellazione profilo incompleta**: restano credenziali, sessioni attive (l'`AuthHandler` accetta utenti senza profilo, `auth.go:71-77`), broadcaster e immagini | `backend/profiles/profiles.go:249`, `backend/admin/admin.go:339` |
| M10 | **`OpenRoom` contraddice il suo commento**: chi ha `member_type='broadcaster'` può aprire stanze anche senza `can_broadcast`. Il match avviene anche su `display_name`, che l'utente può modificare, con rischio di collisioni e blocchi. Il lock advisory resta aperto durante chiamate LiveKit fino a 35 s | `backend/videochat/open_room.go:36`, `:56-66` |
| M11 | **Le GET hanno effetti collaterali**: `ListRooms` e `CreateToken` scrivono in `broadcast_sessions`, che viene "sincronizzata" a ogni lista con scansione completa. Il check-then-insert non è atomico e manca un vincolo unico | `backend/videochat/rooms.go:84-114`, `sessions.go:36-75`, `tokens.go:45` |
| M12 | **Admin senza validazione** su `UpdateProfile` e `UpsertPersonalData`, che bypassano le regole applicate all'utente | `backend/admin/admin.go:124`, `:394` |
| M13 | **Cambio member type non transazionale** su due DB (profiles + broadcasters): un errore a metà lascia stati incoerenti | `backend/admin/admin.go:233-334` |
| M14 | **Validazione orari per confronto di stringhe**: `"9:00" >= "10:00"` risulta vero. Il formato non è verificato | `backend/broadcasters/broadcasters.go:141` |
| M15 | **Errori JSON ignorati** nella lettura dei campi JSONB (`_ = json.Unmarshal`) | `backend/broadcasters/broadcasters.go:365-388` |
| M16 | **Nessun rate limiting** su login, registrazione, reset password e `CheckEmail` (che fa anche un lookup DNS MX a ogni chiamata) | servizi `profiles`, `registration` |

### ⚪ Bassi / qualità

- **L1** – Il servizio `hello` (template Encore) espone un catch-all `/!path` che risponde 200 HTML a qualsiasi rotta sconosciuta e maschera i 404 (`backend/hello/index.go:12`).
- **L2** – La configurazione sito ha valori di prova e refusi: categoria `"test"`, `"prova attributo"`, `"autcall"` (`backend/config/config.go:32`, `:50-51`), `raccomended` (`advertising/promo.go`).
- **L3** – I tipi in `backend/extrahome/data.go` non sono usati, e `ID int64` non corrisponde agli UUID reali dei broadcaster.
- **L4** – Le migrazioni sono state modificate dopo essere state applicate: i commenti di `broadcasters/3_…sql` citano colonne "aggiunte nella migrazione precedente" che invece sono nella 4.
- **L5** – Il vincolo unico su `email` è case-sensitive (`profiles/5_add_email.up.sql`). Meglio `citext` o un indice su `lower(email)`.
- **L6** – Le mail non vengono mai spedite: finiscono come file in `mails/outbox` (con link di reset in chiaro su disco). Il TODO SMTP è ancora aperto.
- **L7** – Codice duplicato: `Profile`, `RegisterParams`, `PersonalData`, `isUniqueViolation` e `normalizeEmail` sono copiati tra `admin`, `profiles`, `registration` e `advertising`.
- **L8** – Frontend: componenti molto grandi (`EditPhotosDialog.vue` 47 KB, `PromosPage.vue` 44 KB, `i18n/index.ts` 119 KB); due dialog quasi omonimi (`ImageDialog.vue` / `ImmageDialog.vue`); refuso `WideochatPage`.
- **L9** – Frontend senza ESLint, senza test unitari e senza guard sul router (il controllo admin è solo in `AdminLayout.vue:100`, il backend comunque protegge).
- **L10** – Messaggi d'errore misti italiano/inglese tra i servizi.

---

## 3. Architettura – osservazioni

1. **I servizi leggono i database degli altri.** `auth` legge `profiles`, `profiles` legge `auth.login_results`, mentre `admin`, `registration` e `videochat` leggono e scrivono su `profiles` e `broadcasters` con `sqldb.Named(...)`. Così i confini tra servizi Encore non valgono più: le migrazioni di un servizio possono rompere gli altri e non ci sono transazioni tra DB. **Proposta:** ogni servizio espone API (anche `private`) per i dati propri, e gli altri le chiamano.
2. **Il login passa da Pub/Sub e polling.** `profiles.Login` pubblica la password in chiaro su un topic, poi interroga il DB di `auth` ogni 50 ms per al massimo 10 s. È più complesso, più lento e meno sicuro di una chiamata diretta. Inoltre `LoginResults` viene pubblicato ma nessuno lo consuma, e una riconsegna "at-least-once" crea sessioni orfane. **Proposta:** chiamare direttamente `auth.Login` (endpoint `private`) passando l'email risolta.
3. **Ci sono due sistemi di chat.** `broadcasters.chatlog` (pubblico, payload JSON libero) e `videochat.messages` (autenticato, con idempotenza) convivono, e anche due componenti Vue fanno la stessa cosa (`BroadcastChat.vue`, `BroadcastStageRight.vue`). **Proposta:** tenere solo `videochat.messages`, inviare via LiveKit solo l'ID del messaggio salvato (oppure usare il server come fonte autorevole) e dismettere `chatlog`.
4. **Le sessioni di trasmissione andrebbero lette dagli eventi LiveKit**, non dedotte "per differenza" a ogni GET. **Proposta:** un endpoint `raw` che riceve i webhook LiveKit (`room_started`, `room_finished`, `participant_left`) e un vincolo `UNIQUE (sid, action)`.
5. **La configurazione del sito è codice Go** (`config/config.go`), mentre esiste già la tabella `extrahome.siteconfig`, non ancora usata. Conviene scegliere una delle due.

---

## 4. Frontend e publicsite

- **F1** – Token di sessione in `localStorage`: qualsiasi XSS (C2, C4) lo può rubare. Il backend supporta già il cookie `session`. **Proposta:** cookie `HttpOnly; Secure; SameSite=Lax` impostato dal login, rimuovendo il token da JS.
- **F2** – URL del backend fisso a `http://localhost:4000` (`Local` in `encore/client.ts:13`, `chat-store.ts:25`, `publicsite/src/pages/IndexPage.vue:129`). In produzione il browser dell'utente chiamerebbe il proprio computer. **Proposta:** base URL da variabile d'ambiente (`process.env.API_BASE_URL` in `quasar.config.ts`) oppure percorso relativo dietro Caddy.
- **F3** – Ogni store crea un proprio `new Client(Local, …)` con la stessa logica di autenticazione. Meglio un unico client in un boot file.
- **F4** – `v-html` va eliminato oppure filtrato con DOMPurify (vedi C4). `stripHtml` con regex non è una sanitizzazione.

---

## 5. Infrastruttura e deploy (`docker/`, `livekit/`)

| ID | Problema |
|---|---|
| D1 | `Dockerfile.backend` avvia `encore run` (server di sviluppo con dashboard su 9400 esposta). Le variabili `POSTGRES_*` del compose **non vengono usate**: `encore run` cerca di creare un proprio Postgres via Docker, che nel container non c'è. Serve `encore build docker` con un file di infrastruttura (`infra-config.json`) che punti al Postgres del compose |
| D2 | `golang:1.24` nel Dockerfile, ma `go.mod` richiede `go 1.26` |
| D3 | Installazione di Encore via `curl … \| bash` senza versione fissata; anche `livekit/livekit-server:latest` non ha versione fissata |
| D4 | Caddy fa proxy solo di `/xxx/*`: restano fuori `/profiles`, `/broadcasters`, `/advertising`, `/images` (senza slash finale), `/image/:uuid`, `/promo`, `/promotions`. Queste richieste ricevono `index.html` |
| D5 | LiveKit nel compose principale espone solo `7882/udp`, ma `livekit.yaml` usa il range `50000-50020/udp`, quindi i flussi audio/video non passeranno. Mancano anche `node_ip`/`use_external_ip` per un host reale |
| D6 | Postgres esposto sull'host (`5432`) con password di default; dashboard Encore esposta (`9400`) |
| D7 | Chiavi LiveKit `devkey/secret` come default ovunque |
| D8 | `Dockerfile.frontend` usa `npm install`, ma il progetto usa pnpm: il lockfile viene ignorato e la build non è riproducibile |
| D9 | Caddy con `auto_https off` ma porte 443 esposte; mancano header di sicurezza (CSP, `X-Content-Type-Options`, `Referrer-Policy`) |
| D10 | Solo `backend/` è un repository git: `frontend/`, `publicsite/`, `docker/` e `livekit/` non risultano versionati (almeno non in `extra/` o in `CODE/`) |

---

## 6. Piano d'azione proposto

**Fase 0 – subito (prima di qualunque esposizione pubblica)**
1. Rimuovere `backend/.secrets.local copy.cue` dal repository e dalla storia git, correggere `.gitignore` (`.secrets*.cue`), generare un nuovo pepper e togliere tutti i default dai file Docker (C5). Il pepper entra nell'hash, quindi la rotazione richiede una strategia: vedi proposta P5.
2. Mettere `auth` + controllo ruolo admin su promo, immagini, inserzionisti; rendere `chatlog` autenticato oppure eliminarlo (C1, C3, H2, H6).
3. Validare gli URL (`http`/`https` soltanto) lato backend e lato publicsite (C2).
4. Eliminare `v-html` o sanitizzare con DOMPurify (C4).
5. Ignorare `enabled`, `used` e `certified` se chi chiama non è admin (C6).

**Fase 1 – entro il prossimo rilascio**
- Controllare lo stato del profilo in login e `AuthHandler` (H1); revocare le sessioni su cambio/reset password e ban (M6, H9).
- Correggere M1, M2, M4, M5 (piccoli interventi localizzati).
- Chat: identità dal server, `CanPublishData=false` per i viewer, rimozione di `kind: "system"` lato client (H3, H4, H5).
- Nascondere telefono e profili disabilitati negli endpoint pubblici (H7); limiti all'upload (H8).

**Fase 2 – deploy**
- Build di produzione Encore + infra config, base URL frontend configurabile, rotte Caddy corrette, porte LiveKit, segreti solo da `.env` non versionato (D1–D9).

**Fase 3 – struttura**
- Togliere gli accessi cross-DB, semplificare il login, unificare la chat, usare i webhook LiveKit, aggiungere rate limiting, ESLint e test frontend.

---

## 7. Note sul metodo

- **File letti integralmente:** tutto il codice Go non generato dei servizi (`auth`, `profiles`, `admin`, `registration`, `mails`, `broadcasters`, `videochat`, `images`, `advertising`, `extrahome`, `config`, `hello`), tutte le migrazioni SQL, i template email, i file Docker/LiveKit/Caddy. Dal frontend: store, router, layout, pagine e componenti principali (chat, foto, broadcast). Dal publicsite: home, router e renderer SSG.
- **Non letti in dettaglio:** `encore.gen.go` e `encore/client.ts` (generati), `i18n/index.ts`, le pagine minori del frontend, la cartella `documments/`, gli asset in `mockup/`.
- **Non eseguito:** test (`encore test ./...`), `vue-tsc`, build. Le verifiche sono statiche. L'unica verifica "dinamica" è stata leggere l'indice git di `backend/` per confermare che il file dei segreti è tracciato.
- I valori dei segreti **non sono riportati** in questo report.
