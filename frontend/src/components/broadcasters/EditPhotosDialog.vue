<template>
  <q-dialog
    :model-value="modelValue"
    persistent
    @update:model-value="(val: boolean) => emit('update:modelValue', val)"
    @show="onShow"
  >
    <q-card style="width: 780px; max-width: 95vw; max-height: 90vh" class="column no-wrap">
      <!-- Header Principale -->
      <q-card-section class="row items-center justify-between bg-deep-orange-7 text-white q-py-sm" style="flex-shrink: 0">
        <div class="row items-center q-gutter-sm">
          <q-icon name="photo_library" size="24px" />
          <div>
            <div class="text-h6 text-weight-bold">{{ t('broadcasters.photosDialog.title') }}</div>
            <div class="text-caption opacity-80" v-if="broadcaster">
              {{ broadcaster.display_name }} (ID #{{ broadcaster.id }})
            </div>
          </div>
        </div>
        <q-btn icon="close" flat round dense color="white" @click="closeDialog" />
      </q-card-section>

      <!-- Corpo del Dialog -->
      <q-card-section class="col scroll q-pa-md">
        <q-banner v-if="localError" class="bg-negative text-white q-mb-md rounded-borders" dense>
          {{ localError }}
        </q-banner>

        <!-- Barra azioni foto: Solo Aggiungi con Cropper (nessuna scelta da galleria) -->
        <div class="row items-center justify-between q-mb-md bg-grey-2 q-pa-sm rounded-borders">
          <div class="row items-center q-gutter-xs">
            <q-badge color="deep-orange" text-color="white" class="q-py-xs q-px-sm">
              {{ t('broadcasters.photosDialog.cropFormat', { proportion: imageConfig.proportion || '2:3', width: imageConfig.width || 600 }) }}
            </q-badge>
            <span class="text-caption text-grey-7">{{ t('broadcasters.photosDialog.coverHint') }}</span>
          </div>
          <div>
            <q-btn
              color="deep-orange-7"
              icon="add_photo_alternate"
              :label="t('broadcasters.photosDialog.addPhoto')"
              unelevated
              dense
              class="q-px-sm text-weight-bold"
              @click="triggerPickFile"
            />
            <!-- File input per aggiungere nuova foto -->
            <input
              ref="fileInputRef"
              type="file"
              accept="image/*"
              class="hidden"
              @change="onFileSelected"
            />
            <!-- File input per sostituire foto esistente -->
            <input
              ref="replaceFileInputRef"
              type="file"
              accept="image/*"
              class="hidden"
              @change="onReplaceFileSelected"
            />
          </div>
        </div>

        <!-- Stato Vuoto -->
        <div v-if="photosList.length === 0" class="text-center q-pa-xl text-grey-6">
          <q-icon name="photo_camera" size="52px" class="q-mb-sm text-deep-orange-4" />
          <div class="text-body1 text-weight-medium">{{ t('broadcasters.photosDialog.emptyTitle') }}</div>
          <div class="text-caption q-mb-md">
            {{ t('broadcasters.photosDialog.emptySubtitle', { proportion: imageConfig.proportion || '2:3' }) }}
          </div>
          <q-btn
            color="deep-orange-7"
            icon="add_photo_alternate"
            :label="t('broadcasters.photosDialog.uploadAndCrop')"
            unelevated
            @click="triggerPickFile"
          />
        </div>

        <!-- Griglia Foto -->
        <div v-else class="row q-col-gutter-md">
          <div
            v-for="(photo, index) in photosList"
            :key="index"
            class="col-12 col-sm-6 col-md-4"
          >
            <q-card flat bordered class="photo-card column full-height">
              <!-- Immagine con Aspect Ratio 2:3 -->
              <div class="photo-ratio-box">
                <img
                  :src="photo.image_url"
                  :alt="stripHtml(photo.alt) || 'Foto Broadcaster'"
                  class="photo-img"
                  loading="lazy"
                  @error="onImgError($event)"
                />
                <!-- Badge Copertina / Posizione -->
                <div class="photo-badges">
                  <q-badge
                    v-if="index === 0"
                    color="positive"
                    :label="t('broadcasters.photosDialog.coverBadge')"
                    class="text-weight-bold q-px-xs"
                  />
                  <q-badge color="dark" class="opacity-80"> #{{ index + 1 }} </q-badge>
                </div>

                <!-- Div in posizione assoluta che può essere spostato (drag) all'interno di photo-ratio-box -->
                <div
                  v-if="photo.caption_enabled"
                  class="caption-drag-box"
                  :class="{ 'caption-drag-box--dragging': captionDragState.active && captionDragState.photoIndex === index }"
                  :style="{
                    left: `${getCaptionLeftPercent(photo)}%`,
                    top: `${getCaptionTopPercent(photo)}%`,
                  }"
                  @pointerdown="startCaptionDrag($event, index)"
                  @pointermove="onCaptionDrag($event)"
                  @pointerup="stopCaptionDrag($event)"
                  @pointercancel="stopCaptionDrag($event)"
                  @dblclick.stop="openCaptionDialog(index)"
                >
                  <div class="caption-drag-inner">
                    <q-icon name="drag_indicator" size="14px" class="caption-drag-handle" />
                    <span v-if="stripHtml(photo.alt)" class="caption-drag-text" v-html="photo.alt" />
                    <span v-else class="caption-drag-placeholder">Didascalia</span>
                  </div>
                  <q-tooltip anchor="top middle" self="bottom middle" :offset="[0, 6]">
                    Trascina per posizionare (X: {{ photo.box_x ?? 0 }}%, Y: {{ photo.box_y ?? 0 }}%) • Doppio clic per modificare
                  </q-tooltip>
                </div>
              </div>

              <!-- Didascalia Foto con Anteprima e Pulsante Modifica -->
              <q-card-section class="q-pa-sm col column justify-between photo-caption-section">
                <div>
                  <div class="row items-center justify-between q-mb-xs">
                    <div class="row items-center q-gutter-xs">
                      <q-toggle
                        v-model="photo.caption_enabled"
                        dense
                        size="xs"
                        color="deep-orange-7"
                        @update:model-value="onToggleCaption(photo, $event)"
                      />
                      <span class="text-caption text-weight-bold text-grey-8">{{ t('broadcasters.photosDialog.captionLabel') }}</span>
                    </div>
                    <div v-if="photo.caption_enabled" class="row items-center q-gutter-xs">
                      <span class="text-caption text-grey-6 text-weight-medium">(X: {{ photo.box_x ?? 0 }}%, Y: {{ photo.box_y ?? 0 }}%)</span>
                      <q-btn
                        flat
                        round
                        dense
                        size="sm"
                        color="deep-orange-7"
                        icon="edit_note"
                        @click="openCaptionDialog(index)"
                      >
                        <q-tooltip>{{ t('broadcasters.photosDialog.editCaption') }}</q-tooltip>
                      </q-btn>
                    </div>
                  </div>

                  <!-- Se abilitata: anteprima cliccabile -->
                  <div
                    v-if="photo.caption_enabled"
                    class="photo-caption-preview cursor-pointer"
                    @click="openCaptionDialog(index)"
                  >
                    <div
                      v-if="stripHtml(photo.alt)"
                      class="text-caption text-grey-9 caption-preview-text"
                      v-html="photo.alt"
                    />
                    <div v-else class="text-caption text-italic text-grey-5 flex items-center q-gutter-xs">
                      <span>{{ t('broadcasters.photosDialog.noCaption') }}</span>
                    </div>
                  </div>
                  <!-- Se disabilitata: non cliccabile / non editabile -->
                  <div
                    v-else
                    class="photo-caption-preview photo-caption-preview--disabled"
                  >
                    <div class="text-caption text-grey-5 flex items-center q-gutter-xs">
                      <q-icon name="subtitles_off" size="14px" color="grey-5" />
                      <span>{{ t('broadcasters.photosDialog.captionDisabled') }}</span>
                    </div>
                  </div>
                </div>
              </q-card-section>

              <!-- Barra Azioni Foto (Sposta Su, Sposta Giù, Fai Copertina, Modifica, Elimina) -->
              <q-separator />
              <q-card-actions align="between" class="q-pa-xs bg-grey-1">
                <div class="row q-gutter-xs">
                  <q-btn
                    flat
                    round
                    dense
                    size="sm"
                    icon="arrow_upward"
                    :disable="index === 0"
                    @click="movePhoto(index, -1)"
                  >
                    <q-tooltip>{{ t('broadcasters.photosDialog.moveUp') }}</q-tooltip>
                  </q-btn>
                  <q-btn
                    flat
                    round
                    dense
                    size="sm"
                    icon="arrow_downward"
                    :disable="index === photosList.length - 1"
                    @click="movePhoto(index, 1)"
                  >
                    <q-tooltip>{{ t('broadcasters.photosDialog.moveDown') }}</q-tooltip>
                  </q-btn>
                  <q-btn
                    v-if="index !== 0"
                    flat
                    dense
                    size="xs"
                    color="primary"
                    :label="t('broadcasters.photosDialog.makeCover')"
                    @click="makeCover(index)"
                  />
                </div>

                <div class="row q-gutter-xs items-center">
                  <!-- Bottone Modifica Foto (con menu contestuale e click rapido) -->
                  <q-btn-dropdown
                    flat
                    dense
                    size="sm"
                    color="primary"
                    icon="edit"
                    no-caps
                    auto-close
                    split
                    :loading="loadingPhotoIndex === index"
                    @click="openEditCurrentPhoto(index)"
                  >
                    <template #label>
                      <span class="q-ml-xs text-caption gt-xs">{{ t('broadcasters.photosDialog.editPhoto') }}</span>
                    </template>
                    <q-list dense style="min-width: 220px">
                      <q-item clickable @click="openEditCurrentPhoto(index)">
                        <q-item-section avatar>
                          <q-icon name="crop" color="deep-orange-7" size="20px" />
                        </q-item-section>
                        <q-item-section>
                          <q-item-label>{{ t('broadcasters.photosDialog.cropCurrentPhoto') }}</q-item-label>
                          <q-item-label caption>Usa il cropper sulla foto attuale</q-item-label>
                        </q-item-section>
                      </q-item>
                      <q-item clickable @click="triggerReplaceDirectly(index)">
                        <q-item-section avatar>
                          <q-icon name="add_photo_alternate" color="primary" size="20px" />
                        </q-item-section>
                        <q-item-section>
                          <q-item-label>{{ t('broadcasters.photosDialog.replaceWithNewPhoto') }}</q-item-label>
                          <q-item-label caption>Carica una nuova immagine da sostituire</q-item-label>
                        </q-item-section>
                      </q-item>
                    </q-list>
                  </q-btn-dropdown>

                  <!-- Bottone Modifica Didascalia Rapido (solo se abilitata) -->
                  <q-btn
                    v-if="photo.caption_enabled"
                    flat
                    round
                    dense
                    size="sm"
                    color="primary"
                    icon="edit_note"
                    @click="openCaptionDialog(index)"
                  >
                    <q-tooltip>{{ t('broadcasters.photosDialog.editCaption') }}</q-tooltip>
                  </q-btn>

                  <!-- Bottone Elimina Foto -->
                  <q-btn
                    flat
                    round
                    dense
                    size="sm"
                    color="negative"
                    icon="delete"
                    @click="confirmRemovePhoto(index)"
                  >
                    <q-tooltip>{{ t('broadcasters.photosDialog.deletePhoto') }}</q-tooltip>
                  </q-btn>
                </div>
              </q-card-actions>
            </q-card>
          </div>
        </div>
      </q-card-section>

      <q-separator style="flex-shrink: 0" />

      <!-- Footer Buttons Dialog Principale -->
      <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
        <q-btn flat :label="t('actions.cancel')" :disable="loading" @click="closeDialog" />
        <q-btn
          color="deep-orange-7"
          icon="save"
          :loading="loading"
          :label="t('broadcasters.photosDialog.save')"
          unelevated
          @click="onSave"
        />
      </q-card-actions>
    </q-card>

    <!-- ======================================================== -->
    <!-- DIALOG RITAGLIO FOTO TRAMITE CROPPER                    -->
    <!-- ======================================================== -->
    <q-dialog v-model="cropperDialogOpen" persistent @show="onCropperShow" @hide="onCropperHide">
      <q-card style="width: 700px; max-width: 95vw; max-height: 92vh" class="column no-wrap">
        <!-- Header Cropper -->
        <q-card-section class="row items-center justify-between bg-deep-orange-7 text-white q-py-sm" style="flex-shrink: 0">
          <div class="row items-center q-gutter-sm">
            <q-icon :name="editingPhotoIndex !== null ? 'edit' : 'crop'" size="22px" />
            <div>
              <div class="text-subtitle1 text-weight-bold">
                {{ editingPhotoIndex !== null ? t('broadcasters.photosDialog.editTitle', { index: editingPhotoIndex + 1 }) : t('broadcasters.photosDialog.cropTitle') }}
              </div>
              <div class="text-caption text-orange-1">
                {{ t('broadcasters.photosDialog.cropFormatSubtitle', { proportion: imageConfig.proportion || '2:3', width: targetCropWidth, height: targetCropHeight }) }}
              </div>
            </div>
          </div>
          <div class="row items-center q-gutter-xs">
            <!-- Pulsante Sostituisci Immagine dentro il Cropper -->
            <q-btn
              color="white"
              text-color="deep-orange-9"
              icon="add_photo_alternate"
              :label="t('broadcasters.photosDialog.replaceWithNewPhoto')"
              dense
              unelevated
              size="sm"
              class="q-px-sm text-weight-bold q-mr-xs"
              :disable="croppingLoading"
              @click="triggerReplaceFile"
            />
            <q-badge v-if="photoNaturalDimensions" color="white" text-color="deep-orange-9" class="q-py-xs q-px-sm text-weight-bold">
              {{ t('broadcasters.photosDialog.naturalDimensions', { width: photoNaturalDimensions.width, height: photoNaturalDimensions.height }) }}
            </q-badge>
            <q-btn icon="close" flat round dense color="white" v-close-popup :disable="croppingLoading" />
          </div>
        </q-card-section>

        <!-- Corpo Cropper -->
        <q-card-section class="col scroll q-pa-md">
          <div v-if="editingPhotoIndex !== null" class="q-mb-sm">
            <q-banner dense rounded class="bg-orange-1 text-deep-orange-10 text-caption rounded-borders">
              <template #avatar>
                <q-icon name="info" color="deep-orange-7" size="20px" />
              </template>
              Stai modificando la foto #{{ editingPhotoIndex + 1 }}. Puoi riposizionarla/ritagliarla, oppure cliccare su <strong>"{{ t('broadcasters.photosDialog.replaceWithNewPhoto') }}"</strong> per sceglierne un'altra.
            </q-banner>
          </div>
          <div v-else class="text-caption text-grey-7 q-mb-sm text-center">
            {{ t('broadcasters.photosDialog.cropInstructions') }}
          </div>

          <div class="cropper-container row justify-center items-center q-mb-md">
            <CropperImage
              v-if="cropperSource && cropperReady"
              ref="cropperRef"
              :src="cropperSource"
              :width="targetCropWidth"
              :height="targetCropHeight"
              :proportion="imageConfig.proportion || '2:3'"
              :show-controls="false"
              :show-preview="false"
              :show-zoom-controls="true"
              class="cropper-box"
              @image-loaded="onCropperImageLoaded"
            />
          </div>

          <!-- Didascalia Foto con q-editor -->
          <div class="q-mb-xs">
            <div class="row items-center justify-between q-mb-xs">
              <div class="row items-center q-gutter-xs">
                <q-toggle
                  v-model="newPhotoCaptionEnabled"
                  dense
                  size="xs"
                  color="deep-orange-7"
                />
                <span class="text-caption text-weight-bold text-grey-8">
                  {{ t('broadcasters.photosDialog.captionCropLabel') }}
                </span>
              </div>
            </div>
            <q-editor
              v-model="newPhotoAlt"
              :disable="!newPhotoCaptionEnabled"
              min-height="4.5rem"
              max-height="9rem"
              :placeholder="newPhotoCaptionEnabled ? t('broadcasters.photosDialog.captionCropPlaceholder') : t('broadcasters.photosDialog.captionDisabled')"
              :toolbar="editorToolbar"
              :fonts="editorFonts"
              dense
              class="photo-editor"
            />
          </div>
        </q-card-section>

        <q-separator style="flex-shrink: 0" />

        <!-- Footer Cropper -->
        <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
          <q-btn flat :label="t('actions.cancel')" v-close-popup :disable="croppingLoading" />
          <q-btn
            color="deep-orange-7"
            :icon="editingPhotoIndex !== null ? 'save' : 'crop'"
            :label="editingPhotoIndex !== null ? t('broadcasters.photosDialog.confirmAndSave') : t('broadcasters.photosDialog.confirmAndAdd')"
            unelevated
            :loading="croppingLoading"
            @click="onCropConfirm"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- ======================================================== -->
    <!-- DIALOG MODIFICA DIDASCALIA FOTO                         -->
    <!-- ======================================================== -->
    <q-dialog v-model="captionDialogOpen" persistent>
      <q-card style="width: 580px; max-width: 95vw" class="column no-wrap">
        <!-- Header Dialog Didascalia -->
        <q-card-section class="row items-center justify-between bg-deep-orange-7 text-white q-py-sm" style="flex-shrink: 0">
          <div class="row items-center q-gutter-sm">
            <q-icon name="edit_note" size="22px" />
            <div>
              <div class="text-subtitle1 text-weight-bold">
                {{ editingCaptionIndex !== null ? t('broadcasters.photosDialog.captionDialogTitle', { index: editingCaptionIndex + 1 }) : t('broadcasters.photosDialog.captionLabel') }}
              </div>
              <div class="text-caption text-orange-1">
                {{ t('broadcasters.photosDialog.captionDialogSubtitle') }}
              </div>
            </div>
          </div>
          <q-btn icon="close" flat round dense color="white" v-close-popup />
        </q-card-section>

        <!-- Corpo con Thumbnail della foto ed Editor -->
        <q-card-section class="q-pa-md">
          <div
            v-if="currentEditingCaptionPhoto !== null && editingCaptionIndex !== null"
            class="row items-center q-gutter-md q-mb-md q-pa-sm bg-grey-1 rounded-borders"
          >
            <div style="width: 54px; height: 76px; overflow: hidden; border-radius: 4px; flex-shrink: 0" class="shadow-1">
              <img
                :src="currentEditingCaptionPhoto.image_url"
                alt="Thumbnail"
                style="width: 100%; height: 100%; object-fit: cover; display: block"
              />
            </div>
            <div class="col">
              <div class="text-subtitle2 text-weight-bold text-grey-9">
                Foto #{{ editingCaptionIndex + 1 }}
                <q-badge v-if="editingCaptionIndex === 0" color="positive" class="q-ml-xs">
                  {{ t('broadcasters.photosDialog.coverBadge') }}
                </q-badge>
              </div>
              <div class="text-caption text-grey-6">
                {{ t('broadcasters.photosDialog.captionDialogHint') }}
              </div>
            </div>
          </div>

          <div class="row items-center justify-between q-mb-sm">
            <div class="row items-center q-gutter-xs">
              <q-toggle
                v-if="currentEditingCaptionPhoto"
                v-model="currentEditingCaptionPhoto.caption_enabled"
                dense
                size="xs"
                color="deep-orange-7"
                @update:model-value="onToggleCaption(currentEditingCaptionPhoto, $event)"
              />
              <span class="text-caption text-weight-bold text-grey-8">
                {{ t('broadcasters.photosDialog.captionLabel') }}
              </span>
            </div>
            <span
              v-if="currentEditingCaptionPhoto?.caption_enabled"
              class="text-caption text-grey-6 text-weight-medium"
            >
              (X: {{ currentEditingCaptionPhoto?.box_x ?? 0 }}%, Y: {{ currentEditingCaptionPhoto?.box_y ?? 0 }}%)
            </span>
          </div>
          <q-editor
            v-model="captionDraft"
            :disable="!currentEditingCaptionPhoto?.caption_enabled"
            min-height="6.5rem"
            max-height="15rem"
            :placeholder="currentEditingCaptionPhoto?.caption_enabled ? t('broadcasters.photosDialog.captionPlaceholder') : t('broadcasters.photosDialog.captionDisabled')"
            :toolbar="editorToolbar"
            :fonts="editorFonts"
            dense
            autofocus
            class="photo-editor"
          />
        </q-card-section>

        <q-separator style="flex-shrink: 0" />

        <!-- Footer Azioni -->
        <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
          <q-btn flat :label="t('actions.cancel')" v-close-popup />
          <q-btn
            color="deep-orange-7"
            icon="check"
            :label="t('actions.save')"
            unelevated
            @click="saveCaptionDraft"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Dialog, Notify } from 'quasar';
import {
  useBroadcastersStore,
  type Broadcaster,
  type BroadcasterPhoto,
} from '@/stores/broadcasters-store';
import { useImagesStore } from '@/stores/images-store';
import CropperImage from '@/components/CropperImage.vue';

export type BroadcasterPhotoItem = BroadcasterPhoto & {
  caption_enabled?: boolean;
};

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
const imagesStore = useImagesStore();

const loading = ref(false);
const localError = ref<string | null>(null);

const photosList = ref<BroadcasterPhotoItem[]>([]);

// Stato Cropper
const fileInputRef = ref<HTMLInputElement | null>(null);
const replaceFileInputRef = ref<HTMLInputElement | null>(null);
const cropperRef = ref<InstanceType<typeof CropperImage> | null>(null);
const cropperDialogOpen = ref(false);
const cropperReady = ref(false);
const cropperSource = ref<string | null>(null);
const photoNaturalDimensions = ref<{ width: number; height: number } | null>(null);
const newPhotoCaptionEnabled = ref(false);
const newPhotoAlt = ref('');
const croppingLoading = ref(false);

const editingPhotoIndex = ref<number | null>(null);
const loadingPhotoIndex = ref<number | null>(null);

// Stato Dialog Didascalia
const captionDialogOpen = ref(false);
const editingCaptionIndex = ref<number | null>(null);
const captionDraft = ref('');
const currentEditingCaptionPhoto = computed<BroadcasterPhotoItem | null>(() => {
  if (editingCaptionIndex.value === null) return null;
  return photosList.value[editingCaptionIndex.value] ?? null;
});

// Stato Drag Riquadro Didascalia su Photo-Ratio-Box
interface CaptionDragState {
  active: boolean;
  pointerId: number;
  startX: number;
  startY: number;
  initialLeftPct: number;
  initialTopPct: number;
  containerRect: DOMRect | null;
  boxRect: DOMRect | null;
  photoIndex: number;
}

const captionDragState = reactive<CaptionDragState>({
  active: false,
  pointerId: -1,
  startX: 0,
  startY: 0,
  initialLeftPct: 0,
  initialTopPct: 0,
  containerRect: null,
  boxRect: null,
  photoIndex: -1,
});

function getCaptionLeftPercent(photo: BroadcasterPhotoItem): number {
  const x = photo.box_x ?? 0;
  if (x > 100) {
    const baseW = imageConfig.value.width || 600;
    return Math.min(Math.max((x / baseW) * 100, 0), 85);
  }
  return Math.min(Math.max(x, 0), 85);
}

function getCaptionTopPercent(photo: BroadcasterPhotoItem): number {
  const y = photo.box_y ?? 0;
  if (y > 100) {
    const baseH = targetCropHeight.value || 900;
    return Math.min(Math.max((y / baseH) * 100, 0), 85);
  }
  return Math.min(Math.max(y, 0), 85);
}

function startCaptionDrag(event: PointerEvent, index: number) {
  event.preventDefault();
  event.stopPropagation();

  const target = event.currentTarget as HTMLElement;
  const container = target.closest('.photo-ratio-box') as HTMLElement;
  if (!container) return;

  target.setPointerCapture(event.pointerId);

  const containerRect = container.getBoundingClientRect();
  const boxRect = target.getBoundingClientRect();

  captionDragState.active = true;
  captionDragState.pointerId = event.pointerId;
  captionDragState.startX = event.clientX;
  captionDragState.startY = event.clientY;
  captionDragState.containerRect = containerRect;
  captionDragState.boxRect = boxRect;
  captionDragState.photoIndex = index;

  const currentLeftPx = boxRect.left - containerRect.left;
  const currentTopPx = boxRect.top - containerRect.top;

  captionDragState.initialLeftPct = containerRect.width > 0 ? (currentLeftPx / containerRect.width) * 100 : 0;
  captionDragState.initialTopPct = containerRect.height > 0 ? (currentTopPx / containerRect.height) * 100 : 0;
}

function onCaptionDrag(event: PointerEvent) {
  if (!captionDragState.active || event.pointerId !== captionDragState.pointerId) return;
  if (!captionDragState.containerRect || !captionDragState.boxRect) return;

  const photo = photosList.value[captionDragState.photoIndex];
  if (!photo || !photo.caption_enabled) return;

  const dx = event.clientX - captionDragState.startX;
  const dy = event.clientY - captionDragState.startY;

  const cWidth = captionDragState.containerRect.width;
  const cHeight = captionDragState.containerRect.height;
  const bWidth = captionDragState.boxRect.width;
  const bHeight = captionDragState.boxRect.height;

  if (cWidth <= 0 || cHeight <= 0) return;

  const startLeftPx = (captionDragState.initialLeftPct / 100) * cWidth;
  const startTopPx = (captionDragState.initialTopPct / 100) * cHeight;

  const maxLeftPx = Math.max(0, cWidth - bWidth);
  const maxTopPx = Math.max(0, cHeight - bHeight);

  const newLeftPx = Math.min(Math.max(startLeftPx + dx, 0), maxLeftPx);
  const newTopPx = Math.min(Math.max(startTopPx + dy, 0), maxTopPx);

  photo.box_x = Math.round((newLeftPx / cWidth) * 100);
  photo.box_y = Math.round((newTopPx / cHeight) * 100);
}

function stopCaptionDrag(event: PointerEvent) {
  if (!captionDragState.active || event.pointerId !== captionDragState.pointerId) return;

  const target = event.currentTarget as HTMLElement;
  if (target.hasPointerCapture(event.pointerId)) {
    target.releasePointerCapture(event.pointerId);
  }

  captionDragState.active = false;
  captionDragState.photoIndex = -1;
  captionDragState.containerRect = null;
  captionDragState.boxRect = null;
}

const imageConfig = computed(() => broadcastersStore.imageConfig);

const editorFonts = {
  arial: 'Arial',
  arial_black: 'Arial Black',
  comic_sans: 'Comic Sans MS',
  courier_new: 'Courier New',
  georgia: 'Georgia',
  impact: 'Impact',
  lucida_grande: 'Lucida Grande',
  times_new_roman: 'Times New Roman',
  verdana: 'Verdana',
};

const editorToolbar = [
  [
    {
      label: 'Font',
      icon: 'font_download',
      fixedIcon: true,
      list: 'no-icons',
      options: [
        'default_font',
        'arial',
        'arial_black',
        'comic_sans',
        'courier_new',
        'georgia',
        'impact',
        'lucida_grande',
        'times_new_roman',
        'verdana',
      ],
    },
    {
      label: 'Dimensione',
      icon: 'format_size',
      fixedIcon: true,
      list: 'no-icons',
      options: [
        'size-1',
        'size-2',
        'size-3',
        'size-4',
        'size-5',
        'size-6',
        'size-7',
      ],
    },
    'bold',
    'italic',
    'underline',
    'removeFormat',
  ],
];

function stripHtml(html: string): string {
  if (!html) return '';
  return html.replace(/<[^>]*>?/gm, '').trim();
}

const targetCropWidth = computed(() => imageConfig.value.width || 600);
const targetCropHeight = computed(() => {
  if (imageConfig.value.height && imageConfig.value.height > 0) {
    return imageConfig.value.height;
  }
  const prop = imageConfig.value.proportion || '2:3';
  const parts = prop.split(':').map((s) => parseFloat(s.trim()));
  const pw = parts[0];
  const ph = parts[1];
  if (parts.length === 2 && pw !== undefined && ph !== undefined && !isNaN(pw) && !isNaN(ph) && pw > 0 && ph > 0) {
    return Math.round((targetCropWidth.value / pw) * ph);
  }
  return Math.round((targetCropWidth.value * 3) / 2);
});

async function onShow() {
  localError.value = null;
  cropperDialogOpen.value = false;
  cropperReady.value = false;
  editingPhotoIndex.value = null;
  loadingPhotoIndex.value = null;
  photoNaturalDimensions.value = null;
  captionDialogOpen.value = false;
  editingCaptionIndex.value = null;
  captionDraft.value = '';
  newPhotoCaptionEnabled.value = false;
  newPhotoAlt.value = '';
  clearCropperSource();

  await Promise.all([
    broadcastersStore.fetchImageConfig(),
    imagesStore.fetchImages(),
  ]);

  if (props.broadcaster?.photos) {
    photosList.value = props.broadcaster.photos.map((p) => {
      const uuidMatch = p.image_url.match(/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}/);
      const matched = imagesStore.images.find((img) => img.url === p.image_url || (uuidMatch && img.uuid === uuidMatch[0]));
      const pDescText = p.description
        ? (typeof p.description === 'string' ? p.description : p.description.text)
        : '';
      const pDescBoxX = typeof p.description === 'object' ? p.description.box_x : undefined;
      const pDescBoxY = typeof p.description === 'object' ? p.description.box_y : undefined;

      const descBoxX = typeof matched?.description === 'object' ? matched.description.box_x : 0;
      const descBoxY = typeof matched?.description === 'object' ? matched.description.box_y : 0;
      const descText = matched?.description
        ? (typeof matched.description === 'string' ? matched.description : matched.description.text)
        : '';
      const altText = p.alt || pDescText || descText || '';
      const finalBoxX = p.box_x ?? pDescBoxX ?? descBoxX ?? 0;
      const finalBoxY = p.box_y ?? pDescBoxY ?? descBoxY ?? 0;
      const isEnabled = stripHtml(altText).length > 0 || finalBoxX > 0 || finalBoxY > 0;
      return {
        image_url: p.image_url,
        alt: isEnabled ? altText : '',
        box_x: isEnabled ? finalBoxX : 0,
        box_y: isEnabled ? finalBoxY : 0,
        caption_enabled: isEnabled,
        description: {
          text: isEnabled ? altText : '',
          box_x: isEnabled ? finalBoxX : 0,
          box_y: isEnabled ? finalBoxY : 0,
        },
      };
    });
  } else {
    photosList.value = [];
  }
}

function closeDialog() {
  emit('update:modelValue', false);
}

function triggerPickFile() {
  editingPhotoIndex.value = null;
  newPhotoCaptionEnabled.value = false;
  newPhotoAlt.value = '';
  fileInputRef.value?.click();
}

function triggerReplaceFile() {
  replaceFileInputRef.value?.click();
}

function triggerReplaceDirectly(index: number) {
  editingPhotoIndex.value = index;
  const target = photosList.value[index];
  newPhotoCaptionEnabled.value = target?.caption_enabled ?? false;
  newPhotoAlt.value = target?.caption_enabled ? (target?.alt || '') : '';
  triggerReplaceFile();
}

function onToggleCaption(photo: BroadcasterPhotoItem, val: boolean) {
  if (!val) {
    photo.box_x = 0;
    photo.box_y = 0;
    photo.alt = '';
    if (editingCaptionIndex.value !== null && photosList.value[editingCaptionIndex.value] === photo) {
      captionDraft.value = '';
    }
  }
}

function openCaptionDialog(index: number) {
  const photo = photosList.value[index];
  if (!photo || !photo.caption_enabled) return;
  editingCaptionIndex.value = index;
  captionDraft.value = photo.alt || '';
  captionDialogOpen.value = true;
}

function saveCaptionDraft() {
  if (editingCaptionIndex.value !== null) {
    const target = photosList.value[editingCaptionIndex.value];
    if (target) {
      if (!target.caption_enabled) {
        target.alt = '';
        target.box_x = 0;
        target.box_y = 0;
      } else {
        target.alt = captionDraft.value;
      }
    }
  }
  captionDialogOpen.value = false;
  editingCaptionIndex.value = null;
}

function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  clearCropperSource();
  cropperReady.value = false;
  photoNaturalDimensions.value = null;
  cropperSource.value = URL.createObjectURL(file);
  newPhotoCaptionEnabled.value = false;
  newPhotoAlt.value = '';

  // Pre-lettura dimensioni dell'immagine originale
  const img = new Image();
  img.onload = () => {
    photoNaturalDimensions.value = {
      width: img.naturalWidth,
      height: img.naturalHeight,
    };
  };
  img.src = cropperSource.value;

  cropperDialogOpen.value = true;
  input.value = '';
}

function onReplaceFileSelected(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  clearCropperSource();
  cropperReady.value = false;
  photoNaturalDimensions.value = null;
  cropperSource.value = URL.createObjectURL(file);

  if (editingPhotoIndex.value !== null) {
    const existing = photosList.value[editingPhotoIndex.value];
    newPhotoCaptionEnabled.value = existing?.caption_enabled ?? false;
    newPhotoAlt.value = existing?.caption_enabled ? (existing.alt || '') : '';
  } else {
    newPhotoCaptionEnabled.value = false;
    newPhotoAlt.value = '';
  }

  const img = new Image();
  img.onload = () => {
    photoNaturalDimensions.value = {
      width: img.naturalWidth,
      height: img.naturalHeight,
    };
  };
  img.src = cropperSource.value;

  cropperDialogOpen.value = true;
  cropperReady.value = true;
  void nextTick(() => {
    setTimeout(() => {
      cropperRef.value?.updateImageLayout(true);
      cropperRef.value?.resetToCenterAndZoom();
    }, 50);
  });
  input.value = '';
}

async function openEditCurrentPhoto(index: number) {
  editingPhotoIndex.value = index;
  const targetPhoto = photosList.value[index];
  if (!targetPhoto) return;

  newPhotoCaptionEnabled.value = targetPhoto.caption_enabled ?? false;
  newPhotoAlt.value = targetPhoto.caption_enabled ? (targetPhoto.alt || '') : '';
  loadingPhotoIndex.value = index;

  try {
    const res = await fetch(targetPhoto.image_url);
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}`);
    }
    const blob = await res.blob();
    clearCropperSource();
    cropperReady.value = false;
    photoNaturalDimensions.value = null;
    cropperSource.value = URL.createObjectURL(blob);

    const img = new Image();
    img.onload = () => {
      photoNaturalDimensions.value = {
        width: img.naturalWidth,
        height: img.naturalHeight,
      };
    };
    img.src = cropperSource.value;

    cropperDialogOpen.value = true;
  } catch (err) {
    console.warn('Fallback caricamento diretto URL:', err);
    clearCropperSource();
    cropperReady.value = false;
    cropperSource.value = targetPhoto.image_url;
    cropperDialogOpen.value = true;
  } finally {
    loadingPhotoIndex.value = null;
  }
}

function onCropperShow() {
  cropperReady.value = true;
  void nextTick(() => {
    setTimeout(() => {
      cropperRef.value?.updateImageLayout(true);
      cropperRef.value?.resetToCenterAndZoom();
    }, 60);
  });
}

function onCropperImageLoaded(data: { naturalWidth: number; naturalHeight: number }) {
  photoNaturalDimensions.value = {
    width: data.naturalWidth,
    height: data.naturalHeight,
  };
}

function clearCropperSource() {
  if (cropperSource.value && cropperSource.value.startsWith('blob:')) {
    URL.revokeObjectURL(cropperSource.value);
  }
  cropperSource.value = null;
}

function onCropperHide() {
  cropperReady.value = false;
  photoNaturalDimensions.value = null;
  clearCropperSource();
  editingPhotoIndex.value = null;
}

async function onCropConfirm() {
  const blob = await cropperRef.value?.getCroppedBlob('image/png');
  if (!blob) {
    Notify.create({
      type: 'warning',
      message: t('broadcasters.photosDialog.cropError'),
    });
    return;
  }

  croppingLoading.value = true;
  try {
    // Conversione Blob in Base64 Data URL
    const base64 = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });

    const isCaptionActive = newPhotoCaptionEnabled.value;
    const desc = isCaptionActive ? newPhotoAlt.value.trim() : '';
    const prevPhoto = editingPhotoIndex.value !== null ? photosList.value[editingPhotoIndex.value] : null;
    const initialBoxX = prevPhoto?.box_x ?? 0;
    const initialBoxY = prevPhoto?.box_y ?? 0;

    // Salvataggio dell'immagine ritagliata nel servizio images
    const saved = await imagesStore.saveImage({
      uuid: '',
      description: {
        text: desc,
        box_x: initialBoxX,
        box_y: initialBoxY,
      },
      data: base64,
      content_type: 'image/png',
      proportion: imageConfig.value.proportion || '2:3',
      width: targetCropWidth.value,
      height: targetCropHeight.value,
    });

    if (editingPhotoIndex.value !== null && editingPhotoIndex.value >= 0 && editingPhotoIndex.value < photosList.value.length) {
      // Modifica foto esistente
      photosList.value[editingPhotoIndex.value] = {
        image_url: saved.url,
        alt: desc,
        box_x: initialBoxX,
        box_y: initialBoxY,
        caption_enabled: isCaptionActive,
        description: {
          text: desc,
          box_x: initialBoxX,
          box_y: initialBoxY,
        },
      };
      Notify.create({
        type: 'positive',
        message: t('broadcasters.photosDialog.photoEditedToast') || 'Foto modificata con successo',
      });
    } else {
      // Aggiunta nuova foto alla lista
      photosList.value.push({
        image_url: saved.url,
        alt: desc,
        box_x: 0,
        box_y: 0,
        caption_enabled: isCaptionActive,
        description: {
          text: desc,
          box_x: 0,
          box_y: 0,
        },
      });
      Notify.create({
        type: 'positive',
        message: t('broadcasters.photosDialog.photoAddedToast'),
      });
    }

    cropperDialogOpen.value = false;
    clearCropperSource();
    editingPhotoIndex.value = null;
  } catch (err: unknown) {
    Notify.create({
      type: 'negative',
      message: err instanceof Error ? err.message : t('broadcasters.photosDialog.saveCropError'),
    });
  } finally {
    croppingLoading.value = false;
  }
}

function confirmRemovePhoto(index: number) {
  Dialog.create({
    title: t('broadcasters.photosDialog.confirmDeleteTitle') || 'Elimina foto',
    message: t('broadcasters.photosDialog.confirmDeleteMessage') || 'Sei sicuro di voler eliminare questa foto dalla galleria?',
    cancel: {
      label: t('actions.cancel') || 'Annulla',
      flat: true,
    },
    ok: {
      label: t('broadcasters.photosDialog.deletePhoto') || 'Elimina',
      color: 'negative',
      unelevated: true,
    },
    persistent: true,
  }).onOk(() => {
    removePhoto(index);
  });
}

function removePhoto(index: number) {
  photosList.value.splice(index, 1);
  Notify.create({
    type: 'info',
    message: t('broadcasters.photosDialog.photoRemovedToast'),
    timeout: 1500,
  });
}

function movePhoto(index: number, direction: number) {
  const target = index + direction;
  if (target < 0 || target >= photosList.value.length) return;
  const temp = photosList.value[index];
  if (!temp) return;
  photosList.value.splice(index, 1);
  photosList.value.splice(target, 0, temp);
}

function makeCover(index: number) {
  if (index <= 0) return;
  const temp = photosList.value[index];
  if (!temp) return;
  photosList.value.splice(index, 1);
  photosList.value.unshift(temp);
}

function onImgError(e: Event) {
  const target = e.target as HTMLImageElement;
  if (target) {
    target.src = 'https://placehold.co/600x900?text=Immagine+non+disponibile';
  }
}

async function onSave() {
  if (!props.broadcaster?.id) return;

  loading.value = true;
  localError.value = null;

  try {
    const cleaned = photosList.value
      .map((p) => {
        const isEnabled = !!p.caption_enabled;
        const text = isEnabled ? p.alt.trim() : '';
        const box_x = isEnabled ? Math.round(p.box_x || 0) : 0;
        const box_y = isEnabled ? Math.round(p.box_y || 0) : 0;
        return {
          image_url: p.image_url.trim(),
          alt: text,
          box_x,
          box_y,
          description: {
            text,
            box_x,
            box_y,
          },
        };
      })
      .filter((p) => p.image_url.length > 0);

    const updated = await broadcastersStore.updatePhotos(props.broadcaster.id, cleaned);

    // Sincronizza anche la descrizione e coordinate Box_x, Box_y nel servizio images
    for (const p of cleaned) {
      const uuidMatch = p.image_url.match(/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}/);
      const matched = imagesStore.images.find((img) => img.url === p.image_url || (uuidMatch && img.uuid === uuidMatch[0]));
      const targetUuid = matched?.uuid || uuidMatch?.[0];
      if (targetUuid) {
        try {
          await imagesStore.saveImage({
            uuid: targetUuid,
            description: {
              text: p.alt,
              box_x: p.box_x,
              box_y: p.box_y,
            },
            data: '',
            content_type: '',
            proportion: matched?.proportion || imageConfig.value.proportion || '2:3',
            width: matched?.width || targetCropWidth.value || 0,
            height: matched?.height || targetCropHeight.value || 0,
          });
        } catch (imgErr) {
          console.warn('Sync image description failed for:', p.image_url, imgErr);
        }
      }
    }
    Notify.create({
      type: 'positive',
      message: t('broadcasters.photosDialog.saveSuccess'),
    });
    emit('saved', updated);
    closeDialog();
  } catch (err: unknown) {
    localError.value = err instanceof Error ? err.message : String(err);
    Notify.create({
      type: 'negative',
      message: localError.value || t('broadcasters.photosDialog.saveError'),
    });
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.photo-card {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  border-radius: 8px;
  overflow: hidden;
}

.photo-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.photo-ratio-box {
  width: 100%;
  aspect-ratio: 2 / 3;
  position: relative;
  background-color: #f1f5f9;
  overflow: hidden;
}

.photo-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.photo-badges {
  position: absolute;
  top: 8px;
  left: 8px;
  display: flex;
  gap: 4px;
}

.cropper-container {
  background: #0f172a;
  border-radius: 8px;
  padding: 8px;
  overflow: hidden;
}

.cropper-box {
  width: 100% !important;
  max-width: 580px;
  padding: 0 !important;
  border: none !important;
  background: transparent !important;
}

.photo-editor {
  border-radius: 4px;
  font-size: 13px;
}

.photo-editor :deep(.q-editor__toolbars-container) {
  background: #f8fafc;
  padding: 2px 4px;
}

.photo-caption-section {
  min-height: 72px;
  background-color: #fafafa;
  border-top: 1px solid #f0f0f0;
}

.photo-caption-preview {
  padding: 6px 8px;
  background: #ffffff;
  border: 1px dashed #d1d5db;
  border-radius: 6px;
  min-height: 42px;
  max-height: 64px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.photo-caption-preview:hover {
  border-color: #ff5722;
  background: #fffbf9;
}

.photo-caption-preview--disabled {
  background: #f1f5f9;
  border: 1px dashed #cbd5e1;
  cursor: not-allowed !important;
  opacity: 0.8;
}

.photo-caption-preview--disabled:hover {
  border-color: #cbd5e1 !important;
  background: #f1f5f9 !important;
}

.caption-preview-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 12px;
  line-height: 1.35;
}

/* Div Assoluto Spostabile (Drag) con Didascalia */
.caption-drag-box {
  position: absolute;
  z-index: 5;
  cursor: grab;
  touch-action: none;
  user-select: none;
  max-width: 85%;
  pointer-events: auto;
  transition: box-shadow 0.15s ease, border-color 0.15s ease;
}

.caption-drag-box:active,
.caption-drag-box--dragging {
  cursor: grabbing !important;
  z-index: 10;
}

.caption-drag-inner {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  color: #ffffff;
  border: 1px dashed rgba(255, 255, 255, 0.7);
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 11px;
  line-height: 1.3;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.9);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
  transition: all 0.2s ease;
}

.caption-drag-box:hover .caption-drag-inner,
.caption-drag-box--dragging .caption-drag-inner {
  border-color: #ff5722;
  background: transparent;
  box-shadow: 0 0 0 2px rgba(255, 87, 34, 0.45), 0 4px 12px rgba(0, 0, 0, 0.4);
}

.caption-drag-handle {
  color: #ff7043;
  flex-shrink: 0;
}

.caption-drag-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-word;
}

.caption-drag-placeholder {
  font-style: italic;
  color: rgba(255, 255, 255, 0.65);
}
</style>
