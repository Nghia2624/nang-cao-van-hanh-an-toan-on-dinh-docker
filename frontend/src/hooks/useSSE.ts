import { useEffect, useRef } from 'react'

interface UseSSEOptions {
  enabled?: boolean
  onMessage?: (data: string) => void
  onError?: (err: Event) => void
  maxRetries?: number
}

export function useSSE(url: string | null, options: UseSSEOptions = {}) {
  const { enabled = true, onMessage, onError, maxRetries = 3 } = options
  const eventSourceRef = useRef<EventSource | null>(null)
  const retriesRef = useRef(0)
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const onMessageRef = useRef(onMessage)
  const onErrorRef = useRef(onError)

  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  useEffect(() => {
    onErrorRef.current = onError
  }, [onError])

  useEffect(() => {
    if (!enabled || !url) return

    function connect() {
      const es = new EventSource(url!)
      eventSourceRef.current = es

      es.onopen = () => {
        retriesRef.current = 0 // Reset retries on successful connection
      }

      es.onmessage = (e) => {
        onMessageRef.current?.(e.data)
      }

      es.onerror = (err) => {
        onErrorRef.current?.(err)
        es.close()
        eventSourceRef.current = null

        // Auto-reconnect with exponential backoff
        if (retriesRef.current < maxRetries) {
          const delay = Math.min(1000 * Math.pow(2, retriesRef.current), 8000)
          retriesRef.current++
          retryTimerRef.current = setTimeout(connect, delay)
        }
      }
    }

    connect()

    return () => {
      if (retryTimerRef.current) {
        clearTimeout(retryTimerRef.current)
        retryTimerRef.current = null
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      retriesRef.current = 0
    }
  }, [url, enabled])
}
