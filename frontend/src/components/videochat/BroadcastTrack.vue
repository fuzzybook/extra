<template>
  <div v-if="audio" ref="container" class="broadcast-audio" />
  <div v-else ref="wrapperRef" class="broadcast-video-wrapper" :style="wrapperStyle">
    <div
      ref="container"
      class="broadcast-video cursor-pointer"
      :style="containerStyle"
      @dblclick="toggleFullscreen"
    >
      <!-- Corner resize handle (drag to resize while preserving aspect ratio) -->
      <div
        class="video-resize-handle"
        :title="t('broadcast.resizeTooltip')"
        @mousedown.prevent="startDragResize"
        @touchstart.prevent="startTouchResize"
      >
        <q-icon name="south_east" size="14px" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Track } from 'livekit-client';

const props = defineProps<{ track: Track; local?: boolean; audio?: boolean }>();
const { t } = useI18n();

const container = ref<HTMLDivElement>();
const wrapperRef = ref<HTMLDivElement>();
let attached: { track: Track; element: HTMLMediaElement } | undefined;

// Format & aspect ratio detection
const detectedWidth = ref(0);
const detectedHeight = ref(0);
const aspectRatioNum = ref(16 / 9);
const aspectRatioStr = ref('16 / 9');

// Sizing preferences (clamped between 25% and 100%)
const savedSize = localStorage.getItem('broadcast_video_size');
const parsedSaved = savedSize ? Number(savedSize) : NaN;
const sizePercent = ref(!Number.isNaN(parsedSaved) && parsedSaved >= 25 && parsedSaved <= 100 ? parsedSaved : 100);

const isFullscreen = ref(false);
const isDragging = ref(false);
let startX = 0;
let startWidth = 0;
let containerParentWidth = 0;



const wrapperStyle = computed(() => {
  return {
    width: `${sizePercent.value}%`,
    maxWidth: `min(100%, calc(78vh * ${aspectRatioNum.value}))`,
    marginLeft: '0',
    marginRight: 'auto',
  };
});

const containerStyle = computed(() => {
  return {
    aspectRatio: aspectRatioStr.value,
    width: '100%',
  };
});

function updateAspectRatio() {
  if (attached?.element instanceof HTMLVideoElement) {
    const video = attached.element;
    if (video.videoWidth && video.videoHeight) {
      detectedWidth.value = video.videoWidth;
      detectedHeight.value = video.videoHeight;
      aspectRatioNum.value = video.videoWidth / video.videoHeight;
      aspectRatioStr.value = `${video.videoWidth} / ${video.videoHeight}`;
    }
  }
}



function saveSizePreference() {
  try {
    localStorage.setItem('broadcast_video_size', String(sizePercent.value));
  } catch {
    // Ignore localStorage write errors
  }
}

// Drag-to-resize handlers
function startDragResize(e: MouseEvent) {
  isDragging.value = true;
  startX = e.clientX;
  if (wrapperRef.value) {
    startWidth = wrapperRef.value.getBoundingClientRect().width;
    containerParentWidth = wrapperRef.value.parentElement?.clientWidth || window.innerWidth;
  }
  document.body.style.cursor = 'nwse-resize';
  document.body.style.userSelect = 'none';
  window.addEventListener('mousemove', onDragMove);
  window.addEventListener('mouseup', stopDrag);
}

function startTouchResize(e: TouchEvent) {
  if (!e.touches[0]) return;
  isDragging.value = true;
  startX = e.touches[0].clientX;
  if (wrapperRef.value) {
    startWidth = wrapperRef.value.getBoundingClientRect().width;
    containerParentWidth = wrapperRef.value.parentElement?.clientWidth || window.innerWidth;
  }
  window.addEventListener('touchmove', onTouchMove, { passive: false });
  window.addEventListener('touchend', stopTouch);
}

function onDragMove(e: MouseEvent) {
  if (!isDragging.value || !containerParentWidth) return;
  const deltaX = e.clientX - startX;
  const newWidth = Math.max(200, startWidth + deltaX);
  const effectiveMaxWidth = Math.min(
    containerParentWidth,
    window.innerHeight * 0.78 * aspectRatioNum.value
  );
  const pct = Math.min(100, Math.max(25, Math.round((newWidth / effectiveMaxWidth) * 100)));
  sizePercent.value = pct;
}

function onTouchMove(e: TouchEvent) {
  if (!isDragging.value || !containerParentWidth || !e.touches[0]) return;
  e.preventDefault();
  const deltaX = e.touches[0].clientX - startX;
  const newWidth = Math.max(200, startWidth + deltaX);
  const effectiveMaxWidth = Math.min(
    containerParentWidth,
    window.innerHeight * 0.78 * aspectRatioNum.value
  );
  const pct = Math.min(100, Math.max(25, Math.round((newWidth / effectiveMaxWidth) * 100)));
  sizePercent.value = pct;
}

function stopDrag() {
  if (isDragging.value) {
    isDragging.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('mousemove', onDragMove);
    window.removeEventListener('mouseup', stopDrag);
    saveSizePreference();
  }
}

function stopTouch() {
  if (isDragging.value) {
    isDragging.value = false;
    window.removeEventListener('touchmove', onTouchMove);
    window.removeEventListener('touchend', stopTouch);
    saveSizePreference();
  }
}

// Fullscreen
function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement;
}

async function toggleFullscreen() {
  if (!container.value) return;
  try {
    if (!document.fullscreenElement) {
      if (container.value.requestFullscreen) {
        await container.value.requestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        await document.exitFullscreen();
      }
    }
  } catch (err) {
    console.warn('Fullscreen request failed:', err);
  }
}

// Attach/Detach
const videoEvents = ['loadedmetadata', 'loadeddata', 'canplay', 'playing', 'resize', 'timeupdate'];

function detach() {
  if (!attached) return;
  if (attached.element instanceof HTMLVideoElement) {
    for (const ev of videoEvents) {
      attached.element.removeEventListener(ev, updateAspectRatio);
    }
  }
  attached.track.detach(attached.element);
  attached.element.remove();
  attached = undefined;
}

function attach() {
  detach();
  if (!container.value) return;
  const element = props.track.attach();
  element.autoplay = true;
  if (element instanceof HTMLVideoElement) {
    element.playsInline = true;
    for (const ev of videoEvents) {
      element.addEventListener(ev, updateAspectRatio);
    }
    if (element.videoWidth && element.videoHeight) {
      updateAspectRatio();
    }
  }
  element.muted = !!props.local;
  container.value.appendChild(element);
  attached = { track: props.track, element };
}

onMounted(() => {
  attach();
  document.addEventListener('fullscreenchange', onFullscreenChange);
});

watch(() => props.track, attach);

onBeforeUnmount(() => {
  detach();
  stopDrag();
  stopTouch();
  document.removeEventListener('fullscreenchange', onFullscreenChange);
});
</script>

<style scoped>
.broadcast-video-wrapper {
  display: block;
  transition: width 0.15s ease-out;
}

.broadcast-video {
  position: relative;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.2);
}

.broadcast-video :deep(video) {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
}

.video-resize-handle {
  position: absolute;
  right: 6px;
  bottom: 6px;
  width: 26px;
  height: 26px;
  background: rgba(0, 0, 0, 0.65);
  color: #fff;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: nwse-resize;
  user-select: none;
  opacity: 0.75;
  transition: opacity 0.2s, background 0.2s, transform 0.15s;
  z-index: 10;
}

.video-resize-handle:hover,
.video-resize-handle:active {
  opacity: 1;
  background: var(--q-primary, #1976d2);
  transform: scale(1.1);
}

:fullscreen .video-resize-handle,
:-webkit-full-screen .video-resize-handle {
  display: none !important;
}



.broadcast-audio {
  display: none;
}
</style>
