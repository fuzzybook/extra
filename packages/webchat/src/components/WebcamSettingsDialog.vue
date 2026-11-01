<template>
  <q-dialog :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)">
    <!-- Colore del testo esplicito: il dialog finisce nel body, che sul sito pubblico ha testo chiaro -->
    <q-card style="width: 620px; max-width: 95vw;" class="q-pa-sm bg-white text-dark">
      <q-card-section class="row items-center justify-between q-pb-none">
        <div class="row items-center q-gutter-sm">
          <q-icon name="videocam" color="primary" size="28px" />
          <div class="text-h6">{{ t('webchat.webcam.title') }}</div>
        </div>
        <q-btn flat round dense icon="close" v-close-popup />
      </q-card-section>

      <q-card-section class="q-pt-md">
        <q-form @submit.prevent="onSave">
          <div class="q-gutter-y-md">
            <!-- Preset Selection -->
            <div>
              <q-select
                popup-content-class="bg-white text-dark"
                v-model="form.preset"
                :options="presetOptions"
                emit-value
                map-options
                outlined
                dense
                :label="t('webchat.webcam.presetsLabel')"
                @update:model-value="applyPresetValues"
              >
                <template #prepend>
                  <q-icon name="auto_awesome" color="primary" />
                </template>
                <template #append>
                  <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                    <q-popup-proxy transition-show="scale" transition-hide="scale">
                      <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                        <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                          <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                          {{ t('webchat.webcam.presetsLabel') }}
                        </div>
                        <div class="text-caption text-grey-3">{{ t('webchat.webcam.presetsHelp') }}</div>
                      </q-card>
                    </q-popup-proxy>
                  </q-btn>
                </template>
              </q-select>
            </div>

            <!-- Camera Device Selection -->
            <div>
              <q-select
                popup-content-class="bg-white text-dark"
                v-model="form.deviceId"
                :options="cameraOptions"
                emit-value
                map-options
                outlined
                dense
                :label="t('webchat.webcam.device')"
                @update:model-value="onManualChange"
              >
                <template #prepend>
                  <q-icon name="photo_camera" color="primary" />
                </template>
                <template #append>
                  <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                    <q-popup-proxy transition-show="scale" transition-hide="scale">
                      <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                        <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                          <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                          {{ t('webchat.webcam.device') }}
                        </div>
                        <div class="text-caption text-grey-3">{{ t('webchat.webcam.deviceHelp') }}</div>
                      </q-card>
                    </q-popup-proxy>
                  </q-btn>
                </template>
              </q-select>
            </div>

            <!-- Width & Aspect Ratio (Proportions) -->
            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-input
                  v-model.number="form.width"
                  type="number"
                  min="160"
                  max="3840"
                  step="10"
                  outlined
                  dense
                  :label="t('webchat.webcam.width')"
                  @update:model-value="onWidthChange"
                >
                  <template #prepend>
                    <q-icon name="aspect_ratio" color="primary" />
                  </template>
                  <template #append>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.width') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.resolutionHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-input>
              </div>

              <div class="col-12 col-sm-6">
                <q-select
                  popup-content-class="bg-white text-dark"
                  v-model="form.aspectRatio"
                  :options="aspectRatioOptions"
                  emit-value
                  map-options
                  outlined
                  dense
                  :label="t('webchat.webcam.aspectRatio')"
                  @update:model-value="onAspectRatioChange"
                >
                  <template #prepend>
                    <q-icon name="crop" color="primary" />
                  </template>
                  <template #append>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.aspectRatio') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.aspectRatioHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-select>
              </div>
            </div>

            <!-- Height (Always calculated from aspect ratio) & Frame Rate -->
            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-input
                  :model-value="form.height"
                  type="number"
                  readonly
                  outlined
                  dense
                  bg-color="grey-2"
                  :label="t('webchat.webcam.heightCalculated')"
                >
                  <template #prepend>
                    <q-icon name="lock" size="18px" color="grey-7" />
                  </template>
                  <template #append>
                    <q-badge color="grey-4" text-color="grey-9" class="q-mr-xs text-weight-bold">
                      {{ form.width }} × {{ form.height }}
                    </q-badge>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.heightCalculated') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.heightCalculatedHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-input>
              </div>

              <div class="col-12 col-sm-6">
                <q-select
                  popup-content-class="bg-white text-dark"
                  v-model="form.frameRate"
                  :options="frameRateOptions"
                  emit-value
                  map-options
                  outlined
                  dense
                  :label="t('webchat.webcam.frameRate')"
                  @update:model-value="onManualChange"
                >
                  <template #prepend>
                    <q-icon name="speed" color="primary" />
                  </template>
                  <template #append>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.frameRate') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.frameRateHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-select>
              </div>
            </div>

            <!-- Bitrate (Transfer rate) & Video Codec -->
            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-input
                  v-model.number="form.bitrateKbps"
                  type="number"
                  min="200"
                  max="12000"
                  step="100"
                  outlined
                  dense
                  :label="t('webchat.webcam.bitrate')"
                  suffix="kbps"
                  @update:model-value="onManualChange"
                >
                  <template #prepend>
                    <q-icon name="network_check" color="primary" />
                  </template>
                  <template #append>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.bitrate') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.bitrateHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-input>
              </div>

              <div class="col-12 col-sm-6">
                <q-select
                  popup-content-class="bg-white text-dark"
                  v-model="form.codec"
                  :options="codecOptions"
                  emit-value
                  map-options
                  outlined
                  dense
                  :label="t('webchat.webcam.codec')"
                  @update:model-value="onManualChange"
                >
                  <template #prepend>
                    <q-icon name="memory" color="primary" />
                  </template>
                  <template #append>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.codec') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.codecHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-select>
              </div>
            </div>

            <!-- Degradation Strategy & Simulcast -->
            <div class="row q-col-gutter-sm items-center">
              <div class="col-12 col-sm-7">
                <q-select
                  popup-content-class="bg-white text-dark"
                  v-model="form.degradationPreference"
                  :options="degradationOptions"
                  emit-value
                  map-options
                  outlined
                  dense
                  :label="t('webchat.webcam.degradation')"
                  @update:model-value="onManualChange"
                >
                  <template #prepend>
                    <q-icon name="tune" color="primary" />
                  </template>
                  <template #append>
                    <q-btn round flat dense size="xs" icon="help_outline" color="primary" @click.stop>
                      <q-popup-proxy transition-show="scale" transition-hide="scale">
                        <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                          <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                            <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                            {{ t('webchat.webcam.degradation') }}
                          </div>
                          <div class="text-caption text-grey-3">{{ t('webchat.webcam.degradationHelp') }}</div>
                        </q-card>
                      </q-popup-proxy>
                    </q-btn>
                  </template>
                </q-select>
              </div>

              <div class="col-12 col-sm-5 row items-center justify-between no-wrap bg-grey-1 q-pa-xs rounded-borders">
                <q-toggle
                  v-model="form.simulcast"
                  :label="t('webchat.webcam.simulcast')"
                  dense
                  @update:model-value="onManualChange"
                />
                <q-btn round flat dense size="xs" icon="help_outline" color="primary" class="q-ml-xs">
                  <q-popup-proxy transition-show="scale" transition-hide="scale">
                    <q-card class="q-pa-md bg-grey-9 text-white shadow-6" style="max-width: 320px">
                      <div class="text-subtitle2 text-bold q-mb-xs row items-center">
                        <q-icon name="info" size="18px" color="primary" class="q-mr-xs" />
                        {{ t('webchat.webcam.simulcast') }}
                      </div>
                      <div class="text-caption text-grey-3">{{ t('webchat.webcam.simulcastHelp') }}</div>
                    </q-card>
                  </q-popup-proxy>
                </q-btn>
              </div>
            </div>
          </div>

          <q-card-actions align="right" class="q-mt-lg q-px-none">
            <q-btn flat :label="t('webchat.cancel')" v-close-popup />
            <q-btn type="submit" color="primary" icon="save" :label="t('webchat.webcam.apply')" />
          </q-card-actions>
        </q-form>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { calculateHeight, defaultWebcamSettings, type WebcamSettings } from '../webcam';

const props = defineProps<{
  modelValue: boolean;
  initialSettings?: Partial<WebcamSettings>;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'apply', settings: WebcamSettings): void;
}>();

const { t } = useI18n();

const availableCameras = ref<{ label: string; value: string }[]>([]);

const form = reactive<WebcamSettings>({ ...defaultWebcamSettings });

const presetOptions = computed(() => [
  { label: t('webchat.webcam.presets.balanced'), value: 'balanced' },
  { label: t('webchat.webcam.presets.high'), value: 'high' },
  { label: t('webchat.webcam.presets.low'), value: 'low' },
  { label: t('webchat.webcam.presets.custom'), value: 'custom' },
]);

const cameraOptions = computed(() => [
  { label: t('webchat.webcam.defaultDevice'), value: '' },
  ...availableCameras.value,
]);

const aspectRatioOptions = [
  { label: '16:9 (Widescreen)', value: '16:9' },
  { label: '4:3 (Standard)', value: '4:3' },
  { label: '1:1 (Quadrato)', value: '1:1' },
  { label: '9:16 (Verticale)', value: '9:16' },
];

const frameRateOptions = [
  { label: '15 FPS', value: 15 },
  { label: '20 FPS', value: 20 },
  { label: '24 FPS (Cinematic)', value: 24 },
  { label: '30 FPS (Standard)', value: 30 },
  { label: '60 FPS (Ultra Fluid)', value: 60 },
];

const codecOptions = [
  { label: 'VP8 (Massima compatibilità)', value: 'vp8' },
  { label: 'H.264 (Accelerazione Hardware)', value: 'h264' },
  { label: 'VP9 (Alta Efficienza)', value: 'vp9' },
  { label: 'AV1 (Compressione Avanzata)', value: 'av1' },
];

const degradationOptions = computed(() => [
  { label: t('webchat.webcam.degradationOptions.maintainFramerate'), value: 'maintain-framerate' },
  { label: t('webchat.webcam.degradationOptions.maintainResolution'), value: 'maintain-resolution' },
  { label: t('webchat.webcam.degradationOptions.balanced'), value: 'balanced' },
]);

function applyPresetValues(preset: string) {
  if (preset === 'balanced') {
    form.width = 1280;
    form.aspectRatio = '16:9';
    form.height = calculateHeight(1280, '16:9');
    form.frameRate = 30;
    form.bitrateKbps = 1700;
    form.codec = 'vp8';
    form.simulcast = true;
    form.degradationPreference = 'maintain-framerate';
  } else if (preset === 'high') {
    form.width = 1920;
    form.aspectRatio = '16:9';
    form.height = calculateHeight(1920, '16:9');
    form.frameRate = 30;
    form.bitrateKbps = 3000;
    form.codec = 'vp8';
    form.simulcast = true;
    form.degradationPreference = 'maintain-framerate';
  } else if (preset === 'low') {
    form.width = 640;
    form.aspectRatio = '4:3';
    form.height = calculateHeight(640, '4:3');
    form.frameRate = 20;
    form.bitrateKbps = 600;
    form.codec = 'vp8';
    form.simulcast = true;
    form.degradationPreference = 'balanced';
  }
}

function onManualChange() {
  if (form.preset !== 'custom') {
    form.preset = 'custom';
  }
}

function onWidthChange() {
  onManualChange();
  form.height = calculateHeight(form.width, form.aspectRatio);
}

function onAspectRatioChange(ratio: '16:9' | '4:3' | '1:1' | '9:16') {
  onManualChange();
  form.aspectRatio = ratio;
  form.height = calculateHeight(form.width, ratio);
}

async function enumerateCameras() {
  if (!navigator.mediaDevices?.enumerateDevices) return;
  try {
    const devices = await navigator.mediaDevices.enumerateDevices();
    availableCameras.value = devices
      .filter((d) => d.kind === 'videoinput')
      .map((d, index) => ({
        label: d.label || `${t('webchat.webcam.camera')} ${index + 1}`,
        value: d.deviceId,
      }));
  } catch (err) {
    console.error('Failed to enumerate media devices', err);
  }
}

function onSave() {
  form.height = calculateHeight(form.width, form.aspectRatio);
  emit('apply', { ...form });
  emit('update:modelValue', false);
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      if (props.initialSettings) {
        Object.assign(form, props.initialSettings);
        form.height = calculateHeight(form.width, form.aspectRatio);
      }
      void enumerateCameras();
    }
  },
  { immediate: true }
);

onMounted(() => {
  void enumerateCameras();
});
</script>
