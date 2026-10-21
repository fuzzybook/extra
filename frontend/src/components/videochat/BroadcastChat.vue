<template>
  <div class="broadcast-chat-card">
    <!-- Chat Header -->
    <div class="chat-header row items-center justify-between q-px-md q-py-sm">
      <div class="row items-center q-gutter-xs">
        <q-icon name="chat" color="primary" size="20px" />
        <span class="text-subtitle2 text-weight-bold">{{ t('broadcast.chatTitle') }}</span>
        <q-badge v-if="filteredMessages.length" color="primary" rounded dense class="q-ml-xs">
          {{ filteredMessages.length }}
        </q-badge>
      </div>
      <div class="row items-center q-gutter-xs">
        <q-btn
          v-if="isAdminOrAdvisor"
          flat
          round
          dense
          size="sm"
          icon="admin_panel_settings"
          color="negative"
          :aria-label="t('admin.sendAlert')"
          @click="openAdminAlertDialog"
        >
          <q-tooltip>{{ t('admin.sendAlert') }}</q-tooltip>
        </q-btn>
        <q-btn
          v-if="filteredMessages.length"
          flat
          round
          dense
          size="sm"
          icon="delete_sweep"
          color="grey-7"
          :aria-label="t('broadcast.chatClear')"
          @click="clearChat"
        >
          <q-tooltip>{{ t('broadcast.chatClear') }}</q-tooltip>
        </q-btn>
      </div>
    </div>

    <!-- Messages Container -->
    <div ref="messagesContainer" class="chat-messages q-pa-sm" @scroll="onScroll">
      <div v-if="!filteredMessages.length" class="chat-empty column flex-center text-grey-6 q-pa-md">
        <q-icon name="chat_bubble_outline" size="48px" color="grey-4" class="q-mb-sm" />
        <p class="text-caption text-center q-mb-none">{{ t('broadcast.chatEmpty') }}</p>
      </div>

      <div v-else class="q-gutter-y-sm">
        <div
          v-for="msg in filteredMessages"
          :key="msg.id"
          class="chat-message-item"
          :class="{
            'is-self': isSelf(msg),
            'is-owner': msg.isOwner,
          }"
        >
          <div class="message-content">
            <!-- Messaggio dell'utente loggato (a destra, verde) -->
            <div
              v-if="isSelf(msg)"
              class="row items-center justify-end q-gutter-x-xs no-wrap q-mb-xs"
            >
              <span class="text-caption text-grey-5 message-time q-mr-xs">
                {{ formatTime(msg.timestamp) }}
              </span>
              <span class="text-caption text-weight-bold text-ellipsis sender-name text-positive">
                {{ msg.senderName }}
              </span>
              <q-avatar size="20px" color="positive" text-color="white" class="text-weight-bold">
                <img v-if="msg.senderAvatar" :src="msg.senderAvatar" alt="" />
                <span v-else>{{ (msg.senderName || 'U').charAt(0).toUpperCase() }}</span>
              </q-avatar>
            </div>

            <!-- Messaggi di tutti gli altri (a sinistra; arancione se owner, altrimenti standard) -->
            <div
              v-else
              class="row items-center q-gutter-x-xs no-wrap q-mb-xs"
            >
              <q-avatar
                size="20px"
                :color="msg.isOwner ? 'amber-8' : 'primary'"
                text-color="white"
                class="text-weight-bold"
              >
                <img v-if="msg.senderAvatar" :src="msg.senderAvatar" alt="" />
                <span v-else>{{ (msg.senderName || 'U').charAt(0).toUpperCase() }}</span>
              </q-avatar>
              <span
                class="text-caption text-weight-bold text-ellipsis sender-name"
                :class="msg.isOwner ? 'text-amber-9' : 'text-grey-9'"
              >
                {{ msg.senderName }}
              </span>
              <span class="text-caption text-grey-5 q-ml-auto message-time">
                {{ formatTime(msg.timestamp) }}
              </span>
            </div>

            <div class="message-bubble">
              {{ msg.text }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Scroll to Bottom Button -->
    <transition name="fade">
      <q-btn
        v-if="showScrollBottom"
        round
        dense
        color="primary"
        icon="keyboard_arrow_down"
        class="scroll-bottom-btn shadow-3"
        size="sm"
        @click="scrollToBottom(true)"
      />
    </transition>

    <!-- Chat Input Fixed at Bottom -->
    <div class="chat-input-area q-pa-sm">
      <q-form @submit.prevent="sendMessage">
        <q-input
          v-model="inputText"
          outlined
          dense
          bg-color="white"
          :placeholder="t('broadcast.chatPlaceholder')"
          maxlength="500"
          :disable="!connected"
          autocomplete="off"
          @keydown.enter.prevent="sendMessage"
        >
          <template #append>
            <q-btn
              round
              dense
              flat
              color="primary"
              icon="send"
              :disable="!connected || !inputText.trim()"
              :aria-label="t('broadcast.chatSend')"
              @click="sendMessage"
            />
          </template>
        </q-input>
      </q-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, onActivated, ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Room as LiveRoom, RemoteParticipant } from 'livekit-client';
import { RoomEvent } from 'livekit-client';
import { useProfilesStore } from '@/stores/profiles-store';
import { useQuasar, Dialog } from 'quasar';
import { useChatStore, type ChatMessage } from '@/stores/chat-store';

const $q = useQuasar();

const props = defineProps<{
  room?: LiveRoom | undefined;
  connected: boolean;
  isOwner: boolean;
  roomName?: string | undefined;
}>();

const { t } = useI18n();
const profiles = useProfilesStore();
const chatStore = useChatStore();

const messagesContainer = ref<HTMLDivElement>();
const showScrollBottom = ref(false);
const currentUserId = ref('');

function isSelf(msg: ChatMessage): boolean {
  const current = currentUserId.value || props.room?.localParticipant?.identity || profiles.profile?.user_id;
  return Boolean(current && msg.senderId === current);
}

// Quando l'utente digita un nuovo testo, il contenuto passa direttamente allo store Pinia
const inputText = computed({
  get: () => chatStore.inputText,
  set: (val: string) => chatStore.setInputText(val),
});

// Messaggi filtrati (lo store include già il supporto futuro a filtri)
const filteredMessages = computed(() => chatStore.filteredMessages);

const roomIdentifier = computed(() => {
  return props.roomName || (props.room as any)?.name || (props.room as any)?.sid || '';
});

function formatTime(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function scrollToBottom(smooth = false) {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTo({
        top: messagesContainer.value.scrollHeight,
        behavior: smooth ? 'smooth' : 'auto',
      });
      showScrollBottom.value = false;
    }
  });
}

function onScroll() {
  if (!messagesContainer.value) return;
  const { scrollTop, scrollHeight, clientHeight } = messagesContainer.value;
  const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);
  showScrollBottom.value = distanceFromBottom > 80;
}

function clearChat() {
  chatStore.clearMessages();
}

const openAdminAlertDialog = async () => {
  if (!props.room) return;
  const ownerId = props.room.localParticipant.identity || '';
  const msgDialog = Dialog.create({
    title: t('admin.sendAlert'),
    prompt: { model: '', label: t('admin.message') },
    cancel: true,
    persistent: true,
  });
  msgDialog.onOk(async (msgText) => {
    await sendAdminAlert(ownerId, msgText);
    ;(msgDialog as any).hide();
  });
};

async function sendAdminAlert(targetId: string, alertText: string) {
  if (!props.room || !props.connected) return;
  const localParticipant = props.room.localParticipant;
  const senderId = localParticipant.identity || profiles.profile?.user_id || 'me';
  const msg: ChatMessage = {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    senderId,
    senderName: 'Admin',
    text: alertText,
    timestamp: Date.now(),
    isOwner: props.isOwner,
    type: 'ADMIN_ALERT',
    targetId,
  };
  const encoder = new TextEncoder();
  const data = encoder.encode(JSON.stringify(msg));
  await localParticipant.publishData(data, { reliable: true, topic: 'chat' });
  // Salva anche l'alert nello store
  chatStore.saveMessage(roomIdentifier.value, msg);
}

// Computed flag per admin o chat advisor
const isAdminOrAdvisor = computed(() => {
  const role = profiles.profile?.role?.toLowerCase();
  return role === 'admin' || role === 'chatadvisor';
});

async function sendMessage() {
  const text = inputText.value.trim();
  if (!text || !props.room || !props.connected) return;

  const localParticipant = props.room.localParticipant;
  const senderId = localParticipant.identity || profiles.profile?.user_id || 'me';
  currentUserId.value = senderId;

  const senderName =
    profiles.profile?.display_name ||
    localParticipant.name ||
    profiles.profile?.email?.split('@')[0] ||
    localParticipant.identity ||
    'User';

  const senderAvatar = profiles.profile?.avatar_url || undefined;

  const msg: ChatMessage = {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    senderId,
    senderName,
    ...(senderAvatar ? { senderAvatar } : {}),
    text,
    timestamp: Date.now(),
    isOwner: props.isOwner,
    type: 'MESSAGE',
  };

  try {
    const encoder = new TextEncoder();
    const data = encoder.encode(JSON.stringify(msg));
    await localParticipant.publishData(data, { reliable: true, topic: 'chat' });

    // Lo store salva il messaggio su DB (tabella chatlog) e lo aggiunge in basso alla lista
    chatStore.saveMessage(roomIdentifier.value, msg);
    chatStore.setInputText('');
    scrollToBottom(true);
  } catch (err) {
    console.error('Failed to send chat message:', err);
  }
}

function handleDataReceived(
  payload: Uint8Array,
  _participant?: RemoteParticipant,
  _kind?: unknown,
  topic?: string
) {
  if (topic && topic !== 'chat') return;
  try {
    const decoder = new TextDecoder();
    const str = decoder.decode(payload);
    const msg = JSON.parse(str) as ChatMessage;

    // Gestione alert amministratore
    if (msg.type === 'ADMIN_ALERT') {
      if (['VIDEO_ON', 'VIDEO_OFF', 'MIC_ON', 'MIC_OFF'].includes(msg.text)) {
        return;
      }
      if (msg.targetId && !props.isOwner) {
        return;
      }
      $q.notify({
        type: 'negative',
        message: t('admin.notifications.adminAlert'),
        caption: msg.text,
        timeout: 0,
        actions: [{ label: t('close'), handler: () => {} }],
      });
      if (props.isOwner) {
        const adminMsg: ChatMessage = {
          id: msg.id,
          senderId: msg.senderId,
          senderName: msg.senderName,
          text: msg.text,
          timestamp: msg.timestamp,
          isSystem: true,
        };
        chatStore.addMessage(adminMsg);
      }
      return;
    }

    // Aggiunge alla lista (in basso, append)
    chatStore.addMessage(msg);
    if (!showScrollBottom.value) {
      scrollToBottom(true);
    }
  } catch (err) {
    console.warn('Failed to parse incoming chat message:', err);
  }
}

function setupRoomListeners(room?: LiveRoom) {
  if (!room) return;
  room.on(RoomEvent.DataReceived, handleDataReceived);
  if (room.localParticipant?.identity) {
    currentUserId.value = room.localParticipant.identity;
  }
}

function teardownRoomListeners(room?: LiveRoom) {
  if (!room) return;
  room.off(RoomEvent.DataReceived, handleDataReceived);
}

// Inizializzazione della chat (preleva gli ultimi 30 messaggi dal DB e li ordina)
async function initializeChat() {
  const rid = roomIdentifier.value;
  if (rid) {
    await chatStore.initChat(rid, 30);
    scrollToBottom();
  }
}

watch(
  () => props.room,
  (newRoom, oldRoom) => {
    teardownRoomListeners(oldRoom);
    setupRoomListeners(newRoom);
  }
);

watch(
  roomIdentifier,
  (newId) => {
    if (newId) {
      void initializeChat();
    }
  },
  { immediate: true }
);

watch(
  () => props.connected,
  (connected) => {
    if (connected) {
      if (props.room?.localParticipant?.identity) {
        currentUserId.value = props.room.localParticipant.identity;
      }
      void initializeChat();
    }
  },
  { immediate: true }
);

onMounted(() => {
  if (props.room) {
    setupRoomListeners(props.room);
  }
  if (roomIdentifier.value) {
    void initializeChat();
  }
});

onBeforeUnmount(() => {
  teardownRoomListeners(props.room);
});

onActivated(() => {
  // Quando il pannello riprende il focus (grazie a keep-alive di q-tab-panels), scrolla in basso
  if (chatStore.messages.length) {
    scrollToBottom();
  }
});
</script>

<style scoped>
.broadcast-chat-card {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background: #ffffff;
  border: none;
  border-radius: 0;
  overflow: hidden;
  box-shadow: none;
  position: relative;
}

.chat-header {
  flex-shrink: 0;
  background: #f8f9fa;
  border-bottom: 1px solid #ebebeb;
}

.chat-messages {
  flex: 1 1 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: #ffffff;
}

.chat-empty {
  height: 100%;
  min-height: 200px;
}

.chat-message-item {
  width: 100%;
}

.message-content {
  display: flex;
  flex-direction: column;
}

.sender-name {
  max-width: 140px;
}

.message-time {
  font-size: 10px;
}

.message-bubble {
  background: #f1f3f5;
  color: #212529;
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 13px;
  line-height: 1.4;
  word-break: break-word;
  white-space: pre-wrap;
  width: fit-content;
  max-width: 100%;
}

/* Tutti gli altri a SINISTRA */
.chat-message-item:not(.is-self) {
  align-items: flex-start;
}

.chat-message-item:not(.is-self) .message-content {
  align-items: flex-start;
}

.chat-message-item:not(.is-self) .message-bubble {
  background: #f1f3f5;
  color: #212529;
}

/* Il proprietario della room in ARANCIONE */
.chat-message-item:not(.is-self).is-owner .message-bubble {
  background: #fff8e1 !important;
  color: #e65100 !important;
  border-left: 3px solid #ff9800 !important;
  border-right: none !important;
}

/* L'utente loggato a DESTRA in VERDE */
.chat-message-item.is-self {
  align-items: flex-end;
}

.chat-message-item.is-self .message-content {
  align-items: flex-end;
}

.chat-message-item.is-self .message-bubble {
  background: #e8f5e9 !important;
  color: #1b5e20 !important;
  border-right: 3px solid #2e7d32 !important;
  border-left: none !important;
}

.chat-input-area {
  flex-shrink: 0;
  background: #f8f9fa;
  border-top: 1px solid #ebebeb;
}

.scroll-bottom-btn {
  position: absolute;
  bottom: 60px;
  right: 16px;
  z-index: 10;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
