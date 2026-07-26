<template>
  <q-dialog v-model="open" @show="focusFirstField">
    <q-card style="min-width: 350px">
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
        <q-tab name="profile" :label="t('admin.profile')" />
        <q-tab name="personal" :label="t('admin.personalData')" />
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
      </q-tab-panels>

      <q-card-section v-if="adminStore.error">
        <q-banner class="bg-negative text-white" rounded>{{ adminStore.error }}</q-banner>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn v-close-popup flat :label="t('actions.cancel')" />
        <q-btn color="primary" :label="t('actions.save')" :loading="adminStore.loading" @click="save" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAdminStore, type PersonalDataParams, type Profile, type ProfileParams } from '@/stores/admin-store';
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
type EditTab = 'profile' | 'personal';
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
  if (!props.profile) return;
  try {
    if (editTab.value === 'profile') {
      const parsed = validateEditForm();
      if (!parsed) return;
      await adminStore.updateProfile(props.profile.user_id, parsed);
    } else {
      const parsed = validatePersonalForm();
      if (!parsed) return;
      await adminStore.upsertPersonalData(props.profile.user_id, parsed);
    }
    open.value = false;
  } catch {
    // adminStore.error already holds the failure message
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
  } else {
    firstNameInputRef.value?.focus();
  }
}
</script>
