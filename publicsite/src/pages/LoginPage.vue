<template>
  <q-page class="account-page">
    <div class="account-grid">
      <!-- Login / recupero password -->
      <section class="account-panel" aria-labelledby="login-title">
        <h1 id="login-title" class="account-title">
          {{ resetMode ? t('login.resetTitle') : t('login.title') }}
        </h1>

        <form v-if="!resetMode" class="account-form" novalidate @submit.prevent="onLogin">
          <label class="account-label" for="login-email">{{ t('login.emailLabel') }}</label>
          <input
            id="login-email"
            v-model.trim="loginEmail"
            class="account-input"
            type="email"
            autocomplete="username"
            :placeholder="t('login.emailPlaceholder')"
            required
          />

          <label class="account-label" for="login-password">{{ t('login.passwordLabel') }}</label>
          <input
            id="login-password"
            v-model="loginPassword"
            class="account-input"
            type="password"
            autocomplete="current-password"
            required
          />

          <p v-if="loginError" class="account-message account-message--error" role="alert">{{ loginError }}</p>

          <button class="account-button" type="submit" :disabled="loginLoading">
            {{ t('login.button') }}
          </button>

          <button class="account-link" type="button" @click="openReset">
            {{ t('login.forgotPassword') }}
          </button>
        </form>

        <form v-else class="account-form" novalidate @submit.prevent="onReset">
          <p class="account-hint">{{ t('login.resetHint') }}</p>

          <label class="account-label" for="reset-email">{{ t('login.emailLabel') }}</label>
          <input
            id="reset-email"
            v-model.trim="resetEmail"
            class="account-input"
            type="email"
            autocomplete="email"
            :placeholder="t('login.emailPlaceholder')"
            required
          />

          <p v-if="resetError" class="account-message account-message--error" role="alert">{{ resetError }}</p>
          <p v-if="resetSent" class="account-message" role="status">{{ t('login.resetSent') }}</p>

          <button class="account-button" type="submit" :disabled="resetLoading">
            {{ t('login.resetButton') }}
          </button>

          <button class="account-link" type="button" @click="resetMode = false">
            {{ t('login.backToLogin') }}
          </button>
        </form>
      </section>

      <!-- Registrazione: prima l'email, poi nome e password -->
      <section class="account-panel" aria-labelledby="register-title">
        <h2 id="register-title" class="account-title">{{ t('login.registerTitle') }}</h2>

        <p v-if="registerDone" class="account-message" role="status">{{ t('login.registerSuccess') }}</p>

        <form v-else class="account-form" novalidate @submit.prevent="onRegister">
          <label class="account-label" for="register-email">{{ t('login.emailLabel') }}</label>
          <input
            id="register-email"
            v-model.trim="registerEmail"
            class="account-input"
            type="email"
            autocomplete="email"
            :placeholder="t('login.emailPlaceholder')"
            required
            @input="registerDetails = false"
          />

          <template v-if="registerDetails">
            <label class="account-label" for="register-language">{{ t('login.languageLabel') }}</label>
            <select id="register-language" ref="registerLanguageInput" v-model="registerLanguage" class="account-input account-select">
              <option v-for="lang in LANGUAGES" :key="lang.value" :value="lang.value">{{ lang.label }}</option>
            </select>

            <!-- Status: tutte le opzioni visibili -->
            <fieldset class="account-radios">
              <legend class="account-label">{{ t('login.statusLabel') }}</legend>
              <label v-for="type in MEMBER_TYPES" :key="type" class="account-check account-radio">
                <input v-model="registerStatus" type="radio" name="register-status" :value="type" required />
                <span>{{ t(`login.memberTypes.${type}`) }}</span>
              </label>
            </fieldset>
            <p v-if="registerStatus === 'broadcaster'" class="account-field-hint">{{ t('login.broadcasterHint') }}</p>
            <p v-else-if="contactRequired" class="account-field-hint">{{ t('login.contactHint') }}</p>

            <label class="account-label" for="register-nickname">{{ t('login.nicknameLabel') }}</label>
            <input
              id="register-nickname"
              v-model.trim="registerNickname"
              class="account-input"
              type="text"
              autocomplete="nickname"
              minlength="2"
              maxlength="32"
              required
            />

            <label class="account-label" for="register-name">{{ t('login.nameLabel') }}</label>
            <input
              id="register-name"
              v-model.trim="registerPrivateName"
              class="account-input"
              type="text"
              autocomplete="name"
              maxlength="120"
              required
            />

            <label class="account-label" for="register-password">{{ t('login.passwordLabel') }}</label>
            <input
              id="register-password"
              v-model="registerPassword"
              class="account-input"
              type="password"
              autocomplete="new-password"
              required
            />

            <label class="account-label" for="register-confirm">{{ t('login.confirmPasswordLabel') }}</label>
            <input
              id="register-confirm"
              v-model="registerConfirm"
              class="account-input"
              type="password"
              autocomplete="new-password"
              required
            />

            <label class="account-label" for="register-phone">
              {{ t('login.phoneLabel') }} <span v-if="!contactRequired" class="account-optional">{{ t('login.optional') }}</span>
            </label>
            <input
              id="register-phone"
              v-model.trim="registerPhone"
              class="account-input"
              type="tel"
              autocomplete="tel"
              maxlength="30"
              :required="contactRequired"
            />

            <label class="account-label" for="register-country">
              {{ t('login.countryLabel') }} <span v-if="!contactRequired" class="account-optional">{{ t('login.optional') }}</span>
            </label>
            <select
              id="register-country"
              v-model="registerCountry"
              class="account-input account-select"
              autocomplete="country"
              :required="contactRequired"
            >
              <option value=""></option>
              <option v-for="c in countries" :key="c.code" :value="c.code">{{ c.name }}</option>
            </select>

            <label class="account-label" for="register-address">
              {{ t('login.addressLabel') }} <span v-if="!contactRequired" class="account-optional">{{ t('login.optional') }}</span>
            </label>
            <textarea
              id="register-address"
              v-model="registerAddress"
              class="account-input account-textarea"
              autocomplete="street-address"
              maxlength="500"
              rows="3"
              :required="contactRequired"
            />

            <label class="account-check">
              <input v-model="registerNewsletter" type="checkbox" />
              <span>{{ t('login.newsletter') }}</span>
            </label>

            <label class="account-check">
              <input v-model="registerTerms" type="checkbox" required />
              <i18n-t keypath="login.acceptTerms" tag="span" scope="global">
                <template #terms>
                  <a :href="termsHref" target="_blank" rel="noopener" class="account-inline-link">{{ t('login.termsLink') }}</a>
                </template>
              </i18n-t>
            </label>
          </template>

          <p v-if="registerError" class="account-message account-message--error" role="alert">{{ registerError }}</p>

          <button class="account-button" type="submit" :disabled="registerLoading">
            {{ t('login.registerButton') }}
          </button>
        </form>
      </section>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import { APIError, ErrCode } from '@/encore/client';
import { api } from '@/encore/api';
import { useSession } from '@/composables/useSession';
import { countryOptions } from '@/utils/countries';

const { t, locale } = useI18n();
const router = useRouter();
const session = useSession();
const route = useRoute();

useMeta(() => ({
  title: `${t('login.title')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: t('login.subtitle'),
    },
  },
}));

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// Messaggio dell'errore del backend solo per gli input non validi (es. regole
// della password); per il resto un messaggio generico.
function errorMessage(err: unknown): string {
  if (err instanceof APIError && err.code === ErrCode.InvalidArgument && err.message) {
    return err.message;
  }
  return t('login.genericError');
}

// --- Login ---
const loginEmail = ref('');
const loginPassword = ref('');
const loginLoading = ref(false);
const loginError = ref('');

// Dopo il login si torna alla pagina di provenienza (es. una diretta), solo
// se è un percorso interno del sito.
function redirectTarget(): string {
  const target = typeof route.query.redirect === 'string' ? route.query.redirect : '';
  return target.startsWith('/') && !target.startsWith('//') && !target.includes('\\') ? target : '/dashboard';
}

async function onLogin() {
  loginError.value = '';
  if (!EMAIL_PATTERN.test(loginEmail.value) || !loginPassword.value) {
    loginError.value = t('login.invalidCredentials');
    return;
  }
  loginLoading.value = true;
  try {
    // La sessione arriva come cookie HttpOnly.
    await api.profiles.Login({ user_email: loginEmail.value, password: loginPassword.value });
    loginPassword.value = '';
    await session.refresh();
    await router.push(redirectTarget());
  } catch (err) {
    loginError.value =
      err instanceof APIError && err.code === ErrCode.Unauthenticated ? t('login.invalidCredentials') : errorMessage(err);
  } finally {
    loginLoading.value = false;
  }
}

// --- Recupero password ---
const resetMode = ref(false);
const resetEmail = ref('');
const resetLoading = ref(false);
const resetError = ref('');
const resetSent = ref(false);

function openReset() {
  resetEmail.value = loginEmail.value;
  resetError.value = '';
  resetSent.value = false;
  resetMode.value = true;
}

async function onReset() {
  resetError.value = '';
  resetSent.value = false;
  if (!EMAIL_PATTERN.test(resetEmail.value)) {
    resetError.value = t('login.emailInvalid');
    return;
  }
  resetLoading.value = true;
  try {
    await api.registration.RequestPasswordReset({ email: resetEmail.value });
    // Stessa risposta che l'email esista o no.
    resetSent.value = true;
  } catch (err) {
    resetError.value = errorMessage(err);
  } finally {
    resetLoading.value = false;
  }
}

// --- Registrazione ---
// Lingue del sito (nomi nella propria lingua) e tipi di membro richiedibili.
const LANGUAGES = [
  { value: 'en', label: 'English' },
  { value: 'it', label: 'Italiano' },
  { value: 'fr', label: 'Français' },
  { value: 'de', label: 'Deutsch' },
] as const;
const MEMBER_TYPES = ['visitor', 'broadcaster', 'affiliate', 'advertiser', 'club'] as const;
// Inserzionisti e club devono indicare telefono, paese e indirizzo.
const CONTACT_REQUIRED_TYPES: readonly string[] = ['advertiser', 'club'];
const PHONE_PATTERN = /^\+?[0-9 ()./-]{5,30}$/;

const siteLanguage = computed(() => {
  const lang = String(route.params.locale || locale.value).slice(0, 2);
  return LANGUAGES.some((l) => l.value === lang) ? lang : 'en';
});
const termsHref = computed(() => `/${siteLanguage.value}/terms`);
const countries = computed(() => countryOptions(siteLanguage.value));

const registerEmail = ref('');
const registerLanguage = ref<string>(siteLanguage.value);
const registerStatus = ref('');
const registerNickname = ref('');
const registerPrivateName = ref('');
const registerPassword = ref('');
const registerConfirm = ref('');
const registerPhone = ref('');
const registerCountry = ref('');
const registerAddress = ref('');
const registerNewsletter = ref(false);
const registerTerms = ref(false);
const registerDetails = ref(false);
const contactRequired = computed(() => CONTACT_REQUIRED_TYPES.includes(registerStatus.value));
const registerLoading = ref(false);
const registerError = ref('');
const registerDone = ref(false);
const registerLanguageInput = ref<HTMLSelectElement | null>(null);

// Controlli lato client del secondo passo; il backend li ripete.
function registerFormError(): string {
  if (!registerStatus.value) return t('login.statusRequired');
  if (registerPassword.value !== registerConfirm.value) return t('login.passwordMismatch');
  if (contactRequired.value && (!registerPhone.value || !registerCountry.value || !registerAddress.value.trim())) {
    return t('login.contactRequired');
  }
  if (registerPhone.value && !PHONE_PATTERN.test(registerPhone.value)) return t('login.phoneInvalid');
  if (!registerTerms.value) return t('login.termsRequired');
  return '';
}

async function onRegister() {
  registerError.value = '';
  if (!EMAIL_PATTERN.test(registerEmail.value)) {
    registerError.value = t('login.emailInvalid');
    return;
  }
  registerLoading.value = true;
  try {
    if (!registerDetails.value) {
      // Primo passo: verifica l'email, poi chiede nome e password.
      const res = await api.registration.CheckEmail(registerEmail.value);
      if (!res.mx_valid) {
        registerError.value = t('login.emailDomainInvalid');
      } else if (!res.available) {
        registerError.value = t('login.emailUnavailable');
      } else {
        registerDetails.value = true;
        await nextTick();
        registerLanguageInput.value?.focus();
      }
      return;
    }
    const formError = registerFormError();
    if (formError) {
      registerError.value = formError;
      return;
    }
    await api.registration.Register({
      email: registerEmail.value,
      display_name: registerNickname.value,
      avatar_url: '',
      password: registerPassword.value,
      language: registerLanguage.value,
      member_type: registerStatus.value,
      name: registerPrivateName.value,
      ...(registerPhone.value ? { phone: registerPhone.value } : {}),
      ...(registerCountry.value ? { country: registerCountry.value } : {}),
      ...(registerAddress.value.trim() ? { address: registerAddress.value.trim() } : {}),
      newsletter: registerNewsletter.value,
      accept_terms: registerTerms.value,
    });
    registerPassword.value = '';
    registerConfirm.value = '';
    registerDone.value = true;
  } catch (err) {
    registerError.value = errorMessage(err);
  } finally {
    registerLoading.value = false;
  }
}
</script>

<style scoped>
.account-page {
  min-height: calc(100vh - 180px);
  padding: 32px 16px;
  background-color: #1b1b1b;
  color: #ffffff;
  font-family: 'Open Sans', 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.account-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  max-width: 1100px;
  margin: 0 auto;
}

.account-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-height: 420px;
  padding: 56px 24px 48px;
  background-color: #000000;
}

.account-title {
  margin: 0 0 28px;
  font-size: 2.5rem;
  font-weight: 400;
  line-height: 1.2;
  letter-spacing: 0;
  text-align: center;
}

.account-form {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 320px;
}

.account-label {
  margin-bottom: 8px;
  font-size: 1.05rem;
}

.account-input {
  height: 46px;
  margin-bottom: 20px;
  padding: 0 10px;
  border: 3px solid #555555;
  border-radius: 0;
  background-color: #1a1a1a;
  color: #ffffff;
  font: inherit;
  font-size: 1rem;
  outline: none;
}

.account-input::placeholder {
  color: #8a8a8a;
}

.account-input:focus {
  border-color: #8f8f8f;
}

/* Select con la freccia piena a destra, come nel modello */
.account-select {
  padding-right: 40px;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='10' viewBox='0 0 12 10'%3E%3Cpath d='M0 0h12L6 10z' fill='%23cfcfcf'/%3E%3C/svg%3E");
  background-position: right 14px center;
  background-repeat: no-repeat;
  cursor: pointer;
}

.account-select option {
  background-color: #1a1a1a;
  color: #ffffff;
}

.account-textarea {
  height: auto;
  min-height: 90px;
  padding: 8px 10px;
  resize: vertical;
}

.account-optional {
  color: #9a9a9a;
  font-size: 0.8rem;
}

.account-field-hint {
  margin: -12px 0 20px;
  color: #9a9a9a;
  font-size: 0.85rem;
  line-height: 1.4;
}

.account-check {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 18px;
  font-size: 1.05rem;
  line-height: 1.4;
  cursor: pointer;
}

.account-check input {
  flex: none;
  width: 30px;
  height: 30px;
  margin: 0;
  border: 3px solid #555555;
  background-color: #1a1a1a;
  appearance: none;
  cursor: pointer;
}

.account-check input:checked {
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 16 16'%3E%3Cpath d='M2 8.5l4 4 8-9' fill='none' stroke='%23ffffff' stroke-width='2.5'/%3E%3C/svg%3E");
  background-position: center;
  background-repeat: no-repeat;
}

.account-check input:focus-visible {
  border-color: #8f8f8f;
  outline: none;
}

.account-check span {
  padding-top: 3px;
}

.account-radios {
  display: flex;
  flex-direction: column;
  margin: 0 0 20px;
  padding: 0;
  border: 0;
}

.account-radios legend {
  padding: 0;
}

.account-radio {
  margin-bottom: 10px;
}

.account-radio input {
  border-radius: 50%;
}

.account-radio input:checked {
  background-image: none;
  box-shadow: inset 0 0 0 6px #1a1a1a;
  background-color: #ffffff;
}

.account-inline-link {
  color: #ffffff;
  text-decoration: underline;
}

.account-button {
  height: 56px;
  margin-top: 4px;
  border: 3px solid #555555;
  border-radius: 0;
  background-color: #000000;
  color: #a6a6a6;
  font: inherit;
  font-size: 1.25rem;
  font-weight: 600;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.account-button:hover:not(:disabled),
.account-button:focus-visible {
  border-color: #8f8f8f;
  color: #ffffff;
  outline: none;
}

.account-button:disabled {
  cursor: wait;
  opacity: 0.6;
}

.account-link {
  align-self: center;
  margin-top: 36px;
  padding: 0;
  border: 0;
  background: none;
  color: #ffffff;
  font: inherit;
  font-size: 0.95rem;
  cursor: pointer;
}

.account-link:hover,
.account-link:focus-visible {
  text-decoration: underline;
  outline: none;
}

.account-hint {
  margin: 0 0 20px;
  color: #bdbdbd;
  font-size: 0.95rem;
  line-height: 1.5;
}

.account-message {
  max-width: 320px;
  margin: 0 0 16px;
  color: #d0d0d0;
  font-size: 0.95rem;
  line-height: 1.5;
  text-align: center;
}

.account-message--error {
  color: #ff8a80;
  text-align: left;
}

@media (max-width: 760px) {
  .account-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .account-panel {
    min-height: 0;
    padding: 40px 16px;
  }

  .account-title {
    font-size: 2rem;
  }
}
</style>
