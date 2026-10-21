<template>
  <q-dialog :model-value="modelValue" persistent @update:model-value="(val: boolean) => emit('update:modelValue', val)">
    <q-card style="min-width: 420px; max-width: 90vw;">
      <q-card-section class="row items-center q-pb-none">
        <div class="text-h6">
          {{ initialImage ? 'Modifica Immagine' : 'Nuova Immagine' }}
        </div>
        <q-space />
        <q-btn icon="close" flat round dense @click="closeDialog" />
      </q-card-section>

      <q-card-section class="q-pt-md">
        <!-- Area Selezione e Preview Immagine con proporzioni mantenute -->
        <div class="row items-center q-gutter-md q-mb-md">
          <div
            class="rounded-borders bg-grey-3 flex flex-center overflow-hidden"
            :style="previewContainerStyle"
          >
            <q-img
              v-if="previewSrc"
              :src="previewSrc"
              :ratio="previewRatio"
              fit="contain"
              class="full-width full-height rounded-borders"
            />
            <q-icon v-else name="image" size="36px" color="grey-6" />
          </div>

          <div class="column q-gutter-xs">
            <div class="row items-center q-gutter-xs">
              <q-btn round flat color="primary" icon="upload" :aria-label="'Scegli file'" @click="pickFile">
                <q-tooltip>Scegli file immagine</q-tooltip>
              </q-btn>
              <q-btn
                v-if="previewSrc"
                round
                flat
                color="warning"
                icon="delete"
                :aria-label="'Rimuovi'"
                @click="clearImage"
              >
                <q-tooltip>Rimuovi immagine</q-tooltip>
              </q-btn>
            </div>
            <div v-if="previewDimensionsText" class="text-caption text-grey-7">
              {{ previewDimensionsText }}
            </div>
          </div>

          <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onFileSelected" />
        </div>

        <!-- Descrizione obbligatoria -->
        <q-input
          v-model="description"
          label="Descrizione immagine *"
          outlined
          dense
          :rules="[(val: string) => !!val?.trim() || 'La descrizione è obbligatoria']"
          class="q-mb-sm"
        />

        <!-- Selezione Categoria basata sull'Enum TypeScript derivato da Go const -->
        <q-select
          v-model="category"
          :options="IMAGE_CATEGORY_OPTIONS"
          emit-value
          map-options
          outlined
          dense
          label="Categoria"
          class="q-mb-sm"
        >
          <template #option="scope">
            <q-item v-bind="scope.itemProps">
              <q-item-section avatar>
                <q-icon :name="scope.opt.icon" :color="scope.opt.color" />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ scope.opt.label }}</q-item-label>
              </q-item-section>
            </q-item>
          </template>
        </q-select>

        <div v-if="errorMessage" class="text-negative text-caption q-mt-xs">
          {{ errorMessage }}
        </div>
      </q-card-section>

      <q-card-actions align="right" class="q-pa-md">
        <q-btn flat label="Annulla" @click="closeDialog" />
        <q-btn
          color="primary"
          label="Salva"
          :loading="saving"
          :disable="!canSave"
          @click="submit"
        />
      </q-card-actions>
    </q-card>

    <!-- Dialog Cropper con vue-advanced-cropper (come AvatarUpload) -->
    <q-dialog v-model="cropperOpen" @show="cropperReady = true" @hide="onCropperHide">
      <q-card style="min-width: 350px; max-width: 90vw; position: relative;">
        <q-card-section>
          <div class="text-h6">Ritaglia Immagine</div>
          <div v-if="targetWidth" class="text-caption text-grey-7">
            <template v-if="targetHeight && targetHeight !== 0">
              Dimensioni richieste: {{ targetWidth }}x{{ targetHeight }}px
            </template>
            <template v-else-if="proportion">
              Dimensioni richieste: {{ targetWidth }}px (proporzione {{ proportion }})
            </template>
          </div>
        </q-card-section>

        <q-card-section>
          <CropperImage
            v-if="cropperSource && cropperReady"
            ref="cropperRef"
            class="cropper"
            :src="cropperSource"
            :width="targetWidth ?? 230"
            :height="targetHeight ?? 0"
            :proportion="proportion ?? '2:3'"
            :show-controls="false"
            :show-preview="false"
          />
        </q-card-section>

        <q-card-actions align="right">
          <q-btn v-close-popup flat label="Annulla" />
          <q-btn color="primary" label="Conferma Ritaglio" @click="applyCrop" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-dialog>
</template>

<style scoped>
.cropper {
  width: min(420px, 82vw);
  padding: 0;
  border: 0;
  background: #ddd;
}
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import CropperImage from '@/components/CropperImage.vue';
import { ImageCategory, IMAGE_CATEGORY_OPTIONS } from '@/types/images';
import { useImagesStore, type ImageItem } from '@/stores/images-store';

const props = defineProps<{
  modelValue: boolean;
  initialImage?: ImageItem | null | undefined;
  initialCategory?: ImageCategory | undefined;
  targetWidth?: number | undefined;
  targetHeight?: number | undefined;
  proportion?: string | undefined;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'saved': [image: ImageItem];
}>();

const imagesStore = useImagesStore();

const fileInput = ref<HTMLInputElement | null>(null);
const cropperRef = ref<InstanceType<typeof CropperImage> | null>(null);

const description = ref('');
const category = ref<ImageCategory>(props.initialCategory ?? ImageCategory.General);
const previewSrc = ref('');
const pendingBase64 = ref('');
const saving = ref(false);
const errorMessage = ref<string | null>(null);

const previewRatio = computed<number>(() => {
  // Se height != 0 si utilizza l'altezza
  if (props.targetHeight && props.targetHeight !== 0 && props.targetWidth && props.targetWidth !== 0) {
    return props.targetWidth / props.targetHeight;
  }
  // Se height == 0 si usa la proporzione
  if (props.proportion && props.proportion.includes(':')) {
    const parts = props.proportion.split(':').map((s) => parseFloat(s.trim()));
    const pw = parts[0];
    const ph = parts[1];
    if (parts.length === 2 && pw !== undefined && ph !== undefined && !isNaN(pw) && !isNaN(ph) && pw > 0 && ph > 0) {
      return pw / ph;
    }
  }
  if (props.targetWidth && props.targetHeight && props.targetHeight !== 0) {
    return props.targetWidth / props.targetHeight;
  }
  return 2 / 3;
});

const previewContainerStyle = computed(() => {
  const ratio = previewRatio.value;
  const baseHeight = 100;
  let w = Math.round(baseHeight * ratio);
  let h = baseHeight;
  if (w > 160) {
    w = 160;
    h = Math.round(w / ratio);
  }
  return {
    width: `${w}px`,
    height: `${h}px`,
    border: '1px solid #dcdcdc',
    backgroundColor: '#f5f5f5',
  };
});

const previewDimensionsText = computed(() => {
  if (props.targetHeight && props.targetHeight !== 0 && props.targetWidth) {
    return `${props.targetWidth}x${props.targetHeight} px`;
  }
  if (props.proportion) {
    return `Prop. ${props.proportion}`;
  }
  return '';
});

// Gestione Cropper
const cropperOpen = ref(false);
const cropperReady = ref(false);
const cropperSource = ref<string | null>(null);

watch(
  () => props.initialImage,
  (img) => {
    if (img) {
      description.value = typeof img.description === 'string' ? img.description : (img.description?.text ?? '');
      previewSrc.value = img.url;
      pendingBase64.value = '';
    } else {
      resetForm();
    }
  },
  { immediate: true }
);

const canSave = computed(() => {
  if (!description.value.trim()) return false;
  // Per nuova immagine, deve esserci un'immagine caricata
  if (!props.initialImage && !pendingBase64.value && !previewSrc.value) return false;
  return true;
});

function resetForm() {
  description.value = '';
  category.value = props.initialCategory ?? ImageCategory.General;
  previewSrc.value = '';
  pendingBase64.value = '';
  errorMessage.value = null;
}

function pickFile() {
  fileInput.value?.click();
}

function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  clearCropperSource();
  cropperSource.value = URL.createObjectURL(file);
  cropperOpen.value = true;
  input.value = '';
}

function clearCropperSource() {
  if (!cropperSource.value) return;
  URL.revokeObjectURL(cropperSource.value);
  cropperSource.value = null;
}

function onCropperHide() {
  cropperReady.value = false;
  clearCropperSource();
}

async function applyCrop() {
  const blob = await cropperRef.value?.getCroppedBlob('image/png');
  if (!blob) return;

  const reader = new FileReader();
  reader.onload = () => {
    const base64 = reader.result as string;
    pendingBase64.value = base64;
    previewSrc.value = base64;
    cropperOpen.value = false;
  };
  reader.readAsDataURL(blob);
}

function clearImage() {
  previewSrc.value = '';
  pendingBase64.value = '';
}

function closeDialog() {
  emit('update:modelValue', false);
}

async function submit() {
  if (!canSave.value) return;
  saving.value = true;
  errorMessage.value = null;

  try {
    const params = {
      description: {
        text: description.value.trim(),
        box_x: typeof props.initialImage?.description === 'object' ? props.initialImage.description.box_x : 0,
        box_y: typeof props.initialImage?.description === 'object' ? props.initialImage.description.box_y : 0,
      },
      data: pendingBase64.value,
      content_type: 'image/png',
      width: props.targetWidth || 0,
      height: props.targetHeight || 0,
      proportion: props.proportion || '',
      uuid: props.initialImage?.uuid || '',
    };

    const saved = await imagesStore.saveImage(params);

    emit('saved', saved);
    closeDialog();
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Errore durante il salvataggio';
  } finally {
    saving.value = false;
  }
}
</script>
