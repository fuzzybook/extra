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
