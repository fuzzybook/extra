import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { Local } from '@/encore/client'

export type ChatMessage = {
  id: string
  senderId: string
  senderName: string
  senderAvatar?: string | undefined
  text: string
  timestamp: number
  isOwner?: boolean | undefined
  isSystem?: boolean | undefined
  type?: string | undefined
  targetId?: string | undefined
}

export type ChatFilters = {
  type?: string | undefined // es. 'ALL', 'MESSAGE', 'ALLERT', 'ACTION', 'CMD'
  senderId?: string | undefined
  search?: string | undefined
}

// In ambiente browser usiamo Local (http://localhost:4000) oppure fallback relativo
const API_BASE = Local || 'http://localhost:4000'

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])
  const inputText = ref<string>('')
  const currentRoomId = ref<string>('')
  const loading = ref<boolean>(false)

  // Predisposizione filtri per la chat
  const filters = ref<ChatFilters>({
    type: 'ALL',
    senderId: '',
    search: '',
  })

  const filteredMessages = computed(() => {
    let list = messages.value
    // Nella chat devono essere visualizzati solo i messaggi di tipo MESSAGE
    list = list.filter((m) => m.type === 'MESSAGE')

    if (filters.value.senderId) {
      list = list.filter((m) => m.senderId === filters.value.senderId)
    }
    if (filters.value.search?.trim()) {
      const q = filters.value.search.trim().toLowerCase()
      list = list.filter(
        (m) => m.text.toLowerCase().includes(q) || m.senderName.toLowerCase().includes(q)
      )
    }
    return list
  })

  function setInputText(text: string) {
    inputText.value = text
  }

  function addMessage(msg: ChatMessage) {
    if (!messages.value.some((m) => m.id === msg.id)) {
      messages.value.push(msg)
    }
  }

  function clearMessages() {
    messages.value = []
  }

  function setFilters(newFilters: Partial<ChatFilters>) {
    filters.value = { ...filters.value, ...newFilters }
  }

  function clearFilters() {
    filters.value = {
      type: 'ALL',
      senderId: '',
      search: '',
    }
  }

  /**
   * Salva asincronamente il messaggio nella tabella chatlog del backend
   */
  async function saveMessage(roomId: string, msg: ChatMessage): Promise<void> {
    addMessage(msg)
    const targetRoom = roomId?.trim() || currentRoomId.value?.trim()
    if (!targetRoom) {
      console.warn('saveMessage: roomId non specificato, impossibile persistere su chatlog')
      return
    }

    try {
      const url = `${API_BASE}/broadcasters/id/${encodeURIComponent(targetRoom)}/chatlog`
      const resp = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: msg.type === 'ADMIN_ALERT' ? 'MESSAGE' : (msg.type || (msg.isSystem ? 'ALLERT' : 'MESSAGE')),
          data: JSON.stringify(msg),
        }),
      })
      if (!resp.ok) {
        console.warn('Errore HTTP salvataggio chatlog:', resp.status, await resp.text())
      }
    } catch (err) {
      console.warn('Errore salvataggio asincrono chatlog:', err)
    }
  }

  /**
   * Registra un evento multimediale (VIDEO_ON, VIDEO_OFF, MIC_ON, MIC_OFF) nel chatlog.
   * Assicura che nel payload il tipo sia esattamente "ADMIN_ALERT", mentre nel DB il record è salvato come "MESSAGE".
   */
  async function logMediaEvent(
    roomId: string,
    eventText: 'VIDEO_ON' | 'VIDEO_OFF' | 'MIC_ON' | 'MIC_OFF',
    senderId?: string,
    senderName?: string
  ): Promise<void> {
    const targetRoom = roomId?.trim() || currentRoomId.value?.trim()
    if (!targetRoom) return

    const msg: ChatMessage = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      senderId: senderId || 'broadcaster',
      senderName: senderName || 'Broadcaster',
      text: eventText,
      timestamp: Date.now(),
      isOwner: true,
      type: 'ADMIN_ALERT',
    }

    await saveMessage(targetRoom, msg)
  }

  /**
   * Inizializza la chat recuperando gli ultimi 30 messaggi cronologici dal DB.
   * I messaggi vengono visualizzati con i più recenti in basso (append ordinato).
   */
  async function initChat(roomId: string, limit = 30): Promise<void> {
    const targetRoom = roomId?.trim()
    if (!targetRoom) return

    if (currentRoomId.value !== targetRoom) {
      currentRoomId.value = targetRoom
      clearMessages()
    } else if (messages.value.length > 0) {
      return
    }

    loading.value = true
    try {
      const url = `${API_BASE}/broadcasters/id/${encodeURIComponent(targetRoom)}/chatlog?limit=${limit}&type=MESSAGE`
      const resp = await fetch(url)
      if (!resp.ok) {
        console.warn('Errore HTTP caricamento chatlog:', resp.status, await resp.text())
        return
      }

      const data = await resp.json()
      const entries: any[] = Array.isArray(data) ? data : (data.messages || [])

      for (const entry of entries) {
        let parsed: Partial<ChatMessage> | null = null
        try {
          parsed = JSON.parse(entry.data)
        } catch {
          parsed = null
        }

        const msg: ChatMessage = {
          id: parsed?.id || `db-${entry.id}`,
          senderId: parsed?.senderId || (entry.type === 'MESSAGE' ? 'user' : 'system'),
          senderName: parsed?.senderName || (entry.type === 'MESSAGE' ? 'User' : 'System'),
          senderAvatar: parsed?.senderAvatar,
          text: parsed?.text || entry.data,
          timestamp: parsed?.timestamp || new Date(entry.timestamp).getTime(),
          isOwner: parsed?.isOwner || false,
          isSystem: parsed?.isSystem ?? (entry.type !== 'MESSAGE'),
          type: parsed?.type || entry.type,
          targetId: parsed?.targetId,
        }

        addMessage(msg)
      }
    } catch (err) {
      console.warn('Errore caricamento cronologia chatlog:', err)
    } finally {
      loading.value = false
    }
  }

  return {
    messages,
    filteredMessages,
    inputText,
    filters,
    loading,
    setInputText,
    addMessage,
    clearMessages,
    setFilters,
    clearFilters,
    saveMessage,
    logMediaEvent,
    initChat,
  }
})
