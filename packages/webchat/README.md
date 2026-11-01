# @extra/webchat

Diretta video (LiveKit) e chat condivise dal pannello (`frontend/`) e dal sito pubblico (`publicsite/`).

Il pacchetto è solo sorgente: non ha `node_modules` né build. Ogni app lo importa come `@extra/webchat` e fornisce `vue`, `quasar`, `vue-i18n` e `livekit-client` (dipendenze peer). Il collegamento è in `quasar.config.ts` di ogni app:

- `build.alias` mappa `@extra/webchat` su `packages/webchat/src`;
- `resolve.dedupe` fa risolvere le dipendenze peer dalla cartella dell'app, così c'è una sola copia di Vue;
- `build.typescript.extendTsConfig` fa lo stesso per TypeScript;
- `server.fs.allow` permette al server di sviluppo di leggere la cartella.

## Contenuto

- `useLiveRoom()`: collegamento alla stanza, eventi, riconnessione, tracce del proprietario, sblocco dell'audio.
- `useRoomChat()`: storico, invio con messaggio ottimistico, messaggi inoltrati dal server sul data channel (topic `chat`). Senza `source.send` la chat è in sola lettura.
- `LiveStage`, `LiveTrack`: video e audio del proprietario.
- `ChatPanel`: chat; i colori si cambiano con le variabili CSS `--webchat-*` (default chiaro).
- `webchatMessages`: testi in en, it, fr, de, es da unire ai messaggi dell'app sotto la chiave `webchat`.

Il pacchetto non usa Pinia, il router o i plugin di Quasar: l'utente, la sorgente della chat e le notifiche arrivano dall'app.
