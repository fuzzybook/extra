<template>
  <div class="cropper-image">
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
        @load="updateImageLayout"
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

    <div v-if="showControls" class="cropper-controls">
      <button type="button" @click="cropImage">{{ t('actions.crop') }}</button>
    </div>

    <canvas
      ref="canvasRef"
      class="cropper-preview"
      :class="{ 'cropper-preview--hidden': !showPreview }"
      :width="width"
      :height="height"
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
    src?: string;
    width?: number;
    height?: number;
    showControls?: boolean;
    showPreview?: boolean;
  }>(),
  {
    src: cropperImageUrl,
    width: 45,
    height: 45,
    showControls: true,
    showPreview: true,
  },
);

const imageSrc = computed(() => props.src);
const stageRef = ref<HTMLDivElement | null>(null);
const imageRef = ref<HTMLImageElement | null>(null);
const canvasRef = ref<HTMLCanvasElement | null>(null);
let resizeObserver: ResizeObserver | null = null;
let baseImageScale = 1;

const imageLayout = reactive({
  left: 0,
  top: 0,
  width: 0,
  height: 0,
  zoom: 1,
});

const selection = reactive({
  x: 20,
  y: 15,
  width: 45,
  height: 45,
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

const maxX = computed(() => 100 - selection.width);
const maxY = computed(() => 100 - selection.height);

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
  () => [props.width, props.height],
  () => {
    updateSelectionSizeFromProps();
  },
);

watch(
  () => props.src,
  async () => {
    imageLayout.zoom = 1;
    imageLayout.width = 0;
    imageLayout.height = 0;
    await nextTick();
    updateImageLayout();
  },
);

onMounted(async () => {
  await nextTick();
  updateImageLayout();

  if (stageRef.value) {
    resizeObserver = new ResizeObserver(() => {
      updateImageLayout();
    });
    resizeObserver.observe(stageRef.value);
  }
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
});

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

  const stageRect = stage.getBoundingClientRect();
  const maskLeft = (selection.x / 100) * stageRect.width;
  const maskTop = (selection.y / 100) * stageRect.height;
  const maskWidth = (selection.width / 100) * stageRect.width;
  const maskHeight = (selection.height / 100) * stageRect.height;

  const sx = ((maskLeft - imageLayout.left) / imageLayout.width) * image.naturalWidth;
  const sy = ((maskTop - imageLayout.top) / imageLayout.height) * image.naturalHeight;
  const sw = (maskWidth / imageLayout.width) * image.naturalWidth;
  const sh = (maskHeight / imageLayout.height) * image.naturalHeight;

  const context = canvas.getContext('2d');
  if (!context) return;

  context.clearRect(0, 0, canvas.width, canvas.height);
  context.drawImage(image, sx, sy, sw, sh, 0, 0, canvas.width, canvas.height);
}

async function getCroppedBlob(type = 'image/png', quality?: number): Promise<Blob | null> {
  cropImage();
  const canvas = canvasRef.value;
  if (!canvas) return null;

  return await new Promise((resolve) => {
    canvas.toBlob((blob) => resolve(blob), type, quality);
  });
}

defineExpose({
  cropImage,
  getCroppedBlob,
});

function updateImageLayout() {
  const image = imageRef.value;
  const stage = stageRef.value;
  if (!image || !stage || !image.naturalWidth || !image.naturalHeight) return;

  const rect = stage.getBoundingClientRect();
  if (!rect.width || !rect.height) return;

  const hadLayout = imageLayout.width > 0 && imageLayout.height > 0;
  const visibleCenterX = hadLayout ? rect.width / 2 - imageLayout.left : 0;
  const visibleCenterY = hadLayout ? rect.height / 2 - imageLayout.top : 0;
  const centerRatioX = hadLayout ? visibleCenterX / imageLayout.width : 0.5;
  const centerRatioY = hadLayout ? visibleCenterY / imageLayout.height : 0.5;

  baseImageScale = Math.min(1, rect.width / image.naturalWidth, rect.height / image.naturalHeight);
  const scale = baseImageScale * imageLayout.zoom;
  imageLayout.width = image.naturalWidth * scale;
  imageLayout.height = image.naturalHeight * scale;
  imageLayout.left = hadLayout
    ? rect.width / 2 - imageLayout.width * centerRatioX
    : (rect.width - imageLayout.width) / 2;
  imageLayout.top = hadLayout
    ? rect.height / 2 - imageLayout.height * centerRatioY
    : (rect.height - imageLayout.height) / 2;

  clampImagePosition();
  updateSelectionSizeFromProps();
}

function zoomImage(event: WheelEvent) {
  const zoomStep = event.deltaY < 0 ? 1.1 : 0.9;
  imageLayout.zoom = clamp(imageLayout.zoom * zoomStep, 1, 6);
  updateImageLayout();
}

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

  const rect = stage.getBoundingClientRect();
  const dx = ((event.clientX - dragState.startX) / rect.width) * 100;
  const dy = ((event.clientY - dragState.startY) / rect.height) * 100;

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

function updateSelectionSizeFromProps() {
  const rect = stageRef.value?.getBoundingClientRect();
  if (!rect?.width || !rect.height) return;

  const maxMaskWidth = imageLayout.width ? Math.min(rect.width, imageLayout.width) : rect.width;
  const maxMaskHeight = imageLayout.height ? Math.min(rect.height, imageLayout.height) : rect.height;

  selection.width = sizePixelsToPercent(props.width, rect.width, maxMaskWidth);
  selection.height = sizePixelsToPercent(props.height, rect.height, maxMaskHeight);
  clampSelectionToImage();
  cropImage();
}

function sizePixelsToPercent(value: number, total: number, maxValue = total) {
  return (clamp(value, 10, maxValue) / total) * 100;
}

function positionPixelsToPercent(value: number, total: number) {
  return (clamp(value, 0, total) / total) * 100;
}

function clampSelectionToImage() {
  const rect = stageRef.value?.getBoundingClientRect();
  if (!rect?.width || !rect.height) return;

  if (!imageLayout.width || !imageLayout.height) {
    selection.x = Math.min(selection.x, maxX.value);
    selection.y = Math.min(selection.y, maxY.value);
    return;
  }

  const maskWidth = (selection.width / 100) * rect.width;
  const maskHeight = (selection.height / 100) * rect.height;
  const minX = clamp(imageLayout.left, 0, rect.width - maskWidth);
  const minY = clamp(imageLayout.top, 0, rect.height - maskHeight);
  const maxXPosition = clamp(
    imageLayout.left + imageLayout.width - maskWidth,
    minX,
    rect.width - maskWidth,
  );
  const maxYPosition = clamp(
    imageLayout.top + imageLayout.height - maskHeight,
    minY,
    rect.height - maskHeight,
  );

  selection.x = positionPixelsToPercent(
    clamp((selection.x / 100) * rect.width, minX, maxXPosition),
    rect.width,
  );
  selection.y = positionPixelsToPercent(
    clamp((selection.y / 100) * rect.height, minY, maxYPosition),
    rect.height,
  );
}

function canDragImage() {
  const rect = stageRef.value?.getBoundingClientRect();
  if (!rect?.width || !rect.height) return false;

  return imageLayout.width > rect.width || imageLayout.height > rect.height;
}

function clampImagePosition() {
  const rect = stageRef.value?.getBoundingClientRect();
  if (!rect?.width || !rect.height || !imageLayout.width || !imageLayout.height) return;

  if (imageLayout.width <= rect.width) {
    imageLayout.left = (rect.width - imageLayout.width) / 2;
  } else {
    imageLayout.left = clamp(imageLayout.left, rect.width - imageLayout.width, 0);
  }

  if (imageLayout.height <= rect.height) {
    imageLayout.top = (rect.height - imageLayout.height) / 2;
  } else {
    imageLayout.top = clamp(imageLayout.top, rect.height - imageLayout.height, 0);
  }
}
</script>

<style scoped>
.cropper-image {
  display: grid;
  gap: 16px;
  width: min(640px, 90vw);
  padding: 16px;
  border: 1px solid #d0d0d0;
  background: #fff;
}

.cropper-stage {
  position: relative;
  height: 420px;
  overflow: hidden;
  background: #ddd;
  cursor: grab;
  touch-action: none;
  user-select: none;
}

.cropper-stage:active {
  cursor: grabbing;
}

.cropper-source {
  position: absolute;
  display: block;
  max-width: none;
  user-select: none;
}

.cropper-mask {
  position: absolute;
  border: 2px solid #1976d2;
  box-shadow: 0 0 0 9999px rgb(0 0 0 / 45%);
  cursor: move;
  touch-action: none;
  user-select: none;
}

.cropper-controls {
  display: grid;
  gap: 10px;
}

.cropper-controls button {
  justify-self: start;
  padding: 8px 14px;
  border: 0;
  color: #fff;
  background: #1976d2;
  cursor: pointer;
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
