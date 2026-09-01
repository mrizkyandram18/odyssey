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

  const pendingCount = submissions.filter((s) => s.status === 'PENDING').length

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
            <span>Antrean Verifikasi Bukti Tugas ({pendingCount} Menunggu / {submissions.length} Total)</span>
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

      {/* Pagination Controls */}
      {(pagination.page > 1 || pagination.has_next) && (
        <div className="flex items-center justify-between p-3 rounded-2xl bg-surface border border-border-subtle">
          <button
            type="button"
            disabled={pagination.page <= 1 || isFetching}
            onClick={() => fetchSubmissions(filter, pagination.page - 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1 cursor-pointer"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Sebelumnya</span>
          </button>

          <span className="text-xs font-bold text-text-secondary font-mono">
            Halaman {pagination.page}
          </span>

          <button
            type="button"
            disabled={!pagination.has_next || isFetching}
            onClick={() => fetchSubmissions(filter, pagination.page + 1)}
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
