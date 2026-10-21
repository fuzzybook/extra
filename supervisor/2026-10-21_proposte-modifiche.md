# Proposte di modifica – progetto `extra`

**Data:** 21 ottobre 2026 · da leggere insieme a `2026-10-21_report-supervisione.md`

Queste sono **proposte**: nessun file del progetto è stato toccato. Ogni proposta indica il problema del report che risolve (ID tra parentesi), i file coinvolti e uno snippet indicativo. Gli snippet sono scritti per essere applicati a mano o da un agente; vanno compilati e testati (`encore test ./...`, `pnpm typecheck`).

---

## P1 – Proteggere gli endpoint di scrittura (C1, C3, H2, H6)

**Idea:** un unico helper esportato da `auth` per il controllo admin, usato ovunque al posto delle copie locali (`videochat/livekit.go:requireAdmin`, `profiles/personal_setup.go`, `broadcasters.authedUser`).

`backend/auth/roles.go` – aggiungere:

```go
// RequireAdmin returns an error unless the current request is authenticated as admin.
func RequireAdmin() error {
	if _, ok := auth.UserID(); !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	data, _ := auth.Data().(*AuthData)
	if data == nil || !data.Role.IsAdmin() {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin role required"}
	}
	return nil
}
```

Poi cambiare le annotazioni e aggiungere il controllo:

| Endpoint | Da | A |
|---|---|---|
| `advertising/promo.go` `CreatePromo`, `UpdatePromo` | `public` | `auth` + `authsvc.RequireAdmin()` |
| `advertising/advertising.go` `List`, `Create` | `public` | `auth` + `RequireAdmin()` (se serve un modulo pubblico "diventa inserzionista", farlo come endpoint separato che non restituisce la lista) |
| `images/images.go` `SaveImage` | `public` | `auth`; impostare `owner = uid` all'INSERT; in UPDATE consentire solo a `owner` o admin |
| `images/images.go` `ListImages` | `public` | restituire pubblicamente solo `approved = TRUE`; la lista completa solo agli admin |
| `broadcasters/chatlog.go` `AddChatLog`, `ListChatLog` | `public` | da **eliminare** a favore di `videochat.messages` (vedi P12). Se resta: `auth`, limite di `len(Data)` (es. 4 KB), tipo `CMD` riservato al proprietario della stanza |

In `SaveImage`, accettare solo tipi immagine raster e limitare la dimensione:

```go
var allowedImageTypes = map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}
const maxImageBytes = 5 << 20 // 5 MB

if len(rawBytes) > maxImageBytes { return nil, &errs.Error{Code: errs.InvalidArgument, Message: "immagine troppo grande"} }
detected := http.DetectContentType(rawBytes) // non fidarsi del data: URL
if !allowedImageTypes[detected] { return nil, &errs.Error{Code: errs.InvalidArgument, Message: "formato non supportato"} }
```

Applicare la stessa regola a `profiles/avatars.go` (H8), leggendo il body con `http.MaxBytesReader(w, req.Body, maxImageBytes)`.

---

## P2 – Validare gli URL salvati e mostrati (C2)

**Backend** – helper condiviso (ad es. nuovo pacchetto `backend/validate`):

```go
func HTTPURL(raw string) error {
	if raw == "" { return nil }
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "URL non valido: ammessi solo http/https"}
	}
	return nil
}
```

Da usare in `CreatePromoParams.Validate`/`UpdatePromoParams.Validate` (per ogni `PromoContent.URL`, `ImageURL`, `MobileImageURL`), in `BroadcasterPhoto.ImageURL` e in `ProfileParams.AvatarURL` (oggi `url.ParseRequestURI` accetta anche `javascript:`).

**Publicsite** – difesa in profondità in `src/pages/IndexPage.vue`:

```ts
function safeUrl(u?: string): string | undefined {
  try { const p = new URL(u ?? '', window.location.origin); return ['http:', 'https:'].includes(p.protocol) ? p.href : undefined; }
  catch { return undefined; }
}
```
```html
<a v-if="safeUrl(c.url)" :href="safeUrl(c.url)" target="_blank" rel="noopener noreferrer">
```

---

## P3 – Eliminare l'XSS delle didascalie (C4, F4)

In `frontend/src/components/broadcasters/EditPhotosDialog.vue:127` e `:175`:

- **Opzione A (consigliata):** usare testo semplice, cioè `{{ stripHtml(photo.alt) }}` al posto di `v-html`, e salvare solo testo.
- **Opzione B:** se serve la formattazione (grassetto/corsivo dal `q-editor`), `pnpm add dompurify` e:

```ts
import DOMPurify from 'dompurify';
const cleanHtml = (h: string) => DOMPurify.sanitize(h, { ALLOWED_TAGS: ['b', 'i', 'u', 'strong', 'em', 'br'], ALLOWED_ATTR: [] });
```
```html
<span v-html="cleanHtml(photo.alt)" />
```

Sanitizzare anche **lato server** in `UpdateBroadcasterPhotos` (ad es. con `github.com/microcosm-cc/bluemonday` `StrictPolicy()` o `UGCPolicy()`), perché i dati potrebbero essere mostrati altrove, per esempio sul sito pubblico.

---

## P4 – Campi riservati agli admin nei broadcaster (C6)

In `backend/broadcasters/broadcasters.go`, `CreateBroadcaster` e `UpdateBroadcaster`, subito dopo `authedUser()`:

```go
if !isAdmin {
	// Campi gestiti solo dall'amministrazione
	p.Enabled, p.Used = nil, nil
	p.Certified = existing.Certified // in Update; in Create: p.Certified = false
}
```

In `CreateBroadcaster` per i non-admin forzare `enabled=false` e `certified=false`, e valutare se permettere la creazione solo a chi ha `member_type='broadcaster'`. Aggiornare anche `profiles.broadcaster` (oggi lo fa solo `admin.UpdateProfileMemberType`), oppure riservare la creazione all'admin.

Nella stessa occasione (H7): `ListBroadcasters`/`GetBroadcaster` pubblici devono filtrare `WHERE enabled = TRUE` e **non** restituire `phone` e `user_id`. Si può definire un `PublicBroadcaster` senza quei campi.

---

## P5 – Segreti e rotazione del pepper (C5, D7)

1. `backend/.gitignore`: sostituire le due righe `.secrets.local.cue` con
   ```
   .secrets*.cue
   *.cue.bak
   ```
2. `git rm --cached ".secrets.local copy.cue"` e ripulire la storia (`git filter-repo --path ".secrets.local copy.cue" --invert-paths`), se il repo è stato condiviso o pushato.
3. Togliere i valori di default da `docker/entrypoint-backend.sh` e `docker/docker-compose.yml`, facendo fallire l'avvio se manca la variabile:
   ```bash
   : "${PASSWORD_PEPPER:?PASSWORD_PEPPER non impostato}"
   ```
   In `.env.example` lasciare un segnaposto (`PASSWORD_PEPPER=__generare_con_openssl_rand_hex_32__`).
4. **Rotazione senza bloccare gli utenti:** l'hash è `bcrypt(sha256(password+pepper))`, quindi serve una fase di transizione:
   ```go
   var secrets struct {
   	PasswordPepper    string
   	PasswordPepperOld string // opzionale, solo durante la migrazione
   }
   // in Login: se verifyPassword(hash, pw) fallisce e PasswordPepperOld != "",
   // riprovare con il vecchio pepper; se combacia → ri-hashare con il nuovo e salvare.
   ```
   Dopo qualche settimana, chi non ha fatto login riceve un reset password forzato e `PasswordPepperOld` viene rimosso.
5. In `auth.pepperedPassword` valutare `HMAC-SHA256(pepper, password)` al posto di `sha256(password+pepper)`: è la costruzione standard per un pepper.

---

## P6 – Applicare lo stato del profilo (H1, M9)

In `backend/auth/auth.go`, `AuthHandler`: leggere anche `status` e rifiutare i non attivi; rifiutare anche i profili mancanti.

```go
var roleStr string
var status int
err = profilesDB.QueryRow(ctx, `SELECT role, status FROM user_profiles WHERE user_id = $1`, userID).Scan(&roleStr, &status)
if errors.Is(err, sqldb.ErrNoRows) || (err == nil && status != 0 /* Active */) {
	return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "account non attivo"}
}
```

Fare lo stesso controllo in `profiles.Login`, prima di pubblicare la richiesta. Quando l'admin cambia lo stato verso Suspended/Banned/Deleted, cancellare le sessioni dell'utente (oggi `admin` non può farlo: serve un endpoint `private` in `auth`, per esempio `RevokeUserSessions(userID)`).

Per `DeleteProfile` (utente e admin): chiamare `auth.DeleteCredentials(userID)` (nuovo endpoint `private`), che cancella credenziali e sessioni, più la pulizia di broadcaster e immagini.

---

## P7 – Rollback sicuro delle transazioni (M1)

In `backend/auth/credentials.go`, sostituire le due `defer func(){ if err != nil { tx.Rollback() } }()` con:

```go
tx, err := db.Begin(ctx)
if err != nil { … }
defer tx.Rollback() // no-op dopo un Commit riuscito
```

In questo modo tutti i `return` anticipati (password riutilizzata, `RowsAffected()==0`) rilasciano la connessione e il lock `FOR UPDATE`.

---

## P8 – Alzata di mano: richieste scadute (M2, M3)

In `backend/videochat/speak_requests.go`, `CreateSpeakRequest`, prima del controllo di esistenza:

```go
_, _ = db.Exec(ctx, `
	UPDATE videochat_speak_requests
	SET status = 'expired', updated_at = NOW()
	WHERE room_name = $1 AND requester_user_id = $2 AND status = 'pending' AND expires_at <= NOW()
`, roomName, string(uid))
```

Gestire anche la corsa concorrente: se l'INSERT fallisce con violazione unica (`23505`), rileggere la richiesta pending e restituirla.

Per il lease (M3): un `cron.NewJob` ogni minuto che trova gli slot `active`/`reserved` con `lease_expires_at < NOW()`, chiama `UpdateParticipant(CanPublish:false)` e libera lo slot. In alternativa, la stessa logica può girare all'inizio di `ApproveSpeakRequest` e `GetSpeakerState`.

---

## P9 – Ultimi messaggi in chat (M4)

`backend/videochat/messages.go:226`, con lo stesso schema già usato in `chatlog.go`:

```sql
SELECT * FROM (
  SELECT id, room_name, sender_user_id, sender_name, sender_avatar, kind, body, client_message_id, created_at
  FROM videochat_messages
  WHERE room_name = $1
  ORDER BY created_at DESC, id DESC
  LIMIT $2
) t ORDER BY created_at ASC, id ASC
```

Aggiungere anche `rows.Err()` dopo il ciclo (manca).

---

## P10 – Email normalizzata ovunque (M5, L5)

- `profiles/login.go`: `email, err := normalizeEmailValue(p.UserEmail)` prima della query.
- `registration.CheckEmail`: usare `normalizeEmail`.
- `admin.InsertProfile`: chiamare `(&profilessvc.RegisterParams{…}).Validate()`, o meglio delegare a `profiles.Insert`.
- Migrazione `profiles/14_email_lower_unique.up.sql`:
  ```sql
  UPDATE user_profiles SET email = lower(btrim(email));
  DROP INDEX IF EXISTS user_profiles_email_idx;
  CREATE UNIQUE INDEX user_profiles_email_lower_idx ON user_profiles (lower(email));
  ```
  Prima verificare che non esistano duplicati che differiscono solo per maiuscole.

---

## P11 – Reset password atomico e revoca sessioni (M6, H9)

`backend/registration/registration.go`, `ConfirmPasswordReset`: consumare il token **prima** di cambiare la password, con un solo UPDATE condizionato:

```go
err := db.QueryRow(ctx, `
	UPDATE password_reset_tokens SET used_at = NOW()
	WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()
	RETURNING user_id, email
`, token).Scan(&userID, &email)
if errors.Is(err, sqldb.ErrNoRows) { return nil, &errs.Error{Code: errs.InvalidArgument, Message: "token non valido o scaduto"} }
```

In `auth.replacePassword` e `setPasswordWithCurrentCheck`, nella stessa transazione:
```sql
DELETE FROM sessions WHERE user_id = $1  -- (nel cambio volontario: AND token <> sessione corrente)
```

Frontend (`profiles-store.ts:101`): `logout()` deve chiamare `client.auth.Logout()` prima di `setToken(null)`.

---

## P12 – Chat affidabile (H3, H4, H5, architettura §3.3)

1. `videochat/livekit.go:158`: `grant.SetCanPublishData(owner)`, cioè i viewer non inviano data message.
2. I viewer inviano i messaggi **solo** tramite `POST /videochat/rooms/:room/messages`. Il server salva il messaggio e lo inoltra nella stanza con `client.SendData(...)` del Room Service, dichiarando mittente e nome dal DB. In questo modo il server è l'unica fonte dell'identità.
3. `messages.go:145`: ignorare `p.Kind` dal client (sempre `"text"`); i messaggi `system` li crea solo il server.
4. `SendCommand`: `requireRoomOwner(...)` o `auth.RequireAdmin()`, e un elenco chiuso di comandi ammessi.
5. Frontend: in `handleDataReceived` usare `participant?.identity` per riconoscere il mittente, e mostrare `ADMIN_ALERT` solo se il mittente è l'owner della stanza (o arriva dal server, `participant === undefined`).
6. Dismettere `broadcasters/chatlog.go` e `chat-store.saveMessage`; tenere un solo componente chat al posto dei due attuali.

---

## P13 – Login senza Pub/Sub (architettura §3.2)

Sostituire `profiles.Login` con una chiamata diretta:

```go
resp, err := authsvc.Login(ctx, &authsvc.LoginParams{UserID: userID, Password: p.Password})
if err != nil { return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"} }
```

Poi rimuovere `LoginRequests`, `LoginResults`, la subscription, la tabella `login_results` (nuova migrazione `DROP TABLE`), il polling e il cron di pulizia. La password non viaggia più su un topic e la latenza del login scende.

---

## P14 – Deploy di produzione (D1–D9, F2)

**Backend:** sostituire `encore run` con la build ufficiale:
```bash
encore build docker --config infra-config.json extra-backend:VERSIONE
```
In `infra-config.json` definire `sql_servers` (host `postgres`, credenziali da env), `secrets` (da env) e `object_storage`. Dopo questo, `Dockerfile.backend` ed `entrypoint-backend.sh` non servono più. Allineare comunque `golang:1.26` se si mantiene una build custom.

**Frontend/publicsite:** base URL configurabile. In `quasar.config.ts` → `build.env: { API_BASE_URL: process.env.API_BASE_URL ?? '' }`, e creare il client con `new Client(process.env.API_BASE_URL || window.location.origin, …)`. Nel `Dockerfile.frontend` usare `corepack enable && pnpm install --frozen-lockfile`.

**Caddy:** è più semplice dare a tutte le API un prefisso unico. In alternativa, elencare anche le rotte esatte:
```
@backend path /auth/* /profiles /profiles/* /admin/* /registration /registration/* /broadcasters /broadcasters/* /advertising /advertising/* /promo /promo/* /promotions /images /image/* /videochat/* /extrahome/*
header {
	X-Content-Type-Options nosniff
	Referrer-Policy strict-origin-when-cross-origin
	Content-Security-Policy "default-src 'self'; img-src 'self' https: data:; connect-src 'self' wss:; frame-ancestors 'none'"
}
```

**LiveKit:** nel compose principale aggiungere `"50000-50020:50000-50020/udp"`, impostare `rtc.use_external_ip: true` (o `node_ip`) in produzione, chiavi da env, immagine con versione fissata.

**Rete:** togliere `ports` da `postgres` e la porta `9400`, così restano raggiungibili solo sulla rete interna.

**CORS:** con l'app servita da un dominio diverso dall'API, configurare `global_cors.allow_origins_with_credentials` in `encore.app`. Da verificare dopo la scelta del dominio.

---

## P15 – Pulizia e qualità (L1–L10)

- Rimuovere il servizio `hello` (o almeno il catch-all `/!path`).
- Ripulire `config/config.go` dai valori di test e correggere `autcall` → `outcall`. Attenzione: è un valore salvato nel DB, quindi serve una migrazione dati.
- Estrarre i tipi e le funzioni duplicate (`normalizeEmail`, `isUniqueViolation`, `Profile`…) in un pacchetto condiviso non-servizio, per esempio `backend/internal/...`.
- Non modificare più migrazioni già applicate: ogni modifica va in una nuova migrazione.
- Frontend: aggiungere ESLint (`@vue/eslint-config-typescript`) e Vitest. Partire dai test dello store `profiles` e da `stripHtml`/`safeUrl`.
- Spezzare `EditPhotosDialog.vue` e `PromosPage.vue` in sotto-componenti; unificare `ImageDialog`/`ImmageDialog`; rinominare `WideochatPage`.
- Mettere sotto git `frontend/`, `publicsite/`, `docker/` e `livekit/` (o un monorepo `extra/`).
- Aggiungere rate limiting (Encore middleware, oppure Caddy `rate_limit` plugin) su `/auth/login`, `/registration*`, `/profiles` POST.

---

### Ordine consigliato di applicazione

P5 → P1 → P2 → P3 → P4 → P6 → P7 → P11 → P10 → P8 → P9 → P12 → P13 → P14 → P15
