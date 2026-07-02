<template>
  <q-dialog v-model="open" @show="focusFirstField">
    <q-card style="min-width: 350px">
      <q-card-section>
        <div class="text-h6">{{ t('actions.newProfile') }}</div>
      </q-card-section>

      <q-card-section class="q-gutter-md">
        <q-input
          ref="emailInputRef"
          v-model="createForm.email"
          :label="t('fields.email')"
          type="email"
          autocomplete="off"
          name="new-profile-email"
        />
        <q-input
          v-model="createForm.password"
          :label="t('fields.password')"
          type="password"
          autocomplete="new-password"
          name="new-profile-password"
        />
        <q-input v-model="createForm.display_name" :label="t('fields.displayName')" />
        <AvatarUpload v-model="createForm.avatar_url" :uploader="adminStore.uploadAvatar" />

        <q-banner v-if="adminStore.error" class="bg-negative text-white" rounded>
          {{ adminStore.error }}
        </q-banner>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn v-close-popup flat :label="t('actions.cancel')" />
        <q-btn color="primary" :label="t('actions.create')" :loading="adminStore.loading" @click="save" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAdminStore, type RegisterParams } from '@/stores/admin-store';
import AvatarUpload from '@/components/AvatarUpload.vue';

type Focusable = { focus: () => void };

const open = defineModel<boolean>({ required: true });
const adminStore = useAdminStore();
const { t } = useI18n();
const emailInputRef = ref<Focusable | null>(null);

const createForm = reactive<RegisterParams>({
  email: '',
  password: '',
  display_name: '',
  avatar_url: '',
});

watch(open, (isOpen) => {
  if (isOpen) {
    resetForm();
  }
});

async function save() {
  try {
    await adminStore.insertProfile({ ...createForm });
    open.value = false;
  } catch {
    // adminStore.error already holds the failure message
  }
}

function resetForm() {
  createForm.email = '';
  createForm.password = '';
  createForm.display_name = '';
  createForm.avatar_url = '';
}

async function focusFirstField() {
  await nextTick();
  emailInputRef.value?.focus();
}
</script>
