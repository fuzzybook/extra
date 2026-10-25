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
            <label class="account-label" for="register-name">{{ t('login.displayNameLabel') }}</label>
            <input
              id="register-name"
              ref="registerNameInput"
              v-model.trim="registerName"
              class="account-input"
              type="text"
              autocomplete="nickname"
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
import { nextTick, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import { APIError, ErrCode } from '@/encore/client';
import { api } from '@/encore/api';

const { t } = useI18n();
const router = useRouter();

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
    await router.push('/dashboard');
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
const registerEmail = ref('');
const registerName = ref('');
const registerPassword = ref('');
const registerConfirm = ref('');
const registerDetails = ref(false);
const registerLoading = ref(false);
const registerError = ref('');
const registerDone = ref(false);
const registerNameInput = ref<HTMLInputElement | null>(null);

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
        registerNameInput.value?.focus();
      }
      return;
    }
    if (registerPassword.value !== registerConfirm.value) {
      registerError.value = t('login.passwordMismatch');
      return;
    }
    await api.registration.Register({
      email: registerEmail.value,
      display_name: registerName.value,
      avatar_url: '',
      password: registerPassword.value,
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
