<template>
  <!-- Stanza, camera e microfono: verdi se attivi, grigi se spenti. -->
  <span class="live-icons">
    <span
      v-for="item in items"
      :key="item.key"
      class="live-icons__item"
      :class="{ 'live-icons__item--on': item.on }"
      role="img"
      :aria-label="item.label"
      :title="item.label"
    >
      <q-icon :name="item.icon" :size="size" />
    </span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { LiveStatus } from '@/composables/useLiveStatus';

const props = withDefaults(defineProps<{ status: LiveStatus; size?: string }>(), { size: '16px' });

const { t } = useI18n();
const items = computed(() => {
  const s = props.status;
  return [
    { key: 'room', icon: s.open ? 'meeting_room' : 'door_front', on: s.open, label: s.open ? t('live.roomOpen') : t('live.roomClosed') },
    { key: 'video', icon: s.video ? 'videocam' : 'videocam_off', on: s.video, label: s.video ? t('live.videoOn') : t('live.videoOff') },
    { key: 'audio', icon: s.audio ? 'mic' : 'mic_off', on: s.audio, label: s.audio ? t('live.audioOn') : t('live.audioOff') },
  ];
});
</script>

<style scoped>
.live-icons {
  display: inline-flex;
  align-items: center;
  gap: 14px;
}

.live-icons__item {
  display: inline-flex;
  color: #9e9e9e;
}

.live-icons__item--on {
  color: #4caf50;
}
</style>
