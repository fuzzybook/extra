import { Local } from './client';

// Mirrors the videochat service until it is included in the generated client.
export interface Room {
  active: boolean;
  sid: string;
  name: string;
  display_name: string;
  owner_id: string;
  mode: 'broadcast';
  num_participants: number;
  max_participants: number;
  creation_time: number;
  avatar_url?: string;
}

export interface CreateRoomParams {
  name: string;
  max_participants: number;
  empty_timeout: number;
}

export interface JoinToken {
  server_url: string;
  participant_token: string;
  room_name: string;
  identity: string;
  owner_id: string;
  role: 'owner' | 'viewer';
  expires_in: number;
}

export interface Participant {
  sid: string;
  identity: string;
  name: string;
  state: string;
  avatar_url?: string;
}

export interface ChatMessage {
  id: string;
  room_name: string;
  sender_user_id: string;
  sender_name: string;
  sender_avatar?: string;
  kind: string;
  body: string;
  client_message_id: string;
  created_at: string;
}

export interface SendMessageParams {
  client_message_id?: string;
  text: string;
  kind?: string;
}

export interface SendCommandParams {
  command: string;
  target_id?: string;
  payload?: string;
}

export function createVideochatClient(getToken: () => string | null) {
  async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const token = getToken();
    const response = await fetch(`${Local}/videochat${path}`, {
      method,
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(body ? { 'Content-Type': 'application/json' } : {}),
      },
      ...(body ? { body: JSON.stringify(body) } : {}),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => null) as { message?: string } | null;
      throw new Error(error?.message || `HTTP ${response.status}`);
    }
    if (method === 'DELETE' || response.status === 204) return undefined as T;
    return response.json() as Promise<T>;
  }
  const roomPath = (name: string) => `/rooms/${encodeURIComponent(name)}`;
  return {
    listRooms: () => request<{ rooms: Room[] }>('GET', '/rooms'),
    createRoom: (params: CreateRoomParams) => request<Room>('POST', '/rooms', params),
    openRoom: (roomName: string) => request<Room>('POST', `${roomPath(roomName)}/open`),
    createToken: (roomName: string) => request<JoinToken>('POST', '/token', { room_name: roomName }),
    deleteRoom: (roomName: string) => request<void>('DELETE', roomPath(roomName)),
    listParticipants: (roomName: string) => request<{ participants: Participant[] }>('GET', `${roomPath(roomName)}/participants`),
    removeParticipant: (roomName: string, identity: string) => request<void>('DELETE', `${roomPath(roomName)}/participants/${encodeURIComponent(identity)}`),
    listMessages: (roomName: string, limit = 50) => request<{ messages: ChatMessage[] }>('GET', `${roomPath(roomName)}/messages?limit=${limit}`),
    sendMessage: (roomName: string, params: SendMessageParams) => request<ChatMessage>('POST', `${roomPath(roomName)}/messages`, params),
    sendCommand: (roomName: string, params: SendCommandParams) => request<{ event_id: string; success: boolean }>('POST', `${roomPath(roomName)}/commands`, params),
  };
}
