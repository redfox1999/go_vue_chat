import { ref, onUnmounted, watch } from 'vue'

export type ConnectionStatus = 'connecting' | 'online' | 'disconnected'

export interface WebSocketMessage {
  action: string
  payload: unknown
}

export function useWebSocket(url: string) {
  const status = ref<ConnectionStatus>('disconnected')
  const lastUpdateTime = ref(Date.now())
  const reconnectAttempts = ref(0)
  const reconnectInterval = ref(1000)
  const maxReconnectInterval = 60000
  
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let pingTimer: ReturnType<typeof setInterval> | null = null
  let pongTimeoutTimer: ReturnType<typeof setTimeout> | null = null
  let stateCheckTimer: ReturnType<typeof setInterval> | null = null
  let messageCallback: ((event: MessageEvent) => void) | null = null

  const connect = () => {
    // 清除之前的定时器
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    
    if (status.value === 'connecting') return
    
    status.value = 'connecting'
    reconnectAttempts.value = 0
    
    console.log(`WebSocket connecting to: ${url}`)
    
    try {
      ws = new WebSocket(url)
      
      ws.onopen = () => {
        console.log('WebSocket opened successfully')
        status.value = 'online'
        reconnectInterval.value = 1000
        reconnectAttempts.value = 0
        startPing()
        startStateCheck()
      }
      
      ws.onclose = (event) => {
        console.log(`WebSocket closed: code=${event.code}, reason="${event.reason}", wasClean=${event.wasClean}`)
        status.value = 'disconnected'
        stopAllTimers()
        
        if (event.code !== 1000) {
          scheduleReconnect()
        }
      }
      
      ws.onerror = (event) => {
        console.error('WebSocket error:', event)
        if (status.value !== 'disconnected') {
          status.value = 'disconnected'
          stopAllTimers()
          scheduleReconnect()
        }
      }
      
      ws.onmessage = (event) => {
        // 处理 pong 响应
        try {
          const data = JSON.parse(event.data)
          if (data.action === 'pong') {
            stopPongTimeout()
          }
        } catch {
          // 非 JSON 消息，忽略
        }
        
        if (messageCallback) {
          messageCallback(event)
        }
      }
      
    } catch (error) {
      console.error('WebSocket connection failed:', error)
      status.value = 'disconnected'
      scheduleReconnect()
    }
  }

  const scheduleReconnect = () => {
    reconnectAttempts.value++
    
    reconnectInterval.value = Math.min(
      reconnectInterval.value * 2,
      maxReconnectInterval
    )
    
    console.log(`WebSocket reconnect attempt ${reconnectAttempts.value} in ${reconnectInterval.value}ms`)
    
    reconnectTimer = setTimeout(() => {
      connect()
    }, reconnectInterval.value)
  }

  const startPing = () => {
    stopPing()
    
    pingTimer = setInterval(() => {
      if (!ws) return
      
      if (ws.readyState !== WebSocket.OPEN) {
        console.warn(`WebSocket ping skipped - readyState is ${ws.readyState}`)
        return
      }
      
      try {
        ws.send(JSON.stringify({ action: 'ping' }))
        startPongTimeout()
      } catch (error) {
        console.error('WebSocket ping failed:', error)
        handleDisconnect()
      }
    }, 30000)
  }

  const stopPing = () => {
    if (pingTimer) {
      clearInterval(pingTimer)
      pingTimer = null
    }
  }

  const startPongTimeout = () => {
    stopPongTimeout()
    
    pongTimeoutTimer = setTimeout(() => {
      console.warn('WebSocket pong timeout')
      handleDisconnect()
    }, 5000)
  }

  const stopPongTimeout = () => {
    if (pongTimeoutTimer) {
      clearTimeout(pongTimeoutTimer)
      pongTimeoutTimer = null
    }
  }

  const startStateCheck = () => {
    stopStateCheck()
    
    stateCheckTimer = setInterval(() => {
      if (!ws) return
      
      // 检查 readyState 是否与我们的状态一致
      if (ws.readyState === WebSocket.CLOSED && status.value !== 'disconnected') {
        console.warn(`WebSocket readyState is CLOSED but status is ${status.value}, forcing disconnect`)
        handleDisconnect()
      } else if (ws.readyState === WebSocket.CONNECTING && status.value === 'online') {
        console.warn('WebSocket unexpectedly connecting again')
        status.value = 'connecting'
      }
    }, 200)
  }

  const stopStateCheck = () => {
    if (stateCheckTimer) {
      clearInterval(stateCheckTimer)
      stateCheckTimer = null
    }
  }

  const stopAllTimers = () => {
    stopPing()
    stopPongTimeout()
    stopStateCheck()
  }

  const handleDisconnect = () => {
    if (status.value === 'disconnected') return
    
    console.log('Handling WebSocket disconnect')
    status.value = 'disconnected'
    stopAllTimers()
    
    if (ws) {
      try {
        ws.close(1011, 'Connection lost')
      } catch (e) {
        console.error('Error closing WebSocket:', e)
      }
      ws = null
    }
    
    scheduleReconnect()
  }

  const send = (message: WebSocketMessage) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      return false
    }
    
    try {
      ws.send(JSON.stringify(message))
      return true
    } catch (error) {
      console.error('WebSocket send failed:', error)
      handleDisconnect()
      return false
    }
  }

  const close = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    stopAllTimers()
    
    if (ws) {
      try {
        ws.close(1000, 'Manual close')
      } catch (e) {
        console.error('Error closing WebSocket:', e)
      }
      ws = null
    }
    
    status.value = 'disconnected'
  }

  const onMessage = (callback: (event: MessageEvent) => void) => {
    messageCallback = callback
  }

  onUnmounted(() => {
    close()
  })

  watch(status, (newStatus, oldStatus) => {
    console.log(`WebSocket status: ${oldStatus} -> ${newStatus}`)
  })

  return {
    status,
    connect,
    close,
    send,
    onMessage,
    reconnectAttempts
  }
}
