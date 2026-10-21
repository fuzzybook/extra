<template>
  <q-badge :color="statusInfo(status).color">
    {{ statusLabel }}
  </q-badge>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { statusInfo, type Status } from '@/stores/profiles-store';

const props = defineProps<{
  status: Status;
}>();

const { t } = useI18n();

const statusLabel = computed(() => {
  const key = `status.${props.status}`;
  const translated = t(key);
  return translated === key ? t('status.unknown', { status: props.status }) : translated;
});
</script>
