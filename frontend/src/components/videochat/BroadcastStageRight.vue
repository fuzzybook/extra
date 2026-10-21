<template>
  <div class="broadcast-stage-right q-pa-sm">
    <q-card flat bordered class="stage-right-card">
      <!-- Tabs: Chat e Partecipanti -->
      <q-tabs
        v-model="currentTab"
        dense
        class="text-grey-7 bg-grey-1"
        active-color="primary"
        indicator-color="primary"
        align="justify"
      >
        <q-tab name="chat" icon="chat">
          <q-badge v-if="filteredMessages.length" color="primary" floating rounded>
            {{ filteredMessages.length }}
          </q-badge>
        </q-tab>
        <q-tab name="participants" icon="people">
          <q-badge v-if="participants.length" color="primary" floating rounded>
            {{ participants.length }}
          </q-badge>
        </q-tab>
      </q-tabs>

      <q-separator />

      <q-tab-panels v-model="currentTab" animated keep-alive class="stage-right-panels">
        <!-- ================================================================= -->
        <!-- TAB PANEL: CHAT                                                   -->
        <!-- ================================================================= -->
        <q-tab-panel name="chat" class="q-pa-none stage-tab-panel">
          <div class="broadcast-chat-wrapper column full-height">
            <!-- Chat Header con azioni (Alert Admin, Pulizia Chat) -->
            <div class="chat-header row items-center justify-between q-px-md q-py-xs bg-grey-1 border-bottom">
              <div class="row items-center q-gutter-xs">
                <span class="text-caption text-weight-bold text-grey-8">{{ t('broadcast.chatTitle') }}</span>
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
                  v-if="messages.length"
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
            <div ref="messagesContainer" class="chat-messages col q-pa-sm" @scroll="onScroll">
              <div v-if="loadingHistory" class="row justify-center q-pa-md">
                <q-spinner color="primary" size="24px" />
              </div>

              <div v-else-if="!filteredMessages.length" class="chat-empty column flex-center text-grey-6 q-pa-md full-height">
                <q-icon name="chat_bubble_outline" size="44px" color="grey-4" class="q-mb-sm" />
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
                    <!-- Messaggio inviato da me (a destra, verde) -->
                    <div v-if="isSelf(msg)" class="row items-center justify-end q-gutter-x-xs no-wrap q-mb-xs">
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

                    <!-- Messaggio inviato da altri (a sinistra) -->
                    <div v-else class="row items-center q-gutter-x-xs no-wrap q-mb-xs">
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

                    <div class="message-bubble" :class="{ 'bubble-self': isSelf(msg), 'bubble-owner': !isSelf(msg) && msg.isOwner }">
                      {{ msg.text }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Pulsante Scroll to Bottom -->
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

            <!-- Input Bar per invio messaggi (passa per Encore Pub/Sub) -->
            <div class="chat-input-area q-pa-sm bg-white border-top">
              <q-form @submit.prevent="sendMessage">
                <q-input
                  v-model="inputText"
                  outlined
                  dense
                  bg-color="white"
                  :placeholder="t('broadcast.chatPlaceholder')"
                  maxlength="500"
                  :disable="!connected || isSending"
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
                      :loading="isSending"
                      :disable="!connected || !inputText.trim()"
                      :aria-label="t('broadcast.chatSend')"
                      @click="sendMessage"
                    />
                  </template>
                </q-input>
              </q-form>
            </div>
          </div>
        </q-tab-panel>

        <!-- ================================================================= -->
        <!-- TAB PANEL: PARTECIPANTI                                           -->
        <!-- ================================================================= -->
        <q-tab-panel name="participants" class="q-pa-none stage-tab-panel column">
          <div v-if="managementError" class="q-pa-xs">
            <q-banner role="alert" dense class="bg-red-1 text-negative text-caption">
              {{ managementError }}
            </q-banner>
          </div>

          <div class="participants-list-container col">
            <q-list separator>
              <q-item v-for="p in participants" :key="p.identity || p.sid || 'participant'" class="q-py-sm">
                <q-item-section avatar>
                  <q-avatar size="36px" color="primary" text-color="white">
                    <img v-if="p.avatar_url" :src="p.avatar_url" alt="" />
                    <span v-else>{{ (p.name || p.identity || 'U').charAt(0).toUpperCase() }}</span>
                  </q-avatar>
                </q-item-section>

                <q-item-section>
                  <q-item-label class="text-weight-bold text-ellipsis">
                    {{ p.name || p.identity }}
                  </q-item-label>
                </q-item-section>

                <q-item-section v-if="canManage" side>
                  <q-btn
                    flat
                    round
                    dense
                    color="negative"
                    icon="person_remove"
                    :aria-label="t('broadcast.remove')"
                    :loading="managementBusy === p.identity"
                    :disable="!!managementBusy"
                    @click="emit('remove-participant', p.identity)"
                  >
                    <q-tooltip>{{ t('broadcast.remove') }}</q-tooltip>
                  </q-btn>
                </q-item-section>
              </q-item>

              <q-item v-if="!participants.length" class="q-py-xl">
                <q-item-section class="text-center text-grey-6">
                  <q-icon name="people_outline" size="48px" color="grey-4" class="q-mx-auto q-mb-sm" />
                  <div>{{ t('broadcast.noParticipants') }}</div>
                </q-item-section>
              </q-item>
            </q-list>
          </div>

          <div v-if="canManage" class="q-pa-sm text-caption text-grey-6 bg-grey-1 border-top">
            {{ t('broadcast.removeHint') }}
          </div>
        </q-tab-panel>
      </q-tab-panels>
    </q-card>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import type { Room as LiveRoom, RemoteParticipant } from 'livekit-client';
import { RoomEvent } from 'livekit-client';
import { useProfilesStore } from '@/stores/profiles-store';
import { createVideochatClient } from '@/encore/videochat';

export interface ParticipantItem {
  sid?: string | undefined;
  identity: string;
  name?: string | undefined;
  avatar_url?: string | undefined;
}

export interface InternalChatMessage {
  id: string;
  senderId: string;
  senderName: string;
  senderAvatar?: string | undefined;
  text: string;
  timestamp: number;
  isOwner?: boolean | undefined;
  type?: string | undefined;
  targetId?: string | undefined;
}

const props = withDefaults(
  defineProps<{
    room?: LiveRoom | null | undefined;
    connected: boolean;
    isOwner: boolean;
    canManage?: boolean | undefined;
    roomName: string;
    participants?: ParticipantItem[] | undefined;
    managementBusy?: string | undefined;
    managementError?: string | undefined;
    modelValue?: 'chat' | 'participants' | undefined;
  }>(),
  {
    room: undefined,
    canManage: false,
    participants: () => [],
    managementBusy: '',
    managementError: '',
    modelValue: 'chat',
  }
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: 'chat' | 'participants'): void;
  (e: 'remove-participant', identity: string): void;
}>();

const { t } = useI18n();
const $q = useQuasar();
const profiles = useProfilesStore();
const videochatClient = createVideochatClient(() => profiles.token);

const currentTab = computed({
  get: () => props.modelValue,
  set: (val: 'chat' | 'participants') => emit('update:modelValue', val),
});

// Stato gestione messaggi
const messages = ref<InternalChatMessage[]>([]);
const inputText = ref('');
const isSending = ref(false);
const loadingHistory = ref(false);
const messagesContainer = ref<HTMLDivElement>();
const showScrollBottom = ref(false);

const filteredMessages = computed(() => {
  return messages.value.filter((m) => m.type !== 'ADMIN_ALERT' || props.isOwner);
});

const isAdminOrAdvisor = computed(() => {
  const role = profiles.profile?.role?.toLowerCase();
  return role === 'admin' || role === 'chatadvisor';
});

function isSelf(msg: InternalChatMessage): boolean {
  const current = props.room?.localParticipant?.identity || profiles.profile?.user_id;
  return Boolean(current && msg.senderId === current);
}

function formatTime(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function scrollToBottom(smooth = false) {
  void nextTick(() => {
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
  showScrollBottom.value = distanceFromBottom > 100;
}

function clearChat() {
  messages.value = [];
}

/**
 * Carica lo storico iniziale dei messaggi della stanza dal backend Encore
 */
async function loadHistory(roomName: string) {
  if (!roomName || !profiles.isAuthenticated) return;
  loadingHistory.value = true;
  try {
    const res = await videochatClient.listMessages(roomName, 50);
    const loaded: InternalChatMessage[] = (res.messages || []).map((m) => ({
      id: m.id,
      senderId: m.sender_user_id,
      senderName: m.sender_name || 'Utente',
      senderAvatar: m.sender_avatar || undefined,
      text: m.body,
      timestamp: new Date(m.created_at).getTime(),
      isOwner: false,
      type: 'MESSAGE',
    }));

    // Deduplica e mantieni ordinati
    const map = new Map<string, InternalChatMessage>();
    loaded.forEach((item) => map.set(item.id, item));
    messages.value.forEach((item) => map.set(item.id, item));
    messages.value = Array.from(map.values()).sort((a, b) => a.timestamp - b.timestamp);
    scrollToBottom();
  } catch (err) {
    console.warn('Impossibile caricare lo storico chat da Encore:', err);
  } finally {
    loadingHistory.value = false;
  }
}

/**
 * Invia un messaggio:
 * 1. Passa per il backend Encore (che valida, memorizza e diffonde tramite Pub/Sub)
 * 2. Invia in tempo reale sul Data Channel LiveKit per la ricezione istantanea dei peer
 */
async function sendMessage() {
  const text = inputText.value.trim();
  if (!text || !props.roomName) return;

  isSending.value = true;
  const clientMessageId = `msg-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  const senderId = props.room?.localParticipant?.identity || profiles.profile?.user_id || 'me';
  const senderName = profiles.profile?.display_name || props.room?.localParticipant?.name || 'Io';
  const senderAvatar = profiles.profile?.avatar_url || undefined;

  const localMsg: InternalChatMessage = {
    id: clientMessageId,
    senderId,
    senderName,
    senderAvatar,
    text,
    timestamp: Date.now(),
    isOwner: props.isOwner,
    type: 'MESSAGE',
  };

  // Aggiungi immediatamente in locale per UX reattiva
  messages.value.push(localMsg);
  inputText.value = '';
  scrollToBottom(true);

  try {
    // 1. Invio a Encore (salvataggio DB + Pub/Sub topic per filtering/audit)
    if (profiles.isAuthenticated) {
      await videochatClient.sendMessage(props.roomName, {
        client_message_id: clientMessageId,
        text,
      });
    }

    // 2. Diffusione via LiveKit Data Channel per tutti i client connessi
    if (props.room && props.connected) {
      const payload = JSON.stringify(localMsg);
      const encoded = new TextEncoder().encode(payload);
      await props.room.localParticipant.publishData(encoded, { reliable: true, topic: 'chat' });
    }
  } catch (err) {
    console.error('Errore durante la pubblicazione del messaggio:', err);
  } finally {
    isSending.value = false;
  }
}

/**
 * Invia un comando / alert amministrativo con pubblicazione su Pub/Sub Encore
 */
function openAdminAlertDialog() {
  const targetId = props.room?.localParticipant?.identity || '';
  const msgDialog = $q.dialog({
    title: t('admin.sendAlert'),
    prompt: { model: '', label: t('admin.message') },
    cancel: true,
    persistent: true,
  });

  msgDialog.onOk(async (alertText: string) => {
    if (!alertText.trim() || !props.roomName) return;
    try {
      // Pubblica comando su Pub/Sub Encore
      await videochatClient.sendCommand(props.roomName, {
        command: 'alert',
        target_id: targetId,
        payload: alertText.trim(),
      });

      // Diffondi via LiveKit
      if (props.room && props.connected) {
        const adminMsg: InternalChatMessage = {
          id: `cmd-${Date.now()}`,
          senderId: profiles.profile?.user_id || 'admin',
          senderName: 'Admin',
          text: alertText.trim(),
          timestamp: Date.now(),
          isOwner: props.isOwner,
          type: 'ADMIN_ALERT',
          targetId,
        };
        const encoded = new TextEncoder().encode(JSON.stringify(adminMsg));
        await props.room.localParticipant.publishData(encoded, { reliable: true, topic: 'chat' });
      }

      $q.notify({ type: 'positive', message: 'Alert inviato con successo' });
    } catch (err) {
      console.warn('Errore invio alert amministratore:', err);
    }
  });
}

/**
 * Gestore ricezione dati LiveKit Data Channel
 */
function handleDataReceived(payload: Uint8Array, _participant?: RemoteParticipant, _kind?: unknown, topic?: string) {
  if (topic && topic !== 'chat') return;
  try {
    const text = new TextDecoder().decode(payload);
    const msg = JSON.parse(text) as InternalChatMessage;

    // Gestione alert amministratore
    if (msg.type === 'ADMIN_ALERT') {
      if (msg.targetId && !props.isOwner && msg.targetId !== profiles.profile?.user_id) {
        return;
      }
      $q.notify({
        type: 'negative',
        message: t('admin.notifications.adminAlert'),
        caption: msg.text,
        timeout: 0,
        actions: [{ label: t('close'), handler: () => {} }],
      });
      return;
    }

    // Aggiungi se non già presente
    if (!messages.value.some((m) => m.id === msg.id)) {
      messages.value.push(msg);
      scrollToBottom(true);
    }
  } catch {
    // Ignora payload non JSON
  }
}

// Watcher per aggancio listener LiveKit quando la room cambia o si connette
watch(
  () => props.room,
  (newRoom, oldRoom) => {
    if (oldRoom) {
      oldRoom.off(RoomEvent.DataReceived, handleDataReceived);
    }
    if (newRoom) {
      newRoom.on(RoomEvent.DataReceived, handleDataReceived);
    }
  },
  { immediate: true }
);

// Watcher per caricare lo storico messaggi all'avvio o al cambio stanza
watch(
  () => props.roomName,
  (newRoomName) => {
    if (newRoomName) {
      void loadHistory(newRoomName);
    }
  },
  { immediate: true }
);

onMounted(() => {
  if (props.roomName) {
    void loadHistory(props.roomName);
  }
});

onBeforeUnmount(() => {
  if (props.room) {
    props.room.off(RoomEvent.DataReceived, handleDataReceived);
  }
});
</script>

<style scoped>
.broadcast-stage-right {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  height: 100%;
  min-height: 0;
  overflow: auto;
  box-sizing: border-box;
}

.stage-right-card {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  height: 100%;
  min-height: 0;
  border-radius: 8px;
  overflow: auto;
  background: #ffffff;
}

.stage-right-panels {
  flex: 1 1 0;
  min-height: 0;
  height: 100%;
  position: relative;
  background: #ffffff;
}

.stage-right-panels :deep(.q-panel) {
  display: flex !important;
  flex: 1 1 0 !important;
  width: 100% !important;
  height: 100% !important;
}

.stage-right-panels :deep(.q-panel > div) {
  width: 100% !important;
  height: 100% !important;
}

.stage-tab-panel,
.stage-right-panels :deep(.q-tab-panel) {
  position: absolute !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  bottom: 0 !important;
  width: 100% !important;
  height: 100% !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: auto !important;
  box-sizing: border-box !important;
}

.broadcast-chat-wrapper {
  position: relative;
  overflow: hidden;
}

.chat-messages {
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 0;
}

.chat-message-item {
  display: flex;
  flex-direction: column;
}

.chat-message-item.is-self .message-content {
  align-items: flex-end;
}

.message-bubble {
  padding: 8px 12px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.4;
  word-break: break-word;
  background: #f1f5f9;
  color: #1e293b;
  display: inline-block;
  max-width: 85%;
}

.bubble-self {
  background: #0284c7;
  color: #ffffff;
  border-bottom-right-radius: 2px;
}

.bubble-owner {
  background: #fef3c7;
  border-left: 3px solid #f59e0b;
}

.message-time {
  font-size: 10px;
}

.sender-name {
  max-width: 140px;
  font-size: 11px;
}

.scroll-bottom-btn {
  position: absolute;
  bottom: 64px;
  right: 16px;
  z-index: 10;
}

.participants-list-container {
  overflow-y: auto;
  min-height: 0;
}

.border-bottom {
  border-bottom: 1px solid #e2e8f0;
}

.border-top {
  border-top: 1px solid #e2e8f0;
}
</style>
