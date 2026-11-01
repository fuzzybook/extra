<template>
  <q-dialog :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)" @show="reset">
    <q-card class="bg-white text-dark" style="width: 440px; max-width: 95vw">
      <q-card-section class="row items-center justify-between bg-primary text-white q-py-sm">
        <div class="text-subtitle1 text-weight-bold">{{ t('userProfile.avatarTitle') }}</div>
        <q-btn v-close-popup flat round dense icon="close" :aria-label="t('brEditor.actions.cancel')" />
      </q-card-section>

      <q-card-section>
        <!-- Stesso caricamento del pannello: file o fotocamera, con ritaglio -->
        <AvatarUpload v-model="avatarUrl" :uploader="upload" />
        <div v-if="error" class="text-negative text-caption q-mt-sm" role="alert">{{ error }}</div>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn v-close-popup flat :label="t('brEditor.actions.cancel')" />
        <q-btn color="primary" unelevated :label="t('brEditor.actions.save')" :loading="saving" :disable="!changed" @click="save" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Notify } from 'quasar';
import { useI18n } from 'vue-i18n';
import { AvatarUpload } from '@extra/broadcaster-editor';
import { api } from '@/encore/api';
import { useSession } from '@/composables/useSession';

defineProps<{ modelValue: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>();

const { t } = useI18n();
const session = useSession();

const avatarUrl = ref('');
const saving = ref(false);
const error = ref('');
const changed = computed(() => avatarUrl.value !== (session.me.value?.avatar_url ?? ''));

function reset() {
  avatarUrl.value = session.me.value?.avatar_url ?? '';
  error.value = '';
}

// Carica l'immagine ritagliata e restituisce il suo URL; il profilo si
// aggiorna solo con "Salva".
async function upload(image: Blob): Promise<string> {
  const resp = await api.profiles.UploadAvatar('POST', image);
  if (!resp.ok) throw new Error(`avatar upload failed (${resp.status})`);
  return ((await resp.json()) as { url: string }).url;
}

async function save() {
  const me = session.me.value;
  if (!me) return;
  saving.value = true;
  error.value = '';
  try {
    session.me.value = await api.profiles.Update({ display_name: me.display_name, avatar_url: avatarUrl.value });
    Notify.create({ type: 'positive', message: t('userProfile.saved') });
    emit('update:modelValue', false);
  } catch {
    error.value = t('userProfile.saveError');
  } finally {
    saving.value = false;
  }
}
</script>
