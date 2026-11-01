<template>
  <q-page class="webcam-page">
    <div class="webcam-container">
      <!-- Titolo solo per screen reader e motori di ricerca: la sezione è già
           evidenziata dal pulsante attivo nell'header. -->
      <h1 class="visually-hidden">{{ t('webcam.title') }}</h1>

      <p v-if="loading" class="webcam-status">{{ t('common.loading') }}</p>
      <p v-else-if="error" class="webcam-status">{{ t('webcam.error') }}</p>
      <p v-else-if="webcams.length === 0" class="webcam-status">{{ t('webcam.empty') }}</p>
      <BroadcasterGrid v-else :broadcasters="webcams" live-when-open />
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import type { broadcasters } from '@/encore/client';
import { api } from '@/encore/api';
import BroadcasterGrid from '@/components/BroadcasterGrid.vue';
import { useLiveStatus } from '@/composables/useLiveStatus';

const { t } = useI18n();
const { liveStatus, refresh, ready } = useLiveStatus();

useMeta(() => ({
  title: `${t('webcam.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('webcam.subtitle'),
    },
  },
}));

const list = ref<broadcasters.Broadcaster[]>([]);
const loading = ref(true);
const error = ref(false);

// Come One, ma solo i broadcaster con il broadcasting abilitato nel personal
// setup: la stessa richiesta dello stato della diretta dice quali sono.
onMounted(async () => {
  try {
    const [res] = await Promise.all([
      api.broadcasters.ListBroadcasters({ category: '', certified: false, availability: '', search: '' }),
      refresh(),
    ]);
    list.value = res.broadcasters;
    error.value = !ready.value;
  } catch {
    error.value = true;
  } finally {
    loading.value = false;
  }
});

const webcams = computed(() => list.value.filter((b) => b.enabled && b.short_id && liveStatus(b.short_id) !== null));
</script>

<style scoped>
.webcam-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.webcam-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 32px 16px 64px;
}

.webcam-status {
  color: #bdbdbd;
  text-align: center;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
}
</style>
