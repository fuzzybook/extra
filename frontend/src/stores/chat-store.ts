import { defineStore } from 'pinia'
import { API_BASE_URL } from '@/encore/base-url'

export type MediaEvent = 'VIDEO_ON' | 'VIDEO_OFF' | 'MIC_ON' | 'MIC_OFF'

/**
 * Registro degli eventi della stanza (chatlog del backend). I messaggi di chat
 * passano invece da videochat (vedi BroadcastStageRight.vue).
 */
export const useChatStore = defineStore('chat', () => {
  /**
   * Registra un evento multimediale del broadcaster (camera e microfono) nel
   * chatlog come evento ACTION. La sessione viaggia nel cookie HttpOnly.
   */
  async function logMediaEvent(
    roomId: string,
    eventText: MediaEvent,
    senderId?: string,
    senderName?: string
  ): Promise<void> {
    const targetRoom = roomId?.trim()
    if (!targetRoom) return

    try {
      const resp = await fetch(`${API_BASE_URL}/broadcasters/id/${encodeURIComponent(targetRoom)}/chatlog`, {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'ACTION',
          data: JSON.stringify({ event: eventText, senderId, senderName, timestamp: Date.now() }),
        }),
      })
      if (!resp.ok) {
        console.warn('Errore HTTP registrazione evento:', resp.status, await resp.text())
      }
    } catch (err) {
      console.warn('Errore registrazione evento:', err)
    }
  }

  return { logMediaEvent }
})
