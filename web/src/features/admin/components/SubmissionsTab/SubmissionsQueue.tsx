import React from 'react'
import { CheckCircle2, ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react'
import { useAdminSubmissions } from '../../hooks/useAdminSubmissions'
import { SubmissionCard } from './SubmissionCard'
import { EditSubmissionModal } from './EditSubmissionModal'
import { ImagePreviewModal } from '../shared/ImagePreviewModal'

export const SubmissionsQueue: React.FC = () => {
  const {
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
  } = useAdminSubmissions()

  // Total is the exact backend total (pagination.total), never submissions.length.
  // Menunggu is the exact global pending count (pendingTotal), never the
  // current page's pending rows.
  const pagePendingCount = submissions.filter((s) => s.status === 'PENDING').length
  const pendingDisplay = pendingTotal ?? pagePendingCount

  const totalPages = pagination.limit > 0 && pagination.total > 0
    ? Math.max(1, Math.ceil(pagination.total / pagination.limit))
    : 1
  const showPagination = totalPages > 1 || pagination.page > 1 || pagination.has_next

  const gotoPage = (page: number) => {
    if (isFetching) return
    if (page < 1 || page === pagination.page) return
    if (totalPages > 1 && page > totalPages) return
    fetchSubmissions(filter, page)
  }

  // Compact windowed page numbers: 1 … c-1 c c+1 … N (max ~7 slots).
  const pageItems: (number | 'ellipsis-left' | 'ellipsis-right')[] = (() => {
    if (totalPages <= 7) {
      return Array.from({ length: totalPages }, (_, i) => i + 1)
    }
    const current = pagination.page
    const window = new Set<number>([1, 2, current - 1, current, current + 1, totalPages - 1, totalPages])
    const sorted = [...window].filter((n) => n >= 1 && n <= totalPages).sort((a, b) => a - b)
    const out: (number | 'ellipsis-left' | 'ellipsis-right')[] = []
    let prev = 0
    for (const n of sorted) {
      if (prev !== 0 && n - prev > 1) {
        out.push(prev === 1 ? 'ellipsis-left' : 'ellipsis-right')
      }
      out.push(n)
      prev = n
    }
    return out
  })()

  return (
    <div className="space-y-3.5">
      {error && (
        <div className="p-3.5 rounded-2xl bg-status-error/10 border border-status-error/20 text-status-error text-xs flex items-center justify-between">
          <span>{error}</span>
          <button
            type="button"
            onClick={() => fetchSubmissions(filter, pagination.page)}
            className="font-bold underline ml-2 cursor-pointer"
          >
            Coba lagi
          </button>
        </div>
      )}

      {/* Toolbar & Filter */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 p-3 rounded-2xl bg-surface border border-border-subtle shadow-xs">
        <div>
          <h3 className="font-bold text-text-primary text-xs flex items-center gap-2">
            <span>Antrean Verifikasi Bukti Tugas ({pendingDisplay} Menunggu / {pagination.total} Total)</span>
            {isFetching && <RefreshCw className="w-3.5 h-3.5 animate-spin text-text-secondary" />}
          </h3>
          <p className="text-[11px] text-text-secondary mt-0.5">
            Approve memberi koin otomatis. Submission otomatis kuis langsung tercatat di sini.
          </p>
        </div>

        {/* Status filter pills */}
        <div className="flex items-center gap-1 p-1 bg-surface-elevated rounded-xl border border-border-subtle text-xs overflow-x-auto shrink-0">
          <button
            type="button"
            onClick={() => setFilter('ALL')}
            className={`px-3 py-1 rounded-lg font-bold text-xs transition-colors whitespace-nowrap cursor-pointer ${
              filter === 'ALL'
                ? 'bg-accent-magic text-white shadow-xs'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Semua
          </button>
          <button
            type="button"
            onClick={() => setFilter('PENDING')}
            className={`px-3 py-1 rounded-lg font-bold text-xs transition-colors whitespace-nowrap cursor-pointer ${
              filter === 'PENDING'
                ? 'bg-accent-gold text-white shadow-xs'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Menunggu
          </button>
          <button
            type="button"
            onClick={() => setFilter('APPROVED')}
            className={`px-3 py-1 rounded-lg font-bold text-xs transition-colors whitespace-nowrap cursor-pointer ${
              filter === 'APPROVED'
                ? 'bg-status-success text-white shadow-xs'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Disetujui
          </button>
          <button
            type="button"
            onClick={() => setFilter('REJECTED')}
            className={`px-3 py-1 rounded-lg font-bold text-xs transition-colors whitespace-nowrap cursor-pointer ${
              filter === 'REJECTED'
                ? 'bg-status-error text-white shadow-xs'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Ditolak
          </button>
        </div>
      </div>

      {/* Queue items */}
      {submissions.length === 0 && !isFetching ? (
        <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
          <div className="w-10 h-10 mx-auto rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center">
            <CheckCircle2 className="w-5 h-5" />
          </div>
          <p className="font-bold text-text-primary text-sm">Tidak Ada Antrean Verifikasi</p>
          <p className="text-xs text-text-secondary max-w-xs mx-auto">
            {filter === 'ALL'
              ? 'Semua tugas dari anggota sudah selesai diperiksa.'
              : `Tidak ada data submission dengan status ${filter}.`}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {submissions.map((sub) => (
            <SubmissionCard
              key={sub.id}
              submission={sub}
              processingId={processingId}
              actionNote={actionNotes[sub.id] || ''}
              onNoteChange={(note) => setNote(sub.id, note)}
              actionPenalty={actionPenalties[sub.id] || 0}
              onPenaltyChange={(coins) => setPenalty(sub.id, coins)}
              onVerify={handleVerify}
              onOpenEdit={openEditModal}
              onPreviewImage={setPreviewImage}
            />
          ))}
        </div>
      )}

      {/* Pagination Controls — server-side: each page fetches only that page */}
      {showPagination && (
        <div className="flex items-center justify-between gap-2 p-3 rounded-2xl bg-surface border border-border-subtle">
          <button
            type="button"
            disabled={pagination.page <= 1 || isFetching}
            onClick={() => gotoPage(pagination.page - 1)}
            aria-label="Halaman sebelumnya"
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1 cursor-pointer"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Sebelumnya</span>
          </button>

          <div className="flex items-center gap-1 overflow-x-auto" role="navigation" aria-label="Navigasi halaman verifikasi">
            {pageItems.map((item) =>
              typeof item === 'number' ? (
                <button
                  key={item}
                  type="button"
                  disabled={isFetching}
                  onClick={() => gotoPage(item)}
                  aria-label={`Halaman ${item}`}
                  aria-current={item === pagination.page ? 'page' : undefined}
                  className={`min-w-8 px-2 py-1.5 rounded-xl text-xs font-bold transition-colors cursor-pointer disabled:cursor-not-allowed ${
                    item === pagination.page
                      ? 'bg-accent-magic text-white shadow-xs'
                      : 'bg-surface-elevated border border-border-subtle text-text-primary hover:bg-surface disabled:opacity-40'
                  }`}
                >
                  {item}
                </button>
              ) : (
                <span key={item} className="px-1 text-xs font-bold text-text-secondary" aria-hidden="true">
                  …
                </span>
              )
            )}
          </div>

          <button
            type="button"
            disabled={(!pagination.has_next && pagination.page >= totalPages) || isFetching}
            onClick={() => gotoPage(pagination.page + 1)}
            aria-label="Halaman selanjutnya"
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1 cursor-pointer"
          >
            <span>Selanjutnya</span>
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* Modals */}
      <EditSubmissionModal
        submission={editingSubmission}
        formPayload={editPayloadForm}
        setFormPayload={setEditPayloadForm}
        adminNotes={editSubmissionNotes}
        setAdminNotes={setEditSubmissionNotes}
        isSaving={isSavingSubmission}
        onClose={closeEditModal}
        onSave={saveEditSubmission}
      />

      <ImagePreviewModal
        imageUrl={previewImage}
        onClose={() => setPreviewImage(null)}
      />
    </div>
  )
}
