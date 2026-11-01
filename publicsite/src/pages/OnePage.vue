<template>
  <q-page class="one-page">
    <div class="one-container">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <span class="dot color-elite">•</span>
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('one.title') }}</h1>
      </div>

      <p v-if="loading" class="one-status">{{ t('common.loading') }}</p>
      <p v-else-if="error" class="one-status">{{ t('one.error') }}</p>
      <p v-else-if="enabled.length === 0" class="one-status">{{ t('one.empty') }}</p>
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
  title: `${t('one.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('one.subtitle'),
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
.one-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.one-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 32px 16px 64px;
}

.one-status {
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
