<template>
  <q-dialog v-model="open" @show="focusFirstField">
    <q-card style="min-width: 350px">
      <q-card-section>
        <div class="text-h6">{{ t('actions.updateStatus') }}</div>
      </q-card-section>

      <q-card-section v-if="profile" class="q-gutter-md">
        <div class="text-subtitle1">{{ profile.email }}</div>

        <q-select
          ref="statusSelectRef"
          v-model="selectedStatus"
          :options="adminStore.updatableStatuses"
          option-label="name"
          option-value="value"
          emit-value
          map-options
          :label="t('fields.status')"
        />

        <q-banner v-if="adminStore.error" class="bg-negative text-white" rounded>
          {{ adminStore.error }}
        </q-banner>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn v-close-popup flat :label="t('actions.cancel')" />
        <q-btn color="primary" :label="t('actions.save')" :loading="adminStore.loading" @click="save" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAdminStore, type Profile } from '@/stores/admin-store';
import type { Status } from '@/stores/profiles-store';

type Focusable = { focus: () => void };

const props = defineProps<{
  profile: Profile | null;
}>();

const open = defineModel<boolean>({ required: true });
const adminStore = useAdminStore();
const { t } = useI18n();
const selectedStatus = ref<Status>(0);
const statusSelectRef = ref<Focusable | null>(null);

watch(
  () => [open.value, props.profile] as const,
  ([isOpen, profile]) => {
    if (isOpen && profile) {
      selectedStatus.value = profile.status;
    }
  },
  { immediate: true },
);

async function save() {
  if (!props.profile) return;
  try {
    await adminStore.updateProfileStatus(props.profile.user_id, selectedStatus.value);
    open.value = false;
  } catch {
    // adminStore.error already holds the failure message
  }
}

async function focusFirstField() {
  await nextTick();
  statusSelectRef.value?.focus();
}
</script>
