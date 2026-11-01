import { onBeforeUnmount, onMounted, ref } from 'vue';
import { api } from '@/encore/api';

// Stato delle dirette condiviso fra tutte le foto della pagina: una sola
// richiesta (ListLive) ogni 30 secondi finché almeno una foto è visibile.
// Solo lato client: durante il rendering SSG lo stato resta sconosciuto.
const enabledIds = ref<Set<string> | null>(null);
const openIds = ref<Set<string>>(new Set());
const videoIds = ref<Set<string>>(new Set());
const audioIds = ref<Set<string>>(new Set());
export interface LiveStatus {
  open: boolean;
  video: boolean;
  audio: boolean;
}

let users = 0;
let timer: ReturnType<typeof setInterval> | undefined;
let pending: Promise<void> | null = null;

function refresh(): Promise<void> {
  pending ??= api.videochat
    .ListLive()
    .then((list) => {
      enabledIds.value = new Set(list.enabled);
      openIds.value = new Set(list.open);
      videoIds.value = new Set(list.video);
      audioIds.value = new Set(list.audio);
    })
    .catch(() => {
      // Stato precedente mantenuto: nuovo tentativo al prossimo giro.
    })
    .finally(() => {
      pending = null;
    });
  return pending;
}

export function useLiveStatus() {
  // Stato aggiornato a ogni caricamento o cambio di pagina: le foto montate
  // insieme condividono la stessa richiesta. Poi polling ogni 30 secondi.
  onMounted(() => {
    void refresh();
    if (users++ === 0) timer = setInterval(() => void refresh(), 30000);
  });
  onBeforeUnmount(() => {
    if (--users === 0) clearInterval(timer);
  });

  /**
   * Stanza aperta, camera e microfono accesi per i broadcaster con il
   * broadcasting abilitato nel personal setup; null per gli altri e finché non
   * è arrivata la prima risposta (nessuno stato da mostrare).
   */
  function liveStatus(shortId: string | undefined): LiveStatus | null {
    const id = shortId?.toLowerCase();
    if (!id || !enabledIds.value?.has(id)) return null;
    return { open: openIds.value.has(id), video: videoIds.value.has(id), audio: audioIds.value.has(id) };
  }

  return { liveStatus, refresh };
}
