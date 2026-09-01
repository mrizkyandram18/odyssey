import React from 'react'
import { CheckCircle2, ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react'
import { useAdminClaims } from '../../hooks/useAdminClaims'
import { ClaimCard } from './ClaimCard'

export const ClaimsQueue: React.FC = () => {
  const {
    claims,
    pagination,
    filter,
    setFilter,
    isFetching,
    error,
    processingId,
    actionNotes,
    setNote,
    handleProcess,
    fetchClaims,
  } = useAdminClaims()

  return (
    <div className="space-y-4">
      {error && (
        <div className="p-4 rounded-2xl bg-status-error/10 border border-status-error/20 text-status-error text-xs flex items-center justify-between">
          <span>{error}</span>
          <button
            type="button"
            onClick={() => fetchClaims(filter, pagination.page)}
            className="font-bold underline ml-2"
          >
            Coba lagi
          </button>
        </div>
      )}

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
            <span>Permintaan Pencairan Koin</span>
            {isFetching && <RefreshCw className="w-3.5 h-3.5 animate-spin text-text-secondary" />}
          </h3>
          <span className="text-xs text-text-secondary">Transfer / Top-up saldo anggota</span>
        </div>

        {/* Status filter pills */}
        <div className="flex items-center gap-1 p-1 bg-surface rounded-xl border border-border-subtle text-xs overflow-x-auto">
          <button
            type="button"
            onClick={() => setFilter('ALL')}
            className={`px-3 py-1 rounded-lg font-bold transition-colors whitespace-nowrap ${
              filter === 'ALL'
                ? 'bg-accent-magic text-white shadow-sm'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Semua
          </button>
          <button
            type="button"
            onClick={() => setFilter('PENDING')}
            className={`px-3 py-1 rounded-lg font-bold transition-colors whitespace-nowrap ${
              filter === 'PENDING'
                ? 'bg-accent-gold text-white shadow-sm'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Menunggu
          </button>
          <button
            type="button"
            onClick={() => setFilter('APPROVED')}
            className={`px-3 py-1 rounded-lg font-bold transition-colors whitespace-nowrap ${
              filter === 'APPROVED'
                ? 'bg-status-success text-white shadow-sm'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Disetujui
          </button>
          <button
            type="button"
            onClick={() => setFilter('REJECTED')}
            className={`px-3 py-1 rounded-lg font-bold transition-colors whitespace-nowrap ${
              filter === 'REJECTED'
                ? 'bg-status-error text-white shadow-sm'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Ditolak
          </button>
        </div>
      </div>

      {claims.length === 0 && !isFetching ? (
        <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
          <div className="w-10 h-10 mx-auto rounded-xl bg-status-success/10 text-status-success flex items-center justify-center">
            <CheckCircle2 className="w-5 h-5" />
          </div>
          <p className="font-bold text-text-primary text-sm">Tidak Ada Klaim</p>
          <p className="text-xs text-text-secondary max-w-xs mx-auto">
            {filter === 'ALL'
              ? 'Semua pengajuan penukaran koin sudah selesai diproses.'
              : `Tidak ada data klaim dengan status ${filter}.`}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {claims.map((claim) => (
            <ClaimCard
              key={claim.id}
              claim={claim}
              processingId={processingId}
              actionNote={actionNotes[claim.id] || ''}
              onNoteChange={(note) => setNote(claim.id, note)}
              onProcess={handleProcess}
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
            onClick={() => fetchClaims(filter, pagination.page - 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Sebelumnya</span>
          </button>

          <span className="text-xs font-bold text-text-secondary">
            Halaman {pagination.page}
          </span>

          <button
            type="button"
            disabled={!pagination.has_next || isFetching}
            onClick={() => fetchClaims(filter, pagination.page + 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1"
          >
            <span>Selanjutnya</span>
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        </div>
      )}
    </div>
  )
}
