<template>
  <q-page class="flex flex-center">
    <q-card class="q-pa-md" style="width: 100%; max-width: 400px">
      <q-card-section>
        <div class="text-h6">{{ pageTitle }}</div>
      </q-card-section>

      <q-card-section>
        <q-form v-if="!token" class="column q-gutter-md" @submit.prevent="requestReset">
          <q-banner v-if="requestSent" class="bg-positive text-white" rounded>
            {{ t('passwordReset.requestSent') }}
          </q-banner>

          <q-input
            v-if="!requestSent"
            v-model="email"
            :label="t('fields.email')"
            type="email"
            autocomplete="email"
            outlined
            :rules="[(val) => !!val || t('fields.email')]"
          />

          <q-banner v-if="error" class="bg-negative text-white" rounded>
            {{ error }}
          </q-banner>

          <q-btn
            v-if="!requestSent"
            type="submit"
            color="primary"
            :label="t('passwordReset.sendLink')"
            :loading="loading"
            class="full-width"
            no-caps
          />

          <q-btn flat no-caps to="/login" :label="t('passwordReset.backToLogin')" />
        </q-form>

        <q-form v-else class="column q-gutter-md" @submit.prevent="confirmReset">
          <q-banner v-if="resetDone" class="bg-positive text-white" rounded>
            {{ t('passwordReset.resetDone') }}
          </q-banner>

          <q-input
            v-if="!resetDone"
            v-model="password"
            :label="t('fields.password')"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="new-password"
            outlined
            :rules="[(val) => !!val || t('fields.password')]"
          >
            <template v-slot:append>
              <q-icon
                :name="showPassword ? 'visibility_off' : 'visibility'"
                class="cursor-pointer"
                @click="showPassword = !showPassword"
              />
            </template>
          </q-input>

          <q-banner v-if="error" class="bg-negative text-white" rounded>
            {{ error }}
          </q-banner>

          <q-btn
            v-if="!resetDone"
            type="submit"
            color="primary"
            :label="t('passwordReset.updatePassword')"
            :loading="loading"
            class="full-width"
            no-caps
          />

          <q-btn flat no-caps to="/login" :label="t('passwordReset.backToLogin')" />
        </q-form>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import Client, { Local } from '@/encore/client';

const { t } = useI18n();
const route = useRoute();
const client = new Client(Local);

const token = computed(() => (typeof route.query.token === 'string' ? route.query.token : ''));
const pageTitle = computed(() => (token.value ? t('passwordReset.newPasswordTitle') : t('passwordReset.title')));

const email = ref(typeof route.query.email === 'string' ? route.query.email : '');
const password = ref('');
const showPassword = ref(false);
const loading = ref(false);
const error = ref('');
const requestSent = ref(false);
const resetDone = ref(false);

async function withLoading(fn: () => Promise<void>) {
  loading.value = true;
  error.value = '';
  try {
    await fn();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function requestReset() {
  await withLoading(async () => {
    await client.registration.RequestPasswordReset({ email: email.value });
    requestSent.value = true;
  });
}

async function confirmReset() {
  await withLoading(async () => {
    await client.registration.ConfirmPasswordReset(token.value, { password: password.value });
    resetDone.value = true;
    password.value = '';
  });
}
</script>
