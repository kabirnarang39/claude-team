import { useState, useEffect, useCallback } from 'react'
import { getInvoices, Invoice } from '../api/billing'

export function useInvoices(initialPage = 1): {
  invoices: Invoice[]
  total: number
  page: number
  perPage: number
  loading: boolean
  error: Error | null
  setPage: (page: number) => void
} {
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPageState] = useState(initialPage)
  const [perPage, setPerPage] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchInvoices = useCallback(async (pageNum: number) => {
    setLoading(true)
    try {
      const data = await getInvoices(pageNum)
      setInvoices(data.invoices)
      setTotal(data.total)
      setPageState(data.page)
      setPerPage(data.per_page)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)))
    } finally {
      setLoading(false)
    }
  }, [])

  const setPage = useCallback(
    (newPage: number) => {
      fetchInvoices(newPage)
    },
    [fetchInvoices]
  )

  useEffect(() => {
    fetchInvoices(initialPage)
  }, [fetchInvoices, initialPage])

  return { invoices, total, page, perPage, loading, error, setPage }
}
