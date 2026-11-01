<template>
  <q-page class="live-page">
    <header class="live-header">
      <router-link :to="profileRoute" class="live-back">‹ {{ t('live.backToProfile') }}</router-link>
      <h1>
        {{ status?.display_name || t('live.liveNow') }}
        <span v-if="live.videos.value.length" class="live-badge">● {{ t('live.liveNow') }}</span>
      </h1>
    </header>

    <p v-if="loading" class="live-status">{{ t('broadcaster.loading') }}</p>

    <div v-else-if="notFound" class="live-status">
      <p>{{ t('broadcaster.notFound') }}</p>
      <router-link :to="homeRoute" class="live-link">{{ t('broadcaster.backHome') }}</router-link>
    </div>

    <div v-else class="live-layout">
      <section class="live-stage" :aria-label="t('live.liveNow')">
        <div v-if="live.reconnecting.value" class="live-banner">{{ t('live.reconnecting') }}</div>
        <div v-if="live.needsAudio.value" class="live-banner">
          {{ t('live.audioHint') }}
          <button type="button" class="live-button" @click="live.startAudio">{{ t('live.enableAudio') }}</button>
        </div>
        <div v-if="loadError || live.error.value" class="live-banner live-banner--error" role="alert">
          {{ t('live.error') }}
          <button type="button" class="live-button" @click="refresh">{{ t('live.retry') }}</button>
        </div>

        <LiveStage :videos="live.videos.value" :audios="live.audios.value" :resizable="false" :captions="false">
          <template #placeholder>
            <q-icon name="live_tv" size="56px" />
            <p v-if="live.connecting.value">{{ t('live.connecting') }}</p>
            <p v-else-if="live.connected.value">{{ t('webchat.waiting') }}</p>
            <p v-else-if="live.ended.value">{{ t('live.ended') }}</p>
            <p v-else>{{ t('live.offline', { name: status?.display_name || '' }) }}</p>
          </template>
        </LiveStage>
      </section>

      <!-- Chat: chi ha fatto login scrive, i visitatori leggono soltanto -->
      <aside class="live-chat">
        <ChatPanel
          :messages="chat.messages.value"
          :self-id="chatUser?.id"
          :loading="chat.loading.value"
          :sending="chat.sending.value"
          :can-write="!!chatUser"
          :disabled="!live.connected.value"
          :send="chat.send"
        >
          <template #readonly>
            <router-link :to="loginRoute" class="live-link">{{ t('live.loginToChat') }}</router-link>
          </template>
        </ChatPanel>
      </aside>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ChatPanel, LiveStage, useLiveRoom, useRoomChat, type ChatUser } from '@extra/webchat';
import { APIError, ErrCode, type profiles, type videochat } from '@/encore/client';
import { api } from '@/encore/api';

const { t } = useI18n();
const route = useRoute();

const shortId = computed(() => String(route.params.shortId || '').toLowerCase());
const localePrefix = computed(() => (route.params.locale ? `/${String(route.params.locale)}` : ''));
const homeRoute = computed(() => localePrefix.value || '/');
const profileRoute = computed(() => `${localePrefix.value}/br/${shortId.value}`);
const loginRoute = computed(() => `${localePrefix.value}/login?redirect=${encodeURIComponent(route.fullPath)}`);

const status = ref<videochat.LiveStatus | null>(null);
const me = ref<profiles.Profile | null>(null);
const loading = ref(true);
const notFound = ref(false);
const loadError = ref(false);

useMeta(() => ({
  title: `${status.value?.display_name ? t('live.title', { name: status.value.display_name }) : t('live.liveNow')} - PublicSite`,
}));

const live = useLiveRoom();

const chatUser = computed<ChatUser | null>(() =>
  me.value ? { id: me.value.user_id, name: me.value.display_name || t('webchat.user'), avatar: me.value.avatar_url || undefined } : null,
);

// La chat della diretta è quella della stanza (chiave: SID). I visitatori la
// leggono dall'endpoint pubblico, chi ha fatto login da quello della stanza.
const chat = useRoomChat({
  room: live.room,
  chatKey: computed(() => (status.value?.live ? status.value.room_sid : '')),
  source: {
    history: async (key) =>
      me.value
        ? (await api.videochat.ListMessages(key, { limit: 50 })).messages
        : (await api.videochat.ListLiveMessages(shortId.value, { limit: 50 })).messages,
    send: (key, params) => api.videochat.SendMessage(key, { ...params, kind: 'text' }),
  },
  ownerId: live.ownerIdentity,
  user: chatUser,
  fallbackName: () => t('webchat.user'),
});

async function currentUser(): Promise<profiles.Profile | null> {
  try {
    return (await api.profiles.Me()) || null;
  } catch {
    return null;
  }
}

// Visitatori: token ospite (solo visione, nascosto agli altri); con login:
// token da spettatore con la propria identità, per poter scrivere in chat.
async function joinLive() {
  const current = status.value;
  if (!current?.live || live.connected.value || live.connecting.value) return;
  await live.join(() =>
    me.value ? api.videochat.CreateToken({ room_name: current.room_name }) : api.videochat.CreateGuestToken(shortId.value),
  );
}

let loadRequest = 0;

async function load() {
  const request = ++loadRequest;
  loading.value = true;
  notFound.value = false;
  loadError.value = false;
  await live.leave();
  live.ended.value = false;
  try {
    const [liveStatus, user] = await Promise.all([api.videochat.GetLive(shortId.value), currentUser()]);
    if (request !== loadRequest) return;
    status.value = liveStatus;
    me.value = user;
  } catch (err) {
    if (request !== loadRequest) return;
    status.value = null;
    if (err instanceof APIError && err.code === ErrCode.NotFound) notFound.value = true;
    else loadError.value = true;
  } finally {
    if (request === loadRequest) loading.value = false;
  }
  if (request === loadRequest) await joinLive();
}

// Fuori onda: controlla periodicamente se la diretta è iniziata.
async function poll() {
  if (loading.value || notFound.value || live.connected.value || live.connecting.value) return;
  try {
    const liveStatus = await api.videochat.GetLive(shortId.value);
    if (liveStatus.live && !status.value?.live) live.ended.value = false;
    status.value = liveStatus;
    await joinLive();
  } catch {
    // nuovo tentativo al prossimo giro
  }
}

function refresh() {
  void load();
}

let pollTimer: ReturnType<typeof setInterval> | undefined;

onMounted(() => {
  void load();
  pollTimer = setInterval(() => void poll(), 20000);
});

watch(shortId, () => void load());

// Fine della diretta: lo stato torna "fuori onda" finché non riparte.
watch(live.ended, (ended) => {
  if (ended && status.value) status.value = { ...status.value, live: false };
});

onBeforeUnmount(() => {
  ++loadRequest;
  clearInterval(pollTimer);
});
</script>

<style scoped>
.live-page {
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

.live-header {
  margin-bottom: 16px;
}

.live-back,
.live-link {
  color: #f4cf00;
  text-decoration: none;
}

.live-back:hover,
.live-link:hover {
  text-decoration: underline;
}

.live-header h1 {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin: 8px 0 0;
  font-size: 32px;
  font-weight: 400;
  line-height: 1.2;
  letter-spacing: 0;
}

.live-badge {
  padding: 2px 10px;
  border-radius: 4px;
  background: #d32f2f;
  color: #ffffff;
  font-size: 14px;
  font-weight: 700;
  text-transform: uppercase;
}

.live-status {
  padding: 64px 16px;
  text-align: center;
}

.live-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 16px;
  align-items: stretch;
}

.live-stage {
  min-width: 0;
}

.live-chat {
  height: min(70vh, 640px);
  min-height: 360px;
  overflow: hidden;
  border: 1px solid #252525;
  border-radius: 4px;
}

.live-banner {
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

.live-banner--error {
  background: #3a1414;
  color: #ffcdd2;
}

.live-button {
  padding: 4px 12px;
  border: 1px solid currentColor;
  border-radius: 4px;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.live-button:focus-visible {
  outline: 2px solid #bbbbbb;
  outline-offset: 2px;
}

@media (max-width: 900px) {
  .live-layout {
    grid-template-columns: 1fr;
  }

  .live-chat {
    height: 420px;
  }

  .live-header h1 {
    font-size: 26px;
  }
}
</style>
