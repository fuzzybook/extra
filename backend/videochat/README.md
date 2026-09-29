# Broadcast LiveKit (uno a molti)

Configurare i secret Encore `LiveKitURL`, `LiveKitAPIKey` e `LiveKitAPISecret` sul server.
Il frontend usa solo il token partecipante, mai la API secret.

## Accesso

- Solo gli amministratori possono creare stanze (`POST /videochat/rooms`).
- Il creatore è l'unico proprietario; anche un altro amministratore entra come spettatore.
- Tutti gli utenti autenticati possono elencare i broadcast e richiedere un token.
- Il proprietario pubblica audio, videocamera e schermo. Gli spettatori hanno
  `canPublish=false`, `canPublishData=false`, `canSubscribe=true` nel JWT firmato.
- Nessun token frontend ha permessi amministrativi o modifica metadati.
- Elenco partecipanti, disconnessione partecipanti e chiusura richiedono il proprietario.

## API

| Metodo | Percorso | Accesso |
| --- | --- | --- |
| GET | `/videochat/status` | Pubblico |
| POST | `/videochat/rooms` | Admin (diventa proprietario) |
| GET | `/videochat/rooms` | Autenticato |
| POST | `/videochat/token` | Autenticato |
| DELETE | `/videochat/rooms/:roomName` | Proprietario |
| GET | `/videochat/rooms/:roomName/participants` | Proprietario |
| DELETE | `/videochat/rooms/:roomName/participants/:identity` | Proprietario |

La lista include i broadcast presenti su LiveKit (`active: true`) e le stanze
di `personal_setup` con `can_broadcast: true` non ancora presenti (`active: false`).
Il confronto usa il nome tecnico oppure proprietario e nome visualizzato.
Le stanze inattive riportano il nome configurato, il proprietario e `mode: "broadcast"`;
SID, partecipanti e data di creazione restano vuoti o zero. Non vengono create
su LiveKit e non possono ricevere un token finché non sono avviate.

Le richieste usano `Authorization: Bearer <session-token>` o il cookie `session`.
Creazione: `{"name":"Titolo della diretta","max_participants":0,"empty_timeout":300}`.
Zero nei campi numerici usa i valori predefiniti di LiveKit; il massimo comprende
anche il proprietario. Il timeout è in secondi.

La risposta stanza aggiunge `display_name`, `owner_id`, `mode: "broadcast"`.
`name` è un identificativo univoco generato dal server: usarlo nei percorsi API e
in `POST /videochat/token` con `{"room_name":"<name restituito>"}`.
La risposta token include `role: "owner" | "viewer"`, `owner_id`, `server_url`,
`participant_token`, `identity`, `room_name`, `expires_in`.

## Ciclo di vita

Le stanze sono effimere. Titolo e proprietario sono conservati nei metadati
LiveKit per la durata della stanza, anche se Encore viene riavviato. Quando
LiveKit elimina una stanza vuota, occorre crearne una nuova; non è un archivio
persistente delle trasmissioni. I nomi interni non vengono riutilizzati.
Le vecchie stanze senza metadati broadcast non compaiono nella lista e non
ricevono nuovi token tramite questa API; non vengono eliminate automaticamente.
I vecchi token meeting già emessi mantengono i permessi fino alla loro scadenza.

La disconnessione di uno spettatore non è un ban: può rientrare. La chiusura
elimina la stanza e disconnette tutti. Su LiveKit self-hosted i JWT già emessi
non sono revocabili da queste API: fino alla scadenza (10 minuti) possono ancora
consentire accesso o ricreazione della vecchia stanza interna. Tale stanza non
avrà metadati broadcast, non apparirà nella lista e non riceverà nuovi token.
Per un blocco irrevocabile servono controllo dell'ammissione/revoca aggiuntivi.
