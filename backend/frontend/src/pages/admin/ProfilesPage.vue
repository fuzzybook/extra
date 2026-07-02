<template>
  <q-page class="q-pa-md">
    <div class="row items-center q-mb-md">
      <div class="text-h5">{{ t('admin.profiles') }}</div>
      <q-space />
      <q-btn color="primary" icon="add" :label="t('actions.newProfile')" @click="openCreate" />
    </div>

    <q-banner v-if="adminStore.error" class="bg-negative text-white q-mb-md" rounded>
      {{ adminStore.error }}
    </q-banner>

    <q-table
      :rows="adminStore.profiles"
      :columns="columns"
      row-key="user_id"
      :loading="adminStore.loading"
      flat
      bordered
    >
      <template v-slot:body-cell-status="props">
        <q-td :props="props">
          <q-badge :color="statusInfo(props.value).color">{{ statusLabel(props.value) }}</q-badge>
        </q-td>
      </template>

      <template v-slot:body-cell-is_artist="props">
        <q-td :props="props">
          <q-icon
            :name="props.value ? 'check_circle' : 'cancel'"
            :color="props.value ? 'positive' : 'grey-5'"
            size="sm"
          >
            <q-tooltip>{{ props.value ? t('admin.artist') : t('admin.notArtist') }}</q-tooltip>
          </q-icon>
        </q-td>
      </template>

      <template v-slot:body-cell-avatar_url="props">
        <q-td :props="props" auto-width>
          <q-avatar size="32px">
            <img v-if="props.value" :src="props.value" />
            <q-icon v-else name="person" />
          </q-avatar>

          <q-btn flat round dense icon="more_vert" size="sm" class="q-ml-xs">
            <q-menu>
              <q-list>
                <q-item v-close-popup clickable @click="openEdit(props.row)">
                  <q-item-section>{{ t('actions.editProfile') }}</q-item-section>
                </q-item>
                <q-item v-close-popup clickable @click="openRoleEdit(props.row)">
                  <q-item-section>{{ t('actions.updateRole') }}</q-item-section>
                </q-item>
                <q-item v-close-popup clickable @click="openStatusEdit(props.row)">
                  <q-item-section>{{ t('actions.updateStatus') }}</q-item-section>
                </q-item>
              </q-list>
            </q-menu>
          </q-btn>
        </q-td>
      </template>
    </q-table>

    <EditProfileDialog v-model="editDialogOpen" :profile="editingProfile" />
    <UpdateRoleDialog v-model="roleDialogOpen" :profile="roleEditingProfile" />
    <UpdateStatusDialog v-model="statusDialogOpen" :profile="statusEditingProfile" />
    <CreateProfileDialog v-model="createDialogOpen" />
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { QTableColumn } from 'quasar';
import { useAdminStore, type Profile } from '@/stores/admin-store';
import { statusInfo } from '@/stores/profiles-store';
import CreateProfileDialog from './dialogs/CreateProfileDialog.vue';
import EditProfileDialog from './dialogs/EditProfileDialog.vue';
import UpdateRoleDialog from './dialogs/UpdateRoleDialog.vue';
import UpdateStatusDialog from './dialogs/UpdateStatusDialog.vue';

const adminStore = useAdminStore();
const { t } = useI18n();

const editDialogOpen = ref(false);
const editingProfile = ref<Profile | null>(null);

function openEdit(profile: Profile) {
  editingProfile.value = profile;
  editDialogOpen.value = true;
}

const createDialogOpen = ref(false);

function openCreate() {
  createDialogOpen.value = true;
}

const roleDialogOpen = ref(false);
const roleEditingProfile = ref<Profile | null>(null);

function openRoleEdit(profile: Profile) {
  roleEditingProfile.value = profile;
  roleDialogOpen.value = true;
}

const statusDialogOpen = ref(false);
const statusEditingProfile = ref<Profile | null>(null);

function openStatusEdit(profile: Profile) {
  statusEditingProfile.value = profile;
  statusDialogOpen.value = true;
}

const columns = computed<QTableColumn[]>(() => [
  { name: 'avatar_url', label: '', field: 'avatar_url', align: 'left' },
  { name: 'display_name', label: t('admin.columns.name'), field: 'display_name', align: 'left', sortable: true },
  { name: 'email', label: t('admin.columns.email'), field: 'email', align: 'left', sortable: true },
  { name: 'role', label: t('admin.columns.role'), field: 'role', align: 'left', sortable: true },
  { name: 'status', label: t('admin.columns.status'), field: 'status', align: 'left', sortable: true },
  { name: 'is_artist', label: t('admin.columns.artist'), field: 'is_artist', align: 'center', sortable: true },
  {
    name: 'created_at',
    label: t('admin.columns.created'),
    field: 'created_at',
    align: 'left',
    sortable: true,
    format: (val: string) => new Date(val).toLocaleString(),
  },
  {
    name: 'updated_at',
    label: t('admin.columns.updated'),
    field: 'updated_at',
    align: 'left',
    sortable: true,
    format: (val: string) => new Date(val).toLocaleString(),
  },
]);

function statusLabel(status: number) {
  const key = `status.${status}`;
  const translated = t(key);
  return translated === key ? t('status.unknown', { status }) : translated;
}

onMounted(() => {
  void adminStore.listProfiles();
  void adminStore.fetchSystemOptions();
});
</script>
