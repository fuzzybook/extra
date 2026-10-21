import { defineStore } from 'pinia';
import { ref } from 'vue';
import { broadcasters as BroadcastersNS, config as ConfigNS } from '@/encore/client';
import { api } from '@/encore/api';

export type Category = ConfigNS.Category;
export type Broadcaster = BroadcastersNS.Broadcaster;
export type BroadcasterAttribute = BroadcastersNS.BroadcasterAttribute;
export type BroadcasterTimeSlot = BroadcastersNS.BroadcasterTimeSlot;
export type BroadcasterHours = BroadcastersNS.BroadcasterHours;
export type BroadcasterPhoto = BroadcastersNS.BroadcasterPhoto;
export type BroadcasterImage = BroadcastersNS.BroadcasterImage;
export type ListBroadcastersParams = BroadcastersNS.ListBroadcastersParams;
export type CreateBroadcasterParams = BroadcastersNS.CreateBroadcasterParams;
export type UpdateBroadcasterParams = BroadcastersNS.UpdateBroadcasterParams;
export type UpdateBroadcasterAttributesParams = BroadcastersNS.UpdateBroadcasterAttributesParams;
export type UpdateBroadcasterHoursParams = BroadcastersNS.UpdateBroadcasterHoursParams;
export type UpdateBroadcasterPhotosParams = BroadcastersNS.UpdateBroadcasterPhotosParams;

export const useBroadcastersStore = defineStore('broadcasters', () => {
  const broadcasters = ref<Broadcaster[]>([]);
  const categories = ref<Category[]>([]);
  const attributePresets = ref<string[]>([]);
  const availabilityOptions = ref<string[]>([]);
  const servicesPresets = ref<string[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const imageConfig = ref<BroadcasterImage>({
    width: 600,
    height: 0,
    proportion: '2:3',
  });


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

  async function fetchBroadcasters(params?: Partial<ListBroadcastersParams>): Promise<Broadcaster[]> {
    return withLoading(async () => {
      const fullParams: ListBroadcastersParams = {
        category: params?.category ?? '',
        certified: params?.certified ?? false,
        availability: params?.availability ?? '',
        search: params?.search ?? '',
      };
      const res = await client.broadcasters.ListBroadcasters(fullParams);
      broadcasters.value = res.broadcasters ?? [];
      return broadcasters.value;
    });
  }

  async function fetchImageConfig(): Promise<BroadcasterImage> {
    try {
      const config = await client.broadcasters.GetBroadcasterImageConfig();
      if (config) {
        imageConfig.value = config;
      }
      return imageConfig.value;
    } catch {
      return imageConfig.value;
    }
  }

  async function getBroadcaster(id: string): Promise<Broadcaster> {
    return withLoading(async () => {
      const b = await client.broadcasters.GetBroadcaster(id);
      const idx = broadcasters.value.findIndex((item) => item.id === id);
      if (idx >= 0) {
        broadcasters.value[idx] = b;
      }
      return b;
    });
  }

  async function createBroadcaster(params: CreateBroadcasterParams): Promise<Broadcaster> {
    return withLoading(async () => {
      const created = await client.broadcasters.CreateBroadcaster(params);
      broadcasters.value.unshift(created);
      return created;
    });
  }

  async function updateBroadcaster(id: string, params: UpdateBroadcasterParams): Promise<Broadcaster> {
    return withLoading(async () => {
      const updated = await client.broadcasters.UpdateBroadcaster(id, params);
      const idx = broadcasters.value.findIndex((b) => b.id === id);
      if (idx >= 0) {
        broadcasters.value[idx] = updated;
      }
      return updated;
    });
  }

  async function updateAttributes(id: string, attributes: BroadcasterAttribute[]): Promise<Broadcaster> {
    return withLoading(async () => {
      const updated = await client.broadcasters.UpdateBroadcasterAttributes(id, { attributes });
      const idx = broadcasters.value.findIndex((b) => b.id === id);
      if (idx >= 0) {
        broadcasters.value[idx] = updated;
      }
      return updated;
    });
  }

  async function updateHours(id: string, hours: BroadcasterHours[]): Promise<Broadcaster> {
    return withLoading(async () => {
      const updated = await client.broadcasters.UpdateBroadcasterHours(id, { hours });
      const idx = broadcasters.value.findIndex((b) => b.id === id);
      if (idx >= 0) {
        broadcasters.value[idx] = updated;
      }
      return updated;
    });
  }

  async function updatePhotos(id: string, photos: BroadcasterPhoto[]): Promise<Broadcaster> {
    return withLoading(async () => {
      const updated = await client.broadcasters.UpdateBroadcasterPhotos(id, { photos });
      const idx = broadcasters.value.findIndex((b) => b.id === id);
      if (idx >= 0) {
        broadcasters.value[idx] = updated;
      }
      return updated;
    });
  }

  async function fetchCategories(): Promise<Category[]> {
    try {
      const res = await client.broadcasters.GetCategories();
      if (res?.categories) {
        categories.value = res.categories;
      }
      return categories.value;
    } catch {
      return categories.value;
    }
  }

  async function fetchAttributePresets(): Promise<string[]> {
    try {
      const res = await client.broadcasters.GetBroadcasterAttributePresets();
      if (res?.presets) {
        attributePresets.value = res.presets;
      }
      return attributePresets.value;
    } catch {
      return attributePresets.value;
    }
  }

  async function fetchAvailabilityOptions(): Promise<string[]> {
    try {
      const res = await client.broadcasters.GetAvailabilityOptions();
      if (res?.options) {
        availabilityOptions.value = res.options;
      }
      return availabilityOptions.value;
    } catch {
      return availabilityOptions.value;
    }
  }

  async function fetchServicesPresets(): Promise<string[]> {
    try {
      const res = await client.broadcasters.GetBroadcasterServicesPresets();
      if (res?.services) {
        servicesPresets.value = res.services;
      }
      return servicesPresets.value;
    } catch {
      return servicesPresets.value;
    }
  }

  async function deleteBroadcaster(id: string): Promise<void> {
    return withLoading(async () => {
      await client.broadcasters.DeleteBroadcaster(id);
      broadcasters.value = broadcasters.value.filter((b) => b.id !== id);
    });
  }

  return {
    broadcasters,
    categories,
    attributePresets,
    availabilityOptions,
    servicesPresets,
    loading,
    error,
    imageConfig,
    fetchBroadcasters,
    fetchImageConfig,
    fetchCategories,
    fetchAttributePresets,
    fetchAvailabilityOptions,
    fetchServicesPresets,
    getBroadcaster,
    createBroadcaster,
    updateBroadcaster,
    updateAttributes,
    updateHours,
    updatePhotos,
    deleteBroadcaster,
  };
});
