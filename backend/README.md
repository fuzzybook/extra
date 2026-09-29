# encore-test

Test del framework Encore con backend Go e frontend Quasar.

## Prerequisiti

- Encore: `brew install encoredev/tap/encore`
- Go
- Node.js e pnpm per il frontend

## Avvio locale

Backend:

```bash
encore run
```

Frontend:

```bash
cd frontend
pnpm install
pnpm dev
```

## Test

Backend:

```bash
go test ./...
```

Frontend:

```bash
cd frontend
pnpm typecheck
```

## Note Encore

Con `encore run` attivo, il dashboard locale e' disponibile su:

```text
http://localhost:9400/
```

Le API locali sono esposte dall'ambiente Encore locale. Consulta il Service Catalog nel dashboard per endpoint e schema aggiornati.

## Videochat

Il servizio `videochat` integra LiveKit per gestire stanze, partecipanti e token di accesso.
Configurazione e API: [videochat/README.md](videochat/README.md).

### Setup personale

`GET /profiles/personal-setup` restituisce il setup dell'utente autenticato:
`user_id`, `can_broadcast` (bool) e `room_name` (string).
In assenza di setup, restituisce `can_broadcast: false` e `room_name: ""`.

Gli amministratori possono impostarlo con `PUT /profiles/personal-setup/:userID`:

```json
{"can_broadcast": true, "room_name": "stanza-utente"}
```

Il nome stanza è obbligatorio quando il broadcasting è abilitato e viene cancellato
quando è disabilitato. Il setup è salvato nel database `profiles` dalla migrazione 11.
Queste API gestiscono i dati del setup; l'applicazione del permesso ai token LiveKit
e la creazione della stanza non sono automatiche.
