import { getAccessToken } from '@shared/api/session'
import { API_CONFIG } from '@shared/config/api'
import type { RoomState } from '../api'

export type PlayerAction = 'play' | 'pause' | 'seek' | 'change_track'

export interface ClientMessage {
  type: 'player_action'
  room_id: string
  user_id: string
  action: PlayerAction
  payload: Record<string, unknown>
  timestamp?: number
}

export interface ServerMessage {
  type: 'sync_state' | 'error' | 'pong'
  room_id?: string
  state?: RoomState
  error?: string
  timestamp?: number
}

type MessageHandler = (message: ServerMessage) => void

const inferWsGatewayUrl = () => {
  if (API_CONFIG.wsGateway) {
    return API_CONFIG.wsGateway
  }
  const httpUrl = API_CONFIG.gateway
  if (httpUrl.startsWith('https://')) {
    return httpUrl.replace('https://', 'wss://')
  }
  if (httpUrl.startsWith('http://')) {
    return httpUrl.replace('http://', 'ws://').replace(':8080', ':8001')
  }
  return `ws://${httpUrl}`
}

export class SessionWebSocket {
  private ws: WebSocket | null = null
  private roomId: string | null = null
  private userId: string | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map()
  private connectionHandlers: Set<(connected: boolean) => void> = new Set()
  private wsGatewayUrl: string

  constructor(wsGatewayUrl?: string) {
    this.wsGatewayUrl = wsGatewayUrl ?? inferWsGatewayUrl()
  }

  connect(roomId: string, userId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN && this.roomId === roomId) {
        resolve()
        return
      }

      this.roomId = roomId
      this.userId = userId

      const token = getAccessToken()
      if (!token) {
        reject(new Error('No access token available'))
        return
      }

      const url = `${this.wsGatewayUrl}/ws?token=${encodeURIComponent(token)}&room_id=${encodeURIComponent(roomId)}`

      try {
        this.ws = new WebSocket(url)

        this.ws.onopen = () => {
          this.reconnectAttempts = 0
          this.notifyConnectionHandlers(true)
          resolve()
        }

        this.ws.onmessage = event => {
          try {
            const message = JSON.parse(event.data) as ServerMessage
            console.log('[WebSocket] Received message:', message)
            this.handleMessage(message)
          } catch (error) {
            console.error('[WebSocket] Failed to parse WebSocket message:', error, event.data)
          }
        }

        this.ws.onerror = error => {
          console.error('WebSocket error:', error)
          reject(error)
        }

        this.ws.onclose = () => {
          this.notifyConnectionHandlers(false)
          this.attemptReconnect()
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts || !this.roomId || !this.userId) {
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      if (this.roomId && this.userId) {
        this.connect(this.roomId, this.userId).catch(() => {
          // Reconnection failed, will try again
        })
      }
    }, delay)
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.roomId = null
    this.userId = null
    this.reconnectAttempts = 0
    this.notifyConnectionHandlers(false)
  }

  sendAction(action: PlayerAction, payload: Record<string, unknown> = {}) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN || !this.roomId || !this.userId) {
      console.warn('[WebSocket] Cannot send action: WebSocket not connected', {
        ws: !!this.ws,
        readyState: this.ws?.readyState,
        roomId: this.roomId,
        userId: this.userId,
      })
      return
    }

    const message: ClientMessage = {
      type: 'player_action',
      room_id: this.roomId,
      user_id: this.userId,
      action,
      payload,
      timestamp: Date.now() / 1000,
    }

    console.log('[WebSocket] Sending player action:', message)
    this.ws.send(JSON.stringify(message))
    console.log('[WebSocket] Player action sent successfully')
  }

  onMessage(type: string, handler: (message: ServerMessage) => void) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set())
    }
    this.messageHandlers.get(type)!.add(handler)

    return () => {
      this.messageHandlers.get(type)?.delete(handler)
    }
  }

  onConnectionChange(handler: (connected: boolean) => void) {
    this.connectionHandlers.add(handler)
    return () => {
      this.connectionHandlers.delete(handler)
    }
  }

  private handleMessage(message: ServerMessage) {
    const handlers = this.messageHandlers.get(message.type)
    if (handlers) {
      handlers.forEach(handler => handler(message))
    }

    // Also notify 'all' handlers
    const allHandlers = this.messageHandlers.get('all')
    if (allHandlers) {
      allHandlers.forEach(handler => handler(message))
    }
  }

  private notifyConnectionHandlers(connected: boolean) {
    this.connectionHandlers.forEach(handler => handler(connected))
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  get currentRoomId(): string | null {
    return this.roomId
  }
}
