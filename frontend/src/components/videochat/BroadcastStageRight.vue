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
          <q-badge v-if="chat.messages.value.length" color="primary" floating rounded>
            {{ chat.messages.value.length }}
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
          <!-- Chat condivisa con il sito pubblico (packages/webchat); i messaggi passano dal server -->
          <ChatPanel
            :messages="chat.messages.value"
            :self-id="chatUser?.id"
            :loading="chat.loading.value"
            :sending="chat.sending.value"
            :can-write="true"
            :disabled="!connected"
            :send="chat.send"
          >
            <template #actions>
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
                v-if="chat.messages.value.length"
                flat
                round
                dense
                size="sm"
                icon="delete_sweep"
                color="grey-7"
                :aria-label="t('broadcast.chatClear')"
                @click="chat.clear"
              >
                <q-tooltip>{{ t('broadcast.chatClear') }}</q-tooltip>
              </q-btn>
            </template>
          </ChatPanel>
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
import { computed, toRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import type { Room as LiveRoom } from 'livekit-client';
import { ChatPanel, useRoomChat, type ChatUser } from '@extra/webchat';
import { api } from '@/encore/api';
import { useProfilesStore } from '@/stores/profiles-store';

export interface ParticipantItem {
  sid?: string | undefined;
  identity: string;
  name?: string | undefined;
  avatar_url?: string | undefined;
}

const props = withDefaults(
  defineProps<{
    room?: LiveRoom | null | undefined;
    connected: boolean;
    isOwner: boolean;
    ownerId?: string | undefined;
    canManage?: boolean | undefined;
    roomName: string;
    participants?: ParticipantItem[] | undefined;
    managementBusy?: string | undefined;
    managementError?: string | undefined;
    modelValue?: 'chat' | 'participants' | undefined;
  }>(),
  {
    room: undefined,
    ownerId: '',
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

const currentTab = computed({
  get: () => props.modelValue,
  set: (val: 'chat' | 'participants') => emit('update:modelValue', val),
});

const isAdminOrAdvisor = computed(() => {
  const role = profiles.profile?.role?.toLowerCase();
  return role === 'admin' || role === 'chatadvisor';
});

const chatUser = computed<ChatUser | null>(() => {
  const profile = profiles.profile;
  if (!profile) return null;
  return {
    id: props.room?.localParticipant?.identity || profile.user_id,
    name: profile.display_name || props.room?.localParticipant?.name || t('webchat.user'),
    avatar: profile.avatar_url || undefined,
  };
});

// Il server valida, salva e inoltra i messaggi a tutti i partecipanti con
// l'identità del mittente presa dal DB.
const chat = useRoomChat({
  room: computed(() => props.room ?? undefined),
  chatKey: toRef(props, 'roomName'),
  source: {
    history: async (key) => (await api.videochat.ListMessages(key, { limit: 50 })).messages,
    send: (key, params) => api.videochat.SendMessage(key, { ...params, kind: 'text' }),
  },
  ownerId: computed(() => props.ownerId ?? ''),
  user: chatUser,
  fallbackName: () => t('webchat.user'),
  onAlert: (body) => {
    $q.notify({
      type: 'negative',
      message: t('admin.notifications.adminAlert'),
      caption: body,
      timeout: 0,
      actions: [{ label: t('close'), handler: () => {} }],
    });
  },
  onError: (message) => {
    $q.notify({ type: 'negative', message });
  },
});

/**
 * Invia un alert amministrativo: il server verifica i permessi e lo recapita
 * al proprietario della stanza.
 */
function openAdminAlertDialog() {
  const msgDialog = $q.dialog({
    title: t('admin.sendAlert'),
    prompt: { model: '', label: t('admin.message') },
    cancel: true,
    persistent: true,
  });

  msgDialog.onOk(async (alertText: string) => {
    if (!alertText.trim() || !props.roomName) return;
    try {
      await api.videochat.SendCommand(props.roomName, {
        command: 'alert',
        target_id: '',
        payload: alertText.trim(),
      });
      $q.notify({ type: 'positive', message: 'Alert inviato con successo' });
    } catch (err) {
      console.warn('Errore invio alert amministratore:', err);
      $q.notify({ type: 'negative', message: err instanceof Error ? err.message : String(err) });
    }
  });
}
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

.participants-list-container {
  overflow-y: auto;
  min-height: 0;
}

.border-top {
  border-top: 1px solid #e2e8f0;
}
</style>
