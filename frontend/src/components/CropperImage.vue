<template>
  <div class="cropper-image">
    <!-- Barra Info Dimensioni e Zoom -->
    <div class="cropper-info-bar row items-center justify-between text-caption q-px-xs q-pb-xs">
      <div class="row items-center q-gutter-xs">
        <q-badge color="grey-8" text-color="white" class="q-py-xs q-px-sm">
          <q-icon name="image" size="14px" class="q-mr-xs" />
          <span>Foto: {{ photoDimensionsText }}</span>
        </q-badge>
        <q-badge color="primary" text-color="white" class="q-py-xs q-px-sm">
          <q-icon name="crop" size="14px" class="q-mr-xs" />
          <span>Output: {{ targetDimensions.width }} × {{ targetDimensions.height }} px ({{ proportionText }})</span>
        </q-badge>
      </div>
      <div>
        <q-badge color="blue-grey-8" text-color="white" class="q-py-xs q-px-sm">
          Zoom: {{ Math.round(imageLayout.zoom * 100) }}%
        </q-badge>
      </div>
    </div>

    <!-- Stage di Ritaglio -->
    <div
      ref="stageRef"
      class="cropper-stage"
      @wheel.prevent="zoomImage"
      @pointerdown="startImageDrag"
      @pointermove="dragImage"
      @pointerup="stopImageDrag"
      @pointercancel="stopImageDrag"
    >
      <img
        ref="imageRef"
        class="cropper-source"
        :style="imageStyle"
        :src="imageSrc"
        alt="Crop source"
        draggable="false"
        @load="onImageLoad"
      />

      <div
        class="cropper-mask"
        :style="selectionStyle"
        @pointerdown="startDrag"
        @pointermove="dragSelection"
        @pointerup="stopDrag"
        @pointercancel="stopDrag"
      />
    </div>

    <!-- Barra Controlli Zoom e Posizionamento -->
    <div v-if="showZoomControls" class="cropper-zoom-toolbar row items-center justify-between q-pt-xs">
      <div class="row items-center q-gutter-xs">
        <q-btn
          flat
          round
          dense
          color="primary"
          icon="zoom_out"
          :disable="imageLayout.zoom <= 0.2"
          @click="zoomOut"
        >
          <q-tooltip>Riduci Zoom (-)</q-tooltip>
        </q-btn>

        <div style="width: 140px" class="q-mx-xs">
          <q-slider
            :model-value="imageLayout.zoom"
            :min="0.2"
            :max="3.5"
            :step="0.05"
            dense
            color="primary"
            @update:model-value="onSliderZoomChange"
          />
        </div>

        <q-btn
          flat
          round
          dense
          color="primary"
          icon="zoom_in"
          :disable="imageLayout.zoom >= 3.5"
          @click="zoomIn"
        >
          <q-tooltip>Aumenta Zoom (+)</q-tooltip>
        </q-btn>
      </div>

      <div class="row items-center q-gutter-xs">
        <q-btn
          flat
          dense
          size="sm"
          color="primary"
          icon="fit_screen"
          label="Adatta"
          @click="fitToStage"
        >
          <q-tooltip>Mostra intera foto</q-tooltip>
        </q-btn>

        <q-btn
          flat
          dense
          size="sm"
          color="primary"
          icon="fullscreen"
          label="Riempi"
          @click="fillCrop"
        >
          <q-tooltip>Riempi maschera di ritaglio</q-tooltip>
        </q-btn>

        <q-btn
          flat
          round
          dense
          color="secondary"
          icon="center_focus_strong"
          @click="resetToCenterAndZoom"
        >
          <q-tooltip>Ripristina posizione e zoom</q-tooltip>
        </q-btn>
      </div>
    </div>

    <!-- Controlli classici con pulsante ritaglio se richiesto -->
    <div v-if="showControls" class="cropper-controls">
      <q-btn color="primary" unelevated :label="t('actions.crop')" @click="cropImage" />
    </div>

    <!-- Canvas Preview nascosto o visibile -->
    <canvas
      ref="canvasRef"
      class="cropper-preview"
      :class="{ 'cropper-preview--hidden': !showPreview }"
      :width="targetDimensions.width"
      :height="targetDimensions.height"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import cropperImageUrl from '@/assets/pexels-photo-4323307.jpg';

const { t } = useI18n();

const props = withDefaults(
  defineProps<{
    src?: string | undefined;
    width?: number | undefined;
    height?: number | undefined;
    proportion?: string | undefined;
    showControls?: boolean | undefined;
    showPreview?: boolean | undefined;
    showZoomControls?: boolean | undefined;
  }>(),
  {
    src: cropperImageUrl,
    width: 230,
    height: 0,
    proportion: '2:3',
    showControls: true,
    showPreview: true,
    showZoomControls: true,
  },
);

const emit = defineEmits<{
  (e: 'image-loaded', data: { naturalWidth: number; naturalHeight: number; targetWidth: number; targetHeight: number }): void;
  (e: 'crop', blob: Blob | null): void;
}>();

const naturalWidth = ref(0);
const naturalHeight = ref(0);

const photoDimensionsText = computed(() => {
  if (naturalWidth.value > 0 && naturalHeight.value > 0) {
    return `${naturalWidth.value} × ${naturalHeight.value} px`;
  }
  return 'Caricamento...';
});

const proportionText = computed(() => {
  return props.proportion && props.proportion.trim().length > 0 ? props.proportion : '2:3';
});

const targetDimensions = computed(() => {
  const w = props.width && props.width > 0 ? props.width : 230;
  // Se height != 0 significa che si deve utilizzare l'altezza specificata
  if (props.height !== undefined && props.height !== 0) {
    return { width: w, height: props.height };
  }
  // Se height == 0 significa che si deve usare la proporzione
  if (props.proportion && props.proportion.trim().length > 0) {
    const parts = props.proportion.split(':').map((s) => parseFloat(s.trim()));
    const pw = parts[0];
    const ph = parts[1];
    if (parts.length === 2 && pw !== undefined && ph !== undefined && !isNaN(pw) && !isNaN(ph) && pw > 0 && ph > 0) {
      const h = Math.round((w / pw) * ph);
      return { width: w, height: h };
    }
  }
  return { width: w, height: Math.round((w * 3) / 2) };
});

const imageSrc = computed(() => props.src);
const stageRef = ref<HTMLDivElement | null>(null);
const imageRef = ref<HTMLImageElement | null>(null);
const canvasRef = ref<HTMLCanvasElement | null>(null);
let resizeObserver: ResizeObserver | null = null;
let baseImageScale = 1;
let retryAnimationFrameId: number | null = null;
let retryCount = 0;

const imageLayout = reactive({
  left: 0,
  top: 0,
  width: 0,
  height: 0,
  zoom: 1,
});

const selection = reactive({
  x: 10,
  y: 10,
  width: 80,
  height: 80,
});

const dragState = reactive({
  active: false,
  pointerId: 0,
  startX: 0,
  startY: 0,
  selectionX: 0,
  selectionY: 0,
});

const imageDragState = reactive({
  active: false,
  pointerId: 0,
  startX: 0,
  startY: 0,
  imageLeft: 0,
  imageTop: 0,
});

const maxX = computed(() => Math.max(0, 100 - selection.width));
const maxY = computed(() => Math.max(0, 100 - selection.height));

const imageStyle = computed(() => ({
  left: `${imageLayout.left}px`,
  top: `${imageLayout.top}px`,
  width: `${imageLayout.width}px`,
  height: `${imageLayout.height}px`,
}));

const selectionStyle = computed(() => ({
  left: `${selection.x}%`,
  top: `${selection.y}%`,
  width: `${selection.width}%`,
  height: `${selection.height}%`,
}));

watch(
  () => [selection.x, selection.y, selection.width, selection.height],
  () => {
    selection.x = Math.min(selection.x, maxX.value);
    selection.y = Math.min(selection.y, maxY.value);
    cropImage();
  },
);

watch(
  () => [props.width, props.height, props.proportion],
  () => {
    updateImageLayout(true);
  },
);

watch(
  () => props.src,
  async () => {
    imageLayout.zoom = 1;
    imageLayout.width = 0;
    imageLayout.height = 0;
    await nextTick();
    updateImageLayout(true);
  },
);

onMounted(async () => {
  await nextTick();
  updateImageLayout(true);

  if (stageRef.value) {
    resizeObserver = new ResizeObserver(() => {
      updateImageLayout(false);
    });
    resizeObserver.observe(stageRef.value);
  }

  // Se l'immagine è già caricata in cache
  const img = imageRef.value;
  if (img && img.complete && img.naturalWidth > 0) {
    onImageLoad();
  }

  // Safety timeout per assicurare il calcolo a transizione modal ultimata
  setTimeout(() => {
    updateImageLayout(true);
  }, 150);
});

onBeforeUnmount(() => {
  if (retryAnimationFrameId !== null) {
    cancelAnimationFrame(retryAnimationFrameId);
    retryAnimationFrameId = null;
  }
  resizeObserver?.disconnect();
});

function onImageLoad() {
  const image = imageRef.value;
  if (image && image.naturalWidth && image.naturalHeight) {
    naturalWidth.value = image.naturalWidth;
    naturalHeight.value = image.naturalHeight;
    emit('image-loaded', {
      naturalWidth: image.naturalWidth,
      naturalHeight: image.naturalHeight,
      targetWidth: targetDimensions.value.width,
      targetHeight: targetDimensions.value.height,
    });
  }
  updateImageLayout(true);
}

function getStageDimensions(): { width: number; height: number } {
  const stage = stageRef.value;
  if (!stage) return { width: 0, height: 0 };

  const clientW = stage.clientWidth;
  const clientH = stage.clientHeight;
  const rect = stage.getBoundingClientRect();

  // clientWidth e clientHeight rappresentano le dimensioni reali di layout non alterate dalle animazioni di CSS transform (scale)
  const width = clientW > 0 ? clientW : rect.width;
  const height = clientH > 0 ? clientH : rect.height;

  return { width, height };
}

function updateImageLayout(resetToCenter = false) {
  const image = imageRef.value;
  const stage = stageRef.value;
  if (!image || !stage) return;
  if (!image.naturalWidth || !image.naturalHeight) return;

  const stageDim = getStageDimensions();
  // Se le dimensioni dello stage sono ancora minime o non calcolate (durante animazione modale), attendi
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) {
    if (retryCount < 30) {
      retryCount++;
      retryAnimationFrameId = requestAnimationFrame(() => updateImageLayout(resetToCenter));
    }
    return;
  }
  retryCount = 0;

  naturalWidth.value = image.naturalWidth;
  naturalHeight.value = image.naturalHeight;

  if (resetToCenter) {
    imageLayout.zoom = 1;
  }

  updateSelectionSizeFromProps(resetToCenter);

  const hadLayout = imageLayout.width > 0 && imageLayout.height > 0 && !resetToCenter;

  const maskLeft = (selection.x / 100) * stageDim.width;
  const maskTop = (selection.y / 100) * stageDim.height;
  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;
  const maskCenterX = maskLeft + maskWidth / 2;
  const maskCenterY = maskTop + maskHeight / 2;

  const curW = imageLayout.width || 1;
  const curH = imageLayout.height || 1;
  const centerRatioX = hadLayout ? (maskCenterX - imageLayout.left) / curW : 0.5;
  const centerRatioY = hadLayout ? (maskCenterY - imageLayout.top) / curH : 0.5;

  baseImageScale = Math.max(maskWidth / image.naturalWidth, maskHeight / image.naturalHeight);
  const scale = baseImageScale * imageLayout.zoom;
  imageLayout.width = image.naturalWidth * scale;
  imageLayout.height = image.naturalHeight * scale;

  imageLayout.left = hadLayout
    ? maskCenterX - imageLayout.width * centerRatioX
    : maskCenterX - imageLayout.width / 2;
  imageLayout.top = hadLayout
    ? maskCenterY - imageLayout.height * centerRatioY
    : maskCenterY - imageLayout.height / 2;

  clampImagePosition();
  clampSelectionToImage();
  cropImage();
}

function setZoom(newZoom: number, focusX?: number, focusY?: number) {
  const stage = stageRef.value;
  const image = imageRef.value;
  if (!stage || !image || !image.naturalWidth || !image.naturalHeight) return;

  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) return;

  const clampedZoom = clamp(newZoom, 0.2, 4);
  if (Math.abs(clampedZoom - imageLayout.zoom) < 0.0001) return;

  const maskLeft = (selection.x / 100) * stageDim.width;
  const maskTop = (selection.y / 100) * stageDim.height;
  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;
  const maskCenterX = maskLeft + maskWidth / 2;
  const maskCenterY = maskTop + maskHeight / 2;

  const cursorX = focusX !== undefined ? focusX : maskCenterX;
  const cursorY = focusY !== undefined ? focusY : maskCenterY;

  const curW = imageLayout.width || 1;
  const curH = imageLayout.height || 1;
  const relX = (cursorX - imageLayout.left) / curW;
  const relY = (cursorY - imageLayout.top) / curH;

  imageLayout.zoom = clampedZoom;
  const newScale = baseImageScale * clampedZoom;

  imageLayout.width = image.naturalWidth * newScale;
  imageLayout.height = image.naturalHeight * newScale;

  imageLayout.left = cursorX - relX * imageLayout.width;
  imageLayout.top = cursorY - relY * imageLayout.height;

  clampImagePosition();
  clampSelectionToImage();
  cropImage();
}

function zoomImage(event: WheelEvent) {
  const stage = stageRef.value;
  if (!stage) return;
  const rect = stage.getBoundingClientRect();
  if (!rect.width || !rect.height) return;

  const cursorX = event.clientX - rect.left;
  const cursorY = event.clientY - rect.top;
  const zoomFactor = event.deltaY < 0 ? 1.08 : 0.92;
  setZoom(imageLayout.zoom * zoomFactor, cursorX, cursorY);
}

function zoomIn() {
  setZoom(imageLayout.zoom * 1.15);
}

function zoomOut() {
  setZoom(imageLayout.zoom / 1.15);
}

function onSliderZoomChange(val: number | null) {
  if (val !== null && !isNaN(val)) {
    setZoom(val);
  }
}

function resetToCenterAndZoom() {
  updateImageLayout(true);
}

function fillCrop() {
  setZoom(1);
}

function fitToStage() {
  const image = imageRef.value;
  const stage = stageRef.value;
  if (!image || !stage || !image.naturalWidth || !image.naturalHeight) return;
  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) return;

  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;
  const fitScale = Math.min(maskWidth / image.naturalWidth, maskHeight / image.naturalHeight);
  if (baseImageScale > 0) {
    const targetZoom = fitScale / baseImageScale;
    setZoom(targetZoom);
  }
}

function cropImage() {
  const image = imageRef.value;
  const canvas = canvasRef.value;
  const stage = stageRef.value;
  if (
    !image ||
    !canvas ||
    !stage ||
    !image.naturalWidth ||
    !image.naturalHeight ||
    !imageLayout.width ||
    !imageLayout.height
  ) {
    return;
  }

  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) return;

  const maskLeft = (selection.x / 100) * stageDim.width;
  const maskTop = (selection.y / 100) * stageDim.height;
  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;

  canvas.width = targetDimensions.value.width;
  canvas.height = targetDimensions.value.height;

  const context = canvas.getContext('2d');
  if (!context) return;

  context.clearRect(0, 0, canvas.width, canvas.height);

  // Calcolo scaling del canvas per rendering senza problemi di fuori-bordo
  const scaleX = canvas.width / (maskWidth || 1);
  const scaleY = canvas.height / (maskHeight || 1);

  const dx = (imageLayout.left - maskLeft) * scaleX;
  const dy = (imageLayout.top - maskTop) * scaleY;
  const dw = imageLayout.width * scaleX;
  const dh = imageLayout.height * scaleY;

  context.drawImage(image, dx, dy, dw, dh);
}

async function getCroppedBlob(type = 'image/png', quality?: number): Promise<Blob | null> {
  cropImage();
  const canvas = canvasRef.value;
  if (!canvas) return null;

  return await new Promise((resolve) => {
    canvas.toBlob((blob) => {
      emit('crop', blob);
      resolve(blob);
    }, type, quality);
  });
}

defineExpose({
  cropImage,
  getCroppedBlob,
  updateImageLayout,
  setZoom,
  fitToStage,
  fillCrop,
  resetToCenterAndZoom,
});

function startDrag(event: PointerEvent) {
  const stage = stageRef.value;
  if (!stage) return;

  event.preventDefault();
  event.stopPropagation();
  const target = event.currentTarget as HTMLElement;
  target.setPointerCapture(event.pointerId);

  dragState.active = true;
  dragState.pointerId = event.pointerId;
  dragState.startX = event.clientX;
  dragState.startY = event.clientY;
  dragState.selectionX = selection.x;
  dragState.selectionY = selection.y;
}

function startImageDrag(event: PointerEvent) {
  const stage = stageRef.value;
  if (!stage || !canDragImage()) return;

  event.preventDefault();
  stage.setPointerCapture(event.pointerId);

  imageDragState.active = true;
  imageDragState.pointerId = event.pointerId;
  imageDragState.startX = event.clientX;
  imageDragState.startY = event.clientY;
  imageDragState.imageLeft = imageLayout.left;
  imageDragState.imageTop = imageLayout.top;
}

function dragImage(event: PointerEvent) {
  if (!imageDragState.active || event.pointerId !== imageDragState.pointerId) return;

  imageLayout.left = imageDragState.imageLeft + event.clientX - imageDragState.startX;
  imageLayout.top = imageDragState.imageTop + event.clientY - imageDragState.startY;
  clampImagePosition();
  clampSelectionToImage();
  cropImage();
}

function stopImageDrag(event: PointerEvent) {
  const stage = stageRef.value;
  if (!stage || !imageDragState.active || event.pointerId !== imageDragState.pointerId) return;

  if (stage.hasPointerCapture(event.pointerId)) {
    stage.releasePointerCapture(event.pointerId);
  }

  imageDragState.active = false;
}

function dragSelection(event: PointerEvent) {
  const stage = stageRef.value;
  if (!dragState.active || event.pointerId !== dragState.pointerId || !stage) return;

  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height) return;

  const dx = ((event.clientX - dragState.startX) / stageDim.width) * 100;
  const dy = ((event.clientY - dragState.startY) / stageDim.height) * 100;

  selection.x = clamp(dragState.selectionX + dx, 0, maxX.value);
  selection.y = clamp(dragState.selectionY + dy, 0, maxY.value);
  clampSelectionToImage();
}

function stopDrag(event: PointerEvent) {
  if (!dragState.active || event.pointerId !== dragState.pointerId) return;

  const target = event.currentTarget as HTMLElement;
  if (target.hasPointerCapture(event.pointerId)) {
    target.releasePointerCapture(event.pointerId);
  }

  dragState.active = false;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function updateSelectionSizeFromProps(resetPosition = false) {
  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) return;

  const stageWidth = stageDim.width;
  const stageHeight = stageDim.height;

  const targetW = targetDimensions.value.width;
  const targetH = targetDimensions.value.height;

  // Adatta le dimensioni della maschera allo stage
  const maxStageW = stageWidth * 0.88;
  const maxStageH = stageHeight * 0.88;

  const scale = Math.min(maxStageW / targetW, maxStageH / targetH);
  const maskPixelW = targetW * scale;
  const maskPixelH = targetH * scale;

  selection.width = (maskPixelW / stageWidth) * 100;
  selection.height = (maskPixelH / stageHeight) * 100;

  if (resetPosition || selection.x + selection.width > 100 || selection.x < 0) {
    selection.x = Math.max(0, (100 - selection.width) / 2);
  }
  if (resetPosition || selection.y + selection.height > 100 || selection.y < 0) {
    selection.y = Math.max(0, (100 - selection.height) / 2);
  }
}

function positionPixelsToPercent(value: number, total: number) {
  return (clamp(value, 0, total) / total) * 100;
}

function clampSelectionToImage() {
  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) return;

  if (!imageLayout.width || !imageLayout.height) {
    selection.x = Math.min(selection.x, maxX.value);
    selection.y = Math.min(selection.y, maxY.value);
    return;
  }

  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;

  const minMaskX = Math.max(0, imageLayout.left);
  const maxMaskX = Math.min(stageDim.width - maskWidth, imageLayout.left + imageLayout.width - maskWidth);

  const minMaskY = Math.max(0, imageLayout.top);
  const maxMaskY = Math.min(stageDim.height - maskHeight, imageLayout.top + imageLayout.height - maskHeight);

  const curMaskX = (selection.x / 100) * stageDim.width;
  const curMaskY = (selection.y / 100) * stageDim.height;

  const clampedX = maxMaskX >= minMaskX ? clamp(curMaskX, minMaskX, maxMaskX) : minMaskX;
  const clampedY = maxMaskY >= minMaskY ? clamp(curMaskY, minMaskY, maxMaskY) : minMaskY;

  selection.x = positionPixelsToPercent(clampedX, stageDim.width);
  selection.y = positionPixelsToPercent(clampedY, stageDim.height);
}

function canDragImage() {
  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120) return false;

  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;

  return imageLayout.width > maskWidth + 0.5 || imageLayout.height > maskHeight + 0.5;
}

function clampImagePosition() {
  const stageDim = getStageDimensions();
  if (!stageDim.width || !stageDim.height || stageDim.width < 120 || stageDim.height < 120 || !imageLayout.width || !imageLayout.height) return;

  const maskLeft = (selection.x / 100) * stageDim.width;
  const maskTop = (selection.y / 100) * stageDim.height;
  const maskWidth = (selection.width / 100) * stageDim.width;
  const maskHeight = (selection.height / 100) * stageDim.height;

  if (imageLayout.width >= maskWidth) {
    const minLeft = maskLeft + maskWidth - imageLayout.width;
    const maxLeft = maskLeft;
    imageLayout.left = clamp(imageLayout.left, minLeft, maxLeft);
  } else {
    imageLayout.left = maskLeft + (maskWidth - imageLayout.width) / 2;
  }

  if (imageLayout.height >= maskHeight) {
    const minTop = maskTop + maskHeight - imageLayout.height;
    const maxTop = maskTop;
    imageLayout.top = clamp(imageLayout.top, minTop, maxTop);
  } else {
    imageLayout.top = maskTop + (maskHeight - imageLayout.height) / 2;
  }
}
</script>

<style scoped>
.cropper-image {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  max-width: 600px;
  padding: 8px;
  border-radius: 8px;
  background: #1e293b;
  color: #fff;
}

.cropper-info-bar {
  background: rgba(15, 23, 42, 0.6);
  border-radius: 6px;
  padding: 6px 10px;
}

.cropper-stage {
  position: relative;
  height: 420px;
  overflow: hidden;
  background: #0b0f19;
  cursor: grab;
  touch-action: none;
  user-select: none;
  border-radius: 6px;
  box-shadow: inset 0 0 16px rgba(0, 0, 0, 0.6);
}

.cropper-stage:active {
  cursor: grabbing;
}

.cropper-source {
  position: absolute;
  display: block;
  max-width: none;
  user-select: none;
  pointer-events: none;
}

.cropper-mask {
  position: absolute;
  border: 2px solid #ff5722;
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.65);
  cursor: move;
  touch-action: none;
  user-select: none;
  border-radius: 4px;
}

.cropper-zoom-toolbar {
  background: rgba(15, 23, 42, 0.6);
  border-radius: 6px;
  padding: 4px 10px;
}

.cropper-controls {
  display: flex;
  justify-content: flex-end;
}

.cropper-preview {
  width: 220px;
  height: 220px;
  border: 1px solid #d0d0d0;
  background: #f5f5f5;
}

.cropper-preview--hidden {
  display: none;
}
</style>
