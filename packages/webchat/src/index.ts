// Webchat condivisa tra il pannello (frontend/) e il sito pubblico (publicsite/).
// Nessuna dipendenza da store, router o client di un'app: utente, sorgente
// della chat e notifiche arrivano dall'app che la usa.
export { useLiveRoom, errorMessage, type LiveMedia, type LiveRoom, type LiveRoomOptions } from './useLiveRoom';
export { useRoomChat, type RoomChat, type RoomChatOptions } from './useRoomChat';
export { webchatMessages, type WebchatMessages } from './i18n';
export type { ChatItem, ChatSource, ChatUser, JoinToken, RoomDataMessage, ServerChatMessage } from './types';
export { default as ChatPanel } from './components/ChatPanel.vue';
export { default as LiveStage } from './components/LiveStage.vue';
export { default as LiveTrack } from './components/LiveTrack.vue';
export {
  buildWebcamOptions,
  calculateHeight,
  defaultWebcamSettings,
  loadWebcamSettings,
  republishCamera,
  unpublishMedia,
  saveWebcamSettings,
  type WebcamSettings,
} from './webcam';
export { default as WebcamSettingsDialog } from './components/WebcamSettingsDialog.vue';
