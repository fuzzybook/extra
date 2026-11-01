<template>
  <div class="webchat-stage">
    <div v-if="videos.length" class="row q-col-gutter-md">
      <div v-for="item in videos" :key="item.sid" :class="videos.length > 1 ? 'col-12 col-md-6' : 'col-12'">
        <LiveTrack :track="item.track" :local="item.local" :resizable="resizable" />
        <div v-if="captions" class="webchat-stage__caption">{{ t(item.screen ? 'webchat.screen' : 'webchat.camera') }}</div>
      </div>
    </div>
    <div v-else class="webchat-stage__placeholder">
      <slot name="placeholder">
        <q-icon name="live_tv" size="56px" />
        <p>{{ t('webchat.waiting') }}</p>
      </slot>
    </div>
    <LiveTrack v-for="item in audios" :key="item.sid" :track="item.track" audio />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import LiveTrack from './LiveTrack.vue';
import type { LiveMedia } from '../useLiveRoom';

withDefaults(defineProps<{ videos: LiveMedia[]; audios: LiveMedia[]; resizable?: boolean; captions?: boolean }>(), {
  resizable: true,
  captions: true,
});

const { t } = useI18n();
</script>

<style scoped>
.webchat-stage {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  box-sizing: border-box;
}

.webchat-stage__caption {
  margin-top: 4px;
  font-size: 0.75rem;
  opacity: 0.8;
}

.webchat-stage__placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 320px;
  padding: 24px;
  border-radius: 4px;
  background: var(--webchat-stage-bg, #212121);
  color: var(--webchat-stage-text, #ffffff);
  text-align: center;
}

.webchat-stage__placeholder p {
  margin: 12px 0 0;
}
</style>
