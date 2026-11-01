<template>
  <q-page padding>
    <div class="row items-center q-gutter-sm q-mb-md">
      <q-btn flat round icon="arrow_back" to="/wideochat" :aria-label="t('broadcast.back')">
        <q-tooltip>{{ t('broadcast.back') }}</q-tooltip>
      </q-btn>
      <h1 class="text-h5 q-my-none">{{ details?.display_name || t('broadcast.title') }}</h1>
      <q-badge v-if="connected">{{ t(isOwner ? 'broadcast.owner' : 'broadcast.viewer') }}</q-badge>
      <q-space />

      <!-- Owner Media Controls: Icon buttons with tooltips -->
      <template v-if="connected && isOwner">
        <q-btn
          round
          :color="cameraOn ? 'primary' : 'grey-8'"
          :icon="cameraOn ? 'videocam' : 'videocam_off'"
          :aria-label="t('broadcast.camera')"
          :aria-pressed="cameraOn"
          :disable="mediaBusy || reconnecting"
          @click="toggleMedia('camera')"
        >
          <q-tooltip>{{ t('broadcast.camera') }}</q-tooltip>
        </q-btn>

        <q-btn
          round
          :color="micOn ? 'primary' : 'grey-8'"
          :icon="micOn ? 'mic' : 'mic_off'"
          :aria-label="t('broadcast.microphone')"
          :aria-pressed="micOn"
          :disable="mediaBusy || reconnecting"
          @click="toggleMedia('microphone')"
        >
          <q-tooltip>{{ t('broadcast.microphone') }}</q-tooltip>
        </q-btn>

        <q-btn
          round
          :color="screenOn ? 'primary' : 'grey-8'"
          :icon="screenOn ? 'stop_screen_share' : 'screen_share'"
          :aria-label="t('broadcast.screen')"
          :aria-pressed="screenOn"
          :disable="mediaBusy || reconnecting"
          @click="toggleMedia('screen')"
        >
          <q-tooltip>{{ t('broadcast.screen') }}</q-tooltip>
        </q-btn>

        <q-btn
          round
          flat
          icon="tune"
          color="primary"
          :aria-label="t('broadcast.webcamSettings')"
          @click="showWebcamSettings = true"
        >
          <q-tooltip>{{ t('broadcast.webcamSettings') }}</q-tooltip>
        </q-btn>

        <q-btn
          v-if="canManage"
          round
          flat
          icon="refresh"
          color="primary"
          :loading="loadingParticipants"
          :aria-label="t('videochat.refresh')"
          @click="loadParticipants"
        >
          <q-tooltip>{{ t('videochat.refresh') }}</q-tooltip>
        </q-btn>

        <q-btn
          v-if="canManage"
          round
          flat
          icon="cancel"
          color="negative"
          :disable="!!managementBusy"
          :aria-label="t('broadcast.close')"
          @click="confirmClose = true"
        >
          <q-tooltip>{{ t('broadcast.close') }}</q-tooltip>
        </q-btn>
      </template>

      <q-btn v-if="connected" outline icon="logout" :label="t('broadcast.leave')" @click="leave">
        <q-tooltip>{{ t('broadcast.leave') }}</q-tooltip>
      </q-btn>
    </div>

    <q-banner v-if="error" role="alert" class="bg-red-1 text-negative q-mb-md">{{ error }}</q-banner>
    <q-banner v-if="!profiles.isAuthenticated" class="bg-grey-2">
      {{ t('broadcast.loginRequired') }}
      <template #action><q-btn flat to="/login" :label="t('nav.login')" /></template>
    </q-banner>
    <template v-else>
      <q-banner v-if="!connected" class="bg-grey-2 q-mb-md">
        {{ t(ended ? 'broadcast.disconnected' : 'broadcast.joinHint') }}
        <template #action>
          <q-btn color="primary" :loading="connecting || opening" :label="t('broadcast.join')" @click="join" />
        </template>
      </q-banner>
      <template v-if="connected">
        <q-banner v-if="reconnecting" class="bg-orange-1 q-mb-md">{{ t('broadcast.reconnecting') }}</q-banner>
        <q-banner v-if="needsAudio" class="bg-orange-1 q-mb-md">
          {{ t('broadcast.audioHint') }}
          <template #action><q-btn flat :label="t('broadcast.enableAudio')" @click="enableAudio" /></template>
        </q-banner>
        <q-splitter
          v-model="splitterModel"
          :limits="[25, 85]"
          :horizontal="$q.screen.lt.md"
          class="broadcast-splitter"
          separator-class="broadcast-splitter-separator"
          @update:model-value="saveSplitterRatio"
        >
          <!-- Video Section on Left (or Top on mobile) -->
          <template #before>
            <LiveStage class="q-pa-sm" :videos="videos" :audios="audios" />
          </template>

          <template #separator>
            <q-avatar
              color="primary"
              text-color="white"
              size="24px"
              icon="drag_indicator"
              class="shadow-2 cursor-pointer"
            />
          </template>

          <!-- Chat & Participants Section on Right (or Bottom on mobile) via BroadcastStageRight -->
          <template #after>
            <BroadcastStageRight
              v-model="rightTab"
              :room="room"
              :connected="connected"
              :is-owner="isOwner"
              :owner-id="ownerIdentity"
              :can-manage="canManage"
              :room-name="activeRoomId"
              :participants="displayParticipants"
              :management-busy="managementBusy"
              :management-error="managementError"
              @remove-participant="removeParticipant"
            />
          </template>
        </q-splitter>
      </template>
    </template>
    <q-dialog v-model="confirmClose" :persistent="managementBusy === 'close'">
      <q-card style="width: 420px; max-width: 100%">
        <q-card-section class="text-h6">{{ t('broadcast.close') }}</q-card-section>
        <q-card-section>{{ t('broadcast.closeHint') }}</q-card-section>
        <q-card-actions align="right">
          <q-btn flat :label="t('actions.cancel')" :disable="managementBusy === 'close'" v-close-popup />
          <q-btn color="negative" :label="t('broadcast.close')" :loading="managementBusy === 'close'" @click="closeRoom" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <WebcamSettingsDialog
      v-model="showWebcamSettings"
      :initial-settings="webcamSettings"
      @apply="onApplyWebcamSettings"
    />
  </q-page>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
  LiveStage,
  WebcamSettingsDialog,
  buildWebcamOptions,
  errorMessage as message,
  loadWebcamSettings,
  republishCamera,
  saveWebcamSettings,
  useLiveRoom,
  type WebcamSettings,
} from '@extra/webchat';
import BroadcastStageRight from '@/components/videochat/BroadcastStageRight.vue';
import { api as client } from '@/encore/api';
import type { videochat } from '@/encore/client';
import { useProfilesStore } from '@/stores/profiles-store';
import { useChatStore } from '@/stores/chat-store';

const profiles = useProfilesStore();
const chatStore = useChatStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const $q = useQuasar();
const api = client.videochat;
const roomName = computed(() => String(route.params.roomName || ''));
const details = ref<videochat.Room>();
const opening = ref(false);
const activeRoomName = computed(() => details.value?.name || roomName.value);
// Collegamento alla stanza condiviso con il sito pubblico (packages/webchat).
const live = useLiveRoom({
  onLeave: (state) => {
    if (!state.isOwner) return;
    if (state.cameraOn) logMediaEvent('VIDEO_OFF');
    if (state.micOn) logMediaEvent('MIC_OFF');
  },
});
const {
  room, connected, connecting, reconnecting, ended, error, isOwner, ownerIdentity,
  cameraOn, micOn, screenOn, needsAudio, videos, audios, sync: syncMedia, startAudio: enableAudio,
} = live;
const activeRoomId = computed(() => details.value?.sid || room.value?.name || details.value?.name || roomName.value);
const canManage = computed(() => !!details.value && details.value.owner_id === profiles.profile?.user_id);
const mediaBusy = ref(false);
const participants = ref<videochat.Participant[]>([]);
const loadingParticipants = ref(false);
const managementError = ref('');
const managementBusy = ref('');
const confirmClose = ref(false);
const showWebcamSettings = ref(false);
const rightTab = ref<'chat' | 'participants'>('chat');

const savedSplitter = typeof window !== 'undefined' ? localStorage.getItem('broadcast_splitter_ratio') : null;
const parsedSplitter = savedSplitter ? Number(savedSplitter) : NaN;
const splitterModel = ref(!Number.isNaN(parsedSplitter) && parsedSplitter >= 25 && parsedSplitter <= 85 ? parsedSplitter : 70);

function saveSplitterRatio(val: number) {
  try {
    localStorage.setItem('broadcast_splitter_ratio', String(val));
  } catch {
    // ignore
  }
}

const displayParticipants = computed(() => {
  const ownerId = details.value?.owner_id || ownerIdentity.value;
  if (canManage.value && participants.value.length) {
    return participants.value.filter((p) => p.identity !== ownerId);
  }
  const current = room.value;
  if (!current || !connected.value) return [];
  const list: { sid: string; identity: string; name: string; avatar_url?: string | undefined }[] = [];
  if (current.localParticipant && current.localParticipant.identity !== ownerId) {
    list.push({
      sid: current.localParticipant.sid,
      identity: current.localParticipant.identity,
      name: profiles.profile?.display_name || current.localParticipant.name || current.localParticipant.identity,
      avatar_url: profiles.profile?.avatar_url,
    });
  }
  current.remoteParticipants.forEach((p) => {
    if (p.identity !== ownerId) {
      list.push({
        sid: p.sid,
        identity: p.identity,
        name: p.name || p.identity,
        avatar_url: p.metadata && p.metadata.startsWith('http') ? p.metadata : undefined,
      });
    }
  });
  return list;
});

watch(rightTab, (tab) => {
  if (tab === 'participants' && canManage.value) {
    void loadParticipants();
  }
});

const webcamSettings = ref<WebcamSettings>(loadWebcamSettings());

async function onApplyWebcamSettings(newSettings: WebcamSettings) {
  webcamSettings.value = { ...newSettings };
  saveWebcamSettings(newSettings);

  const current = room.value;
  if (cameraOn.value && current && connected.value && isOwner.value) {
    mediaBusy.value = true;
    error.value = '';
    try {
      await republishCamera(current, newSettings);
      syncMedia();
      $q.notify({ type: 'positive', message: t('webchat.webcam.savedSuccess') });
    } catch (reason) {
      error.value = message(reason);
    } finally {
      mediaBusy.value = false;
    }
  } else {
    $q.notify({ type: 'positive', message: t('webchat.webcam.savedSuccess') });
  }
}

let participantRequest = 0;
let refreshTimer: ReturnType<typeof setInterval> | undefined;

function logMediaEvent(eventText: 'VIDEO_ON' | 'VIDEO_OFF' | 'MIC_ON' | 'MIC_OFF') {
  if (!isOwner.value) return;
  const senderId = room.value?.localParticipant?.identity || profiles.profile?.user_id || 'broadcaster';
  const senderName = profiles.profile?.display_name || room.value?.localParticipant?.name || 'Broadcaster';
  const targetRoom = activeRoomId.value || roomName.value;
  if (!targetRoom) return;
  void chatStore.logMediaEvent(targetRoom, eventText, senderId, senderName);
}

async function leave() {
  mediaBusy.value = false;
  await live.leave();
}

async function join() {
  if (opening.value || connecting.value || connected.value || !profiles.isAuthenticated) return;
  const joined = await live.join(async () => {
    const opened = await api.OpenRoom(roomName.value);
    details.value = opened;
    return api.CreateToken({ room_name: opened.name });
  });
  // Pubblicare richiede sempre un'azione esplicita del proprietario.
  if (joined && canManage.value) void loadParticipants();
}

async function toggleMedia(kind: 'camera' | 'microphone' | 'screen') {
  const current = room.value;
  if (!current || !connected.value || !isOwner.value || mediaBusy.value) return;
  mediaBusy.value = true;
  error.value = '';
  try {
    if (kind === 'camera') {
      if (!cameraOn.value) {
        const { captureOptions, publishOptions } = buildWebcamOptions(webcamSettings.value);
        await current.localParticipant.setCameraEnabled(true, captureOptions, publishOptions);
        logMediaEvent('VIDEO_ON');
      } else {
        await current.localParticipant.setCameraEnabled(false);
        logMediaEvent('VIDEO_OFF');
      }
    }
    if (kind === 'microphone') {
      if (!micOn.value) {
        await current.localParticipant.setMicrophoneEnabled(true);
        logMediaEvent('MIC_ON');
      } else {
        await current.localParticipant.setMicrophoneEnabled(false);
        logMediaEvent('MIC_OFF');
      }
    }
    if (kind === 'screen') await current.localParticipant.setScreenShareEnabled(!screenOn.value, { audio: true });
    // Uscita durante l'operazione: la stanza vecchia va chiusa.
    if (room.value !== current) { await current.disconnect(); return; }
    syncMedia();
  } catch (reason) {
    if (room.value === current) error.value = message(reason);
  } finally {
    if (room.value === current) mediaBusy.value = false;
  }
}

async function loadParticipants() {
  if (!canManage.value || loadingParticipants.value) return;
  const request = ++participantRequest;
  loadingParticipants.value = true;
  managementError.value = '';
  try {
    const response = await api.ListParticipants(activeRoomId.value);
    if (request === participantRequest) participants.value = response.participants ?? [];
  } catch (reason) {
    if (request === participantRequest) managementError.value = message(reason);
  } finally {
    if (request === participantRequest) loadingParticipants.value = false;
  }
}

async function removeParticipant(identity: string) {
  if (managementBusy.value || !canManage.value) return;
  const version = pageVersion;
  managementBusy.value = identity;
  managementError.value = '';
  try {
    await api.RemoveParticipant(activeRoomId.value, identity);
    if (version === pageVersion) {
      $q.notify({ type: 'positive', message: t('broadcast.participantRemoved') });
      await loadParticipants();
    }
  }
  catch (reason) { if (version === pageVersion) managementError.value = message(reason); }
  finally { if (version === pageVersion) managementBusy.value = ''; }
}

async function closeRoom() {
  if (managementBusy.value || !canManage.value) return;
  const version = pageVersion;
  managementBusy.value = 'close';
  managementError.value = '';
  try {
    await api.DeleteRoom(activeRoomId.value);
    if (version !== pageVersion) return;
    confirmClose.value = false;
    await leave();
    $q.notify({ type: 'positive', message: t('broadcast.closedSuccess') });
    await router.push('/wideochat');
  } catch (reason) {
    if (version === pageVersion) {
      managementError.value = message(reason);
      confirmClose.value = false;
    }
  } finally { if (version === pageVersion) managementBusy.value = ''; }
}

let pageVersion = 0;
watch([roomName, () => profiles.profile?.user_id], async () => {
  const version = ++pageVersion;
  ++participantRequest;
  loadingParticipants.value = false;
  opening.value = true;
  clearInterval(refreshTimer);
  await leave();
  if (version !== pageVersion) return;
  details.value = undefined;
  participants.value = [];
  confirmClose.value = false;
  managementError.value = '';
  managementBusy.value = '';
  error.value = '';
  ended.value = false;
  if (!profiles.isAuthenticated) { opening.value = false; return; }
  try {
    const opened = await api.OpenRoom(roomName.value);
    if (version !== pageVersion) return;
    details.value = opened;
    refreshTimer = setInterval(() => { if (canManage.value) void loadParticipants(); }, 10000);
    if (canManage.value) void loadParticipants();
  } catch (reason) { if (version === pageVersion) error.value = message(reason); }
  finally { if (version === pageVersion) opening.value = false; }
}, { immediate: true });

watch(canManage, allowed => { if (allowed) void loadParticipants(); });

onBeforeUnmount(() => {
  ++pageVersion;
  ++participantRequest;
  clearInterval(refreshTimer);
});
</script>

<style scoped>
.broadcast-splitter {
  display: flex;
  flex-direction: row;
  flex: 1 1 0;
  min-height: calc(100vh - 160px);
  height: 100%;
  width: 100%;
}

@media (max-width: 1023px) {
  .broadcast-splitter {
    min-height: 850px;
    height: auto;
  }
}

.broadcast-splitter-separator {
  background-color: #e0e0e0;
  transition: background-color 0.2s ease;
}

.broadcast-splitter-separator:hover {
  background-color: var(--q-primary);
}

:deep(.q-splitter__panel.q-splitter__after) {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-height: 0;
  height: 100%;
  height: -webkit-fill-available;
  height: stretch;
}

</style>
