import { computed, ref } from 'vue';
import type { profiles } from '@/encore/client';
import { api } from '@/encore/api';

// Stato condiviso nel browser: l'header e le pagine vedono lo stesso utente.
// Viene caricato solo lato client (onMounted), mai durante il rendering SSG.
const me = ref<profiles.Profile | null>(null);
const isBroadcaster = ref(false);
const canBroadcast = ref(false);
const loaded = ref(false);
let pending: Promise<void> | null = null;

async function load() {
  try {
    me.value = (await api.profiles.Me()) || null;
  } catch {
    me.value = null;
  }
  isBroadcaster.value = false;
  canBroadcast.value = false;
  if (me.value) {
    try {
      // Abilitazione e nome della stanza vengono dal personal setup del profilo.
      const mine = await api.videochat.GetMyLive();
      isBroadcaster.value = mine.is_broadcaster;
      canBroadcast.value = mine.can_broadcast;
    } catch {
      // Senza stato della diretta il comando "Apri stanza" resta nascosto.
    }
  }
  loaded.value = true;
}

/** Ricarica l'utente collegato (dopo login o logout); le chiamate concorrenti condividono la richiesta. */
function refresh(): Promise<void> {
  pending ??= load().finally(() => {
    pending = null;
  });
  return pending;
}

async function logout() {
  try {
    await api.auth.Logout();
  } finally {
    me.value = null;
    isBroadcaster.value = false;
    canBroadcast.value = false;
  }
}

export function useSession() {
  return { me, isBroadcaster, canBroadcast, loaded, loggedIn: computed(() => !!me.value), refresh, logout };
}
