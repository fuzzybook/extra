<template>
  <q-dialog :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)" @show="reset">
    <q-card class="bg-white text-dark" style="width: 460px; max-width: 95vw">
      <q-card-section class="row items-center justify-between bg-primary text-white q-py-sm">
        <div class="text-subtitle1 text-weight-bold">{{ t('userProfile.dataTitle') }}</div>
        <q-btn v-close-popup flat round dense icon="close" :aria-label="t('brEditor.actions.cancel')" />
      </q-card-section>

      <q-form @submit.prevent="save">
        <q-card-section class="q-gutter-md">
          <q-input
            v-model.trim="displayName"
            :label="t('userProfile.nickname')"
            :hint="t('userProfile.nicknameHint')"
            maxlength="32"
            autocomplete="nickname"
            outlined
            autofocus
            :error="Boolean(nicknameError)"
            :error-message="nicknameError"
          />
          <q-input :model-value="session.me.value?.email ?? ''" :label="t('userProfile.email')" outlined readonly />
          <div v-if="error" class="text-negative text-caption" role="alert">{{ error }}</div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn v-close-popup flat :label="t('brEditor.actions.cancel')" />
          <q-btn type="submit" color="primary" unelevated :label="t('brEditor.actions.save')" :loading="saving" :disable="!changed" />
        </q-card-actions>
      </q-form>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Notify } from 'quasar';
import { useI18n } from 'vue-i18n';
import { api } from '@/encore/api';
import { useSession } from '@/composables/useSession';

defineProps<{ modelValue: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>();

const { t } = useI18n();
const session = useSession();

const displayName = ref('');
const nicknameError = ref('');
const saving = ref(false);
const error = ref('');
const changed = computed(() => displayName.value !== (session.me.value?.display_name ?? ''));

function reset() {
  displayName.value = session.me.value?.display_name ?? '';
  nicknameError.value = '';
  error.value = '';
}

async function save() {
  const me = session.me.value;
  if (!me) return;
  // Stesso limite della registrazione (il backend lo ricontrolla).
  const length = [...displayName.value].length;
  nicknameError.value = length < 2 || length > 32 ? t('userProfile.nicknameInvalid') : '';
  if (nicknameError.value) return;
  saving.value = true;
  error.value = '';
  try {
    session.me.value = await api.profiles.Update({ display_name: displayName.value, avatar_url: me.avatar_url });
    Notify.create({ type: 'positive', message: t('userProfile.saved') });
    emit('update:modelValue', false);
  } catch {
    error.value = t('userProfile.saveError');
  } finally {
    saving.value = false;
  }
}
</script>
