<template>
  <q-page class="webcam-page flex flex-center">
    <div class="text-center q-pa-xl content-card">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <span class="dot color-webcam">•</span>
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('webcam.title') }}</h1>
      </div>
      <p class="text-body1 text-grey-4 q-mb-lg">
        {{ t('webcam.subtitle') }}
      </p>

      <div class="webcam-box q-mb-lg flex flex-center">
        <q-icon name="videocam" size="64px" color="blue-4" />
        <div class="text-caption text-grey-4 q-mt-sm">{{ t('webcam.status') }}</div>
      </div>

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
          :label="t('actions.goToDashboard')"
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
  title: `${t('webcam.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('webcam.subtitle'),
    },
  },
}));
</script>

<style scoped>
.webcam-page {
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

.webcam-box {
  width: 100%;
  height: 220px;
  background: #111118;
  border: 1px dashed rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  flex-direction: column;
}

.dot {
  font-size: 32px;
  line-height: 1;
}

.color-webcam {
  color: #0a84ff;
}
</style>
