# Quasar App (frontend)

## Broadcast video

La pagina `/#/wideochat` elenca le dirette per gli utenti autenticati. Solo gli
amministratori possono creare stanze. Il creatore può aprire **Gestisci**, entrare
e attivare videocamera, microfono o condivisione schermo. Gli spettatori usano
**Guarda** e **Entra**, senza richieste di accesso a videocamera o microfono.

Il pannello del proprietario permette di elencare e disconnettere i partecipanti
e chiudere la stanza. **Esci** lascia la diretta senza chiuderla per tutti.
I permessi sono imposti dal server, non soltanto dai pulsanti dell'interfaccia.

È richiesto il servizio Encore `videochat` aggiornato, con i secret LiveKit
configurati. Il client usa l'indirizzo `Local` in `src/encore/client.ts`.
Per videocamera, microfono e condivisione schermo usare HTTPS oppure localhost.
La CSP consente LiveKit Cloud e LiveKit locale sulla porta 7880. Per un server
self-hosted diverso, configurare l'origine prima di avviare o compilare:
`QCLI_LIVEKIT_URL=wss://video.example.com pnpm build` (oppure `pnpm dev`).
Questa variabile autorizza l'origine nella CSP; l'URL effettivo di connessione
arriva sempre dal server Encore con il token.

Le stanze sono effimere e scompaiono secondo i timeout LiveKit. Le stanze create
prima del supporto broadcast vanno ricreate per assegnare un proprietario.
Disconnettere uno spettatore non lo banna. Su LiveKit self-hosted i token già
emessi possono restare validi fino alla scadenza anche dopo una disconnessione
o chiusura: non è implementata una revoca immediata dei JWT.

## Install the dependencies

```bash
pnpm install
# or: yarn/npm/bun install
```

### Start the app in development mode (HMR, error reporting, etc.)

```bash
quasar dev
```

### Build the app for production

```bash
quasar build
```

### Customize the configuration
See [Configuring quasar.config.js](https://v2.quasar.dev/quasar-cli-vite/quasar-config-js).

All'apertura della pagina di una stanza, il frontend chiama
`POST /videochat/rooms/:roomName/open`: il server riutilizza la stanza attiva oppure
crea quella abilitata nei personal settings, con il proprietario configurato.
Lo stesso controllo viene ripetuto prima di richiedere il token di ingresso.
Non vengono create stanze per nomi sconosciuti o configurazioni disabilitate;
i nomi assegnati a più proprietari richiedono una configurazione univoca.
