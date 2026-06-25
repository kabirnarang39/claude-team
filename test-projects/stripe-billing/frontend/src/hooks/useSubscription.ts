import { useState, useEffect, useRef, useCallback } from 'react'
import { getSubscription, Subscription } from '../api/billing'

export function useSubscription(): {
  subscription: Subscription | null
  loading: boolean
  error: Error | null
  refetch: () => void
  startPolling: () => void
} {
  const [subscription, setSubscription] = useState<Subscription | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const pollingInterval = useRef<ReturnType<typeof setInterval> | null>(null)
  const pollingTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)

  const stopPolling = useCallback(() => {
    if (pollingInterval.current !== null) {
      clearInterval(pollingInterval.current)
      pollingInterval.current = null
    }
    if (pollingTimeout.current !== null) {
      clearTimeout(pollingTimeout.current)
      pollingTimeout.current = null
    }
  }, [])

  const fetchSubscription = useCallback(async () => {
    try {
      const data = await getSubscription()
      setSubscription(data)
      setError(null)
      // Stop polling once subscription becomes active
      if (data.status === 'active' || data.status === 'trialing') {
        stopPolling()
      }
      return data
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)))
      return null
    }
  }, [stopPolling])

  const refetch = useCallback(() => {
    setLoading(true)
    fetchSubscription().finally(() => setLoading(false))
  }, [fetchSubscription])

  const startPolling = useCallback(() => {
    // Clear any existing polling
    stopPolling()

    // Poll every 2 seconds
    pollingInterval.current = setInterval(() => {
      fetchSubscription()
    }, 2000)

    // Stop polling after 30 seconds regardless
    pollingTimeout.current = setTimeout(() => {
      stopPolling()
    }, 30_000)
  }, [fetchSubscription, stopPolling])

  useEffect(() => {
    setLoading(true)
    fetchSubscription().finally(() => setLoading(false))

    return () => {
      stopPolling()
    }
  }, [fetchSubscription, stopPolling])

  return { subscription, loading, error, refetch, startPolling }
}
