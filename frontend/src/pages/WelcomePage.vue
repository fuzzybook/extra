<template>
  <q-page class="flex flex-center">
    <q-card class="q-pa-md" style="width: 100%; max-width: 440px">
      <q-card-section>
        <div class="text-h6">{{ t('welcome.title') }}</div>
      </q-card-section>

      <q-card-section>
        <q-inner-loading :showing="loading" />

        <q-banner v-if="success" class="bg-positive text-white" rounded>
          {{ t('welcome.success', { email }) }}
        </q-banner>

        <q-banner v-else-if="error" class="bg-negative text-white" rounded>
          {{ error }}
        </q-banner>

        <q-banner v-else-if="!loading" class="bg-warning text-white" rounded>
          {{ t('welcome.missingToken') }}
        </q-banner>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn color="primary" to="/login" :label="t('nav.login')" no-caps />
      </q-card-actions>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { api } from '@/encore/api';

const { t } = useI18n();
const route = useRoute();
const client = api;
const loading = ref(false);
const success = ref(false);
const error = ref('');
const email = ref('');

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : '';
  if (!token) return;

  loading.value = true;
  try {
    const res = await client.registration.ConfirmWelcome(token);
    success.value = res.confirmed;
    email.value = res.email;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
});
</script>
