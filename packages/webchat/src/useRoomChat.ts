import { getCurrentInstance, onBeforeUnmount, ref, watch, type Ref } from 'vue';
import { RoomEvent, type RemoteParticipant, type Room } from 'livekit-client';
import { errorMessage } from './useLiveRoom';
import type { ChatItem, ChatSource, ChatUser, RoomDataMessage, ServerChatMessage } from './types';

export interface RoomChatOptions {
  /** Stanza LiveKit da cui ricevere i messaggi inoltrati dal server. */
  room: Ref<Room | undefined>;
  /** Chiave della chat (SID della stanza); al cambio viene ricaricato lo storico. */
  chatKey: Ref<string>;
  source: ChatSource;
  ownerId: Ref<string>;
  /** Chi scrive; null per i visitatori, che leggono soltanto. */
  user: Ref<ChatUser | null>;
  /** Testo di riserva per i mittenti senza nome. */
  fallbackName?: (() => string) | undefined;
  /** Alert amministrativo ricevuto dal proprietario della stanza. */
  onAlert?: ((body: string) => void) | undefined;
  onError?: ((message: string) => void) | undefined;
}

/**
 * Chat di una stanza: storico, invio con messaggio ottimistico e ricezione dei
 * messaggi che il server inoltra sul data channel LiveKit (topic "chat").
 * I client non pubblicano mai messaggi direttamente su LiveKit.
 */
export function useRoomChat(options: RoomChatOptions) {
  const messages = ref<ChatItem[]>([]);
  const loading = ref(false);
  const sending = ref(false);
  let historyRequest = 0;

  function toItem(m: ServerChatMessage | RoomDataMessage): ChatItem {
    return {
      id: m.id,
      senderId: m.sender_user_id,
      senderName: m.sender_name || options.fallbackName?.() || '',
      senderAvatar: m.sender_avatar || undefined,
      text: m.body,
      timestamp: new Date(m.created_at).getTime(),
      isOwner: 'is_owner' in m ? m.is_owner : Boolean(options.ownerId.value) && m.sender_user_id === options.ownerId.value,
      clientMessageId: m.client_message_id || undefined,
    };
  }

  async function loadHistory() {
    const key = options.chatKey.value;
    if (!key) return;
    const request = ++historyRequest;
    loading.value = true;
    try {
      const loaded = (await options.source.history(key)).map(toItem);
      if (request !== historyRequest) return;
      // Deduplica: i messaggi arrivati nel frattempo restano.
      const map = new Map<string, ChatItem>();
      loaded.forEach((item) => map.set(item.id, item));
      messages.value.forEach((item) => map.set(item.id, item));
      messages.value = Array.from(map.values()).sort((a, b) => a.timestamp - b.timestamp);
    } catch (reason) {
      if (request === historyRequest) console.warn('Impossibile caricare lo storico chat:', reason);
    } finally {
      if (request === historyRequest) loading.value = false;
    }
  }

  /**
   * Aggiunge o aggiorna un messaggio confermato dal server. Il messaggio
   * ottimistico locale (id = client_message_id) viene sostituito da quello
   * salvato, così l'inoltro del server non crea doppioni.
   */
  function upsert(m: ServerChatMessage | RoomDataMessage) {
    const confirmed = toItem(m);
    const idx = messages.value.findIndex(
      (x) => x.id === confirmed.id || (confirmed.clientMessageId && x.id === confirmed.clientMessageId),
    );
    if (idx >= 0) messages.value.splice(idx, 1, confirmed);
    else messages.value.push(confirmed);
  }

  /** Invia un messaggio; false se non è partito (il testo va restituito all'utente). */
  async function send(text: string): Promise<boolean> {
    const body = text.trim();
    const user = options.user.value;
    const key = options.chatKey.value;
    const sendToServer = options.source.send;
    if (!body || !key || !user || !sendToServer || sending.value) return false;
    sending.value = true;
    const clientMessageId = `msg-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    // Messaggio ottimistico, sostituito dalla conferma del server
    messages.value.push({
      id: clientMessageId,
      senderId: user.id,
      senderName: user.name,
      senderAvatar: user.avatar,
      text: body,
      timestamp: Date.now(),
      isOwner: Boolean(options.ownerId.value) && user.id === options.ownerId.value,
      clientMessageId,
    });
    try {
      upsert(await sendToServer(key, { client_message_id: clientMessageId, text: body }));
      return true;
    } catch (reason) {
      messages.value = messages.value.filter((m) => m.id !== clientMessageId);
      options.onError?.(errorMessage(reason));
      return false;
    } finally {
      sending.value = false;
    }
  }

  function clear() {
    messages.value = [];
  }

  /**
   * Sono accettati solo i pacchetti inviati dal server (participant assente):
   * un partecipante non può spacciarsi per un altro utente o per l'amministrazione.
   */
  function onData(payload: Uint8Array, participant?: RemoteParticipant, _kind?: unknown, topic?: string) {
    if (participant || topic !== 'chat') return;
    let msg: RoomDataMessage;
    try {
      msg = JSON.parse(new TextDecoder().decode(payload)) as RoomDataMessage;
    } catch {
      return;
    }
    if (msg.type === 'admin.alert') options.onAlert?.(msg.body);
    else if (msg.type === 'chat.message') upsert(msg);
  }

  watch(
    options.room,
    (current, previous) => {
      previous?.off(RoomEvent.DataReceived, onData);
      current?.on(RoomEvent.DataReceived, onData);
    },
    { immediate: true },
  );

  watch(
    options.chatKey,
    (key, previous) => {
      if (previous && key !== previous) messages.value = [];
      void loadHistory();
    },
    { immediate: true },
  );

  if (getCurrentInstance()) {
    onBeforeUnmount(() => {
      ++historyRequest;
      options.room.value?.off(RoomEvent.DataReceived, onData);
    });
  }

  return { messages, loading, sending, send, clear, reload: loadHistory };
}

export type RoomChat = ReturnType<typeof useRoomChat>;
