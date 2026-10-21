<template>
  <q-dialog
    :model-value="modelValue"
    position="standard"
    persistent
    @update:model-value="(val: boolean) => emit('update:modelValue', val)"
    @show="onShow"
  >

    <q-card style="width: 650px; min-width: 360px; max-width: 90vw; height: 85vh; max-height: 85vh" class="column">
      <!-- Header -->
      <q-card-section class="row items-center justify-between q-pb-sm">
        <div>
          <div class="text-h6 text-weight-bold">Seleziona Immagine</div>
          <div class="text-caption text-grey-7">Scegli un'immagine dalla galleria o caricane una nuova</div>
        </div>
        <div class="row items-center q-gutter-sm">
          <q-btn
            color="primary"
            icon="add_photo_alternate"
            label="Carica Nuova"
            dense
            unelevated
            class="q-px-sm"
            @click="openUploadDialog = true"
          />
          <q-btn icon="close" flat round dense @click="closeDialog" />
        </div>
      </q-card-section>

      <!-- Barra ricerca e conteggio -->
      <q-card-section class="q-pt-none q-pb-sm">
        <q-input
          v-model="search"
          placeholder="Cerca immagine per descrizione..."
          outlined
          dense
          clearable
        >
          <template #prepend>
            <q-icon name="search" />
          </template>
        </q-input>
      </q-card-section>

      <!-- Contenuto scrollabile con Galleria immagini -->
      <q-card-section class="col scroll q-pt-none">
        <div v-if="imagesStore.loading" class="row justify-center items-center q-pa-xl">
          <q-spinner color="primary" size="48px" />
        </div>

        <div v-else-if="filteredImages.length === 0" class="text-center q-pa-xl text-grey-6">
          <q-icon name="image_not_supported" size="48px" class="q-mb-sm" />
          <div class="text-body1">Nessuna immagine trovata</div>
          <div class="text-caption">Carica la prima immagine con il pulsante in alto a destra.</div>
        </div>

        <!-- Lista immagini con titolo e anteprima proporzionata (max 80px) in riga (row) -->
        <q-list v-else separator class="rounded-borders">
          <q-item
            v-for="img in filteredImages"
            :key="img.uuid"
            clickable
            v-ripple
            class="list-image-item q-py-sm items-center cursor-pointer"
            @click="chooseImage(img)"
          >
            <!-- Immagine a sinistra con altezza massima di 80px e proporzioni mantenute -->
            <q-item-section side class="q-pr-md">
              <div class="image-thumb-box flex flex-center">
                <img
                  :src="img.url"
                  :alt="imageDescriptionText(img.description)"
                  class="image-thumb"
                  loading="lazy"
                />
              </div>
            </q-item-section>

            <!-- Titolo / Descrizione a destra, allineati a sinistra -->
            <q-item-section>
              <q-item-label class="text-subtitle1 text-weight-bold text-grey-9">
                {{ imageDescriptionText(img.description) }}
              </q-item-label>
              <q-item-label caption class="text-grey-6 row items-center q-gutter-sm q-mt-xs">
                <span>{{ formatDate(img.created_at) }}</span>
                <span v-if="img.width && img.height" class="text-caption text-grey-6">
                  • {{ img.width }}x{{ img.height }} px
                </span>
                <span v-else-if="img.proportion" class="text-caption text-grey-6">
                  • Prop. {{ img.proportion }}
                </span>
              </q-item-label>
            </q-item-section>

            <q-item-section side>
              <q-btn
                flat
                dense
                round
                color="primary"
                icon="check_circle"
                @click.stop="chooseImage(img)"
              >
                <q-tooltip>Seleziona immagine</q-tooltip>
              </q-btn>
            </q-item-section>
          </q-item>
        </q-list>
      </q-card-section>

      <!-- Footer -->
      <q-card-actions align="right" class="q-pa-md bg-grey-1">
        <q-btn flat label="Annulla" @click="closeDialog" />
      </q-card-actions>
    </q-card>

    <!-- Dialog di caricamento nuova immagine collegato -->
    <ImmageDialog
      v-model="openUploadDialog"
      :initial-category="ImageCategory.Promo"
      :target-width="targetWidth"
      :target-height="targetHeight"
      :proportion="proportion"
      @saved="onImageUploaded"
    />
  </q-dialog>
</template>

<style scoped>
.list-image-item {
  transition: background-color 0.15s ease;
  border-radius: 6px;
}

.list-image-item:hover {
  background-color: #f0f7ff;
}

.image-thumb-box {
  height: 80px;
  max-height: 80px;
  min-width: 60px;
  max-width: 140px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f7f7f7;
}

.image-thumb {
  max-height: 80px;
  max-width: 140px;
  width: auto;
  height: auto;
  object-fit: contain;
  display: block;
}
</style>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useImagesStore, type ImageItem } from '@/stores/images-store';
import ImmageDialog from '@/components/ImmageDialog.vue';
import { ImageCategory } from '@/types/images';

defineProps<{
  modelValue: boolean;
  targetWidth?: number | undefined;
  targetHeight?: number | undefined;
  proportion?: string | undefined;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'selected': [image: ImageItem];
}>();

const imagesStore = useImagesStore();
const search = ref('');
const openUploadDialog = ref(false);

function imageDescriptionText(desc: any): string {
  if (!desc) return '';
  if (typeof desc === 'string') return desc;
  return desc.text ?? '';
}

const filteredImages = computed(() => {
  const query = search.value.trim().toLowerCase();
  if (!query) {
    return imagesStore.images;
  }
  return imagesStore.images.filter((img) =>
    imageDescriptionText(img.description).toLowerCase().includes(query)
  );
});

function onShow() {
  void imagesStore.fetchImages();
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return isNaN(d.getTime()) ? '' : d.toLocaleDateString();
}

function chooseImage(img: ImageItem) {
  emit('selected', img);
  emit('update:modelValue', false);
}

function onImageUploaded(newImg: ImageItem) {
  openUploadDialog.value = false;
  emit('selected', newImg);
  emit('update:modelValue', false);
}

function closeDialog() {
  emit('update:modelValue', false);
}
</script>
