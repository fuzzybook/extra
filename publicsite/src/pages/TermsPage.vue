<template>
  <q-page class="terms-page">
    <article class="terms-panel">
      <h1 class="terms-title">{{ t('terms.title') }}</h1>
      <!-- Testo segnaposto: da sostituire con i termini definitivi. -->
      <p class="terms-text">{{ t('terms.body') }}</p>
      <p class="terms-updated">{{ t('terms.updated') }}</p>
    </article>
  </q-page>
</template>

<script setup lang="ts">
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

useMeta(() => ({
  title: `${t('terms.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('terms.subtitle'),
    },
  },
}));
</script>

<style scoped>
.terms-page {
  min-height: calc(100vh - 180px);
  padding: 32px 16px;
  background-color: #1b1b1b;
  color: #ffffff;
  font-family: 'Open Sans', 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.terms-panel {
  max-width: 820px;
  margin: 0 auto;
  padding: 48px 32px;
  background-color: #000000;
}

.terms-title {
  margin: 0 0 24px;
  font-size: 2.5rem;
  font-weight: 400;
  line-height: 1.2;
}

.terms-text {
  color: #d0d0d0;
  font-size: 1rem;
  line-height: 1.6;
}

.terms-updated {
  margin-top: 32px;
  color: #8a8a8a;
  font-size: 0.9rem;
}
</style>
