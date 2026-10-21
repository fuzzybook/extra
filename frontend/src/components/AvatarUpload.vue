<template>
  <div class="row items-center q-gutter-md">
    <q-avatar size="64px">
      <img v-if="modelValue" :src="modelValue" />
      <q-icon v-else name="person" size="40px" />
    </q-avatar>

    <div class="row items-center q-gutter-xs">
      <q-btn round flat color="primary" icon="upload" :aria-label="t('actions.chooseFile')" @click="pickFile">
        <q-tooltip>{{ t('actions.chooseFile') }}</q-tooltip>
      </q-btn>
      <q-btn round flat color="primary" icon="photo_camera" :aria-label="t('actions.camera')" @click="openCamera">
        <q-tooltip>{{ t('actions.camera') }}</q-tooltip>
      </q-btn>
      <q-btn
        v-if="modelValue"
        round
        flat
        color="warning"
        icon="delete"
        :aria-label="t('actions.remove')"
        @click="$emit('update:modelValue', '')"
      >
        <q-tooltip>{{ t('actions.remove') }}</q-tooltip>
      </q-btn>
    </div>

    <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onFileSelected" />

    <q-dialog v-model="cameraOpen" @hide="stopCamera">
      <q-card style="min-width: 350px; max-width: 90vw;">
        <q-card-section>
          <div class="text-h6">{{ t('actions.camera') }}</div>
        </q-card-section>

        <q-card-section>
          <video
            ref="videoRef"
            class="camera-preview"
            autoplay
            muted
            playsinline
          />
          <div v-if="cameraError" class="text-negative text-caption q-mt-sm">
            {{ cameraError }}
          </div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn v-close-popup flat :label="t('actions.cancel')" />
          <q-btn color="primary" :label="t('actions.takePhoto')" :disable="!cameraReady" @click="capturePhoto" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <q-dialog v-model="cropperOpen" @show="ready = true" @hide="onHide">
      <q-card style="min-width: 350px; max-width: 90vw; position: relative;">
        <q-card-section>
          <div class="text-h6">{{ t('avatar.cropAvatar') }}</div>
        </q-card-section>

        <q-card-section>
          <CropperImage
            v-if="source && ready"
            ref="cropperRef"
            class="cropper"
            :src="source"
            :width="220"
            :height="220"
            :show-controls="false"
            :show-preview="false"
          />
        </q-card-section>

        <q-card-actions align="right">
          <q-btn v-close-popup flat :label="t('actions.cancel')" />
          <q-btn color="primary" :label="t('actions.save')" :loading="uploading" @click="cropAndUpload" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </div>
</template>

<style scoped>
.cropper {
  width: min(420px, 82vw);
  padding: 0;
  border: 0;
  background: #DDD;
}

.camera-preview {
  display: block;
  width: min(420px, 82vw);
  max-height: 420px;
  background: #222;
}
</style>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import CropperImage from '@/components/CropperImage.vue';

const props = defineProps<{
  modelValue: string;
  /** Uploads the cropped image and resolves to its public URL. */
  uploader: (image: Blob) => Promise<string>;
}>();
const { t } = useI18n();

const emit = defineEmits<{
  'update:modelValue': [url: string];
}>();

const fileInput = ref<HTMLInputElement | null>(null);
const cropperRef = ref<InstanceType<typeof CropperImage> | null>(null);
const videoRef = ref<HTMLVideoElement | null>(null);
const cropperOpen = ref(false);
const cameraOpen = ref(false);
const ready = ref(false);
const uploading = ref(false);
const cameraReady = ref(false);
const cameraError = ref<string | null>(null);
const source = ref<string | null>(null);
let cameraStream: MediaStream | null = null;

function pickFile() {
  fileInput.value?.click();
}

function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  setCropperSource(file);
  // Allow selecting the same file again later.
  input.value = '';
}

function onHide() {
  ready.value = false;
  clearCropperSource();
}

async function openCamera() {
  cameraError.value = null;
  cameraReady.value = false;
  cameraOpen.value = true;
  await nextTick();
  await startCamera();
}

async function startCamera() {
  if (!navigator.mediaDevices?.getUserMedia) {
    cameraError.value = t('avatar.cameraUnavailable');
    return;
  }

  try {
    stopCamera();
    cameraStream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: 'user' },
      audio: false,
    });

    if (!videoRef.value) return;
    videoRef.value.srcObject = cameraStream;
    await videoRef.value.play();
    cameraReady.value = true;
  } catch (err) {
    cameraError.value = err instanceof Error ? err.message : String(err);
    stopCamera();
  }
}

function stopCamera() {
  cameraStream?.getTracks().forEach((track) => track.stop());
  cameraStream = null;
  cameraReady.value = false;

  if (videoRef.value) {
    videoRef.value.srcObject = null;
  }
}

async function capturePhoto() {
  const video = videoRef.value;
  if (!video?.videoWidth || !video.videoHeight) return;

  const canvas = document.createElement('canvas');
  canvas.width = video.videoWidth;
  canvas.height = video.videoHeight;

  const context = canvas.getContext('2d');
  if (!context) return;

  context.drawImage(video, 0, 0, canvas.width, canvas.height);
  const blob = await new Promise<Blob | null>((resolve) => {
    canvas.toBlob((b) => resolve(b), 'image/png');
  });
  if (!blob) return;

  cameraOpen.value = false;
  stopCamera();
  setCropperSource(blob);
}

function setCropperSource(file: Blob) {
  clearCropperSource();
  source.value = URL.createObjectURL(file);
  cropperOpen.value = true;
}

function clearCropperSource() {
  if (!source.value) return;
  URL.revokeObjectURL(source.value);
  source.value = null;
}

async function cropAndUpload() {
  const blob = await cropperRef.value?.getCroppedBlob('image/png');
  if (!blob) return;
  uploading.value = true;
  try {
    const url = await props.uploader(blob);
    emit('update:modelValue', url);
    cropperOpen.value = false;
  } catch {
    // Errors surface through the store's shared error state.
  } finally {
    uploading.value = false;
  }
}

onBeforeUnmount(() => {
  stopCamera();
  clearCropperSource();
});
</script>
