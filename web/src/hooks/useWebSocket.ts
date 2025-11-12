import { useEffect, useRef, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'

export interface WebSocketMessage {
  type: 'position_update' | 'balance_update' | 'order_update' | 'market_update' | 'pnl_update' | 'trade_executed' | 'error'
  trader_id?: string
  data: any
  timestamp: string
}

interface UseWebSocketOptions {
  traderId?: string
  onMessage?: (message: WebSocketMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}

export function useWebSocket({ traderId, onMessage, onConnect, onDisconnect, onError }: UseWebSocketOptions = {}) {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const [connectionError, setConnectionError] = useState<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const { token } = useAuthStore()

  useEffect(() => {
    if (!token) return

    const connect = () => {
      try {
        // Determine WebSocket URL
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host
        const endpoint = traderId ? `/api/ws/traders/${traderId}` : '/api/ws/all'
        const wsUrl = `${protocol}//${host}${endpoint}?token=${token}`

        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          console.log('WebSocket connected')
          setIsConnected(true)
          setConnectionError(null)
          onConnect?.()
        }

        ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            setLastMessage(message)
            onMessage?.(message)
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error)
          }
        }

        ws.onerror = (event) => {
          console.error('WebSocket error:', event)
          setConnectionError('WebSocket connection error')
          onError?.(event)
        }

        ws.onclose = () => {
          console.log('WebSocket disconnected')
          setIsConnected(false)
          onDisconnect?.()

          // Attempt to reconnect after 3 seconds
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log('Attempting to reconnect WebSocket...')
            connect()
          }, 3000)
        }
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error)
        setConnectionError('Failed to connect')
      }
    }

    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [token, traderId, onMessage, onConnect, onDisconnect, onError])

  const sendMessage = (message: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket is not connected')
    }
  }

  return {
    isConnected,
    lastMessage,
    connectionError,
    sendMessage,
  }
}
