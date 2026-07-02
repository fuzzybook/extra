<template>
  <q-page class="flex flex-center">
    <q-card class="q-pa-md" style="width: 100%; max-width: 440px">
      <template v-if="success">
        <q-card-section>
          <div class="text-h6">{{ t('register.welcomeTitle') }}</div>
        </q-card-section>

        <q-card-section>
          <q-banner class="bg-positive text-white" rounded>
            {{ t('register.confirmationSent') }}
          </q-banner>
        </q-card-section>
      </template>

      <template v-else>
        <q-card-section>
          <div class="text-h6">{{ t('register.title') }}</div>
        </q-card-section>

        <q-card-section>
          <q-form class="column q-gutter-md" @submit.prevent="onSubmit">
            <q-input
              ref="emailInputRef"
              v-model="form.email"
              :label="t('fields.email')"
              type="email"
              autocomplete="email"
              :error="Boolean(fieldErrors.email)"
              :error-message="fieldErrors.email"
              :loading="checkingEmail"
              @blur="checkEmailAvailability"
            />

            <q-input
              v-model="form.password"
              :label="t('fields.password')"
              type="password"
              autocomplete="new-password"
              :error="Boolean(fieldErrors.password)"
              :error-message="fieldErrors.password"
            />

            <q-input
              v-model="form.display_name"
              :label="t('fields.displayName')"
              :error="Boolean(fieldErrors.display_name)"
              :error-message="fieldErrors.display_name"
            />

            <q-banner v-if="error" class="bg-negative text-white" rounded>
              {{ error }}
            </q-banner>

            <q-btn
              type="submit"
              color="primary"
              :label="t('actions.register')"
              :loading="loading || checkingEmail"
              class="full-width"
              no-caps
            />

            <q-btn flat no-caps to="/login" :label="t('register.haveAccount')" />
          </q-form>
        </q-card-section>
      </template>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import Client, { Local, type registration } from '@/encore/client';
import { RegistrationRegisterParamsSchema } from '@/encore/zod';

type Focusable = { focus: () => void };

const { t } = useI18n();
const client = new Client(Local);
const emailInputRef = ref<Focusable | null>(null);
const loading = ref(false);
const checkingEmail = ref(false);
const error = ref<string | null>(null);
const success = ref(false);
const fieldErrors = reactive<Partial<Record<keyof registration.RegisterParams, string>>>({});
const form = reactive<registration.RegisterParams>({
  email: '',
  password: '',
  display_name: '',
  avatar_url: '',
});

onMounted(async () => {
  await nextTick();
  emailInputRef.value?.focus();
});

async function onSubmit() {
  const params = validateForm();
  if (!params) return;
  const emailAvailable = await checkEmailAvailability();
  if (!emailAvailable) return;

  loading.value = true;
  error.value = null;
  success.value = false;
  try {
    await client.registration.Register(params);
    success.value = true;
    resetForm();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function checkEmailAvailability(): Promise<boolean> {
  const emailResult = RegistrationRegisterParamsSchema.shape.email.safeParse(form.email);
  if (!emailResult.success) return false;

  checkingEmail.value = true;
  fieldErrors.email = '';
  try {
    const res = await client.registration.CheckEmail(form.email);
    if (!res.mx_valid) {
      fieldErrors.email = t('register.emailDomainInvalid');
      return false;
    }
    if (!res.available) {
      fieldErrors.email = t('register.emailUnavailable');
      return false;
    }
    return true;
  } catch (err) {
    fieldErrors.email = err instanceof Error ? err.message : String(err);
    return false;
  } finally {
    checkingEmail.value = false;
  }
}

function validateForm(): registration.RegisterParams | null {
  clearFieldErrors();
  const result = RegistrationRegisterParamsSchema.safeParse({ ...form });
  if (result.success) return result.data;

  const flattened = result.error.flatten().fieldErrors;
  for (const key of Object.keys(flattened) as (keyof registration.RegisterParams)[]) {
    fieldErrors[key] = flattened[key]?.[0] ?? '';
  }
  return null;
}

function clearFieldErrors() {
  fieldErrors.email = '';
  fieldErrors.password = '';
  fieldErrors.display_name = '';
  fieldErrors.avatar_url = '';
}

function resetForm() {
  form.email = '';
  form.password = '';
  form.display_name = '';
  form.avatar_url = '';
}
</script>
