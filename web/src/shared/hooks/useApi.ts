import { useEffect, useState } from 'react'
import { apiClient } from '../lib/api'
import type { ApiError } from '../types'

export function useApi<T>(path: string): { data: T | null; error: ApiError | null; loading: boolean } {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<ApiError | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    void apiClient
    void path
    void setData
    void setError
    void setLoading
  }, [path])

  return { data, error, loading }
}
