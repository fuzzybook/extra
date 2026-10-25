<template>
  <!-- Promo right: immagine, con il link di destinazione se presente. -->
  <a
    v-if="content.url"
    :href="content.url"
    :target="content.new_window ? '_blank' : undefined"
    :rel="content.new_window ? 'noopener noreferrer' : undefined"
    class="promo-image-link"
  >
    <img :src="content.image_url" :alt="alt || ''" class="promo-image-link__img" loading="lazy" />
  </a>
  <img v-else :src="content.image_url" :alt="alt || ''" class="promo-image-link__img" loading="lazy" />
</template>

<script setup lang="ts">
import type { advertising } from '@/encore/client';

defineProps<{
  content: advertising.PromoContent;
  alt?: string | undefined;
}>();
</script>

<style scoped>
.promo-image-link {
  display: block;
}

.promo-image-link__img {
  display: block;
  width: 100%;
  height: auto;
  border-radius: 4px;
}
</style>
