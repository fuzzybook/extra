<template>
  <q-page class="elite-page">
    <div class="elite-container">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <span class="dot color-elite">•</span>
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('elite.title') }}</h1>
      </div>

      <p v-if="loading" class="elite-status">{{ t('common.loading') }}</p>
      <p v-else-if="error" class="elite-status">{{ t('elite.error') }}</p>
      <p v-else-if="enabled.length === 0" class="elite-status">{{ t('elite.empty') }}</p>
      <BroadcasterGrid v-else :broadcasters="enabled" />
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

const { t } = useI18n();

useMeta(() => ({
  title: `${t('elite.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('elite.subtitle'),
    },
  },
}));

const list = ref<broadcasters.Broadcaster[]>([]);
const loading = ref(true);
const error = ref(false);

// Elenco caricato nel browser (non nel pre-render SSG); solo i profili abilitati.
onMounted(async () => {
  try {
    const res = await api.broadcasters.ListBroadcasters({ category: '', certified: false, availability: '', search: '' });
    list.value = res.broadcasters;
  } catch {
    error.value = true;
  } finally {
    loading.value = false;
  }
});

const enabled = computed(() => list.value.filter((b) => b.enabled && b.short_id));
</script>

<style scoped>
.elite-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.elite-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 32px 16px 64px;
}

.elite-status {
  color: #bdbdbd;
  text-align: center;
}

.dot {
  font-size: 32px;
  line-height: 1;
}

.color-elite {
  color: #ff9500;
}
</style>
