import { getAccessToken } from '@shared/api/session'
import { API_CONFIG } from '@shared/config/api'

type MessagingServerMessage =
  | { type: 'message'; conversation_id?: string; message?: unknown; error?: string }
  | { type: 'conversation_read'; conversation_id?: string; message?: unknown; error?: string }
  | { type: 'error'; error: string }
  | { type: 'pong'; timestamp?: number }

type MessageHandler = (message: MessagingServerMessage) => void

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

export class MessagingWebSocket {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map()
  private wsGatewayUrl: string

  constructor(wsGatewayUrl?: string) {
    this.wsGatewayUrl = wsGatewayUrl ?? inferWsGatewayUrl()
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      const token = getAccessToken()
      if (!token) {
        reject(new Error('No access token available'))
        return
      }

      // Используем room_id=messaging, чтобы ws-gateway принял соединение
      const url = `${this.wsGatewayUrl}/ws?token=${encodeURIComponent(token)}&room_id=messaging`

      try {
        this.ws = new WebSocket(url)

        this.ws.onopen = () => {
          this.reconnectAttempts = 0
          resolve()
        }

        this.ws.onmessage = event => {
          try {
            const message = JSON.parse(event.data) as MessagingServerMessage
            this.handleMessage(message)
          } catch (error) {
            console.error('[Messaging WS] Failed to parse message:', error, event.data)
          }
        }

        this.ws.onerror = error => {
          console.error('[Messaging WS] WebSocket error:', error)
          reject(error)
        }

        this.ws.onclose = () => {
          this.attemptReconnect()
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      this.connect().catch(() => {
        // silent retry
      })
    }, delay)
  }

  onMessage(type: MessagingServerMessage['type'] | 'all', handler: MessageHandler) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set())
    }
    this.messageHandlers.get(type)!.add(handler)

    return () => {
      this.messageHandlers.get(type)?.delete(handler)
    }
  }

  private handleMessage(message: MessagingServerMessage) {
    const handlers = this.messageHandlers.get(message.type)
    if (handlers) {
      handlers.forEach(h => h(message))
    }
    const allHandlers = this.messageHandlers.get('all')
    if (allHandlers) {
      allHandlers.forEach(h => h(message))
    }
  }
}
