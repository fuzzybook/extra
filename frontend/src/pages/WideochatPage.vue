<template>
  <q-page padding>
    <div class="row items-center justify-between q-gutter-sm q-mb-lg">
      <h1 class="text-h4 q-my-none">{{ t('videochat.title') }}</h1>
      <q-btn color="primary" icon="add" :label="t('videochat.create')" :disable="!isAdmin" @click="openCreate">
        <q-tooltip v-if="!isAdmin">{{ t('videochat.adminRequired') }}</q-tooltip>
      </q-btn>
    </div>

    <q-banner v-if="!profiles.isAuthenticated" class="bg-grey-2 q-mb-md">
      {{ t('broadcast.loginRequired') }}
      <template #action>
        <q-btn flat to="/login" :label="t('nav.login')" />
      </template>
    </q-banner>

    <template v-else>
      <q-banner v-if="listError" role="alert" class="bg-red-1 text-negative q-mb-md">
        {{ t('videochat.loadError') }} {{ listError }}
      </q-banner>
      <q-table
        :title="t('videochat.rooms')"
        :rows="rooms"
        :columns="columns"
        :row-key="roomRowKey"
        :loading="loading"
        :no-data-label="t('videochat.empty')"
        :rows-per-page-options="[10, 25, 50]"
        v-model:pagination="pagination"
        wrap-cells
      >
        <template #body-cell-status="props">
          <q-td :props="props" align="center">
            <q-badge
              v-if="props.row.active"
              color="positive"
              class="q-px-sm q-py-xs text-weight-bold shadow-1"
            >
              <q-icon name="sensors" size="14px" class="q-mr-xs pulse-dot" />
              {{ t('videochat.active') }}
            </q-badge>
            <q-badge
              v-else
              color="grey-6"
              outline
              class="q-px-sm q-py-xs text-weight-medium"
            >
              <q-icon name="radio_button_unchecked" size="10px" class="q-mr-xs" />
              {{ t('videochat.inactive') }}
            </q-badge>
          </q-td>
        </template>

        <template #body-cell-name="props">
          <q-td :props="props">
            <div class="row items-center no-wrap">
              <q-avatar size="36px" class="q-mr-sm shadow-1">
                <img v-if="props.row.avatar_url" :src="props.row.avatar_url" alt="" />
                <q-icon v-else name="account_circle" size="36px" color="grey-6" />
              </q-avatar>
              <div>
                <div class="text-subtitle2 text-weight-bold">
                  {{ props.row.display_name || props.row.name }}
                </div>
                <div v-if="props.row.display_name && props.row.display_name !== props.row.name" class="text-caption text-grey-7">
                  {{ props.row.name }}
                </div>
              </div>
            </div>
          </q-td>
        </template>

        <template #body-cell-participants="props">
          <q-td :props="props">
            <q-badge v-if="props.row.active" color="primary" class="q-px-sm q-py-xs">
              <q-icon name="group" size="14px" class="q-mr-xs" />
              {{ props.row.num_participants }}
            </q-badge>
            <span v-else class="text-grey-6">—</span>
          </q-td>
        </template>

        <template #body-cell-maximum="props">
          <q-td :props="props">
            {{ props.row.max_participants ? props.row.max_participants : '—' }}
          </q-td>
        </template>

        <template #body-cell-created="props">
          <q-td :props="props">
            {{ props.row.creation_time ? new Date(props.row.creation_time * 1000).toLocaleString(locale) : '—' }}
          </q-td>
        </template>

        <template #body-cell-actions="props">
          <q-td :props="props">
            <q-btn
              v-if="props.row.owner_id === profiles.profile?.user_id"
              color="primary"
              flat
              icon="settings"
              :label="t('broadcast.manage')"
              :to="{ name: 'broadcast', params: { roomName: props.row.name } }"
            />
            <q-btn
              v-else
              :color="props.row.active ? 'positive' : 'primary'"
              :flat="!props.row.active"
              :unelevated="props.row.active"
              :icon="props.row.active ? 'live_tv' : 'tv'"
              :label="t('broadcast.watch')"
              :to="{ name: 'broadcast', params: { roomName: props.row.name } }"
            />
          </q-td>
        </template>

        <template #top-right>
          <q-btn flat color="primary" icon="refresh" :label="t('videochat.refresh')" :disable="loading || creating" @click="loadRooms" />
        </template>
      </q-table>
    </template>

    <q-dialog v-model="showCreate" :persistent="creating">
      <q-card style="width: 440px; max-width: 100%">
        <q-form @submit="createRoom">
          <q-card-section class="text-h6">{{ t('videochat.create') }}</q-card-section>
          <q-card-section class="q-gutter-md">
            <q-banner v-if="createError" role="alert" class="bg-red-1 text-negative">{{ createError }}</q-banner>
            <q-input v-model="name" :label="t('videochat.name')" :rules="[validName]" :disable="creating" autofocus outlined />
            <q-input v-model.number="maxParticipants" type="number" min="0" max="4294967295" step="1" :label="t('videochat.maxParticipants')" :hint="t('videochat.defaultHint')" :rules="[validNumber]" :disable="creating" outlined />
            <q-input v-model.number="emptyTimeout" type="number" min="0" max="4294967295" step="1" :label="t('videochat.timeout')" :hint="t('videochat.defaultHint')" :rules="[validNumber]" :disable="creating" outlined />
          </q-card-section>
          <q-card-actions align="right">
            <q-btn flat :label="t('actions.cancel')" :disable="creating" v-close-popup />
            <q-btn type="submit" color="primary" :label="t('videochat.create')" :loading="creating" />
          </q-card-actions>
        </q-form>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar, type QTableColumn } from 'quasar';
import { useProfilesStore } from '@/stores/profiles-store';
import { createVideochatClient, type Room } from '@/encore/videochat';

const { t, locale } = useI18n();
const $q = useQuasar();
const profiles = useProfilesStore();
const client = createVideochatClient();
const isAdmin = computed(() => profiles.isAuthenticated && profiles.profile?.role === 'admin');
const rooms = ref<Room[]>([]);
const loading = ref(false);
const creating = ref(false);
const showCreate = ref(false);
const listError = ref('');
const createError = ref('');
const name = ref('');
const maxParticipants = ref(0);
const emptyTimeout = ref(0);
let requestVersion = 0;

const pagination = ref({
  sortBy: 'status',
  descending: true,
  page: 1,
  rowsPerPage: 10,
});

const columns = computed<QTableColumn[]>(() => [
  { name: 'status', label: t('videochat.status'), field: 'active', align: 'center', sortable: true },
  { name: 'name', label: t('videochat.broadcaster'), field: 'display_name', align: 'left', sortable: true },
  { name: 'participants', label: t('videochat.participants'), field: 'num_participants', align: 'left', sortable: true },
  { name: 'maximum', label: t('videochat.maxParticipants'), field: 'max_participants', align: 'left', format: (value: number) => value ? String(value) : '—' },
  { name: 'created', label: t('videochat.created'), field: 'creation_time', align: 'left', sortable: true },
  { name: 'actions', label: t('videochat.actions'), field: 'name', align: 'left' },
]);

function roomRowKey(room: Room) {
  return JSON.stringify([room.owner_id, room.name]);
}

function validName(value: string) {
  const trimmed = value.trim();
  return (trimmed.length > 0 && new TextEncoder().encode(trimmed).length <= 128 && !/[\/\\\p{Cc}]/u.test(trimmed)) || t('videochat.invalidName');
}

function validNumber(value: unknown) {
  return (typeof value === 'number' && Number.isInteger(value) && value >= 0 && value <= 4294967295) || t('videochat.invalidNumber');
}

async function loadRooms() {
  if (!profiles.isAuthenticated) return;
  const version = ++requestVersion;
  loading.value = true;
  listError.value = '';
  try {
    const response = await client.listRooms();
    if (version === requestVersion) rooms.value = response.rooms ?? [];
  } catch (error) {
    if (version === requestVersion) listError.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (version === requestVersion) loading.value = false;
  }
}

function openCreate() {
  name.value = '';
  maxParticipants.value = 0;
  emptyTimeout.value = 0;
  createError.value = '';
  showCreate.value = true;
}

async function createRoom() {
  if (creating.value || !isAdmin.value) return;
  creating.value = true;
  createError.value = '';
  const userId = profiles.profile?.user_id;
  try {
    await client.createRoom({ name: name.value.trim(), max_participants: maxParticipants.value, empty_timeout: emptyTimeout.value });
    if (userId !== profiles.profile?.user_id || !isAdmin.value) return;
    showCreate.value = false;
    $q.notify({ type: 'positive', message: t('videochat.createdSuccess') });
    await loadRooms();
  } catch (error) {
    createError.value = error instanceof Error ? error.message : String(error);
  } finally {
    creating.value = false;
  }
}

watch(() => profiles.profile?.user_id, () => {
  ++requestVersion;
  rooms.value = [];
  listError.value = '';
  loading.value = false;
  showCreate.value = false;
  if (profiles.isAuthenticated) void loadRooms();
}, { immediate: true });
</script>

<style scoped>
@keyframes pulse {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.35;
    transform: scale(0.85);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.pulse-dot {
  animation: pulse 1.8s infinite ease-in-out;
}
</style>
