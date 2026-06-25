import { useState, useEffect, useCallback } from 'react'
import { getUsage, Usage } from '../api/billing'

export function useUsage(): {
  usage: Usage | null
  loading: boolean
  error: Error | null
  refetch: () => void
} {
  const [usage, setUsage] = useState<Usage | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchUsage = useCallback(async () => {
    try {
      const data = await getUsage()
      setUsage(data)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)))
    }
  }, [])

  const refetch = useCallback(() => {
    setLoading(true)
    fetchUsage().finally(() => setLoading(false))
  }, [fetchUsage])

  useEffect(() => {
    setLoading(true)
    fetchUsage().finally(() => setLoading(false))

    // Auto-refresh every 60 seconds
    const interval = setInterval(() => {
      fetchUsage()
    }, 60_000)

    return () => {
      clearInterval(interval)
    }
  }, [fetchUsage])

  return { usage, loading, error, refetch }
}
