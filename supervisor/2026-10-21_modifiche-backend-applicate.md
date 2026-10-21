# Modifiche applicate al backend – progetto `extra`

**Data:** 21 ottobre 2026
**Riferimenti:** `2026-10-21_report-supervisione.md` (ID dei problemi) e `2026-10-21_proposte-modifiche.md` (ID delle proposte P1–P15).
**Perimetro:** solo `backend/`, più tre piccoli adattamenti nel frontend, indispensabili perché il pannello admin e la chat continuino a funzionare (vedi §3).

Nel codice ogni intervento è marcato con un commento `[supervisione 2026-10-21 · <ID>]`, così è facile ritrovarlo con una ricerca. I punti lasciati aperti sono marcati con `TODO(supervisione …)`.

**Verifiche eseguite:** `go build ./...` e `go vet ./...` sono passati senza errori con Go 1.26.8. I test unitari del nuovo pacchetto `pkg/validate` passano.
**Non eseguito:** `encore test ./...`, perché la CLI Encore non era disponibile nell'ambiente di lavoro. I test esistenti vanno lanciati in locale prima del merge. Le modifiche **non sono state committate**.

---

## 1. Interventi applicati

### Fase 0 – sicurezza critica

| ID | Cosa è cambiato | File |
|---|---|---|
| P1 · C1 | `CreatePromo` e `UpdatePromo` passano da `public` ad `auth` e sono riservati agli admin (`authsvc.RequireAdmin()`) | `advertising/promo.go` |
| P1 · H6 | `advertising.List` e `advertising.Create` passano ad `auth` e sono solo per admin (prima esponevano le email degli inserzionisti) | `advertising/advertising.go` |
| P1 · C3 | `SaveImage` passa ad `auth`. L'`owner` viene salvato all'INSERT. Un'immagine esistente può modificarla solo il proprietario o un admin (le immagini "storiche" senza owner solo un admin). Sono accettati solo PNG, JPEG e WebP fino a 5 MB, con il tipo rilevato dai byte e non dichiarato dal client. La descrizione viene sanitizzata | `images/images.go` |
| P1 · C3 | `ListImages` restituisce: agli admin tutte le immagini, agli utenti loggati quelle approvate più le proprie, agli anonimi solo quelle approvate. `GetImage` invia `X-Content-Type-Options: nosniff` | `images/images.go` |
| P1 · H2 | Chatlog: `AddChatLog` e `ListChatLog` passano ad `auth`. Il payload ha un limite di 4 KB e il tipo `CMD` è riservato agli admin | `broadcasters/chatlog.go` |
| P1 · H8 | Upload avatar: massimo 5 MB, solo PNG, JPEG e WebP, con tipo rilevato dai byte | `profiles/avatars.go` |
| P1 | Nuovi helper condivisi `auth.RequireAdmin()` e `auth.IsCallerAdmin()` | `auth/roles.go` |
| P2 · C2 | Nuovo pacchetto `pkg/validate`: `SafeURL`/`HTTPURL` accettano solo `http`/`https` o percorsi che iniziano con `/` (rifiutano `javascript:`, `data:`, `//host`…). Applicato a URL, immagini e immagini mobile delle promo, alle foto dei broadcaster e agli avatar | `pkg/validate/validate.go`, `advertising/promo.go`, `broadcasters/broadcasters.go`, `profiles/profiles.go` |
| P3 · C4 | `validate.SanitizeRichText`: tiene solo i tag di formattazione base (`b, strong, i, em, u, s, strike, br, p, div, span`) senza nessun attributo ed elimina script e style con il loro contenuto. Applicato alle didascalie delle foto (`alt`, `description.text`) e alla descrizione delle immagini | `pkg/validate/richtext.go`, `broadcasters/broadcasters.go`, `images/images.go` |
| P4 · C6 | I non-admin non possono più impostare `enabled`, `used` e `certified`: in creazione sono forzati a `false`/`true`/`false`, in modifica restano i valori salvati | `broadcasters/broadcasters.go` |
| P4 · H7 | Negli endpoint pubblici dei broadcaster, admin e proprietario vedono tutto. Gli altri vedono solo i profili abilitati, senza `phone` e con `user_id` azzerato | `broadcasters/broadcasters.go` |
| P5 · C5 | `.gitignore` ora ignora `.secrets*.cue` e `*.cue.bak` | `.gitignore` |
| P5 | Rotazione del pepper senza bloccare gli utenti: se è impostata la variabile d'ambiente `PASSWORD_PEPPER_OLD`, il login accetta anche il vecchio pepper e ri-hasha la password con quello nuovo | `auth/auth.go`, `auth/login.go` |

### Fase 1 – correzioni

| ID | Cosa è cambiato | File |
|---|---|---|
| P6 · H1 | `AuthHandler` rifiuta le sessioni di profili inesistenti o con stato Suspended (2), Deleted (3) o Banned (6). `auth.Login` rifiuta questi account con il messaggio generico "invalid credentials" | `auth/auth.go`, `auth/login.go` |
| P6 · H1 | Nuovo endpoint private `auth.RevokeUserSessions`. `admin.UpdateProfileStatus` lo chiama quando lo stato diventa bloccante | `auth/login.go`, `admin/admin.go` |
| P6 · M9 | Nuovo endpoint private `auth.DeleteCredentials`, chiamato da `profiles.Delete` e `admin.DeleteProfile`: cancella credenziali e sessioni | `auth/login.go`, `profiles/profiles.go`, `admin/admin.go` |
| P7 · M1 | `defer tx.Rollback()` incondizionato: nessuna transazione resta aperta sui return anticipati | `auth/credentials.go` |
| P11 · M6 | `ConfirmPasswordReset` consuma il token con un unico `UPDATE … WHERE used_at IS NULL AND expires_at > NOW() RETURNING`. Se il cambio password fallisce, il token torna disponibile | `registration/registration.go` |
| P11 · H9 | Reset o impostazione password da admin: tutte le sessioni vengono revocate. Cambio volontario: vengono revocate tutte tranne quella corrente | `auth/credentials.go` |
| P10 · M5 | Login, `CheckEmail` e `RequestPasswordReset` normalizzano l'email e confrontano con `lower(email)` | `profiles/login.go`, `registration/registration.go` |
| P10 · M12 | `admin.InsertProfile`, `admin.UpdateProfile` e `admin.UpsertPersonalData` applicano le stesse validazioni dell'utente | `admin/admin.go` |
| M8 | Registrazione e creazione da admin: se la creazione delle credenziali fallisce, il profilo appena creato viene rimosso | `profiles/profiles.go`, `admin/admin.go` |
| P8 · M2 | Alzata di mano: le richieste `pending` scadute passano a `expired` prima di crearne una nuova. In caso di corsa concorrente (errore `23505`) viene restituita la richiesta esistente invece di un errore 500 | `videochat/speak_requests.go` |
| P8 · M3 | Nuovo cron `videochat-release-expired-speakers` (ogni minuto, endpoint private `ReleaseExpiredSpeakers`): revoca su LiveKit il permesso di pubblicare agli speaker con lease scaduto e libera lo slot. `ApproveSpeakRequest` revoca lo speaker scaduto prima di promuovere quello nuovo | `videochat/speak_requests.go` |
| P9 · M4 | `ListMessages` restituisce gli **ultimi** N messaggi in ordine cronologico. Aggiunto il controllo `rows.Err()` | `videochat/messages.go` |
| P12 · H4 | `SendMessage` ignora `kind` inviato dal client: i messaggi degli utenti sono sempre `text` | `videochat/messages.go` |
| P12 · H5 | `SendCommand` richiede l'owner della stanza o un admin, accetta solo un elenco chiuso di comandi (`alert`, `mute`, `clear`, `filter`) e limita il payload a 2000 caratteri | `videochat/messages.go` |
| M14 | `ValidateHours` converte gli orari in minuti (`HH:MM`, `24:00` ammesso), quindi `9:00` < `10:00` è ora corretto e il formato viene verificato | `broadcasters/broadcasters.go` |
| M15 | Gli errori di `json.Unmarshal` sui campi JSONB dei broadcaster vengono registrati con `rlog.Warn` invece di essere ignorati | `broadcasters/broadcasters.go` |

---

## 2. Azioni manuali necessarie

1. **Togliere il file dei segreti dal repository** (C5). Il file è ancora tracciato da git:
   ```bash
   cd backend
   git rm --cached ".secrets.local copy.cue"
   ```
   Se il repository è stato pushato o condiviso, va ripulita anche la storia (`git filter-repo --path ".secrets.local copy.cue" --invert-paths`). È un'operazione distruttiva, da fare consapevolmente.
2. **Ruotare il pepper** (C5). Generare un nuovo valore (`openssl rand -hex 32`), impostarlo come secret `PasswordPepper` e mettere il valore attuale in `PASSWORD_PEPPER_OLD`. Dopo qualche settimana: reset forzato per chi non ha fatto login, poi rimuovere `PASSWORD_PEPPER_OLD`.
3. **Togliere i valori di default del pepper dai file Docker** (`docker/entrypoint-backend.sh`, `docker-compose.yml`, `.env.example`, `.env`). Non l'ho fatto perché è fuori dalla cartella backend.
4. **Lanciare i test**: `encore test ./...`, poi una prova manuale di login, chat, alzata di mano, promo e upload immagini.

---

## 3. Adattamenti al frontend (minimi, per compatibilità)

Alcuni endpoint sono passati da `public` ad `auth`, e queste parti del frontend chiamavano il backend senza token:

- `frontend/src/pages/PromosPage.vue` – il client Encore ora invia `Authorization: Bearer <token>`.
- `frontend/src/pages/AdvertisingPage.vue` – idem.
- `frontend/src/stores/chat-store.ts` – le due `fetch` verso il chatlog inviano l'header `Authorization`.

Il client TypeScript generato (`src/encore/client.ts`) non va rigenerato: le firme degli endpoint non sono cambiate.

---

## 4. Non applicato (e perché)

| ID | Motivo |
|---|---|
| H3 (`CanPublishData` dei viewer) | Il frontend attuale invia i messaggi di chat dei viewer via LiveKit: disattivarlo ora romperebbe la chat. Prima va spostata la chat su `POST /videochat/rooms/:room/messages` (P12, punti 2, 5 e 6). Il punto è segnato con un `TODO` in `videochat/livekit.go` |
| P13 (login senza Pub/Sub) | È un intervento di struttura (Fase 3): richiede di rimuovere topic, subscription, tabella `login_results` e test |
| L5 / P10 migrazione `lower(email)` unica | Prima di creare l'indice va verificato sul DB reale che non esistano email duplicate che differiscono solo per maiuscole, altrimenti la migrazione bloccherebbe l'avvio. Nel frattempo il codice confronta già con `lower(email)` |
| M7 (`POST /profiles` pubblico) | Il frontend usa ancora `profiles.Insert` in `profiles-store.ts`. Va prima rimosso quell'uso, poi l'endpoint può diventare `private` |
| M9 (pulizia broadcaster e immagini alla cancellazione) | Richiede API dedicate nei servizi `broadcasters` e `images` (§3.1 del report). Segnato con `TODO` |
| M10, M11, M13 | Cambiano la logica di apertura delle stanze e delle sessioni (nomi delle stanze usati dal frontend) e le transazioni tra due DB: vanno progettati insieme ai webhook LiveKit (§3.4) |
| M16 (rate limiting) | Da fare con un middleware Encore o con Caddy. Segnato con `TODO` su `CheckEmail` |
| Fase 2 (deploy, D1–D9) e Fase 3 (struttura) | Fuori dal perimetro di questo intervento |
| C2 e C4 lato frontend (`safeUrl` nel publicsite, DOMPurify) | Il backend ora rifiuta gli URL pericolosi e sanitizza le didascalie. La difesa lato frontend resta consigliata |
| Server MCP di Encore | Non era raggiungibile da questa sessione (gira in locale sul Mac). Le modifiche sono state fatte direttamente sui file |

---

## 5. Note di comportamento da conoscere

- **Immagini già caricate senza owner**: solo un admin può modificarle. Se un broadcaster deve cambiare la didascalia di una foto caricata prima di oggi, l'aggiornamento in `images` viene ignorato. Quello del profilo broadcaster invece funziona.
- **`ListImages` per i non-admin**: nessuna immagine è ancora `approved = TRUE`, quindi i broadcaster vedranno solo le proprie. Serve un flusso di approvazione da parte dell'admin.
- **Stati del profilo**: Inactive (1), Pending (4) e Waiting deletion (5) possono ancora fare login. Se vanno bloccati, basta aggiungerli a `isBlockedStatus` in `auth/auth.go`.
- **Orari dei broadcaster**: valori salvati in formati diversi da `HH:MM` verranno rifiutati al prossimo salvataggio, con un messaggio chiaro.
