<template>
  <q-dialog
    :model-value="modelValue"
    persistent
    @update:model-value="(val: boolean) => emit('update:modelValue', val)"
    @show="onShow"
  >
    <q-card style="width: 640px; max-width: 95vw; max-height: 90vh" class="column no-wrap">
      <q-card-section class="row items-center justify-between bg-primary text-white q-py-sm" style="flex-shrink: 0">
        <div class="row items-center q-gutter-sm">
          <q-icon :name="isEdit ? 'edit' : 'person_add'" size="24px" />
          <div class="text-h6 text-weight-bold">
            {{ isEdit ? t('broadcasters.dialog.editTitle') : t('broadcasters.dialog.newTitle') }}
          </div>
        </div>
        <q-btn icon="close" flat round dense color="white" @click="closeDialog" />
      </q-card-section>

      <q-card-section class="col scroll q-pa-md">
        <q-banner v-if="localError" class="bg-negative text-white q-mb-md" rounded dense>
          {{ localError }}
        </q-banner>

        <q-form ref="formRef" @submit.prevent="onSubmit" class="q-gutter-y-md">
          <!-- Nome d'arte / Display Name -->
          <q-input
            v-model="form.display_name"
            :label="t('broadcasters.dialog.displayNameLabel')"
            outlined
            dense
            :rules="[(val) => (val && val.trim().length > 0) || t('broadcasters.dialog.displayNameRequired')]"
          >
            
          </q-input>

          <!-- Categoria (colore automatico da config Categories) -->
          <div>
            <q-select
              v-model="form.category"
              :options="categoryOptions"
              option-label="label"
              option-value="value"
              emit-value
              map-options
              :label="t('broadcasters.dialog.categoryLabel')"
              outlined
              dense
              :rules="[(val) => !!val || t('broadcasters.dialog.categoryRequired')]"
              @update:model-value="onCategoryChange"
            >
              <template #selected>
                <div v-if="form.category" class="row items-center q-gutter-xs">
                  <div
                    class="color-preview"
                    :style="{ backgroundColor: selectedCategoryColor }"
                  />
                  <span>{{ form.category }}</span>
                </div>
              </template>
              <template #option="scope">
                <q-item v-bind="scope.itemProps">
                  <q-item-section avatar side class="q-pr-sm">
                    <div class="color-preview" :style="{ backgroundColor: scope.opt.color || '#6366f1' }" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ scope.opt.label }}</q-item-label>
                  </q-item-section>
                </q-item>
              </template>
            </q-select>
          </div>

          <!-- Disponibilità e Certificato -->
          <div class="row q-col-gutter-sm items-center">
            <div class="col-12 col-sm-6">
              <q-select
                v-model="form.availability"
                :options="availabilityOptions"
                emit-value
                map-options
                :label="t('broadcasters.dialog.availabilityLabel')"
                outlined
                dense
              >
               
              </q-select>
            </div>
            <div class="col-12 col-sm-6">
              <q-toggle
                v-model="form.certified"
                :label="t('broadcasters.dialog.certifiedLabel')"
                color="positive"
                checked-icon="verified"
                unchecked-icon="verified_user"
              />
            </div>
            <div class="col-12 col-sm-6">
              <q-toggle
                v-model="form.enabled"
                :label="t('broadcasters.dialog.enabledLabel')"
                color="positive"
                checked-icon="check_circle"
                unchecked-icon="cancel"
              />
            </div>
            <div class="col-12 col-sm-6">
              <q-toggle
                v-model="form.used"
                :label="t('broadcasters.dialog.usedLabel')"
                color="primary"
                checked-icon="how_to_reg"
                unchecked-icon="person_outline"
              />
            </div>
          </div>

          <!-- Telefono -->
          <q-input
            v-model="form.phone"
            :label="t('broadcasters.dialog.phoneLabel')"
            outlined
            dense
          >
            
          </q-input>

          <!-- Biografia -->
          <q-input
            v-model="form.bio"
            :label="t('broadcasters.dialog.bioLabel')"
            type="textarea"
            outlined
            dense
            rows="3"
            :placeholder="t('broadcasters.dialog.bioPlaceholder')"
          >
           
          </q-input>

          <!-- Servizi Offerti (Chips) -->
          <div>
            <div class="text-caption text-weight-bold text-grey-8 q-mb-xs">
              {{ t('broadcasters.dialog.servicesLabel') }}
            </div>
            <q-select
              v-model="form.services"
              multiple
              use-chips
              use-input
              hide-dropdown-icon
              new-value-mode="add-unique"
              outlined
              dense
              :placeholder="t('broadcasters.dialog.servicesPlaceholder')"
            />
            <!-- Suggerimenti rapidi da config (come BroadcasterAttributes) -->
            <div v-if="availableServicesPresets.length > 0" class="q-mt-xs">
              <div class="text-caption text-grey-7 q-mb-xs" style="font-size: 11px">
                {{ t('broadcasters.dialog.quickPresetsHint') }}
              </div>
              <div class="row q-gutter-xs">
                <q-chip
                  v-for="preset in availableServicesPresets"
                  :key="preset"
                  clickable
                  outline
                  color="primary"
                  size="sm"
                  icon="add"
                  @click="addServicePreset(preset)"
                >
                  {{ preset }}
                </q-chip>
              </div>
            </div>
          </div>

          <!-- Lingue Parlate (Chips) -->
          <div>
            <div class="text-caption text-weight-bold text-grey-8 q-mb-xs">
              {{ t('broadcasters.dialog.languagesLabel') }}
            </div>
            <q-select
              v-model="form.languages"
              multiple
              use-chips
              use-input
              hide-dropdown-icon
              new-value-mode="add-unique"
              outlined
              dense
              :placeholder="t('broadcasters.dialog.languagesPlaceholder')"
            >
              
            </q-select>
          </div>
        </q-form>
      </q-card-section>

      <q-separator style="flex-shrink: 0" />

      <!-- Footer Buttons -->
      <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
        <q-btn flat :label="t('actions.cancel')" :disable="loading" @click="closeDialog" />
        <q-btn
          color="primary"
          icon="save"
          :loading="loading"
          :label="isEdit ? t('broadcasters.dialog.saveChanges') : t('broadcasters.dialog.createBroadcaster')"
          unelevated
          @click="onSubmit"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { Notify } from 'quasar';
import { useBroadcastersStore, type Broadcaster } from '@/stores/broadcasters-store';

const props = defineProps<{
  modelValue: boolean;
  broadcaster?: Broadcaster | null;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'saved', broadcaster: Broadcaster): void;
}>();

const { t } = useI18n();
const broadcastersStore = useBroadcastersStore();
const formRef = ref();
const loading = ref(false);
const localError = ref<string | null>(null);

const isEdit = computed(() => !!props.broadcaster && !!props.broadcaster.id);

const categoryOptions = computed(() => {
  if (broadcastersStore.categories && broadcastersStore.categories.length > 0) {
    return broadcastersStore.categories.map((c) => ({
      label: c.name,
      value: c.name,
      color: c.color,
    }));
  }
  return [
    { label: 'General', value: 'General', color: '#63f18eff' },
    { label: 'Music', value: 'Music', color: '#f16365ff' },
    { label: 'Beauty', value: 'Beauty', color: '#6366f1' },
    { label: 'test', value: 'test', color: '#ecf163ff' },
  ];
});

const selectedCategoryColor = computed(() => {
  const match = categoryOptions.value.find(
    (c) => c.value.toLowerCase() === (form.category || '').toLowerCase()
  );
  return match?.color || '#63f18eff';
});

function onCategoryChange() {
  form.category_color = selectedCategoryColor.value;
}

function formatAvailabilityLabel(val: string): string {
  const key = `broadcasters.availability.${val.toLowerCase()}`;
  return t(key, val);
}

const availabilityOptions = computed(() => {
  const options = (broadcastersStore.availabilityOptions && broadcastersStore.availabilityOptions.length > 0)
    ? broadcastersStore.availabilityOptions
    : ['available', 'live', 'busy', 'offline'];

  return options.map((opt) => ({
    label: formatAvailabilityLabel(opt),
    value: opt,
  }));
});

const form = reactive({
  display_name: '',
  category: 'General',
  category_color: '#63f18eff',
  bio: '',
  certified: false,
  enabled: false,
  used: true,
  availability: 'available',
  phone: '',
  services: [] as string[],
  languages: [] as string[],
});

const availableServicesPresets = computed(() => {
  const current = form.services || [];
  return (broadcastersStore.servicesPresets || []).filter(
    (preset) => !current.some((s) => s.toLowerCase() === preset.toLowerCase())
  );
});

function addServicePreset(preset: string) {
  if (!form.services.includes(preset)) {
    form.services.push(preset);
  }
}

async function onShow() {
  localError.value = null;
  await Promise.all([
    broadcastersStore.fetchCategories(),
    broadcastersStore.fetchAvailabilityOptions(),
    broadcastersStore.fetchServicesPresets(),
  ]);

  const defaultCat = broadcastersStore.categories[0]?.name || 'General';
  const defaultAvailability = broadcastersStore.availabilityOptions[0] || 'available';

  if (props.broadcaster) {
    form.display_name = props.broadcaster.display_name || '';
    form.category = props.broadcaster.category || defaultCat;
    form.category_color = selectedCategoryColor.value;
    form.bio = props.broadcaster.bio || '';
    form.certified = !!props.broadcaster.certified;
    form.enabled = !!props.broadcaster.enabled;
    form.used = props.broadcaster.used ?? true;
    form.availability = props.broadcaster.availability || defaultAvailability;
    form.phone = props.broadcaster.phone || '';
    form.services = Array.isArray(props.broadcaster.services) ? [...props.broadcaster.services] : [];
    form.languages = Array.isArray(props.broadcaster.languages) ? [...props.broadcaster.languages] : [];
  } else {
    form.display_name = '';
    form.category = defaultCat;
    form.category_color = selectedCategoryColor.value;
    form.bio = '';
    form.certified = false;
    form.enabled = false;
    form.used = true;
    form.availability = defaultAvailability;
    form.phone = '';
    form.services = ['Live Streaming'];
    form.languages = ['Italiano'];
  }
}

function closeDialog() {
  emit('update:modelValue', false);
}

async function onSubmit() {
  if (formRef.value) {
    const valid = await formRef.value.validate();
    if (!valid) return;
  }

  loading.value = true;
  localError.value = null;

  try {
    let saved: Broadcaster;
    if (isEdit.value && props.broadcaster?.id) {
      saved = await broadcastersStore.updateBroadcaster(props.broadcaster.id, {
        display_name: form.display_name.trim(),
        category: form.category.trim(),
        category_color: selectedCategoryColor.value,
        bio: form.bio.trim(),
        certified: form.certified,
        enabled: form.enabled,
        used: form.used,
        availability: form.availability,
        phone: form.phone.trim(),
        services: form.services,
        languages: form.languages,
        attributes: props.broadcaster?.attributes ?? [],
        hours: props.broadcaster?.hours ?? [],
        photos: props.broadcaster?.photos ?? [],
      });
      Notify.create({
        type: 'positive',
        message: t('broadcasters.dialog.updateSuccess'),
      });
    } else {
      saved = await broadcastersStore.createBroadcaster({
        user_id: '',
        display_name: form.display_name.trim(),
        category: form.category.trim(),
        category_color: selectedCategoryColor.value,
        bio: form.bio.trim(),
        certified: form.certified,
        enabled: form.enabled,
        used: form.used,
        availability: form.availability,
        phone: form.phone.trim(),
        services: form.services,
        languages: form.languages,
        attributes: [],
        hours: [],
        photos: [],
      });
      Notify.create({
        type: 'positive',
        message: t('broadcasters.dialog.createSuccess'),
      });
    }

    emit('saved', saved);
    closeDialog();
  } catch (err: unknown) {
    localError.value = err instanceof Error ? err.message : String(err);
    Notify.create({
      type: 'negative',
      message: localError.value || t('broadcasters.dialog.saveError'),
    });
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.color-preview {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid rgba(0, 0, 0, 0.2);
}
</style>
