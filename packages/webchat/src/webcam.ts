import { Track, type Room, type TrackPublishOptions, type VideoCaptureOptions, type VideoCodec } from 'livekit-client';

/** Impostazioni avanzate della webcam del proprietario della stanza. */
export interface WebcamSettings {
  preset: 'balanced' | 'high' | 'low' | 'custom';
  deviceId: string;
  width: number;
  height: number;
  aspectRatio: '16:9' | '4:3' | '1:1' | '9:16';
  frameRate: number;
  bitrateKbps: number;
  codec: 'vp8' | 'h264' | 'vp9' | 'av1';
  simulcast: boolean;
  degradationPreference: 'maintain-framerate' | 'maintain-resolution' | 'balanced';
}

export const defaultWebcamSettings: WebcamSettings = {
  preset: 'balanced',
  deviceId: '',
  width: 1280,
  height: 720,
  aspectRatio: '16:9',
  frameRate: 30,
  bitrateKbps: 1700,
  codec: 'vp8',
  simulcast: true,
  degradationPreference: 'maintain-framerate',
};

// Stessa chiave per pannello e sito pubblico: le impostazioni restano nel browser.
const storageKey = 'videochat_webcam_settings';

export function loadWebcamSettings(): WebcamSettings {
  try {
    const saved = localStorage.getItem(storageKey);
    if (saved) return { ...defaultWebcamSettings, ...(JSON.parse(saved) as Partial<WebcamSettings>) };
  } catch {
    // localStorage non disponibile o valore non valido: valori predefiniti
  }
  return { ...defaultWebcamSettings };
}

export function saveWebcamSettings(settings: WebcamSettings) {
  try {
    localStorage.setItem(storageKey, JSON.stringify(settings));
  } catch (reason) {
    console.warn('Impossibile salvare le impostazioni webcam:', reason);
  }
}

/** Altezza ricavata dalla larghezza e dal formato, per non deformare l'immagine. */
export function calculateHeight(width: number, ratio: string): number {
  const w = Number(width) || 1280;
  if (ratio === '4:3') return Math.round((w * 3) / 4);
  if (ratio === '1:1') return w;
  if (ratio === '9:16') return Math.round((w * 16) / 9);
  return Math.round((w * 9) / 16);
}

/** Opzioni LiveKit di acquisizione e pubblicazione per setCameraEnabled. */
export function buildWebcamOptions(settings: WebcamSettings) {
  const aspectRatios: Record<string, number> = {
    '16:9': 16 / 9,
    '4:3': 4 / 3,
    '1:1': 1,
    '9:16': 9 / 16,
  };

  const captureOptions: VideoCaptureOptions = {
    ...(settings.deviceId ? { deviceId: { exact: settings.deviceId } } : {}),
    resolution: {
      width: settings.width || 1280,
      height: settings.height || 720,
      frameRate: settings.frameRate || 30,
      aspectRatio: aspectRatios[settings.aspectRatio] ?? 16 / 9,
    },
    frameRate: settings.frameRate || 30,
  };

  const publishOptions: TrackPublishOptions = {
    videoEncoding: {
      maxBitrate: (settings.bitrateKbps || 1700) * 1000,
      maxFramerate: settings.frameRate || 30,
    },
    videoCodec: (settings.codec || 'vp8') as VideoCodec,
    simulcast: settings.simulcast !== false,
    degradationPreference: (settings.degradationPreference || 'maintain-framerate') as RTCDegradationPreference,
  };

  return { captureOptions, publishOptions };
}

/** Ripubblica la camera già accesa con le nuove impostazioni. */
export async function republishCamera(room: Room, settings: WebcamSettings) {
  const publication = room.localParticipant.getTrackPublication(Track.Source.Camera);
  if (publication?.track) await room.localParticipant.unpublishTrack(publication.track, true);
  const { captureOptions, publishOptions } = buildWebcamOptions(settings);
  await room.localParticipant.setCameraEnabled(true, captureOptions, publishOptions);
}

/**
 * Spegne camera o microfono togliendo la traccia invece di metterla in muto:
 * LiveKit avvisa il backend (webhook track_unpublished) e la camera si spegne
 * davvero. Il muto non genera webhook, quindi lo stato pubblico non cambierebbe.
 */
export async function unpublishMedia(room: Room, kind: 'camera' | 'microphone') {
  const source = kind === 'camera' ? Track.Source.Camera : Track.Source.Microphone;
  const publication = room.localParticipant.getTrackPublication(source);
  if (publication?.track) await room.localParticipant.unpublishTrack(publication.track, true);
}
