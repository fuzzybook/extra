<template>
  <q-page class="q-pa-md">
    <div class="text-h5 q-mb-md">{{ t('profile.title') }}</div>

    <q-banner v-if="profilesStore.error" class="bg-negative text-white q-mb-md" rounded>
      {{ profilesStore.error }}
    </q-banner>

    <q-card v-if="profilesStore.profile" flat bordered style="max-width: 560px">
      <q-card-section class="row items-center q-gutter-sm">
        <q-avatar size="48px">
          <img v-if="profileForm.avatar_url" :src="profileForm.avatar_url" />
          <q-icon v-else name="person" />
        </q-avatar>
        <div>
          <div class="text-subtitle1">{{ profilesStore.profile.email }}</div>
          <div class="text-subtitle1">{{ displayValue(profileForm.display_name) }}</div>
        </div>
      </q-card-section>

      <q-tabs v-model="profileTab" align="left" class="text-primary" dense>
        <q-tab v-if="isEditMode" name="avatar" :label="t('profile.avatar')" />
        <q-tab name="profile" :label="t('admin.profile')" />
        <q-tab name="personal" :label="t('admin.personalData')" />
        <q-tab v-if="isEditMode" name="password" :label="t('profile.password')" />
      </q-tabs>
      <q-separator />

      <q-tab-panels v-model="profileTab" animated>
        <q-tab-panel v-if="isEditMode" name="avatar" class="q-gutter-md">
          <AvatarUpload v-model="profileForm.avatar_url" :uploader="profilesStore.uploadAvatar" />
          <div v-if="profileFormErrors.avatar_url" class="text-negative text-caption">
            {{ profileFormErrors.avatar_url }}
          </div>
        </q-tab-panel>

        <q-tab-panel name="profile" class="q-gutter-md">
          <q-list separator>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.displayName') }}</q-item-label>
                <q-item-label>{{ displayValue(profileForm.display_name) }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.email') }}</q-item-label>
                <q-item-label>{{ profilesStore.profile.email }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.role') }}</q-item-label>
                <q-item-label>{{ profilesStore.profile.role }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.status') }}</q-item-label>
                <q-item-label>
                  <ProfileStatusBadge :status="profilesStore.profile.status" />
                </q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <q-tab-panel name="personal" class="q-gutter-md">
          <template v-if="isEditMode">
            <q-input
              ref="firstNameInputRef"
              v-model="personalForm.first_name"
              :label="t('fields.firstName')"
              :error="Boolean(personalFormErrors.first_name)"
              :error-message="personalFormErrors.first_name"
            />
            <q-input
              v-model="personalForm.last_name"
              :label="t('fields.lastName')"
              :error="Boolean(personalFormErrors.last_name)"
              :error-message="personalFormErrors.last_name"
            />
            <q-input
              v-model="personalForm.address"
              :label="t('fields.address')"
              :error="Boolean(personalFormErrors.address)"
              :error-message="personalFormErrors.address"
            />
            <q-input
              v-model="personalForm.cap"
              :label="t('fields.cap')"
              :error="Boolean(personalFormErrors.cap)"
              :error-message="personalFormErrors.cap"
            />
            <q-input
              v-model="personalForm.city"
              :label="t('fields.city')"
              :error="Boolean(personalFormErrors.city)"
              :error-message="personalFormErrors.city"
            />
            <q-select
              v-model="personalForm.country"
              :options="filteredCountries"
              option-label="label"
              option-value="value"
              emit-value
              map-options
              use-input
              clearable
              input-debounce="200"
              :label="t('fields.country')"
              :error="Boolean(personalFormErrors.country)"
              :error-message="personalFormErrors.country"
              @filter="filterCountries"
            />
          </template>

          <q-list v-else separator>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.firstName') }}</q-item-label>
                <q-item-label>{{ displayValue(personalForm.first_name) }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.lastName') }}</q-item-label>
                <q-item-label>{{ displayValue(personalForm.last_name) }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.address') }}</q-item-label>
                <q-item-label>{{ displayValue(personalForm.address) }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.cap') }}</q-item-label>
                <q-item-label>{{ displayValue(personalForm.cap) }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.city') }}</q-item-label>
                <q-item-label>{{ displayValue(personalForm.city) }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{ t('fields.country') }}</q-item-label>
                <q-item-label>{{ displayValue(countryLabel) }}</q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <q-tab-panel v-if="isEditMode" name="password" class="q-gutter-md">
          <q-input
            ref="currentPasswordInputRef"
            v-model="passwordForm.current_password"
            :label="t('profile.currentPassword')"
            :type="showCurrentPassword ? 'text' : 'password'"
            autocomplete="current-password"
            :error="Boolean(passwordFormErrors.current_password)"
            :error-message="passwordFormErrors.current_password"
            outlined
          >
            <template v-slot:append>
              <q-icon
                :name="showCurrentPassword ? 'visibility_off' : 'visibility'"
                class="cursor-pointer"
                @click="showCurrentPassword = !showCurrentPassword"
              />
            </template>
          </q-input>

          <q-input
            v-model="passwordForm.password"
            :label="t('profile.newPassword')"
            :type="showNewPassword ? 'text' : 'password'"
            autocomplete="new-password"
            :error="Boolean(passwordFormErrors.password)"
            :error-message="passwordFormErrors.password"
            outlined
          >
            <template v-slot:append>
              <q-icon
                :name="showNewPassword ? 'visibility_off' : 'visibility'"
                class="cursor-pointer"
                @click="showNewPassword = !showNewPassword"
              />
            </template>
          </q-input>

          <q-input
            v-model="passwordConfirmation"
            :label="t('profile.confirmNewPassword')"
            :type="showPasswordConfirmation ? 'text' : 'password'"
            autocomplete="new-password"
            :error="Boolean(passwordConfirmationError)"
            :error-message="passwordConfirmationError"
            outlined
          >
            <template v-slot:append>
              <q-icon
                :name="showPasswordConfirmation ? 'visibility_off' : 'visibility'"
                class="cursor-pointer"
                @click="showPasswordConfirmation = !showPasswordConfirmation"
              />
            </template>
          </q-input>
        </q-tab-panel>
      </q-tab-panels>

      <q-card-actions v-if="isEditMode" align="right">
        <q-btn flat :label="t('actions.cancel')" @click="cancelEdit" />
        <q-btn color="primary" :label="t('actions.save')" :loading="profilesStore.loading" @click="save" />
      </q-card-actions>
    </q-card>

    <div v-else-if="!profilesStore.isAuthenticated" class="text-grey">
      {{ t('profile.loginRequired') }}
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Notify } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { countries } from '@/data/countries';
import { AuthSetPasswordParamsSchema, ProfilesPersonalDataParamsSchema, ProfilesProfileParamsSchema } from '@/encore/zod';
import {
  useProfilesStore,
  type PersonalDataParams,
  type Profile,
  type ProfileParams,
  type SetPasswordParams,
} from '@/stores/profiles-store';
import AvatarUpload from '@/components/AvatarUpload.vue';
import ProfileStatusBadge from '@/components/ProfileStatusBadge.vue';

const { t } = useI18n();
const profilesStore = useProfilesStore();
const route = useRoute();
const router = useRouter();

type CountryOption = { label: string; value: string };
type ProfileTab = 'avatar' | 'profile' | 'personal' | 'password';
type Focusable = { focus: () => void };

const countryOptions: CountryOption[] = countries.map((c) => {
  const [value, label] = Object.entries(c)[0] as [string, string];
  return { label, value };
});

const filteredCountries = ref<CountryOption[]>(countryOptions);
const profileTab = ref<ProfileTab>('profile');
const isEditMode = computed(() => route.query.mode === 'edit');
const countryLabel = computed(() => countryOptions.find((c) => c.value === personalForm.country)?.label ?? personalForm.country);
const firstNameInputRef = ref<Focusable | null>(null);
const currentPasswordInputRef = ref<Focusable | null>(null);

const profileForm = reactive<ProfileParams>({
  display_name: '',
  avatar_url: '',
});
const profileFormErrors = reactive<Partial<Record<keyof ProfileParams, string>>>({});

const personalForm = reactive<PersonalDataParams>({
  first_name: '',
  last_name: '',
  address: '',
  cap: '',
  city: '',
  country: '',
});
const personalFormErrors = reactive<Partial<Record<keyof PersonalDataParams, string>>>({});

const passwordForm = reactive<SetPasswordParams>({
  current_password: '',
  password: '',
});
const passwordFormErrors = reactive<Partial<Record<keyof SetPasswordParams, string>>>({});
const passwordConfirmation = ref('');
const passwordConfirmationError = ref('');
const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showPasswordConfirmation = ref(false);

watch(
  () => profilesStore.profile,
  (profile) => {
    if (profile) {
      void loadProfile(profile);
    }
  },
  { immediate: true },
);

watch(profileTab, () => {
  if (isEditMode.value) {
    void focusFirstField();
  }
});

watch(isEditMode, (editing) => {
  if (editing) {
    profileTab.value = 'avatar';
    resetPasswordForm();
    void focusFirstField();
  } else if (profilesStore.profile) {
    profileTab.value = profileTab.value === 'avatar' || profileTab.value === 'password' ? 'profile' : profileTab.value;
    void loadProfile(profilesStore.profile);
  }
});

function filterCountries(val: string, update: (cb: () => void) => void) {
  update(() => {
    const needle = val.toLowerCase();
    filteredCountries.value = needle
      ? countryOptions.filter(
          (c) => c.label.toLowerCase().includes(needle) || c.value.toLowerCase().includes(needle),
        )
      : countryOptions;
  });
}

async function loadProfile(profile: Profile) {
  profileForm.display_name = profile.display_name;
  profileForm.avatar_url = profile.avatar_url;
  clearProfileFormErrors();
  resetPersonalForm();

  try {
    const data = profilesStore.personalData ?? (await profilesStore.fetchPersonalData());
    resetPersonalForm(data);
  } catch {
    // No personal data yet (404) - leave the form empty for creation.
    profilesStore.error = null;
  }
}

async function save() {
  try {
    if (profileTab.value === 'avatar') {
      const parsed = validateProfileForm();
      if (!parsed) return;
      await profilesStore.update(parsed);
    } else if (profileTab.value === 'personal') {
      const parsed = validatePersonalForm();
      if (!parsed) return;
      await profilesStore.savePersonalData(parsed);
      Notify.create({ type: 'positive', message: t('admin.notifications.personalDataUpdated') });
    } else if (profileTab.value === 'password') {
      const parsed = validatePasswordForm();
      if (!parsed) return;
      await profilesStore.setPassword(parsed);
      resetPasswordForm();
    }
    await router.replace('/profile');
  } catch {
    // profilesStore.error already holds the failure message
  }
}

async function cancelEdit() {
  if (profilesStore.profile) {
    await loadProfile(profilesStore.profile);
  }
  await router.replace('/profile');
}

function resetPersonalForm(data?: PersonalDataParams) {
  personalForm.first_name = data?.first_name ?? '';
  personalForm.last_name = data?.last_name ?? '';
  personalForm.address = data?.address ?? '';
  personalForm.cap = data?.cap ?? '';
  personalForm.city = data?.city ?? '';
  personalForm.country = data?.country ?? '';
  clearPersonalFormErrors();
}

function validateProfileForm(): ProfileParams | null {
  clearProfileFormErrors();
  const result = ProfilesProfileParamsSchema.safeParse({ ...profileForm });
  if (result.success) return result.data;

  const fieldErrors = result.error.flatten().fieldErrors;
  for (const key of Object.keys(fieldErrors) as (keyof ProfileParams)[]) {
    profileFormErrors[key] = fieldErrors[key]?.[0] ?? '';
  }
  return null;
}

function clearProfileFormErrors() {
  profileFormErrors.display_name = '';
  profileFormErrors.avatar_url = '';
}

function validatePersonalForm(): PersonalDataParams | null {
  clearPersonalFormErrors();
  const result = ProfilesPersonalDataParamsSchema.safeParse({ ...personalForm });
  if (result.success) return result.data;

  const fieldErrors = result.error.flatten().fieldErrors;
  for (const key of Object.keys(fieldErrors) as (keyof PersonalDataParams)[]) {
    personalFormErrors[key] = fieldErrors[key]?.[0] ?? '';
  }
  return null;
}

function clearPersonalFormErrors() {
  personalFormErrors.first_name = '';
  personalFormErrors.last_name = '';
  personalFormErrors.address = '';
  personalFormErrors.cap = '';
  personalFormErrors.city = '';
  personalFormErrors.country = '';
}

function validatePasswordForm(): SetPasswordParams | null {
  clearPasswordFormErrors();
  const result = AuthSetPasswordParamsSchema.safeParse({ ...passwordForm });
  if (!result.success) {
    const fieldErrors = result.error.flatten().fieldErrors;
    for (const key of Object.keys(fieldErrors) as (keyof SetPasswordParams)[]) {
      passwordFormErrors[key] = fieldErrors[key]?.[0] ?? '';
    }
    return null;
  }

  if (passwordForm.password !== passwordConfirmation.value) {
    passwordConfirmationError.value = t('profile.passwordMismatch');
    return null;
  }

  return result.data;
}

function clearPasswordFormErrors() {
  passwordFormErrors.current_password = '';
  passwordFormErrors.password = '';
  passwordConfirmationError.value = '';
}

function resetPasswordForm() {
  passwordForm.current_password = '';
  passwordForm.password = '';
  passwordConfirmation.value = '';
  showCurrentPassword.value = false;
  showNewPassword.value = false;
  showPasswordConfirmation.value = false;
  clearPasswordFormErrors();
}

function displayValue(value: string | number | null | undefined) {
  const normalized = String(value ?? '').trim();
  return normalized || '-';
}

async function focusFirstField() {
  await nextTick();
  if (profileTab.value === 'personal') {
    firstNameInputRef.value?.focus();
  } else if (profileTab.value === 'password') {
    currentPasswordInputRef.value?.focus();
  }
}

onMounted(async () => {
  if (profilesStore.isAuthenticated && !profilesStore.profile) {
    try {
      await profilesStore.me();
    } catch {
      // profilesStore.error already holds the failure message
    }
  }
});
</script>
