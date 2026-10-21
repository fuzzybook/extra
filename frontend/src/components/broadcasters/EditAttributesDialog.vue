<template>
  <q-dialog
    :model-value="modelValue"
    persistent
    @update:model-value="(val: boolean) => emit('update:modelValue', val)"
    @show="onShow"
  >
    <q-card style="width: 620px; max-width: 95vw; max-height: 90vh" class="column no-wrap">
      <!-- Header -->
      <q-card-section class="row items-center justify-between bg-indigo-7 text-white q-py-sm" style="flex-shrink: 0">
        <div class="row items-center q-gutter-sm">
          <q-icon name="tune" size="24px" />
          <div>
            <div class="text-h6 text-weight-bold">{{ t('broadcasters.attributesDialog.title') }}</div>
            <div class="text-caption opacity-80" v-if="broadcaster">
              {{ broadcaster.display_name }} (ID #{{ broadcaster.id }})
            </div>
          </div>
        </div>
        <q-btn icon="close" flat round dense color="white" @click="closeDialog" />
      </q-card-section>

      <!-- Corpo -->
      <q-card-section class="col scroll q-pa-md">
        <q-banner v-if="localError" class="bg-negative text-white q-mb-md" rounded dense>
          {{ localError }}
        </q-banner>

        <!-- Presets rapidi -->
        <div class="q-mb-md">
          <div class="text-caption text-grey-7 q-mb-xs text-weight-medium">
            {{ t('broadcasters.attributesDialog.quickPresets') }}
          </div>
          <div class="row q-gutter-xs">
            <q-chip
              v-for="preset in availablePresets"
              :key="preset"
              clickable
              outline
              color="indigo"
              size="sm"
              icon="add"
              @click="addPreset(preset)"
            >
              {{ preset }}
            </q-chip>
          </div>
        </div>

        <!-- Form aggiunta nuovo attributo -->
        <q-card flat bordered class="q-pa-sm q-mb-md bg-grey-1">
          <div class="text-subtitle2 text-weight-bold text-grey-9 q-mb-xs">
            {{ t('broadcasters.attributesDialog.addNewTitle') }}
          </div>
          <div class="row q-col-gutter-sm items-center">
            <div class="col-12 col-sm-5">
              <q-input
                v-model="newAttr.label"
                :label="t('broadcasters.attributesDialog.labelField')"
                :placeholder="t('broadcasters.attributesDialog.labelPlaceholder')"
                outlined
                dense
              />
            </div>
            <div class="col-12 col-sm-5">
              <q-input
                v-model="newAttr.value"
                :label="t('broadcasters.attributesDialog.valueField')"
                :placeholder="t('broadcasters.attributesDialog.valuePlaceholder')"
                outlined
                dense
                @keyup.enter="addAttribute"
              />
            </div>
            <div class="col-12 col-sm-2 text-right">
              <q-btn
                color="indigo"
                icon="add"
                :label="t('broadcasters.attributesDialog.add')"
                unelevated
                dense
                class="full-width q-py-xs"
                :disable="!newAttr.label.trim() || !newAttr.value.trim()"
                @click="addAttribute"
              />
            </div>
          </div>
        </q-card>

        <!-- Lista attributi correnti -->
        <div v-if="attributesList.length === 0" class="text-center q-pa-lg text-grey-6">
          <q-icon name="list_alt" size="40px" class="q-mb-xs" />
          <div class="text-body2">{{ t('broadcasters.attributesDialog.emptyTitle') }}</div>
          <div class="text-caption">{{ t('broadcasters.attributesDialog.emptySubtitle') }}</div>
        </div>

        <q-list v-else separator bordered class="rounded-borders">
          <q-item
            v-for="(attr, index) in attributesList"
            :key="index"
            class="items-center q-py-sm"
          >
            <q-item-section side>
              <q-avatar size="28px" color="indigo-1" text-color="indigo-8" icon="label" />
            </q-item-section>

            <q-item-section>
              <div class="row q-col-gutter-xs items-center">
                <div class="col-12 col-sm-5">
                  <q-input
                    v-model="attr.label"
                    :label="t('broadcasters.attributesDialog.labelField')"
                    dense
                    outlined
                    hide-bottom-space
                  />
                </div>
                <div class="col-12 col-sm-7">
                  <q-input
                    v-model="attr.value"
                    :label="t('broadcasters.attributesDialog.valueField')"
                    dense
                    outlined
                    hide-bottom-space
                  />
                </div>
              </div>
            </q-item-section>

            <q-item-section side>
              <q-btn
                flat
                round
                dense
                color="negative"
                icon="delete"
                @click="removeAttribute(index)"
              >
                <q-tooltip>{{ t('broadcasters.attributesDialog.removeTooltip') }}</q-tooltip>
              </q-btn>
            </q-item-section>
          </q-item>
        </q-list>
      </q-card-section>

      <q-separator style="flex-shrink: 0" />

      <!-- Footer Buttons -->
      <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
        <q-btn flat :label="t('actions.cancel')" :disable="loading" @click="closeDialog" />
        <q-btn
          color="indigo"
          icon="save"
          :loading="loading"
          :label="t('broadcasters.attributesDialog.save')"
          unelevated
          @click="onSave"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Notify } from 'quasar';
import {
  useBroadcastersStore,
  type Broadcaster,
  type BroadcasterAttribute,
} from '@/stores/broadcasters-store';

const props = defineProps<{
  modelValue: boolean;
  broadcaster: Broadcaster | null;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'saved', broadcaster: Broadcaster): void;
}>();

const { t } = useI18n();
const broadcastersStore = useBroadcastersStore();
const loading = ref(false);
const localError = ref<string | null>(null);

const attributesList = ref<BroadcasterAttribute[]>([]);
const newAttr = reactive({ label: '', value: '' });

const presetLabels = computed<string[]>(() => {
  if (broadcastersStore.attributePresets && broadcastersStore.attributePresets.length > 0) {
    return broadcastersStore.attributePresets;
  }
  return [
    'Genere Musicale',
    'Strumento',
    'Anni di Esperienza',
    'Città Base',
    'Instagram',
    'Spotify',
    'YouTube',
    'SoundCloud',
    'Attrezzatura DJ',
    'Tariffa Oraria',
  ];
});

const availablePresets = computed(() => {
  const currentLabels = new Set(attributesList.value.map((a) => a.label.toLowerCase()));
  return presetLabels.value.filter((p) => !currentLabels.has(p.toLowerCase()));
});

async function onShow() {
  localError.value = null;
  newAttr.label = '';
  newAttr.value = '';
  if (!broadcastersStore.attributePresets || broadcastersStore.attributePresets.length === 0) {
    await broadcastersStore.fetchAttributePresets();
  }
  if (props.broadcaster?.attributes) {
    attributesList.value = props.broadcaster.attributes.map((a) => ({
      label: a.label,
      value: a.value,
    }));
  } else {
    attributesList.value = [];
  }
}

function closeDialog() {
  emit('update:modelValue', false);
}

function addPreset(preset: string) {
  newAttr.label = preset;
  newAttr.value = '';
}

function addAttribute() {
  const lbl = newAttr.label.trim();
  const val = newAttr.value.trim();
  if (!lbl || !val) return;

  attributesList.value.push({ label: lbl, value: val });
  newAttr.label = '';
  newAttr.value = '';
}

function removeAttribute(index: number) {
  attributesList.value.splice(index, 1);
}

async function onSave() {
  if (!props.broadcaster?.id) return;

  loading.value = true;
  localError.value = null;

  try {
    const cleaned = attributesList.value
      .map((a) => ({ label: a.label.trim(), value: a.value.trim() }))
      .filter((a) => a.label.length > 0 && a.value.length > 0);

    const updated = await broadcastersStore.updateAttributes(props.broadcaster.id, cleaned);
    Notify.create({
      type: 'positive',
      message: t('broadcasters.attributesDialog.saveSuccess'),
    });
    emit('saved', updated);
    closeDialog();
  } catch (err: unknown) {
    localError.value = err instanceof Error ? err.message : String(err);
    Notify.create({
      type: 'negative',
      message: localError.value || t('broadcasters.attributesDialog.saveError'),
    });
  } finally {
    loading.value = false;
  }
}
</script>
