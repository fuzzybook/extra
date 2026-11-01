<template>
  <!-- Lista dei broadcaster: copertina (prima foto) larga al massimo 120px -->
  <ul class="broadcaster-grid">
    <li v-for="card in cards" :key="card.short_id">
      <router-link :to="broadcasterRoute(card.short_id)" class="broadcaster-card">
        <PromoBroadcaster v-if="card.image_url" :broadcaster="card" :alt="card.display_name" />
        <div v-else class="broadcaster-card__placeholder" aria-hidden="true"></div>
        <div class="broadcaster-card__name">{{ card.display_name }}</div>
        <div v-if="card.category" class="broadcaster-card__category">{{ card.category }}</div>
      </router-link>
    </li>
  </ul>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import type { advertising, broadcasters } from '@/encore/client';
import PromoBroadcaster from '@/components/PromoBroadcaster.vue';

// Profili già filtrati dalla pagina (abilitati e con short_id).
const props = defineProps<{ broadcasters: broadcasters.Broadcaster[] }>();

const route = useRoute();

type Card = advertising.PromoBroadcaster & { category: string };

// La copertina è la prima foto, con la sua didascalia nella posizione salvata.
const cards = computed<Card[]>(() =>
  props.broadcasters.map((b) => {
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
.broadcaster-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(100px, 120px));
  justify-content: center;
  gap: 20px 16px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.broadcaster-card {
  display: block;
  max-width: 120px;
  color: inherit;
  text-decoration: none;
}

.broadcaster-card:focus-visible {
  outline: 2px solid #bbbbbb;
  outline-offset: 3px;
}

.broadcaster-card:hover .broadcaster-card__name {
  color: #ff9500;
}

.broadcaster-card__placeholder {
  aspect-ratio: 2 / 3;
  border-radius: 4px;
  background: linear-gradient(145deg, #292929, #151515);
}

.broadcaster-card__name {
  margin-top: 6px;
  font-size: 0.85rem;
  line-height: 1.3;
  font-weight: 600;
  overflow-wrap: anywhere;
  transition: color 0.15s;
}

.broadcaster-card__category {
  color: #9a9a9a;
  font-size: 0.75rem;
}
</style>
