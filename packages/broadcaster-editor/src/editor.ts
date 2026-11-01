import { inject, reactive, ref, type App, type InjectionKey } from 'vue';
import type {
  Broadcaster,
  BroadcasterAttribute,
  BroadcasterHours,
  BroadcasterImage,
  BroadcasterPhoto,
  Category,
  CreateBroadcasterParams,
  Image,
  SaveImageParams,
  UpdateBroadcasterParams,
} from './types';

/**
 * Le chiamate al backend usate dai dialoghi: ogni app passa il proprio client
 * Encore generato (servizi broadcasters e images), che ha già questi metodi.
 */
export interface BroadcasterEditorApi {
  broadcasters: {
    GetCategories(): Promise<{ categories: Category[] }>;
    GetBroadcasterAttributePresets(): Promise<{ presets: string[] }>;
    GetAvailabilityOptions(): Promise<{ options: string[] }>;
    GetBroadcasterServicesPresets(): Promise<{ services: string[] }>;
    GetBroadcasterImageConfig(): Promise<BroadcasterImage>;
    CreateBroadcaster(params: CreateBroadcasterParams): Promise<Broadcaster>;
    UpdateBroadcaster(id: string, params: UpdateBroadcasterParams): Promise<Broadcaster>;
    UpdateBroadcasterAttributes(id: string, params: { attributes: BroadcasterAttribute[] }): Promise<Broadcaster>;
    UpdateBroadcasterHours(id: string, params: { hours: BroadcasterHours[] }): Promise<Broadcaster>;
    UpdateBroadcasterPhotos(id: string, params: { photos: BroadcasterPhoto[] }): Promise<Broadcaster>;
  };
  images: {
    ListImages(): Promise<{ images: Image[] }>;
    SaveImage(params: SaveImageParams): Promise<Image>;
  };
}

function createState(api: BroadcasterEditorApi) {
  const categories = ref<Category[]>([]);
  const attributePresets = ref<string[]>([]);
  const availabilityOptions = ref<string[]>([]);
  const servicesPresets = ref<string[]>([]);
  const imageConfig = ref<BroadcasterImage>({ width: 600, height: 0, proportion: '2:3' });
  const images = ref<Image[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

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

  // Liste di supporto: se la richiesta fallisce restano i valori precedenti.
  async function fetchCategories() {
    try {
      categories.value = (await api.broadcasters.GetCategories()).categories ?? categories.value;
    } catch {
      // valori precedenti
    }
    return categories.value;
  }
  async function fetchAttributePresets() {
    try {
      attributePresets.value = (await api.broadcasters.GetBroadcasterAttributePresets()).presets ?? attributePresets.value;
    } catch {
      // valori precedenti
    }
    return attributePresets.value;
  }
  async function fetchAvailabilityOptions() {
    try {
      availabilityOptions.value = (await api.broadcasters.GetAvailabilityOptions()).options ?? availabilityOptions.value;
    } catch {
      // valori precedenti
    }
    return availabilityOptions.value;
  }
  async function fetchServicesPresets() {
    try {
      servicesPresets.value = (await api.broadcasters.GetBroadcasterServicesPresets()).services ?? servicesPresets.value;
    } catch {
      // valori precedenti
    }
    return servicesPresets.value;
  }
  async function fetchImageConfig() {
    try {
      const config = await api.broadcasters.GetBroadcasterImageConfig();
      if (config) imageConfig.value = config;
    } catch {
      // valori precedenti
    }
    return imageConfig.value;
  }

  const createBroadcaster = (params: CreateBroadcasterParams) => withLoading(() => api.broadcasters.CreateBroadcaster(params));
  const updateBroadcaster = (id: string, params: UpdateBroadcasterParams) =>
    withLoading(() => api.broadcasters.UpdateBroadcaster(id, params));
  const updateAttributes = (id: string, attributes: BroadcasterAttribute[]) =>
    withLoading(() => api.broadcasters.UpdateBroadcasterAttributes(id, { attributes }));
  const updateHours = (id: string, hours: BroadcasterHours[]) => withLoading(() => api.broadcasters.UpdateBroadcasterHours(id, { hours }));
  const updatePhotos = (id: string, photos: BroadcasterPhoto[]) =>
    withLoading(() => api.broadcasters.UpdateBroadcasterPhotos(id, { photos }));

  async function fetchImages() {
    return withLoading(async () => {
      images.value = (await api.images.ListImages()).images ?? [];
      return images.value;
    });
  }
  async function saveImage(params: SaveImageParams) {
    return withLoading(async () => {
      const saved = await api.images.SaveImage(params);
      const index = images.value.findIndex((img) => img.uuid === saved.uuid);
      if (index >= 0) images.value[index] = saved;
      else images.value.unshift(saved);
      return saved;
    });
  }

  // reactive: nei dialoghi i valori si leggono senza .value, come da uno store.
  return reactive({
    categories,
    attributePresets,
    availabilityOptions,
    servicesPresets,
    imageConfig,
    images,
    loading,
    error,
    fetchCategories,
    fetchAttributePresets,
    fetchAvailabilityOptions,
    fetchServicesPresets,
    fetchImageConfig,
    createBroadcaster,
    updateBroadcaster,
    updateAttributes,
    updateHours,
    updatePhotos,
    fetchImages,
    saveImage,
  });
}

export type BroadcasterEditor = ReturnType<typeof createState>;

const editorKey: InjectionKey<BroadcasterEditor> = Symbol('broadcaster-editor');

/** Plugin da installare nell'app (app.use) con il client Encore dell'app. */
export function createBroadcasterEditor(api: BroadcasterEditorApi) {
  return {
    install(app: App) {
      app.provide(editorKey, createState(api));
    },
  };
}

/** Stato e azioni condivisi dai dialoghi (dati di supporto, salvataggi, immagini). */
export function useBroadcasterEditor(): BroadcasterEditor {
  const editor = inject(editorKey);
  if (!editor) throw new Error('createBroadcasterEditor non installato nell\'app');
  return editor;
}
