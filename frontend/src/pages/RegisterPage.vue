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
              ref="displayNameInputRef"
              v-model="form.display_name"
              :label="t('fields.displayName')"
              autocomplete="name"
              :error="Boolean(fieldErrors.display_name)"
              :error-message="fieldErrors.display_name"
            />

            <q-input
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
              :type="showPassword ? 'text' : 'password'"
              autocomplete="new-password"
              :error="Boolean(fieldErrors.password)"
              :error-message="fieldErrors.password"
            >
              <template v-slot:append>
                <q-icon
                  :name="showPassword ? 'visibility' : 'visibility_off'"
                  class="cursor-pointer"
                  @mousedown="showPassword = true"
                  @mouseup="showPassword = false"
                  @mouseleave="showPassword = false"
                  @touchstart.prevent="showPassword = true"
                  @touchend="showPassword = false"
                  @touchcancel="showPassword = false"
                />
              </template>
            </q-input>

            <q-input
              v-model="form.confirm_password"
              :label="t('fields.confirmPassword')"
              :type="showConfirmPassword ? 'text' : 'password'"
              autocomplete="new-password"
              :error="Boolean(fieldErrors.confirm_password)"
              :error-message="fieldErrors.confirm_password"
            >
              <template v-slot:append>
                <q-icon
                  :name="showConfirmPassword ? 'visibility' : 'visibility_off'"
                  class="cursor-pointer"
                  @mousedown="showConfirmPassword = true"
                  @mouseup="showConfirmPassword = false"
                  @mouseleave="showConfirmPassword = false"
                  @touchstart.prevent="showConfirmPassword = true"
                  @touchend="showConfirmPassword = false"
                  @touchcancel="showConfirmPassword = false"
                />
              </template>
            </q-input>

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
const displayNameInputRef = ref<Focusable | null>(null);
const loading = ref(false);
const checkingEmail = ref(false);
const error = ref<string | null>(null);
const success = ref(false);
const showPassword = ref(false);
const showConfirmPassword = ref(false);

const fieldErrors = reactive<Partial<Record<keyof registration.RegisterParams | 'confirm_password', string>>>({});
const form = reactive({
  display_name: '',
  email: '',
  password: '',
  confirm_password: '',
  avatar_url: '',
});

onMounted(async () => {
  await nextTick();
  displayNameInputRef.value?.focus();
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
  const { confirm_password, ...payload } = form;
  const result = RegistrationRegisterParamsSchema.safeParse(payload);
  if (!result.success) {
    const flattened = result.error.flatten().fieldErrors;
    for (const key of Object.keys(flattened) as (keyof registration.RegisterParams)[]) {
      fieldErrors[key] = flattened[key]?.[0] ?? '';
    }
    return null;
  }
  if (!confirm_password || form.password !== confirm_password) {
    fieldErrors.confirm_password = t('register.passwordMismatch');
    return null;
  }
  return result.data;
}

function clearFieldErrors() {
  fieldErrors.display_name = '';
  fieldErrors.email = '';
  fieldErrors.password = '';
  fieldErrors.confirm_password = '';
  fieldErrors.avatar_url = '';
}

function resetForm() {
  form.display_name = '';
  form.email = '';
  form.password = '';
  form.confirm_password = '';
  form.avatar_url = '';
}
</script>
