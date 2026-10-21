<template>
  <q-page class="login-page flex flex-center">
    <div class="text-center q-pa-xl content-card">
      <div class="row items-center justify-center q-gutter-x-sm q-mb-md">
        <q-icon name="lock" color="cyan" size="36px" />
        <h1 class="text-h4 text-weight-bold text-white q-my-none">{{ t('login.title') }}</h1>
      </div>
      <p class="text-body1 text-grey-4 q-mb-lg">
        {{ t('login.subtitle') }}
      </p>

      <div class="q-mb-lg">
        <q-btn
          to="/dashboard"
          color="cyan-5"
          text-color="dark"
          size="lg"
          unelevated
          icon="login"
          :label="t('login.button')"
          class="full-width rounded-borders text-weight-bold"
        />
      </div>

      <div class="row justify-center">
        <q-btn
          to="/"
          flat
          color="grey-4"
          :label="t('common.backHome')"
          icon="arrow_back"
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
  title: `${t('login.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('login.subtitle'),
    },
  },
}));
</script>

<style scoped>
.login-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.content-card {
  width: 100%;
  max-width: 480px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  backdrop-filter: blur(10px);
}
</style>
