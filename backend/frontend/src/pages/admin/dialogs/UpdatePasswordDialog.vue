<template>
  <q-dialog v-model="open" @show="focusFirstField">
    <q-card style="min-width: 350px">
      <q-card-section>
        <div class="text-h6">{{ t('actions.updatePassword') }}</div>
      </q-card-section>

      <q-card-section v-if="profile" class="q-gutter-md">
        <div class="text-subtitle1">{{ profile.email }}</div>

        <q-input
          ref="passwordInputRef"
          v-model="password"
          :label="t('fields.password')"
          :type="showPassword ? 'text' : 'password'"
          autocomplete="new-password"
          :error="Boolean(passwordError)"
          :error-message="passwordError"
          outlined
        >
          <template v-slot:append>
            <q-icon
              :name="showPassword ? 'visibility_off' : 'visibility'"
              class="cursor-pointer"
              @click="showPassword = !showPassword"
            />
          </template>
        </q-input>

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
import { Notify } from 'quasar';
import { useAdminStore, type Profile } from '@/stores/admin-store';

type Focusable = { focus: () => void };

const props = defineProps<{
  profile: Profile | null;
}>();

const open = defineModel<boolean>({ required: true });
const adminStore = useAdminStore();
const { t } = useI18n();
const password = ref('');
const passwordError = ref('');
const showPassword = ref(false);
const passwordInputRef = ref<Focusable | null>(null);

watch(
  () => [open.value, props.profile] as const,
  ([isOpen]) => {
    if (isOpen) {
      password.value = '';
      passwordError.value = '';
      showPassword.value = false;
    }
  },
  { immediate: true },
);

async function save() {
  if (!props.profile) return;
  passwordError.value = '';

  if (!password.value) {
    passwordError.value = t('fields.password');
    return;
  }

  try {
    await adminStore.updateProfilePassword(props.profile.user_id, { password: password.value });
    Notify.create({ type: 'positive', message: t('admin.notifications.passwordUpdated') });
    open.value = false;
  } catch {
    // adminStore.error already holds the failure message
  }
}

async function focusFirstField() {
  await nextTick();
  passwordInputRef.value?.focus();
}
</script>
