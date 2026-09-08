import { useState, useCallback, useEffect } from 'react'
import { adminTasksApi } from '../../../shared/lib/api'
import type { ClaimView, PaginationMeta } from '../../../shared/types'

export interface UseAdminClaimsOptions {
  enabled?: boolean
}

export function useAdminClaims(options?: UseAdminClaimsOptions) {
  const enabled = options?.enabled ?? true
  const [claims, setClaims] = useState<ClaimView[]>([])
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, limit: 50, total: 0, has_next: false })
  const [pendingTotal, setPendingTotal] = useState<number | null>(null)
  const [filter, setFilter] = useState<'PENDING' | 'APPROVED' | 'REJECTED' | 'ALL'>('PENDING')
  const [isFetching, setIsFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [processingId, setProcessingId] = useState<number | null>(null)
  const [actionNotes, setActionNotes] = useState<Record<number, string>>({})

  const fetchClaims = useCallback(async (statusFilter = filter, page = 1) => {
    setIsFetching(true)
    setError(null)
    try {
      const statusParam = statusFilter === 'ALL' ? undefined : statusFilter
      const res: any = await adminTasksApi.getClaims({ status: statusParam, page, limit: 50 })
      const items: ClaimView[] = Array.isArray(res) ? res : res?.items || []
      const pag: PaginationMeta = Array.isArray(res)
        ? { page, limit: 50, total: items.length, has_next: false }
        : res?.pagination || { page, limit: 50, total: items.length, has_next: false }
      setClaims(items)
      setPagination(pag)
      if (Array.isArray(res)) {
        setPendingTotal(items.filter((c) => c.status === 'PENDING').length)
      } else if (statusFilter === 'PENDING') {
        setPendingTotal(pag.total)
      } else {
        try {
          const countRes: any = await adminTasksApi.getClaims({ status: 'PENDING', page: 1, limit: 1 })
          const countTotal: number | undefined = Array.isArray(countRes)
            ? countRes.filter((c: ClaimView) => c.status === 'PENDING').length
            : countRes?.pagination?.total
          setPendingTotal(typeof countTotal === 'number' ? countTotal : null)
        } catch {
          setPendingTotal(null)
        }
      }
    } catch (err: any) {
      setError(err?.message || 'Gagal memuat daftar klaim reward')
    } finally {
      setIsFetching(false)
    }
  }, [filter])

  useEffect(() => {
    if (enabled) {
      fetchClaims(filter, 1)
    }
  }, [filter, fetchClaims, enabled])

  useEffect(() => {
    if (!enabled) return
    const onClaimsChanged = () => {
      fetchClaims(filter, pagination.page)
    }
    window.addEventListener('odyssey:claims-changed', onClaimsChanged)
    return () => {
      window.removeEventListener('odyssey:claims-changed', onClaimsChanged)
    }
  }, [enabled, fetchClaims, filter, pagination.page])

  const setNote = (id: number, note: string) => {
    setActionNotes((prev) => ({ ...prev, [id]: note }))
  }

  const handleProcess = async (id: number, status: 'APPROVED' | 'REJECTED') => {
    setProcessingId(id)
    try {
      const notes = actionNotes[id]
      await adminTasksApi.processClaim(id, status, notes)
      window.dispatchEvent(new CustomEvent('odyssey:claims-changed'))
      await fetchClaims(filter, pagination.page)
    } catch (err: any) {
      alert(`Gagal memproses klaim: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  return {
    claims,
    pagination,
    pendingTotal,
    filter,
    setFilter,
    isFetching,
    error,
    processingId,
    actionNotes,
    setNote,
    handleProcess,
    fetchClaims,
  }
}
