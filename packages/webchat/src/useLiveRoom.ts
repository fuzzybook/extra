import { computed, getCurrentInstance, onBeforeUnmount, ref, shallowRef } from 'vue';
import { Room, RoomEvent, Track } from 'livekit-client';
import type { JoinToken } from './types';

/** Traccia video o audio del proprietario della stanza. */
export interface LiveMedia {
  sid: string;
  track: Track;
  local: boolean;
  screen: boolean;
}

export interface LiveRoomOptions {
  /**
   * Chiamata all'inizio di leave(), anche quando la stanza si chiude da sola,
   * con lo stato dei dispositivi del proprietario prima dell'uscita.
   */
  onLeave?: ((state: { isOwner: boolean; cameraOn: boolean; micOn: boolean }) => void) | undefined;
}

export function errorMessage(reason: unknown): string {
  return reason instanceof Error ? reason.message : String(reason);
}

/**
 * Collegamento a una stanza LiveKit: token, eventi, riconnessione, tracce del
 * proprietario e sblocco dell'audio. Gli spettatori non accedono mai a
 * camera e microfono; la pubblicazione resta a carico del proprietario.
 */
export function useLiveRoom(options: LiveRoomOptions = {}) {
  const room = shallowRef<Room>();
  const connected = ref(false);
  const connecting = ref(false);
  const reconnecting = ref(false);
  const ended = ref(false);
  const error = ref('');
  const role = ref('');
  const identity = ref('');
  const ownerIdentity = ref('');
  const isOwner = computed(() => role.value === 'owner');
  const cameraOn = ref(false);
  const micOn = ref(false);
  const screenOn = ref(false);
  const needsAudio = ref(false);
  const videos = shallowRef<LiveMedia[]>([]);
  const audios = shallowRef<LiveMedia[]>([]);
  let generation = 0;

  function sync() {
    const current = room.value;
    if (!current) return;
    const video: LiveMedia[] = [];
    const audio: LiveMedia[] = [];
    const owner = isOwner.value ? current.localParticipant : current.remoteParticipants.get(ownerIdentity.value);
    owner?.trackPublications.forEach((publication) => {
      if (!publication.track || publication.isMuted) return;
      const item = {
        sid: publication.trackSid,
        track: publication.track,
        local: isOwner.value,
        screen: publication.source === Track.Source.ScreenShare,
      };
      if (publication.kind === Track.Kind.Video) video.push(item);
      if (publication.kind === Track.Kind.Audio && !isOwner.value) audio.push(item);
    });
    videos.value = video.sort((a, b) => Number(b.screen) - Number(a.screen));
    audios.value = audio;
    cameraOn.value = current.localParticipant.isCameraEnabled;
    micOn.value = current.localParticipant.isMicrophoneEnabled;
    screenOn.value = current.localParticipant.isScreenShareEnabled;
    needsAudio.value = !current.canPlaybackAudio;
  }

  async function leave() {
    options.onLeave?.({ isOwner: isOwner.value, cameraOn: cameraOn.value, micOn: micOn.value });
    ++generation;
    const current = room.value;
    room.value = undefined;
    current?.removeAllListeners();
    connected.value = false;
    connecting.value = false;
    reconnecting.value = false;
    role.value = '';
    cameraOn.value = false;
    micOn.value = false;
    screenOn.value = false;
    videos.value = [];
    audios.value = [];
    needsAudio.value = false;
    await current?.disconnect();
  }

  /**
   * Chiede un token e si collega. Restituisce false se la connessione non è
   * riuscita o se nel frattempo è stata chiamata leave().
   */
  async function join(getToken: () => Promise<JoinToken>): Promise<boolean> {
    if (connecting.value || connected.value) return false;
    const version = ++generation;
    connecting.value = true;
    error.value = '';
    ended.value = false;
    let current: Room | undefined;
    try {
      const token = await getToken();
      if (version !== generation) return false;
      role.value = token.role;
      identity.value = token.identity;
      ownerIdentity.value = token.owner_id;
      current = new Room({ adaptiveStream: true, dynacast: true });
      room.value = current;
      current
        .on(RoomEvent.TrackSubscribed, sync)
        .on(RoomEvent.TrackUnsubscribed, sync)
        .on(RoomEvent.LocalTrackPublished, sync)
        .on(RoomEvent.LocalTrackUnpublished, sync)
        .on(RoomEvent.TrackMuted, sync)
        .on(RoomEvent.TrackUnmuted, sync)
        .on(RoomEvent.AudioPlaybackStatusChanged, sync)
        .on(RoomEvent.ParticipantConnected, sync)
        .on(RoomEvent.ParticipantDisconnected, sync)
        .on(RoomEvent.Reconnecting, () => {
          reconnecting.value = true;
        })
        .on(RoomEvent.Reconnected, () => {
          reconnecting.value = false;
          sync();
        })
        .on(RoomEvent.Disconnected, () => {
          ended.value = true;
          void leave();
        })
        .on(RoomEvent.MediaDevicesError, (reason) => {
          error.value = errorMessage(reason);
        });
      await current.connect(token.server_url, token.participant_token);
      if (version !== generation) {
        await current.disconnect();
        return false;
      }
      connected.value = true;
      sync();
      return true;
    } catch (reason) {
      if (version === generation) {
        error.value = errorMessage(reason);
        await leave();
      } else {
        await current?.disconnect();
      }
      return false;
    } finally {
      if (version === generation) connecting.value = false;
    }
  }

  /** Sblocca l'audio dopo un gesto dell'utente (regole di autoplay del browser). */
  async function startAudio() {
    try {
      await room.value?.startAudio();
      sync();
    } catch (reason) {
      error.value = errorMessage(reason);
    }
  }

  if (getCurrentInstance()) {
    onBeforeUnmount(() => {
      void leave();
    });
  }

  return {
    room,
    connected,
    connecting,
    reconnecting,
    ended,
    error,
    role,
    identity,
    ownerIdentity,
    isOwner,
    cameraOn,
    micOn,
    screenOn,
    needsAudio,
    videos,
    audios,
    sync,
    join,
    leave,
    startAudio,
  };
}

export type LiveRoom = ReturnType<typeof useLiveRoom>;
