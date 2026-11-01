<template>
  <q-page class="studio-page" :class="{ 'studio-page--window': inWindow }">
    <header class="studio-header">
      <h1>
        {{ mine?.room_name || t('studio.title') }}
        <span v-if="live.connected.value" class="studio-badge" :class="{ 'studio-badge--on': onAir }">
          ● {{ onAir ? t('studio.onAir') : t('studio.offAir') }}
        </span>
      </h1>
    </header>

    <p v-if="!session.loaded.value || (opening && !mine)" class="studio-status">{{ t('studio.opening') }}</p>

    <div v-else-if="!session.canBroadcast.value" class="studio-status" role="alert">
      <q-icon name="videocam_off" size="48px" />
      <h2>{{ t('studio.disabledTitle') }}</h2>
      <p>{{ t('studio.disabledText') }}</p>
    </div>

    <!-- Diretta terminata: stanza chiusa, si può riaprire -->
    <div v-else-if="ended" class="studio-status" role="status">
      <q-icon name="stop_circle" size="48px" />
      <p>{{ t('studio.ended') }}</p>
      <button type="button" class="studio-button studio-button--center" :disabled="opening" @click="reopen">{{ t('studio.reopen') }}</button>
    </div>

    <!-- Stanza spostata in una finestra separata: qui resta solo l'avviso -->
    <div v-else-if="movedToWindow" class="studio-status" role="status">
      <q-icon name="open_in_new" size="48px" />
      <p>{{ t('studio.movedToWindow') }}</p>
      <button type="button" class="studio-button studio-button--center" @click="bringBack">{{ t('studio.bringBack') }}</button>
    </div>

    <!-- Video e chat divisi da uno splitter trascinabile, come nel pannello -->
    <q-splitter
      v-else
      v-model="splitter"
      :limits="[25, 85]"
      :horizontal="$q.screen.lt.md"
      class="studio-splitter"
      separator-class="studio-splitter-separator"
      @update:model-value="saveSplitter"
    >
      <template #before>
        <section class="studio-stage" :aria-label="t('studio.title')">
          <div v-if="live.reconnecting.value" class="studio-banner">{{ t('live.reconnecting') }}</div>
          <div v-if="failure" class="studio-banner studio-banner--error" role="alert">
            {{ failure }}
            <button v-if="!live.connected.value" type="button" class="studio-button" @click="open">{{ t('live.retry') }}</button>
          </div>

          <LiveStage :videos="live.videos.value" :audios="[]" :captions="false">
            <template #placeholder>
              <q-icon name="videocam" size="56px" />
              <p>{{ live.connecting.value ? t('live.connecting') : t('studio.preview') }}</p>
            </template>
          </LiveStage>

          <div class="studio-controls">
            <button
              type="button"
              class="studio-button"
              :class="{ 'studio-button--on': live.cameraOn.value }"
              :aria-pressed="live.cameraOn.value"
              :disabled="!live.connected.value || mediaBusy"
              @click="toggle('camera')"
            >
              <q-icon :name="live.cameraOn.value ? 'videocam' : 'videocam_off'" size="18px" />
              {{ t('studio.camera') }}
            </button>
            <button
              type="button"
              class="studio-button"
              :class="{ 'studio-button--on': live.micOn.value }"
              :aria-pressed="live.micOn.value"
              :disabled="!live.connected.value || mediaBusy"
              @click="toggle('microphone')"
            >
              <q-icon :name="live.micOn.value ? 'mic' : 'mic_off'" size="18px" />
              {{ t('studio.microphone') }}
            </button>
            <button
              type="button"
              class="studio-button"
              :aria-label="t('webchat.webcam.open')"
              :title="t('webchat.webcam.open')"
              :disabled="mediaBusy"
              @click="showSettings = true"
            >
              <q-icon name="tune" size="18px" />
            </button>
            <button v-if="live.connected.value" type="button" class="studio-button" :disabled="mediaBusy" @click="stop">
              <q-icon name="stop" size="18px" />
              {{ t('studio.stop') }}
            </button>
            <button v-if="!inWindow" type="button" class="studio-button studio-button--end" :disabled="!mine" @click="openInWindow">
              <q-icon name="open_in_new" size="18px" />
              {{ t('studio.openWindow') }}
            </button>
          </div>
          <p v-if="settingsSaved" class="studio-note" role="status">{{ t('webchat.webcam.savedSuccess') }}</p>
        </section>
      </template>

      <template #separator>
        <q-avatar color="grey-9" text-color="white" size="24px" icon="drag_indicator" class="shadow-2 cursor-pointer" />
      </template>

      <template #after>
        <aside class="studio-chat">
          <ChatPanel
            :messages="chat.messages.value"
            :self-id="chatUser?.id"
            :loading="chat.loading.value"
            :sending="chat.sending.value"
            :can-write="!!chatUser"
            :disabled="!live.connected.value"
            :send="chat.send"
          />
        </aside>
      </template>
    </q-splitter>

    <WebcamSettingsDialog v-model="showSettings" :initial-settings="webcamSettings" @apply="applySettings" />
  </q-page>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMeta, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
  ChatPanel,
  LiveStage,
  WebcamSettingsDialog,
  buildWebcamOptions,
  errorMessage,
  loadWebcamSettings,
  republishCamera,
  saveWebcamSettings,
  useLiveRoom,
  useRoomChat,
  type ChatUser,
  type WebcamSettings,
} from '@extra/webchat';
import { APIError, ErrCode, type videochat } from '@/encore/client';
import { api } from '@/encore/api';
import { API_BASE_URL } from '@/encore/base-url';
import { useSession } from '@/composables/useSession';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const session = useSession();
const $q = useQuasar();

useMeta(() => ({ title: `${t('studio.title')} - PublicSite` }));

const localePrefix = computed(() => (route.params.locale ? `/${String(route.params.locale)}` : ''));
const mine = ref<videochat.MyLive | null>(null);
// Studio aperto con "Apri in finestra" (rotta senza header): il bottone non serve.
const inWindow = route.name === 'studio-window';
const movedToWindow = ref(false);
const ended = ref(false);
// La stanza aperta da questa pagina: va chiusa quando si esce.
let ownedRoomSid = '';
let studioWindow: Window | null = null;
const opening = ref(true);
const mediaBusy = ref(false);
const mediaError = ref('');
const showSettings = ref(false);

// Proporzione video/chat dello splitter, ricordata nel browser.
const splitterKey = 'studio_splitter_ratio';
function loadSplitter(): number {
  try {
    const saved = Number(localStorage.getItem(splitterKey));
    if (saved >= 25 && saved <= 85) return saved;
  } catch {
    // localStorage non disponibile
  }
  return 70;
}
const splitter = ref(loadSplitter());
function saveSplitter(value: number) {
  try {
    localStorage.setItem(splitterKey, String(value));
  } catch {
    // localStorage non disponibile
  }
}
const settingsSaved = ref(false);
const webcamSettings = ref<WebcamSettings>(loadWebcamSettings());

type MediaEvent = 'VIDEO_ON' | 'VIDEO_OFF' | 'MIC_ON' | 'MIC_OFF';

// Registro eventi della stanza, come nel pannello: accensione e spegnimento di
// camera e microfono finiscono nel chatlog con il SID della stanza.
function logMediaEvent(event: MediaEvent) {
  const roomId = mine.value?.room_sid || mine.value?.room_name;
  const me = session.me.value;
  if (!roomId || !live.isOwner.value) return;
  const data = JSON.stringify({ event, senderId: me?.user_id, senderName: me?.display_name, timestamp: Date.now() });
  void api.broadcasters.AddChatLog(roomId, { type: 'ACTION', data }).catch((reason) => {
    console.warn('Errore registrazione evento:', reason);
  });
}

const live = useLiveRoom({
  // Uscita o chiusura della stanza con camera o microfono accesi.
  onLeave: (state) => {
    if (!state.isOwner) return;
    if (state.cameraOn) logMediaEvent('VIDEO_OFF');
    if (state.micOn) logMediaEvent('MIC_OFF');
  },
});
const onAir = computed(() => live.cameraOn.value || live.micOn.value);
const failure = computed(() => mediaError.value || live.error.value);

const chatUser = computed<ChatUser | null>(() => {
  const me = session.me.value;
  return me ? { id: me.user_id, name: me.display_name || t('webchat.user'), avatar: me.avatar_url || undefined } : null;
});

const chat = useRoomChat({
  room: live.room,
  chatKey: computed(() => mine.value?.room_sid ?? ''),
  source: {
    history: async (key) => (await api.videochat.ListMessages(key, { limit: 50 })).messages,
    send: (key, params) => api.videochat.SendMessage(key, { ...params, kind: 'text' }),
  },
  ownerId: live.ownerIdentity,
  user: chatUser,
  fallbackName: () => t('webchat.user'),
});

// Apre la stanza del broadcaster (creata dal server se non esiste) ed entra
// come proprietario. Camera e microfono restano spenti finché non li accende.
async function open() {
  mediaError.value = '';
  opening.value = true;
  try {
    await live.leave();
    const room = await api.videochat.OpenMyLive();
    mine.value = room;
    await live.join(() => api.videochat.CreateToken({ room_name: room.room_name ?? '' }));
    if (live.connected.value) ownedRoomSid = room.room_sid ?? '';
  } catch (reason) {
    // Broadcasting disabilitato nel frattempo: lo stato aggiornato mostra l'avviso.
    if (reason instanceof APIError && reason.code === ErrCode.PermissionDenied) await session.refresh();
    else mediaError.value = errorMessage(reason);
  } finally {
    opening.value = false;
  }
}

async function toggle(kind: 'camera' | 'microphone') {
  const current = live.room.value;
  if (!current || !live.connected.value || mediaBusy.value) return;
  mediaBusy.value = true;
  mediaError.value = '';
  try {
    if (kind === 'camera') {
      const enable = !live.cameraOn.value;
      const { captureOptions, publishOptions } = buildWebcamOptions(webcamSettings.value);
      await current.localParticipant.setCameraEnabled(enable, captureOptions, publishOptions);
      logMediaEvent(enable ? 'VIDEO_ON' : 'VIDEO_OFF');
    } else {
      const enable = !live.micOn.value;
      await current.localParticipant.setMicrophoneEnabled(enable);
      logMediaEvent(enable ? 'MIC_ON' : 'MIC_OFF');
    }
    if (live.room.value === current) live.sync();
  } catch (reason) {
    if (live.room.value === current) mediaError.value = errorMessage(reason);
  } finally {
    mediaBusy.value = false;
  }
}

// Termina la diretta: spegne camera e microfono, chiude la stanza (gli
// spettatori vengono disconnessi) ed esce.
async function stop() {
  const current = live.room.value;
  const roomSid = mine.value?.room_sid;
  if (!current || mediaBusy.value) return;
  mediaBusy.value = true;
  mediaError.value = '';
  try {
    if (live.cameraOn.value) {
      await current.localParticipant.setCameraEnabled(false);
      logMediaEvent('VIDEO_OFF');
    }
    if (live.micOn.value) {
      await current.localParticipant.setMicrophoneEnabled(false);
      logMediaEvent('MIC_OFF');
    }
    if (roomSid) await api.videochat.CloseMyLive({ room_sid: roomSid });
    ownedRoomSid = '';
    await live.leave();
    ended.value = true;
  } catch (reason) {
    mediaError.value = errorMessage(reason);
  } finally {
    mediaBusy.value = false;
  }
}

async function reopen() {
  ended.value = false;
  await open();
}

// Uscita dalla pagina (cambio pagina, ricarica, chiusura della scheda o della
// finestra separata) con la stanza aperta: la diretta termina. keepalive fa
// partire la richiesta anche mentre la pagina viene chiusa.
function closeOnExit() {
  const roomSid = ownedRoomSid;
  if (!roomSid) return;
  // Riportata nella pagina di origine con "Riporta qui": la stanza resta aperta.
  if ((window as Window & { studioHandover?: boolean }).studioHandover) return;
  void fetch(`${API_BASE_URL}/videochat/my-live/close`, {
    method: 'POST',
    keepalive: true,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ room_sid: roomSid }),
  }).catch(() => {});
  ownedRoomSid = '';
}

// Salva le impostazioni e, se la camera è accesa, la ripubblica subito.
async function applySettings(settings: WebcamSettings) {
  webcamSettings.value = { ...settings };
  saveWebcamSettings(settings);
  settingsSaved.value = false;
  const current = live.room.value;
  if (current && live.connected.value && live.cameraOn.value) {
    mediaBusy.value = true;
    mediaError.value = '';
    try {
      await republishCamera(current, settings);
      if (live.room.value === current) live.sync();
    } catch (reason) {
      mediaError.value = errorMessage(reason);
      return;
    } finally {
      mediaBusy.value = false;
    }
  }
  settingsSaved.value = true;
  setTimeout(() => (settingsSaved.value = false), 4000);
}

// Sposta la stanza in una finestra separata, con il nome della stanza come
// nome della finestra (un secondo clic riusa la stessa). Lo stesso utente non
// può stare due volte nella stanza: questa pagina ne esce.
function openInWindow() {
  const roomName = mine.value?.room_name;
  if (!roomName) return;
  const locale = route.params.locale ? { locale: String(route.params.locale) } : {};
  const url = window.location.origin + router.resolve({ name: 'studio-window', params: locale }).href;
  // Stesse dimensioni della pagina di origine.
  studioWindow = window.open(
    url,
    roomName,
    `toolbar=no, location=no, directories=no, status=no, menubar=no, scrollbars=yes, resizable=yes, width=${window.innerWidth}, height=${window.innerHeight}`,
  );
  if (!studioWindow) {
    mediaError.value = t('studio.popupBlocked');
    return;
  }
  movedToWindow.value = true;
  ownedRoomSid = '';
  void live.leave();
}

// Riporta la stanza in questa pagina chiudendo la finestra separata.
async function bringBack() {
  if (studioWindow && !studioWindow.closed) {
    // Stessa origine: la finestra vede il segnale e non chiude la stanza.
    (studioWindow as Window & { studioHandover?: boolean }).studioHandover = true;
    studioWindow.close();
  }
  studioWindow = null;
  movedToWindow.value = false;
  await open();
}

onMounted(async () => {
  window.addEventListener('pagehide', closeOnExit);
  await session.refresh();
  if (!session.me.value) {
    await router.replace(`${localePrefix.value}/login?redirect=${encodeURIComponent(route.fullPath)}`);
    return;
  }
  if (session.canBroadcast.value) await open();
  else opening.value = false;
});

onBeforeUnmount(() => {
  window.removeEventListener('pagehide', closeOnExit);
  closeOnExit();
});
</script>

<style scoped>
.studio-page {
  --webchat-surface: #111111;
  --webchat-header: #161616;
  --webchat-text: #e9e9e9;
  --webchat-muted: #929292;
  --webchat-border: #252525;
  --webchat-bubble: #1f1f1f;
  --webchat-bubble-text: #e9e9e9;
  --webchat-self: #f4cf00;
  --webchat-self-text: #111111;
  --webchat-owner: #2a2410;
  --webchat-owner-text: #f4e7a1;
  --webchat-accent: #ff9500;
  --webchat-input: #0b0b0b;
  --webchat-stage-bg: #151515;
  --webchat-stage-text: #929292;
  padding: 24px 16px 48px;
  background: #050505;
  color: #e9e9e9;
}

.studio-header h1 {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin: 0 0 16px;
  font-size: 32px;
  font-weight: 400;
  line-height: 1.2;
  letter-spacing: 0;
}

.studio-badge {
  padding: 2px 10px;
  border-radius: 4px;
  background: #333333;
  color: #cccccc;
  font-size: 14px;
  font-weight: 700;
  text-transform: uppercase;
}

.studio-badge--on {
  background: #d32f2f;
  color: #ffffff;
}

.studio-status {
  padding: 64px 16px;
  text-align: center;
}

.studio-status h2 {
  margin: 12px 0 8px;
  font-size: 22px;
  font-weight: 500;
  line-height: 1.3;
  letter-spacing: 0;
}

.studio-status p {
  margin: 0;
  color: #b0b0b0;
}

.studio-splitter {
  height: calc(100vh - 260px);
  min-height: 480px;
}

.studio-stage {
  min-width: 0;
  padding-right: 12px;
}

:deep(.studio-splitter-separator) {
  background-color: #252525;
  transition: background-color 0.2s ease;
}

:deep(.studio-splitter-separator:hover) {
  background-color: #f4cf00;
}

.studio-controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}

.studio-note {
  margin: 8px 0 0;
  color: #a5d6a7;
  font-size: 13px;
}

.studio-chat {
  height: 100%;
  margin-left: 12px;
  overflow: hidden;
  border: 1px solid #252525;
  border-radius: 4px;
}

.studio-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background: #2a2410;
  color: #f4e7a1;
}

.studio-banner--error {
  background: #3a1414;
  color: #ffcdd2;
}

.studio-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid #555555;
  border-radius: 4px;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.studio-button--on {
  border-color: #f4cf00;
  background: #f4cf00;
  color: #111111;
}

.studio-button:disabled {
  opacity: 0.5;
  cursor: default;
}

.studio-button:focus-visible {
  outline: 2px solid #bbbbbb;
  outline-offset: 2px;
}

.studio-button--end {
  margin-left: auto;
}

.studio-button--center {
  margin-top: 16px;
}

/* Finestra separata: solo webchat, a tutta altezza */
.studio-page--window {
  padding: 8px 12px 12px;
}

.studio-page--window .studio-header h1 {
  margin-bottom: 8px;
  font-size: 18px;
}

.studio-page--window .studio-splitter {
  height: calc(100vh - 52px);
  min-height: 0;
}

/* Schermi stretti: splitter orizzontale, video sopra e chat sotto */
@media (max-width: 1023px) {
  .studio-splitter,
  .studio-page--window .studio-splitter {
    height: 900px;
  }

  .studio-stage {
    padding: 0 0 12px;
  }

  .studio-chat {
    margin: 12px 0 0;
  }

  .studio-header h1 {
    font-size: 26px;
  }
}
</style>
