import { defineStore } from 'pinia';
import { ref } from 'vue';
import { images as ImagesNS } from '@/encore/client';
import { api } from '@/encore/api';

export type ImageItem = ImagesNS.Image;
export type SaveImageParams = ImagesNS.SaveImageParams;
export type ListImagesResponse = ImagesNS.ListImagesResponse;

export const useImagesStore = defineStore('images', () => {
  const images = ref<ImageItem[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);


  const client = api;

  async function withLoading<T>(fn: () => Promise<T>): Promise<T> {
    loading.value = true;
    error.value = null;
    try {
      return await fn();
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function fetchImages(): Promise<ImageItem[]> {
    return withLoading(async () => {
      const res = await client.images.ListImages();
      images.value = res.images ?? [];
      return images.value;
    });
  }

  async function saveImage(params: SaveImageParams): Promise<ImageItem> {
    return withLoading(async () => {
      const saved = await client.images.SaveImage(params);
      const index = images.value.findIndex((img) => img.uuid === saved.uuid);
      if (index >= 0) {
        images.value[index] = saved;
      } else {
        images.value.unshift(saved);
      }
      return saved;
    });
  }

  return {
    images,
    loading,
    error,
    fetchImages,
    saveImage,
  };
});
