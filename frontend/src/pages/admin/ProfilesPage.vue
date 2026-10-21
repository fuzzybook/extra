<template>
  <q-page class="q-pa-md" ref="profilesPage">
    <div class="row items-center q-mb-md">
      <div class="text-h5">{{ t('admin.profiles') }}</div>
      <q-space />
      <q-btn color="primary" icon="add" :label="t('actions.newProfile')" @click="openCreate" />
    </div>

    <q-banner v-if="adminStore.error" class="bg-negative text-white q-mb-md" rounded>
      {{ adminStore.error }}
    </q-banner>

    <q-table
      :style="{ height: `${profilesTableHeight}px` }"
      class="sticky-header-table"
      :rows="adminStore.profiles"
      :columns="columns"
      row-key="user_id"
      :loading="adminStore.loading"
      flat
      bordered
      ref="profilesTable"
    >
      <template v-slot:body-cell-status="props">
        <q-td :props="props">
          <q-badge :color="statusInfo(props.value).color">{{ statusLabel(props.value) }}</q-badge>
        </q-td>
      </template>

      <template v-slot:body-cell-member_type="props">
        <q-td :props="props">
          <q-badge :color="memberTypeInfo(props.value).color">
            {{ memberTypeInfo(props.value).label }}
          </q-badge>
          <div v-if="props.row.broadcaster" class="text-caption text-grey ellipsis q-mt-xs" style="max-width: 140px">
            ID: {{ props.row.broadcaster.slice(0, 8) }}...
            <q-tooltip>Broadcaster ID: {{ props.row.broadcaster }}</q-tooltip>
          </div>
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
                <q-item v-close-popup clickable @click="openMemberTypeEdit(props.row)">
                  <q-item-section>{{ t('actions.updateMemberType') }}</q-item-section>
                </q-item>
                <q-item v-close-popup clickable @click="openStatusEdit(props.row)">
                  <q-item-section>{{ t('actions.updateStatus') }}</q-item-section>
                </q-item>
                <q-item v-close-popup clickable @click="openPasswordEdit(props.row)">
                  <q-item-section>{{ t('actions.updatePassword') }}</q-item-section>
                </q-item>
              </q-list>
            </q-menu>
          </q-btn>
        </q-td>
      </template>
    </q-table>

    <EditProfileDialog v-model="editDialogOpen" :profile="editingProfile" />
    <UpdateRoleDialog v-model="roleDialogOpen" :profile="roleEditingProfile" />
    <UpdateMemberTypeDialog v-model="memberTypeDialogOpen" :profile="memberTypeEditingProfile" />
    <UpdateStatusDialog v-model="statusDialogOpen" :profile="statusEditingProfile" />
    <UpdatePasswordDialog v-model="passwordDialogOpen" :profile="passwordEditingProfile" />
    <CreateProfileDialog v-model="createDialogOpen" />
  </q-page>
</template>

<style lang="scss">
.sticky-header-table {  /* height or max-height is important */

  .q-table__top,
  .q-table__bottom,
  thead tr:first-child th {    /* bg color is important for th; just specify one */
    background-color: #ffffff;
  }
  thead tr th {
    position: sticky;
    z-index: 1;
  }
  thead tr:first-child th {
    top: 0;
  }
  /* this is when the loading indicator appears */
  &.q-table--loading thead tr:last-child th {    /* height of all previous header rows */
    top: 48px;
  }
  /* prevent scrolling behind sticky top row on focus */
  tbody {    /* height of all previous header rows */
    scroll-margin-top: 48px;
  }
}</style>


<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type ComponentPublicInstance } from 'vue';
import { useI18n } from 'vue-i18n';
import type { QTableColumn } from 'quasar';
import { useAdminStore, type Profile } from '@/stores/admin-store';
import { useLayoutStore } from '@/stores/layout-store';
import { memberTypeInfo, statusInfo } from '@/stores/profiles-store';
import CreateProfileDialog from './dialogs/CreateProfileDialog.vue';
import EditProfileDialog from './dialogs/EditProfileDialog.vue';
import UpdatePasswordDialog from './dialogs/UpdatePasswordDialog.vue';
import UpdateRoleDialog from './dialogs/UpdateRoleDialog.vue';
import UpdateMemberTypeDialog from './dialogs/UpdateMemberTypeDialog.vue';
import UpdateStatusDialog from './dialogs/UpdateStatusDialog.vue';

const adminStore = useAdminStore();
const layoutStore = useLayoutStore();
const { t } = useI18n();

const profilesPage = ref<ComponentPublicInstance | null>(null);
const profilesTable = ref<ComponentPublicInstance | null>(null);
const profilesTableHeight = ref(610);

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

const memberTypeDialogOpen = ref(false);
const memberTypeEditingProfile = ref<Profile | null>(null);

function openMemberTypeEdit(profile: Profile) {
  memberTypeEditingProfile.value = profile;
  memberTypeDialogOpen.value = true;
}

const statusDialogOpen = ref(false);
const statusEditingProfile = ref<Profile | null>(null);

function openStatusEdit(profile: Profile) {
  statusEditingProfile.value = profile;
  statusDialogOpen.value = true;
}

const passwordDialogOpen = ref(false);
const passwordEditingProfile = ref<Profile | null>(null);

function openPasswordEdit(profile: Profile) {
  passwordEditingProfile.value = profile;
  passwordDialogOpen.value = true;
}

const columns = computed<QTableColumn[]>(() => [
  { name: 'avatar_url', label: '', field: 'avatar_url', align: 'left' },
  { name: 'display_name', label: t('admin.columns.name'), field: 'display_name', align: 'left', sortable: true },
  { name: 'email', label: t('admin.columns.email'), field: 'email', align: 'left', sortable: true },
  { name: 'role', label: t('admin.columns.role'), field: 'role', align: 'left', sortable: true },
  { name: 'member_type', label: t('admin.columns.memberType'), field: 'member_type', align: 'center', sortable: true },
  { name: 'status', label: t('admin.columns.status'), field: 'status', align: 'left', sortable: true },
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

function getElement(component: ComponentPublicInstance | null) {
  return component?.$el instanceof HTMLElement ? component.$el : null;
}

function updateProfilesTableHeight() {
  const page = getElement(profilesPage.value);
  const table = getElement(profilesTable.value);

  if (!page || !table) {
    return;
  }

  const tableRect = table.getBoundingClientRect();
  const pageStyle = window.getComputedStyle(page);
  const pagePaddingBottom = parseFloat(pageStyle.paddingBottom) || 0;
  const availableHeight = window.innerHeight - tableRect.top - pagePaddingBottom;

  profilesTableHeight.value = Math.max(0, Math.floor(availableHeight));
}

async function updateProfilesTableHeightAfterRender() {
  await nextTick();
  updateProfilesTableHeight();
}

onMounted(async () => {
  void adminStore.listProfiles();
  void adminStore.fetchSystemOptions();
  layoutStore.startBodyResizeListener();
  await updateProfilesTableHeightAfterRender();
});

watch(
  () => [adminStore.error, layoutStore.bodyWidth, layoutStore.bodyHeight],
  () => {
    void updateProfilesTableHeightAfterRender();
  },
);

onBeforeUnmount(() => {
  layoutStore.stopBodyResizeListener();
});
</script>
