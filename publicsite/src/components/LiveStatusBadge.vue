<template>
  <!-- Stato della diretta in fondo alla foto del broadcaster: stanza, camera e
       microfono, verdi se attivi e grigi se spenti. Il genitore deve essere
       position: relative. -->
  <div v-if="status" class="live-status" :class="{ 'live-status--open': status.open }">
    <span
      v-for="item in items"
      :key="item.icon"
      class="live-status__item"
      :class="{ 'live-status__item--on': item.on }"
      role="img"
      :aria-label="item.label"
      :title="item.label"
    >
      <q-icon :name="item.icon" size="16px" />
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLiveStatus } from '@/composables/useLiveStatus';

const props = defineProps<{ shortId: string | undefined }>();

const { t } = useI18n();
const { liveStatus } = useLiveStatus();
const status = computed(() => liveStatus(props.shortId));
const items = computed(() => {
  const s = status.value;
  if (!s) return [];
  return [
    { icon: s.open ? 'meeting_room' : 'door_front', on: s.open, label: s.open ? t('live.roomOpen') : t('live.roomClosed') },
    { icon: s.video ? 'videocam' : 'videocam_off', on: s.video, label: s.video ? t('live.videoOn') : t('live.videoOff') },
    { icon: s.audio ? 'mic' : 'mic_off', on: s.audio, label: s.audio ? t('live.audioOn') : t('live.audioOff') },
  ];
});
</script>

<style scoped>
.live-status {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 1;
  display: flex;
  justify-content: center;
  gap: 14px;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.65);
}

.live-status__item {
  display: inline-flex;
  color: #9e9e9e;
}

.live-status__item--on {
  color: #4caf50;
}
</style>
