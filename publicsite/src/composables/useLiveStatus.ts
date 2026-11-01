import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import type { videochat } from '@/encore/client';
import { api } from '@/encore/api';
import { API_BASE_URL } from '@/encore/base-url';

// Stato delle dirette condiviso fra tutte le foto della pagina. Arriva in
// tempo reale da una sola connessione SSE (/videochat/live-stream) finché
// almeno una foto è visibile; se lo stream non è collegato si torna al polling
// di ListLive ogni 30 secondi. Solo lato client: durante il rendering SSG lo
// stato resta sconosciuto.
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
let loadedAt = 0;
let source: EventSource | null = null;
let closeTimer: ReturnType<typeof setTimeout> | undefined;

function apply(list: videochat.LiveList) {
  enabledIds.value = new Set(list.enabled);
  openIds.value = new Set(list.open);
  videoIds.value = new Set(list.video);
  audioIds.value = new Set(list.audio);
  loadedAt = Date.now();
}

function refresh(): Promise<void> {
  pending ??= api.videochat
    .ListLive()
    .then(apply)
    .catch(() => {
      // Stato precedente mantenuto: nuovo tentativo al prossimo giro.
    })
    .finally(() => {
      pending = null;
    });
  return pending;
}

const streamOpen = () => source?.readyState === EventSource.OPEN;

// EventSource si riconnette da solo se la connessione cade; il server manda
// lo stato completo a ogni connessione e poi a ogni cambiamento.
function openStream() {
  if (source || typeof EventSource === 'undefined') return;
  source = new EventSource(`${API_BASE_URL}/videochat/live-stream`);
  source.addEventListener('status', (event) => {
    try {
      apply(JSON.parse((event as MessageEvent<string>).data) as videochat.LiveList);
    } catch {
      // evento non valido: si aspetta il prossimo
    }
  });
}

function closeStream() {
  source?.close();
  source = null;
}

export function useLiveStatus() {
  onMounted(() => {
    clearTimeout(closeTimer);
    // Prima foto della pagina: stato subito via REST (le foto montate subito
    // dopo una risposta la riusano), poi aggiornamenti in tempo reale.
    if (!streamOpen() && Date.now() - loadedAt > 2000) void refresh();
    if (users++ === 0) {
      openStream();
      timer = setInterval(() => {
        if (!streamOpen()) void refresh();
      }, 30000);
    }
  });
  onBeforeUnmount(() => {
    if (--users > 0) return;
    clearInterval(timer);
    // Cambio pagina: le foto della nuova pagina si montano subito dopo, quindi
    // la connessione si chiude solo se nessuna foto ricompare a breve.
    closeTimer = setTimeout(closeStream, 5000);
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

  return {
    liveStatus,
    refresh,
    // true dopo la prima risposta: da lì liveStatus distingue abilitati e non.
    ready: computed(() => enabledIds.value !== null),
  };
}
