<template>
  <q-dialog :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)" @show="load">
    <q-card class="bg-white text-dark" style="width: 440px; max-width: 95vw">
      <q-card-section class="row items-center justify-between bg-primary text-white q-py-sm">
        <div class="text-subtitle1 text-weight-bold">{{ t('userProfile.personalTitle') }}</div>
        <q-btn v-close-popup flat round dense icon="close" :aria-label="t('brEditor.actions.cancel')" />
      </q-card-section>

      <q-card-section v-if="loading" class="row justify-center q-pa-xl">
        <q-spinner size="32px" color="primary" />
      </q-card-section>

      <q-form v-else @submit.prevent="save">
        <q-card-section class="q-gutter-sm">
          <!-- Un campo per riga -->
          <q-input v-model="form.first_name" :label="t('userProfile.firstName')" autocomplete="given-name" outlined maxlength="120" autofocus />
          <q-input v-model="form.last_name" :label="t('userProfile.lastName')" autocomplete="family-name" outlined maxlength="120" />
          <q-input v-model="form.address" :label="t('userProfile.address')" autocomplete="street-address" outlined maxlength="500" />
          <q-input v-model="form.cap" :label="t('userProfile.cap')" autocomplete="postal-code" outlined maxlength="10" />
          <q-input v-model="form.city" :label="t('userProfile.city')" autocomplete="address-level2" outlined maxlength="120" />
          <q-select
            v-model="form.country"
            :options="countries"
            :label="t('userProfile.country')"
            option-value="code"
            option-label="name"
            emit-value
            map-options
            outlined
            popup-content-class="bg-white text-dark"
          />
          <q-input
            v-if="phone"
            :model-value="phone"
            :label="t('userProfile.phone')"
            :hint="t('userProfile.phoneHint')"
            outlined
            readonly
          />
          <div v-if="error" class="text-negative text-caption" role="alert">{{ error }}</div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn v-close-popup flat :label="t('brEditor.actions.cancel')" />
          <q-btn type="submit" color="primary" unelevated :label="t('brEditor.actions.save')" :loading="saving" />
        </q-card-actions>
      </q-form>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { Notify } from 'quasar';
import { useI18n } from 'vue-i18n';
import { APIError, ErrCode } from '@/encore/client';
import { api } from '@/encore/api';
import { countryOptions } from '@/utils/countries';

defineProps<{ modelValue: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>();

const { t, locale } = useI18n();
const route = useRoute();

const form = reactive({ first_name: '', last_name: '', address: '', cap: '', city: '', country: '' });
const phone = ref('');
const loading = ref(false);
const saving = ref(false);
const error = ref('');

// Stati europei, come alla registrazione; un paese già salvato fuori elenco
// resta selezionabile.
const countries = computed(() => {
  const lang = String(route.params.locale || locale.value).slice(0, 2);
  const list = countryOptions(lang);
  if (form.country && !list.some((c) => c.code === form.country)) list.push({ code: form.country, name: form.country });
  return list;
});

async function load() {
  error.value = '';
  loading.value = true;
  try {
    const data = await api.profiles.GetPersonalData();
    Object.assign(form, {
      first_name: data.first_name,
      last_name: data.last_name,
      address: data.address,
      cap: data.cap,
      city: data.city,
      country: data.country,
    });
    phone.value = data.phone;
  } catch (err) {
    // NotFound: nessun dato personale ancora, il modulo resta vuoto.
    Object.assign(form, { first_name: '', last_name: '', address: '', cap: '', city: '', country: '' });
    phone.value = '';
    if (!(err instanceof APIError && err.code === ErrCode.NotFound)) error.value = t('userProfile.saveError');
  } finally {
    loading.value = false;
  }
}

async function save() {
  // Stessi controlli del backend: tutti i campi obbligatori, CAP da 4 a 10.
  const values = { ...form, cap: form.cap.trim() };
  if (Object.values(values).some((v) => !v.trim())) {
    error.value = t('userProfile.fieldsRequired');
    return;
  }
  if (values.cap.length < 4 || values.cap.length > 10) {
    error.value = t('userProfile.capInvalid');
    return;
  }
  saving.value = true;
  error.value = '';
  try {
    await api.profiles.UpsertPersonalData(values);
    Notify.create({ type: 'positive', message: t('userProfile.saved') });
    emit('update:modelValue', false);
  } catch {
    error.value = t('userProfile.saveError');
  } finally {
    saving.value = false;
  }
}
</script>
