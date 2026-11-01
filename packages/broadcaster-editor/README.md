# @extra/broadcaster-editor

Dialoghi di modifica del profilo broadcaster condivisi dal pannello (`frontend`) e dal sito pubblico (`publicsite`):

- `EditBroadcasterDialog`: dati del profilo (anche creazione, nel pannello)
- `EditAttributesDialog`: attributi
- `EditHoursDialog`: orari
- `EditPhotosDialog`: foto, con ritaglio (`CropperImage`) e didascalie
- `AvatarUpload`: avatar dell'utente (file o fotocamera, con ritaglio)

## Collegamento a un'app

1. `quasar.config.ts`: `broadcasterEditorSetup(process.cwd())` da `../packages/broadcaster-editor/quasar` (alias, Vite, TypeScript), come per `packages/webchat`.
2. Boot file: `app.use(createBroadcasterEditor(api))` con il client Encore dell'app (servizi `broadcasters` e `images`).
3. Traduzioni: `brEditor: broadcasterEditorMessages.<lingua>` (da `@extra/broadcaster-editor/i18n`) nei messaggi di ogni lingua.

I dialoghi emettono `saved` con il broadcaster aggiornato: l'app aggiorna le proprie liste.

## Permessi

Il backend permette al proprietario di modificare il proprio profilo; alcuni campi (es. `enabled`, `certified`) li cambia solo un amministratore.
