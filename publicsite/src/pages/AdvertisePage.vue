<template>
  <q-page class="advertise-page flex flex-center">
    <div class="text-center q-pa-xl content-card">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <q-icon name="campaign" color="amber" size="36px" />
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('advertise.title') }}</h1>
      </div>
      <p class="text-body1 text-grey-4 q-mb-lg">
        {{ t('advertise.subtitle') }}
      </p>

      <div class="adv-features q-gutter-y-sm text-left q-mb-lg">
        <div class="row items-center text-grey-3">
          <q-icon name="check_circle" color="positive" size="20px" class="q-mr-sm" />
          <span>{{ t('advertise.feature1') }}</span>
        </div>
        <div class="row items-center text-grey-3">
          <q-icon name="check_circle" color="positive" size="20px" class="q-mr-sm" />
          <span>{{ t('advertise.feature2') }}</span>
        </div>
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
          color="warning"
          text-color="dark"
          unelevated
          :label="t('advertise.startNow')"
          icon="rocket_launch"
          class="rounded-borders text-weight-bold"
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
  title: `${t('advertise.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('advertise.subtitle'),
    },
  },
}));
</script>

<style scoped>
.advertise-page {
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

.adv-features {
  background: #111118;
  padding: 16px 20px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
</style>
