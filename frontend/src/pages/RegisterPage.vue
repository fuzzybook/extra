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
              autocomplete="nickname"
              :error="Boolean(fieldErrors.display_name)"
              :error-message="fieldErrors.display_name"
            />

            <q-input
              v-model="form.name"
              :label="t('register.nameLabel')"
              autocomplete="name"
              :error="Boolean(fieldErrors.name)"
              :error-message="fieldErrors.name"
            />

            <q-select
              v-model="form.member_type"
              :options="memberTypeOptions"
              :label="t('register.memberTypeLabel')"
              emit-value
              map-options
              :hint="form.member_type === 'broadcaster' ? t('register.broadcasterHint') : undefined"
              :error="Boolean(fieldErrors.member_type)"
              :error-message="fieldErrors.member_type"
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

            <q-checkbox v-model="form.newsletter" :label="t('register.newsletter')" dense />

            <div>
              <q-checkbox v-model="form.accept_terms" :label="t('register.acceptTerms')" dense />
              <div v-if="fieldErrors.accept_terms" class="text-negative text-caption q-mt-xs">
                {{ fieldErrors.accept_terms }}
              </div>
            </div>

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
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type registration } from '@/encore/client';
import { api } from '@/encore/api';
import { RegistrationRegisterParamsSchema } from '@/encore/zod';

type Focusable = { focus: () => void };

const { t, locale } = useI18n();
const client = api;
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
  name: '',
  member_type: 'visitor',
  newsletter: false,
  accept_terms: false,
});

const MEMBER_TYPES = ['visitor', 'broadcaster', 'affiliate', 'advertiser'] as const;
const memberTypeOptions = computed(() =>
  MEMBER_TYPES.map((value) => ({ value, label: t(`register.memberTypes.${value}`) })),
);

// Lingua del profilo: quella dell'interfaccia, se il sito la supporta.
const SITE_LANGUAGES = ['en', 'it', 'fr', 'de'];
function profileLanguage(): string {
  const lang = String(locale.value).slice(0, 2).toLowerCase();
  return SITE_LANGUAGES.includes(lang) ? lang : 'en';
}

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
  const result = RegistrationRegisterParamsSchema.safeParse({ ...payload, language: profileLanguage() });
  if (!result.success) {
    const flattened = result.error.flatten().fieldErrors;
    for (const key of Object.keys(flattened) as (keyof registration.RegisterParams)[]) {
      fieldErrors[key] = flattened[key]?.[0] ?? '';
    }
    if (fieldErrors.accept_terms) {
      fieldErrors.accept_terms = t('register.termsRequired');
    }
    return null;
  }
  if (!confirm_password || form.password !== confirm_password) {
    fieldErrors.confirm_password = t('register.passwordMismatch');
    return null;
  }
  const { phone, country, address, ...required } = result.data;
  return {
    ...required,
    ...(phone ? { phone } : {}),
    ...(country ? { country } : {}),
    ...(address ? { address } : {}),
  };
}

function clearFieldErrors() {
  fieldErrors.display_name = '';
  fieldErrors.email = '';
  fieldErrors.password = '';
  fieldErrors.confirm_password = '';
  fieldErrors.avatar_url = '';
  fieldErrors.name = '';
  fieldErrors.member_type = '';
  fieldErrors.accept_terms = '';
}

function resetForm() {
  form.display_name = '';
  form.email = '';
  form.password = '';
  form.confirm_password = '';
  form.avatar_url = '';
  form.name = '';
  form.member_type = 'visitor';
  form.newsletter = false;
  form.accept_terms = false;
}
</script>
