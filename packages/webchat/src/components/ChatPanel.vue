<template>
  <div class="webchat-chat">
    <!-- Intestazione: titolo e azioni dell'app (alert, pulizia…) -->
    <div class="webchat-chat__header">
      <span class="webchat-chat__title">{{ t('webchat.chatTitle') }}</span>
      <div class="row items-center q-gutter-xs">
        <slot name="actions" />
      </div>
    </div>

    <div ref="list" class="webchat-chat__messages" aria-live="polite" @scroll="onScroll">
      <div v-if="loading && !messages.length" class="row justify-center q-pa-md">
        <q-spinner size="24px" />
      </div>

      <div v-else-if="!messages.length" class="webchat-chat__empty">
        <q-icon name="chat_bubble_outline" size="44px" />
        <p>{{ t('webchat.chatEmpty') }}</p>
      </div>

      <div v-else class="q-gutter-y-sm">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="webchat-msg"
          :class="{ 'webchat-msg--self': isSelf(msg), 'webchat-msg--owner': !isSelf(msg) && msg.isOwner }"
        >
          <div class="webchat-msg__meta">
            <q-avatar size="20px" class="webchat-msg__avatar">
              <img v-if="msg.senderAvatar" :src="msg.senderAvatar" alt="" />
              <span v-else>{{ (msg.senderName || '?').charAt(0).toUpperCase() }}</span>
            </q-avatar>
            <span class="webchat-msg__name">{{ msg.senderName }}</span>
            <span class="webchat-msg__time">{{ formatTime(msg.timestamp) }}</span>
          </div>
          <div class="webchat-msg__bubble">{{ msg.text }}</div>
        </div>
      </div>
    </div>

    <transition name="fade">
      <q-btn
        v-if="showScrollBottom"
        round
        dense
        size="sm"
        icon="keyboard_arrow_down"
        class="webchat-chat__scroll"
        :aria-label="t('webchat.scrollBottom')"
        @click="scrollToBottom(true)"
      />
    </transition>

    <div class="webchat-chat__footer">
      <form v-if="canWrite" class="webchat-chat__form" @submit.prevent="submit">
        <input
          v-model="text"
          class="webchat-chat__input"
          type="text"
          maxlength="500"
          autocomplete="off"
          :placeholder="t('webchat.chatPlaceholder')"
          :aria-label="t('webchat.chatPlaceholder')"
          :disabled="disabled || sending"
        />
        <q-btn
          type="submit"
          round
          dense
          flat
          icon="send"
          class="webchat-chat__send"
          :loading="sending"
          :disable="disabled || !text.trim()"
          :aria-label="t('webchat.chatSend')"
        />
      </form>
      <!-- Chat in sola lettura (visitatori senza login) -->
      <div v-else class="webchat-chat__readonly">
        <slot name="readonly">{{ t('webchat.readOnly') }}</slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ChatItem } from '../types';

const props = withDefaults(
  defineProps<{
    messages: ChatItem[];
    /** Id di chi guarda, per allineare a destra i propri messaggi. */
    selfId?: string | undefined;
    loading?: boolean;
    sending?: boolean;
    canWrite?: boolean;
    disabled?: boolean;
    /** Invia il testo; false se l'invio non è riuscito e il testo va mantenuto. */
    send?: ((text: string) => Promise<boolean>) | undefined;
  }>(),
  { selfId: '', loading: false, sending: false, canWrite: false, disabled: false, send: undefined },
);

const { t } = useI18n();
const text = ref('');
const list = ref<HTMLDivElement>();
const showScrollBottom = ref(false);

function isSelf(msg: ChatItem): boolean {
  return Boolean(props.selfId) && msg.senderId === props.selfId;
}

function formatTime(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function scrollToBottom(smooth = false) {
  void nextTick(() => {
    if (!list.value) return;
    list.value.scrollTo({ top: list.value.scrollHeight, behavior: smooth ? 'smooth' : 'auto' });
    showScrollBottom.value = false;
  });
}

function onScroll() {
  if (!list.value) return;
  const { scrollTop, scrollHeight, clientHeight } = list.value;
  showScrollBottom.value = scrollHeight - (scrollTop + clientHeight) > 100;
}

async function submit() {
  const value = text.value.trim();
  if (!value || !props.send || props.sending) return;
  text.value = '';
  scrollToBottom(true);
  if (!(await props.send(value))) text.value = value;
}

// Nuovi messaggi: segue la chat solo se si è già in fondo.
watch(
  () => props.messages.length,
  () => {
    if (!showScrollBottom.value) scrollToBottom(true);
  },
);
</script>

<style scoped>
.webchat-chat {
  --wc-surface: var(--webchat-surface, #ffffff);
  --wc-header: var(--webchat-header, #fafafa);
  --wc-text: var(--webchat-text, #1e293b);
  --wc-muted: var(--webchat-muted, #9e9e9e);
  --wc-border: var(--webchat-border, #e2e8f0);
  --wc-bubble: var(--webchat-bubble, #f1f5f9);
  --wc-bubble-text: var(--webchat-bubble-text, #1e293b);
  --wc-self: var(--webchat-self, #0284c7);
  --wc-self-text: var(--webchat-self-text, #ffffff);
  --wc-owner: var(--webchat-owner, #fef3c7);
  --wc-owner-text: var(--webchat-owner-text, #1e293b);
  --wc-accent: var(--webchat-accent, #f59e0b);
  --wc-input: var(--webchat-input, #ffffff);
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  background: var(--wc-surface);
  color: var(--wc-text);
}

.webchat-chat__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 36px;
  padding: 2px 16px;
  border-bottom: 1px solid var(--wc-border);
  background: var(--wc-header);
}

.webchat-chat__title {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--wc-muted);
}

.webchat-chat__messages {
  flex: 1 1 0;
  min-height: 0;
  padding: 8px;
  overflow-x: hidden;
  overflow-y: auto;
}

.webchat-chat__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 16px;
  color: var(--wc-muted);
  text-align: center;
  font-size: 0.75rem;
}

.webchat-chat__empty p {
  margin: 8px 0 0;
}

.webchat-msg {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.webchat-msg--self {
  align-items: flex-end;
}

.webchat-msg__meta {
  display: flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
  margin-bottom: 4px;
  font-size: 11px;
}

.webchat-msg--self .webchat-msg__meta {
  flex-direction: row-reverse;
}

.webchat-msg__avatar {
  background: var(--wc-self);
  color: var(--wc-self-text);
  font-weight: 700;
}

.webchat-msg--owner .webchat-msg__avatar {
  background: var(--wc-accent);
}

.webchat-msg__name {
  max-width: 140px;
  overflow: hidden;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.webchat-msg--owner .webchat-msg__name {
  color: var(--wc-accent);
}

.webchat-msg__time {
  color: var(--wc-muted);
  font-size: 10px;
}

.webchat-msg__bubble {
  display: inline-block;
  max-width: 85%;
  padding: 8px 12px;
  border-radius: 12px;
  background: var(--wc-bubble);
  color: var(--wc-bubble-text);
  font-size: 13px;
  line-height: 1.4;
  word-break: break-word;
  white-space: pre-line;
}

.webchat-msg--self .webchat-msg__bubble {
  border-bottom-right-radius: 2px;
  background: var(--wc-self);
  color: var(--wc-self-text);
}

.webchat-msg--owner .webchat-msg__bubble {
  border-left: 3px solid var(--wc-accent);
  background: var(--wc-owner);
  color: var(--wc-owner-text);
}

.webchat-chat__scroll {
  position: absolute;
  right: 16px;
  bottom: 64px;
  z-index: 10;
  background: var(--wc-self);
  color: var(--wc-self-text);
}

.webchat-chat__footer {
  padding: 8px;
  border-top: 1px solid var(--wc-border);
}

.webchat-chat__form {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 4px 0 12px;
  border: 1px solid var(--wc-border);
  border-radius: 4px;
  background: var(--wc-input);
}

.webchat-chat__form:focus-within {
  border-color: var(--wc-self);
}

.webchat-chat__input {
  flex: 1 1 auto;
  min-width: 0;
  height: 38px;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--wc-text);
  font: inherit;
  font-size: 14px;
}

.webchat-chat__input::placeholder {
  color: var(--wc-muted);
}

.webchat-chat__send {
  color: var(--wc-self);
}

.webchat-chat__readonly {
  padding: 8px;
  color: var(--wc-muted);
  font-size: 13px;
  text-align: center;
}
</style>
