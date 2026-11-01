/** Token LiveKit restituito dal backend (videochat.CreateToken o CreateGuestToken). */
export interface JoinToken {
  server_url: string;
  participant_token: string;
  room_name: string;
  identity: string;
  owner_id: string;
  /** 'owner' | 'viewer' | 'guest' */
  role: string;
}

/** Messaggio di chat salvato dal backend (videochat.ChatMessage). */
export interface ServerChatMessage {
  id: string;
  sender_user_id: string;
  sender_name: string;
  sender_avatar?: string | undefined;
  body: string;
  client_message_id?: string | undefined;
  created_at: string;
}

/**
 * Payload inviato dal server sul data channel LiveKit (topic "chat"), vedi
 * videochat.RoomDataMessage nel backend.
 */
export interface RoomDataMessage extends ServerChatMessage {
  type: 'chat.message' | 'admin.alert';
  is_owner: boolean;
}

/** Messaggio mostrato nella chat. */
export interface ChatItem {
  id: string;
  senderId: string;
  senderName: string;
  senderAvatar?: string | undefined;
  text: string;
  timestamp: number;
  isOwner: boolean;
  clientMessageId?: string | undefined;
}

/** Utente che scrive in chat; null per i visitatori senza login. */
export interface ChatUser {
  id: string;
  name: string;
  avatar?: string | undefined;
}

/**
 * Da dove la chat legge lo storico e dove invia i messaggi. Ogni app la
 * costruisce con il proprio client Encore; senza send la chat è in sola
 * lettura.
 */
export interface ChatSource {
  history(key: string): Promise<ServerChatMessage[]>;
  send?: ((key: string, params: { client_message_id: string; text: string }) => Promise<ServerChatMessage>) | undefined;
}
