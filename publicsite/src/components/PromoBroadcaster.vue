<template>
  <!-- Foto del broadcaster con le sue proporzioni; la didascalia (HTML) è
       posizionata in box_x / box_y solo dopo il rendering dell'immagine e scala
       con la foto (vedi .photo-caption-text in app.css). -->
  <div class="promo-broadcaster" :style="frameStyle">
    <img
      :key="broadcaster.image_url"
      :src="broadcaster.image_url"
      :alt="alt || broadcaster.display_name || ''"
      class="promo-broadcaster__img"
      loading="lazy"
      @load="loaded = true"
      @error="loaded = false"
    />
    <div v-if="loaded && caption" class="promo-broadcaster__layer photo-caption-frame">
      <div class="promo-broadcaster__caption photo-caption-text" :style="captionStyle" v-html="caption" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { advertising } from '@/encore/client';
import { sanitizeInlineHtml } from '@/utils/sanitize';

const props = defineProps<{
  broadcaster: advertising.PromoBroadcaster;
  alt?: string | undefined;
}>();

const loaded = ref(false);
watch(
  () => props.broadcaster.image_url,
  () => {
    loaded.value = false;
  },
);

// Proporzioni dalla foto, così lo spazio è riservato prima del caricamento.
const frameStyle = computed(() => {
  const { width, height } = props.broadcaster;
  return width && height ? { aspectRatio: `${width} / ${height}` } : {};
});

const caption = computed(() => sanitizeInlineHtml(props.broadcaster.description));

// box_x / box_y sono l'offset in percentuale della foto (0-100).
function percent(value: number | undefined): number {
  return Math.min(Math.max(value ?? 0, 0), 100);
}

const captionStyle = computed(() => ({
  left: `${percent(props.broadcaster.box_x)}%`,
  top: `${percent(props.broadcaster.box_y)}%`,
}));
</script>

<style scoped>
.promo-broadcaster {
  position: relative;
  width: 100%;
  overflow: hidden;
  border-radius: 4px;
}

.promo-broadcaster__img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.promo-broadcaster__layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

/* Stessa geometria della didascalia nel pannello (padding in em, bordo 1px) */
.promo-broadcaster__caption {
  position: absolute;
  max-width: 85%;
  padding: 0.35em 0.7em;
  border: 1px solid transparent;
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.85);
  word-break: break-word;
}
</style>
