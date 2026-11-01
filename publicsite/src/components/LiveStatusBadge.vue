<template>
  <!-- Stato della diretta in fondo alla foto del broadcaster: il genitore
       deve essere position: relative. -->
  <div v-if="status" class="live-status">
    <LiveStatusIcons :status="status" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useLiveStatus } from '@/composables/useLiveStatus';
import LiveStatusIcons from '@/components/LiveStatusIcons.vue';

const props = defineProps<{ shortId: string | undefined }>();

const { liveStatus } = useLiveStatus();
const status = computed(() => liveStatus(props.shortId));
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
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.65);
}
</style>
