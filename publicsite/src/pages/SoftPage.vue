<template>
  <q-page class="soft-page flex flex-center">
    <div class="text-center q-pa-xl content-card">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <span class="dot color-private">•</span>
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('soft.title') }}</h1>
      </div>
      <p class="text-body1 text-grey-4 q-mb-lg">
        {{ t('soft.subtitle') }}
      </p>
      <div class="row justify-center q-gutter-md">
        <q-btn
          to="/"
          outline
          color="white"
          :label="t('common.backHome')"
          icon="arrow_back"
          class="rounded-borders"
        />
        <q-btn
          to="/dashboard"
          color="primary"
          unelevated
          :label="t('common.dashboard')"
          icon="dashboard"
          class="rounded-borders"
        />
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

useMeta(() => ({
  title: `${t('soft.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('soft.subtitle'),
    },
  },
}));
</script>

<style scoped>
.soft-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.content-card {
  max-width: 650px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  backdrop-filter: blur(10px);
}

.dot {
  font-size: 32px;
  line-height: 1;
}

.color-private {
  color: #30d158;
}
</style>
