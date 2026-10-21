<template>
  <ImmageDialog
    :model-value="modelValue"
    :initial-image="initialImage"
    :initial-category="initialCategory"
    :target-width="targetWidth"
    :target-height="targetHeight"
    :proportion="proportion"
    @update:model-value="(val: boolean) => emit('update:modelValue', val)"
    @saved="(img: ImageItem) => emit('saved', img)"
  />
</template>

<script setup lang="ts">
import ImmageDialog from '@/components/ImmageDialog.vue';
import type { ImageItem } from '@/stores/images-store';
import type { ImageCategory } from '@/types/images';

defineProps<{
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
</script>
