<template>
  <q-dialog v-model="open" :persistent="saving" @show="focusFirstField">
    <q-card style="width: 560px; max-width: 100%">
      <q-card-section v-if="profile" class="row items-center q-gutter-sm">
        <q-avatar size="48px">
          <img v-if="editForm.avatar_url" :src="editForm.avatar_url" />
          <q-icon v-else name="person" />
        </q-avatar>
        <div>
          <div class="text-subtitle1">{{ profile.email }}</div>
          <div class="text-caption text-grey">{{ profile.role }} · {{ profile.user_id }}</div>
        </div>
      </q-card-section>

      <q-tabs v-model="editTab" align="left" class="text-primary" dense>
        <q-tab name="profile" :label="t('admin.profile')" :disable="saving" />
        <q-tab name="personal" :label="t('admin.personalData')" :disable="saving" />
        <q-tab name="settings" :label="t('personalSettings.title')" :disable="saving" />
      </q-tabs>
      <q-separator />

      <q-tab-panels v-model="editTab" animated>
        <q-tab-panel name="profile" class="q-gutter-md">
          <AvatarUpload v-model="editForm.avatar_url" :uploader="adminStore.uploadAvatar" />
          <div v-if="editFormErrors.avatar_url" class="text-negative text-caption">
            {{ editFormErrors.avatar_url }}
          </div>
          <q-input
            ref="displayNameInputRef"
            v-model="editForm.display_name"
            :label="t('fields.displayName')"
            :error="Boolean(editFormErrors.display_name)"
            :error-message="editFormErrors.display_name"
          />
        </q-tab-panel>

        <q-tab-panel name="personal" class="q-gutter-md">
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
        </q-tab-panel>
        <q-tab-panel name="settings" class="q-gutter-md">
          <q-linear-progress v-if="settingsLoading" indeterminate :aria-label="t('personalSettings.loading')" />
          <q-banner v-if="settingsError" class="bg-negative text-white" role="alert" rounded>
            {{ settingsError }}
            <template #action>
              <q-btn flat :label="t('personalSettings.retry')" :disable="settingsLoading" @click="loadSettings" />
            </template>
          </q-banner>
          <q-toggle
            v-model="settingsForm.can_broadcast"
            :label="t('personalSettings.canBroadcast')"
            :disable="!settingsLoaded || settingsLoading || saving"
          />
          <q-input
            v-if="settingsForm.can_broadcast"
            ref="roomNameInputRef"
            v-model="settingsForm.room_name"
            :label="t('personalSettings.roomName')"
            :hint="t('personalSettings.roomHint')"
            :error="Boolean(settingsRoomError)"
            :error-message="settingsRoomError"
            :disable="!settingsLoaded || settingsLoading || saving"
          />
          <p class="text-caption text-grey-7">{{ t('personalSettings.description') }}</p>
        </q-tab-panel>
      </q-tab-panels>

      <q-card-section v-if="adminStore.error">
        <q-banner class="bg-negative text-white" rounded>{{ adminStore.error }}</q-banner>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn v-close-popup flat :label="t('actions.cancel')" :disable="saving" />
        <q-btn color="primary" :label="t('actions.save')" :loading="saving"
          :disable="adminStore.loading || (editTab === 'settings' && (!settingsLoaded || settingsLoading))" @click="save" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Notify } from 'quasar';
import { useAdminStore, type PersonalDataParams, type PersonalSetupParams, type Profile, type ProfileParams } from '@/stores/admin-store';
import { countries } from '@/data/countries';
import { AdminPersonalDataParamsSchema, AdminProfileParamsSchema } from '@/encore/zod';
import AvatarUpload from '@/components/AvatarUpload.vue';

const props = defineProps<{
  profile: Profile | null;
}>();

const open = defineModel<boolean>({ required: true });
const adminStore = useAdminStore();
const { t } = useI18n();

type CountryOption = { label: string; value: string };
type EditTab = 'profile' | 'personal' | 'settings';
type Focusable = { focus: () => void };

const countryOptions: CountryOption[] = countries.map((c) => {
  const [value, label] = Object.entries(c)[0] as [string, string];
  return { label, value };
});

const filteredCountries = ref<CountryOption[]>(countryOptions);
const editTab = ref<EditTab>('profile');
const displayNameInputRef = ref<Focusable | null>(null);
const firstNameInputRef = ref<Focusable | null>(null);
const editForm = reactive<ProfileParams>({
  display_name: '',
  avatar_url: '',
});
const editFormErrors = reactive<Partial<Record<keyof ProfileParams, string>>>({});
const personalForm = reactive<PersonalDataParams>({
  first_name: '',
  last_name: '',
  address: '',
  cap: '',
  city: '',
  country: '',
});
const personalFormErrors = reactive<Partial<Record<keyof PersonalDataParams, string>>>({});
const settingsForm = reactive<PersonalSetupParams>({ can_broadcast: false, room_name: '' });
const settingsLoading = ref(false);
const settingsLoaded = ref(false);
const settingsError = ref('');
const settingsRoomError = ref('');
const roomNameInputRef = ref<Focusable | null>(null);
const saving = ref(false);
let settingsRequest = 0;

watch(() => [open.value, props.profile?.user_id] as const, () => {
  ++settingsRequest;
  settingsForm.can_broadcast = false;
  settingsForm.room_name = '';
  settingsLoaded.value = false;
  settingsLoading.value = false;
  settingsError.value = '';
  settingsRoomError.value = '';
  if (open.value && props.profile) void loadSettings();
}, { immediate: true });

watch(() => settingsForm.can_broadcast, () => { settingsRoomError.value = ''; });

async function loadSettings() {
  if (!props.profile || !open.value) return;
  const request = ++settingsRequest;
  settingsLoading.value = true;
  settingsLoaded.value = false;
  settingsError.value = '';
  try {
    const setup = await adminStore.getPersonalSetup(props.profile.user_id);
    if (request !== settingsRequest) return;
    settingsForm.can_broadcast = setup.can_broadcast;
    settingsForm.room_name = setup.room_name;
    settingsLoaded.value = true;
  } catch (error) {
    if (request === settingsRequest) settingsError.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (request === settingsRequest) settingsLoading.value = false;
  }
}

watch(
  () => [open.value, props.profile] as const,
  ([isOpen, profile]) => {
    if (isOpen && profile) {
      void loadProfile(profile);
    }
  },
  { immediate: true },
);

watch(editTab, () => {
  if (open.value) {
    void focusFirstField();
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
  editTab.value = 'profile';
  editForm.display_name = profile.display_name;
  editForm.avatar_url = profile.avatar_url;
  clearEditFormErrors();
  resetPersonalForm();
  try {
    const data = await adminStore.getPersonalData(profile.user_id);
    resetPersonalForm(data);
  } catch {
    // No personal data yet (404) — leave the form empty for creation.
    adminStore.error = null;
  }
}

async function save() {
  if (!props.profile || saving.value) return;
  saving.value = true;
  try {
    if (editTab.value === 'profile') {
      const parsed = validateEditForm();
      if (!parsed) return;
      await adminStore.updateProfile(props.profile.user_id, parsed);
      Notify.create({ type: 'positive', message: t('admin.notifications.profileUpdated') });
    } else if (editTab.value === 'personal') {
      const parsed = validatePersonalForm();
      if (!parsed) return;
      await adminStore.upsertPersonalData(props.profile.user_id, parsed);
      Notify.create({ type: 'positive', message: t('admin.notifications.personalDataUpdated') });
    } else {
      if (!settingsLoaded.value || settingsLoading.value) return;
      const roomName = settingsForm.can_broadcast ? settingsForm.room_name.trim() : '';
      settingsRoomError.value = '';
      if (settingsForm.can_broadcast && (!roomName || new TextEncoder().encode(roomName).length > 128 || /[\/\\\p{Cc}]/u.test(roomName))) {
        settingsRoomError.value = t('personalSettings.invalidRoom');
        return;
      }
      await adminStore.upsertPersonalSetup(props.profile.user_id, { can_broadcast: settingsForm.can_broadcast, room_name: roomName });
      Notify.create({ type: 'positive', message: t('personalSettings.saved') });
    }
    open.value = false;
  } catch {
    // adminStore.error already holds the failure message
  } finally {
    saving.value = false;
  }
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

function validateEditForm(): ProfileParams | null {
  clearEditFormErrors();
  const result = AdminProfileParamsSchema.safeParse({ ...editForm });
  if (result.success) return result.data;

  const fieldErrors = result.error.flatten().fieldErrors;
  for (const key of Object.keys(fieldErrors) as (keyof ProfileParams)[]) {
    editFormErrors[key] = fieldErrors[key]?.[0] ?? '';
  }
  return null;
}

function clearEditFormErrors() {
  editFormErrors.display_name = '';
  editFormErrors.avatar_url = '';
}

function validatePersonalForm(): PersonalDataParams | null {
  clearPersonalFormErrors();
  const result = AdminPersonalDataParamsSchema.safeParse({ ...personalForm });
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

async function focusFirstField() {
  await nextTick();
  if (editTab.value === 'profile') {
    displayNameInputRef.value?.focus();
  } else if (editTab.value === 'personal') {
    firstNameInputRef.value?.focus();
  } else {
    roomNameInputRef.value?.focus();
  }
}
</script>
