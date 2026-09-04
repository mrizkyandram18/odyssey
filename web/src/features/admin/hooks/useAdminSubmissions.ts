import { useState, useCallback, useEffect } from 'react'
import { adminTasksApi } from '../../../shared/lib/api'
import type { PendingSubmissionView, PaginationMeta } from '../../../shared/types'

export function useAdminSubmissions() {
  const [submissions, setSubmissions] = useState<PendingSubmissionView[]>([])
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, limit: 50, total: 0, has_next: false })
  const [pendingTotal, setPendingTotal] = useState<number | null>(null)
  const [filter, setFilter] = useState<'PENDING' | 'APPROVED' | 'REJECTED' | 'ALL'>('PENDING')
  const [isFetching, setIsFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [processingId, setProcessingId] = useState<number | null>(null)
  const [actionNotes, setActionNotes] = useState<Record<number, string>>({})
  const [actionPenalties, setActionPenalties] = useState<Record<number, number>>({})

  // Modal edit submission state
  const [editingSubmission, setEditingSubmission] = useState<PendingSubmissionView | null>(null)
  const [editPayloadForm, setEditPayloadForm] = useState<Record<string, any>>({})
  const [editSubmissionNotes, setEditSubmissionNotes] = useState('')
  const [isSavingSubmission, setIsSavingSubmission] = useState(false)

  // Image preview state
  const [previewImage, setPreviewImage] = useState<string | null>(null)

  const fetchSubmissions = useCallback(async (statusFilter = filter, page = 1) => {
    setIsFetching(true)
    setError(null)
    try {
      const statusParam = statusFilter === 'ALL' ? undefined : statusFilter
      const res: any = await adminTasksApi.getSubmissions({ status: statusParam, page, limit: 50 })
      const items: PendingSubmissionView[] = Array.isArray(res) ? res : res?.items || []
      const pag: PaginationMeta = Array.isArray(res)
        ? { page, limit: 50, total: items.length, has_next: false }
        : res?.pagination || { page, limit: 50, total: items.length, has_next: false }
      setSubmissions(items)
      setPagination(pag)
      // Exact global "Menunggu" count: when the current filter is PENDING the
      // page total already IS the pending total. Otherwise fetch it with a
      // lightweight limit=1 count query (1 row transferred, total from backend).
      // Never derive the global pending count from the current page's rows.
      if (Array.isArray(res)) {
        setPendingTotal(items.filter((s) => s.status === 'PENDING').length)
      } else if (statusFilter === 'PENDING') {
        setPendingTotal(pag.total)
      } else {
        try {
          const countRes: any = await adminTasksApi.getSubmissions({ status: 'PENDING', page: 1, limit: 1 })
          const countTotal: number | undefined = Array.isArray(countRes)
            ? countRes.filter((s: PendingSubmissionView) => s.status === 'PENDING').length
            : countRes?.pagination?.total
          setPendingTotal(typeof countTotal === 'number' ? countTotal : null)
        } catch {
          setPendingTotal(null)
        }
      }
    } catch (err: any) {
      setError(err?.message || 'Gagal memuat daftar verifikasi tugas')
    } finally {
      setIsFetching(false)
    }
  }, [filter])

  useEffect(() => {
    fetchSubmissions(filter, 1)
  }, [filter, fetchSubmissions])

  const setNote = (id: number, note: string) => {
    setActionNotes((prev) => ({ ...prev, [id]: note }))
  }

  const setPenalty = (id: number, coins: number) => {
    setActionPenalties((prev) => ({ ...prev, [id]: coins }))
  }

  const handleVerify = async (id: number, status: 'APPROVED' | 'REJECTED') => {
    setProcessingId(id)
    try {
      const notes = actionNotes[id]
      const penaltyCoins = status === 'REJECTED' ? actionPenalties[id] : undefined
      await adminTasksApi.verifySubmission(id, status, notes, penaltyCoins)
      // Refresh current page
      await fetchSubmissions(filter, pagination.page)
    } catch (err: any) {
      alert(`Gagal memverifikasi tugas: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  const openEditModal = (submission: PendingSubmissionView) => {
    setEditingSubmission(submission)
    setEditPayloadForm(submission.payload ? JSON.parse(JSON.stringify(submission.payload)) : {})
    setEditSubmissionNotes(submission.admin_notes || '')
  }

  const closeEditModal = () => {
    setEditingSubmission(null)
    setEditPayloadForm({})
    setEditSubmissionNotes('')
  }

  const saveEditSubmission = async () => {
    if (!editingSubmission) return
    setIsSavingSubmission(true)
    try {
      await adminTasksApi.editSubmission(editingSubmission.id, editPayloadForm, editSubmissionNotes || undefined)
      closeEditModal()
      await fetchSubmissions(filter, pagination.page)
    } catch (err: any) {
      alert(`Gagal menyimpan perubahan submission: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setIsSavingSubmission(false)
    }
  }

  return {
    submissions,
    pagination,
    pendingTotal,
    filter,
    setFilter,
    isFetching,
    error,
    processingId,
    actionNotes,
    setNote,
    actionPenalties,
    setPenalty,
    handleVerify,
    editingSubmission,
    editPayloadForm,
    setEditPayloadForm,
    editSubmissionNotes,
    setEditSubmissionNotes,
    isSavingSubmission,
    openEditModal,
    closeEditModal,
    saveEditSubmission,
    previewImage,
    setPreviewImage,
    fetchSubmissions,
  }
}
