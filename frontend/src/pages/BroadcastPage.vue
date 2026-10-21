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
            <div class="broadcast-stage-left q-pa-sm">
              <div v-if="videos.length" class="row q-col-gutter-md">
                <div v-for="item in videos" :key="item.sid" :class="videos.length > 1 ? 'col-12 col-md-6' : 'col-12'">
                  <BroadcastTrack :track="item.track" :local="item.local" />
                  <div class="text-caption q-mt-xs">{{ t(item.screen ? 'broadcast.screen' : 'broadcast.camera') }}</div>
                </div>
              </div>
              <div v-else class="broadcast-placeholder flex flex-center bg-grey-10 text-white rounded-borders">
                <div class="text-center q-pa-lg"><q-icon name="live_tv" size="56px" /><p>{{ t('broadcast.waiting') }}</p></div>
              </div>
              <BroadcastTrack v-for="item in audios" :key="item.sid" :track="item.track" audio />
            </div>
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
              :room="live"
              :connected="connected"
              :is-owner="isOwner"
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
import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
  Room as LiveRoom,
  RoomEvent,
  Track,
  type VideoCaptureOptions,
  type TrackPublishOptions,
  type VideoCodec,
} from 'livekit-client';
import BroadcastTrack from '@/components/videochat/BroadcastTrack.vue';
import BroadcastStageRight from '@/components/videochat/BroadcastStageRight.vue';
import WebcamSettingsDialog, { type WebcamSettings } from '@/components/videochat/WebcamSettingsDialog.vue';
import { createVideochatClient, type Room, type Participant } from '@/encore/videochat';
import { useProfilesStore } from '@/stores/profiles-store';
import { useChatStore } from '@/stores/chat-store';

const profiles = useProfilesStore();
const chatStore = useChatStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const $q = useQuasar();
const api = createVideochatClient(() => profiles.token);
const roomName = computed(() => String(route.params.roomName || ''));
const details = ref<Room>();
const opening = ref(false);
const activeRoomName = computed(() => details.value?.name || roomName.value);
const activeRoomId = computed(() => details.value?.sid || live.value?.name || details.value?.name || roomName.value);
const live = shallowRef<LiveRoom>();
const connected = ref(false);
const connecting = ref(false);
const reconnecting = ref(false);
const ended = ref(false);
const isOwner = ref(false);
const ownerIdentity = ref('');
const canManage = computed(() => !!details.value && details.value.owner_id === profiles.profile?.user_id);
const error = ref('');
const cameraOn = ref(false);
const micOn = ref(false);
const screenOn = ref(false);
const mediaBusy = ref(false);
const needsAudio = ref(false);
const participants = ref<Participant[]>([]);
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
  const room = live.value;
  if (!room || !connected.value) return [];
  const list: { sid: string; identity: string; name: string; avatar_url?: string | undefined }[] = [];
  if (room.localParticipant && room.localParticipant.identity !== ownerId) {
    list.push({
      sid: room.localParticipant.sid,
      identity: room.localParticipant.identity,
      name: profiles.profile?.display_name || room.localParticipant.name || room.localParticipant.identity,
      avatar_url: profiles.profile?.avatar_url,
    });
  }
  room.remoteParticipants.forEach((p) => {
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

const defaultWebcamSettings: WebcamSettings = {
  preset: 'balanced',
  deviceId: '',
  width: 1280,
  height: 720,
  aspectRatio: '16:9',
  frameRate: 30,
  bitrateKbps: 1700,
  codec: 'vp8',
  simulcast: true,
  degradationPreference: 'maintain-framerate',
};

const webcamSettings = ref<WebcamSettings>(loadInitialWebcamSettings());

function loadInitialWebcamSettings(): WebcamSettings {
  try {
    const saved = localStorage.getItem('videochat_webcam_settings');
    if (saved) return { ...defaultWebcamSettings, ...JSON.parse(saved) };
  } catch {}
  return { ...defaultWebcamSettings };
}

function buildWebcamOptions(settings: WebcamSettings) {
  const aspectRatios: Record<string, number> = {
    '16:9': 16 / 9,
    '4:3': 4 / 3,
    '1:1': 1,
    '9:16': 9 / 16,
  };

  const captureOptions: VideoCaptureOptions = {
    ...(settings.deviceId ? { deviceId: { exact: settings.deviceId } } : {}),
    resolution: {
      width: settings.width || 1280,
      height: settings.height || 720,
      frameRate: settings.frameRate || 30,
      aspectRatio: aspectRatios[settings.aspectRatio] ?? 16 / 9,
    },
    frameRate: settings.frameRate || 30,
  };

  const publishOptions: TrackPublishOptions = {
    videoEncoding: {
      maxBitrate: (settings.bitrateKbps || 1700) * 1000,
      maxFramerate: settings.frameRate || 30,
    },
    videoCodec: (settings.codec || 'vp8') as VideoCodec,
    simulcast: settings.simulcast !== false,
    degradationPreference: (settings.degradationPreference || 'maintain-framerate') as RTCDegradationPreference,
  };

  return { captureOptions, publishOptions };
}

async function onApplyWebcamSettings(newSettings: WebcamSettings) {
  webcamSettings.value = { ...newSettings };
  try {
    localStorage.setItem('videochat_webcam_settings', JSON.stringify(newSettings));
  } catch (e) {
    console.error(e);
  }

  const room = live.value;
  if (cameraOn.value && room && connected.value && isOwner.value) {
    mediaBusy.value = true;
    error.value = '';
    try {
      const camPub = room.localParticipant.getTrackPublication(Track.Source.Camera);
      if (camPub?.track) {
        await room.localParticipant.unpublishTrack(camPub.track, true);
      }
      const { captureOptions, publishOptions } = buildWebcamOptions(newSettings);
      await room.localParticipant.setCameraEnabled(true, captureOptions, publishOptions);
      syncMedia();
      $q.notify({ type: 'positive', message: t('webcam.savedSuccess') });
    } catch (reason) {
      error.value = message(reason);
    } finally {
      mediaBusy.value = false;
    }
  } else {
    $q.notify({ type: 'positive', message: t('webcam.savedSuccess') });
  }
}

type Media = { sid: string; track: Track; local: boolean; screen: boolean };
const videos = shallowRef<Media[]>([]);
const audios = shallowRef<Media[]>([]);
let generation = 0;
let participantRequest = 0;
let refreshTimer: ReturnType<typeof setInterval> | undefined;

function message(reason: unknown) { return reason instanceof Error ? reason.message : String(reason); }

function syncMedia() {
  const room = live.value;
  if (!room) return;
  const video: Media[] = [];
  const audio: Media[] = [];
  const owner = isOwner.value ? room.localParticipant : room.remoteParticipants.get(ownerIdentity.value);
  owner?.trackPublications.forEach(publication => {
    if (!publication.track || publication.isMuted) return;
    const item = { sid: publication.trackSid, track: publication.track, local: isOwner.value,
      screen: publication.source === Track.Source.ScreenShare };
    if (publication.kind === Track.Kind.Video) video.push(item);
    if (publication.kind === Track.Kind.Audio && !isOwner.value) audio.push(item);
  });
  videos.value = video.sort((a, b) => Number(b.screen) - Number(a.screen));
  audios.value = audio;
  cameraOn.value = room.localParticipant.isCameraEnabled;
  micOn.value = room.localParticipant.isMicrophoneEnabled;
  screenOn.value = room.localParticipant.isScreenShareEnabled;
  needsAudio.value = !room.canPlaybackAudio;
}

function logMediaEvent(eventText: 'VIDEO_ON' | 'VIDEO_OFF' | 'MIC_ON' | 'MIC_OFF') {
  if (!isOwner.value) return;
  const senderId = live.value?.localParticipant?.identity || profiles.profile?.user_id || 'broadcaster';
  const senderName = profiles.profile?.display_name || live.value?.localParticipant?.name || 'Broadcaster';
  const targetRoom = activeRoomId.value || roomName.value;
  if (!targetRoom) return;
  void chatStore.logMediaEvent(targetRoom, eventText, senderId, senderName);
}

async function leave() {
  if (isOwner.value) {
    if (cameraOn.value) {
      logMediaEvent('VIDEO_OFF');
    }
    if (micOn.value) {
      logMediaEvent('MIC_OFF');
    }
  }
  ++generation;
  const room = live.value;
  live.value = undefined;
  room?.removeAllListeners();
  connected.value = false;
  connecting.value = false;
  reconnecting.value = false;
  mediaBusy.value = false;
  isOwner.value = false;
  cameraOn.value = false;
  micOn.value = false;
  screenOn.value = false;
  videos.value = [];
  audios.value = [];
  needsAudio.value = false;
  await room?.disconnect();
}

async function join() {
  if (opening.value || connecting.value || connected.value || !profiles.isAuthenticated) return;
  const version = ++generation;
  connecting.value = true;
  error.value = '';
  ended.value = false;
  let room: LiveRoom | undefined;
  try {
    const opened = await api.openRoom(roomName.value);
    if (version !== generation) return;
    details.value = opened;
    const token = await api.createToken(opened.name);
    if (version !== generation) return;
    isOwner.value = token.role === 'owner';
    ownerIdentity.value = token.owner_id;
    room = new LiveRoom({ adaptiveStream: true, dynacast: true });
    live.value = room;
    room.on(RoomEvent.TrackSubscribed, syncMedia)
      .on(RoomEvent.TrackUnsubscribed, syncMedia)
      .on(RoomEvent.LocalTrackPublished, syncMedia)
      .on(RoomEvent.LocalTrackUnpublished, syncMedia)
      .on(RoomEvent.TrackMuted, syncMedia)
      .on(RoomEvent.TrackUnmuted, syncMedia)
      .on(RoomEvent.AudioPlaybackStatusChanged, syncMedia)
      .on(RoomEvent.ParticipantConnected, syncMedia)
      .on(RoomEvent.ParticipantDisconnected, syncMedia)
      .on(RoomEvent.Reconnecting, () => { reconnecting.value = true; })
      .on(RoomEvent.Reconnected, () => { reconnecting.value = false; syncMedia(); })
      .on(RoomEvent.Disconnected, () => { ended.value = true; void leave(); })
      .on(RoomEvent.MediaDevicesError, reason => { error.value = message(reason); });
    await room.connect(token.server_url, token.participant_token);
    if (version !== generation) { await room.disconnect(); return; }
    connected.value = true;
    // Publishing always requires an explicit owner action; viewers never access devices.
    syncMedia();
    if (canManage.value) void loadParticipants();
  } catch (reason) {
    if (version === generation) { error.value = message(reason); await leave(); }
    else { await room?.disconnect(); }
  } finally {
    if (version === generation) connecting.value = false;
  }
}

async function toggleMedia(kind: 'camera' | 'microphone' | 'screen') {
  const room = live.value;
  if (!room || !connected.value || !isOwner.value || mediaBusy.value) return;
  const version = generation;
  mediaBusy.value = true;
  error.value = '';
  try {
    if (kind === 'camera') {
      if (!cameraOn.value) {
        const { captureOptions, publishOptions } = buildWebcamOptions(webcamSettings.value);
        await room.localParticipant.setCameraEnabled(true, captureOptions, publishOptions);
        logMediaEvent('VIDEO_ON');
      } else {
        await room.localParticipant.setCameraEnabled(false);
        logMediaEvent('VIDEO_OFF');
      }
    }
    if (kind === 'microphone') {
      if (!micOn.value) {
        await room.localParticipant.setMicrophoneEnabled(true);
        logMediaEvent('MIC_ON');
      } else {
        await room.localParticipant.setMicrophoneEnabled(false);
        logMediaEvent('MIC_OFF');
      }
    }
    if (kind === 'screen') await room.localParticipant.setScreenShareEnabled(!screenOn.value, { audio: true });
    if (version !== generation) { await room.disconnect(); return; }
    syncMedia();
  } catch (reason) {
    if (version === generation) error.value = message(reason);
  } finally {
    if (version === generation) mediaBusy.value = false;
  }
}

async function enableAudio() {
  try { await live.value?.startAudio(); syncMedia(); }
  catch (reason) { error.value = message(reason); }
}

async function loadParticipants() {
  if (!canManage.value || loadingParticipants.value) return;
  const request = ++participantRequest;
  loadingParticipants.value = true;
  managementError.value = '';
  try {
    const response = await api.listParticipants(activeRoomId.value);
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
    await api.removeParticipant(activeRoomId.value, identity);
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
    await api.deleteRoom(activeRoomId.value);
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
watch([roomName, () => profiles.token], async () => {
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
    const opened = await api.openRoom(roomName.value);
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
  void leave();
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

.broadcast-stage-left {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  box-sizing: border-box;
}

.broadcast-placeholder {
  min-height: 320px;
}
</style>
