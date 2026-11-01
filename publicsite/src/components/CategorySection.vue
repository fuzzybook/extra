<template>
  <q-page class="section-page">
    <div class="section-container">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <span class="dot" :style="{ color }">•</span>
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t(`${section}.title`) }}</h1>
      </div>

      <p v-if="loading" class="section-status">{{ t('common.loading') }}</p>
      <p v-else-if="error" class="section-status">{{ t(`${section}.error`) }}</p>
      <p v-else-if="enabled.length === 0" class="section-status">{{ t(`${section}.empty`) }}</p>
      <BroadcasterGrid v-else :broadcasters="enabled" />
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import type { broadcasters } from '@/encore/client';
import { api } from '@/encore/api';
import BroadcasterGrid from '@/components/BroadcasterGrid.vue';

// Sezione del sito con i broadcaster di una categoria: il nome della sezione
// (chiave i18n) e la categoria come in config.SiteConfig (Elite, Clubs, Soft);
// il backend la confronta senza distinguere maiuscole e minuscole.
const props = defineProps<{ section: 'elite' | 'clubs' | 'soft'; category: string; color: string }>();

const { t } = useI18n();

useMeta(() => ({
  title: `${t(`${props.section}.title`)} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t(`${props.section}.subtitle`),
    },
  },
}));

const list = ref<broadcasters.Broadcaster[]>([]);
const loading = ref(true);
const error = ref(false);

// Elenco caricato nel browser (non nel pre-render SSG), filtrato dal server
// per categoria; solo i profili abilitati.
async function load() {
  loading.value = true;
  error.value = false;
  try {
    const res = await api.broadcasters.ListBroadcasters({ category: props.category, certified: false, availability: '', search: '' });
    list.value = res.broadcasters;
  } catch {
    error.value = true;
  } finally {
    loading.value = false;
  }
}

onMounted(load);
watch(() => props.category, load);

const enabled = computed(() => list.value.filter((b) => b.enabled && b.short_id));
</script>

<style scoped>
.section-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.section-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 32px 16px 64px;
}

.section-status {
  color: #bdbdbd;
  text-align: center;
}

.dot {
  font-size: 32px;
  line-height: 1;
}
</style>
