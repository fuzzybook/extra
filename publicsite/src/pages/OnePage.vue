<template>
  <q-page class="one-page">
    <div class="one-container">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <span class="dot color-elite">•</span>
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('one.title') }}</h1>
      </div>

      <p v-if="loading" class="one-status">{{ t('common.loading') }}</p>
      <p v-else-if="error" class="one-status">{{ t('one.error') }}</p>
      <p v-else-if="cards.length === 0" class="one-status">{{ t('one.empty') }}</p>

      <!-- Lista dei broadcaster: copertina (prima foto) larga al massimo 120px -->
      <ul v-else class="one-list">
        <li v-for="card in cards" :key="card.short_id">
          <router-link :to="broadcasterRoute(card.short_id)" class="one-card">
            <PromoBroadcaster v-if="card.image_url" :broadcaster="card" :alt="card.display_name" />
            <div v-else class="one-card__placeholder" aria-hidden="true"></div>
            <div class="one-card__name">{{ card.display_name }}</div>
            <div v-if="card.category" class="one-card__category">{{ card.category }}</div>
          </router-link>
        </li>
      </ul>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import type { advertising, broadcasters } from '@/encore/client';
import { api } from '@/encore/api';
import PromoBroadcaster from '@/components/PromoBroadcaster.vue';

const { t } = useI18n();
const route = useRoute();

useMeta(() => ({
  title: `${t('one.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('one.subtitle'),
    },
  },
}));

type Card = advertising.PromoBroadcaster & { category: string };

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

// La copertina è la prima foto, con la sua didascalia nella posizione salvata.
const cards = computed<Card[]>(() =>
  list.value
    .filter((b) => b.enabled && b.short_id)
    .map((b) => {
      const cover = b.photos.find((p) => p.image_url);
      return {
        short_id: b.short_id,
        display_name: b.display_name,
        category: b.category,
        image_url: cover?.image_url ?? '',
        description: cover?.description?.text ?? '',
        box_x: cover?.description?.box_x ?? 0,
        box_y: cover?.description?.box_y ?? 0,
      };
    }),
);

function broadcasterRoute(shortId: string): string {
  const locale = route.params.locale ? `/${String(route.params.locale)}` : '';
  return `${locale}/br/${shortId}`;
}
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

.one-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 120px));
  justify-content: center;
  gap: 20px 16px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.one-card {
  display: block;
  max-width: 120px;
  color: inherit;
  text-decoration: none;
}

.one-card:focus-visible {
  outline: 2px solid #bbbbbb;
  outline-offset: 3px;
}

.one-card:hover .one-card__name {
  color: #ff9500;
}

.one-card__placeholder {
  aspect-ratio: 2 / 3;
  border-radius: 4px;
  background: linear-gradient(145deg, #292929, #151515);
}

.one-card__name {
  margin-top: 6px;
  font-size: 0.85rem;
  line-height: 1.3;
  font-weight: 600;
  overflow-wrap: anywhere;
  transition: color 0.15s;
}

.one-card__category {
  color: #9a9a9a;
  font-size: 0.75rem;
}

.dot {
  font-size: 32px;
  line-height: 1;
}

.color-elite {
  color: #ff9500;
}
</style>
