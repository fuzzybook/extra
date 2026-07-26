<template>
  <q-page class="flex flex-center">
    <q-card class="q-pa-md" style="width: 100%; max-width: 400px">
      <q-card-section>
        <div class="text-h6">{{ t('login.title') }}</div>
      </q-card-section>

      <q-card-section>
        <q-form class="column q-gutter-md" @submit.prevent="onSubmit">
          <q-banner v-if="registered" class="bg-positive text-white" rounded>
            {{ t('register.success') }}
          </q-banner>

          <q-input
            v-model="email"
            :label="t('fields.email')"
            type="email"
            autocomplete="email"
            outlined
            :rules="[(val) => !!val || t('fields.email')]"
          />

          <div class="column q-gutter-xs">
            <div class="text-right">
              <router-link to="/password-reset" class="text-primary text-caption text-no-decoration">
                {{ t('login.forgotPassword') }}
              </router-link>
            </div>

            <q-input
              v-model="password"
              :label="t('fields.password')"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
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
          </div>

          <q-banner v-if="store.error" class="bg-negative text-white" rounded>
            {{ store.error }}
          </q-banner>

          <q-btn
            type="submit"
            color="primary"
            :label="t('nav.login')"
            :loading="store.loading"
            class="full-width"
            no-caps
          />

          <q-btn flat no-caps to="/register" :label="t('login.needAccount')" />
        </q-form>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';
import { useProfilesStore } from '@/stores/profiles-store';

const { t } = useI18n();
const email = ref('');
const password = ref('');
const showPassword = ref(false);

const store = useProfilesStore();
const router = useRouter();
const route = useRoute();
const registered = route.query.registered === '1';

if (typeof route.query.email === 'string') {
  email.value = route.query.email;
}

async function onSubmit() {
  try {
    await store.login({ user_email: email.value, password: password.value });
    await router.push('/');
  } catch {
    // store.error already holds the failure message
  }
}
</script>
