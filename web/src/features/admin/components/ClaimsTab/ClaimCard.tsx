import React from 'react'
import {
  Wallet,
  Building2,
  Smartphone,
  Coins,
  Copy,
  XCircle,
  CheckCircle2,
} from 'lucide-react'
import type { ClaimView } from '../../../../shared/types'

interface ClaimCardProps {
  claim: ClaimView
  processingId: number | null
  actionNote: string
  onNoteChange: (note: string) => void
  onProcess: (id: number, status: 'APPROVED' | 'REJECTED') => void
}

export const ClaimCard: React.FC<ClaimCardProps> = ({
  claim,
  processingId,
  actionNote,
  onNoteChange,
  onProcess,
}) => {
  const isPending = claim.status === 'PENDING'
  const isApproved = claim.status === 'APPROVED'
  const isRejected = claim.status === 'REJECTED'

  const handleCopy = () => {
    navigator.clipboard.writeText(claim.target_value)
    alert('Nomor tujuan berhasil disalin!')
  }

  return (
    <div
      className={`p-4 rounded-2xl border shadow-sm space-y-3 ${
        isApproved
          ? 'bg-surface border-status-success/30'
          : isRejected
          ? 'bg-surface border-status-error/30'
          : 'bg-surface border-border-subtle'
      }`}
    >
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-1.5">
            {claim.target_type === 'EWALLET' ? (
              <Wallet className="w-4 h-4 text-accent-magic" />
            ) : claim.target_type === 'BANK' ? (
              <Building2 className="w-4 h-4 text-status-success" />
            ) : (
              <Smartphone className="w-4 h-4 text-accent-cyan" />
            )}
            <h4 className="font-heading font-bold text-text-primary text-base">
              Pencairan {claim.target_type}
            </h4>
          </div>
          <p className="text-xs text-text-secondary mt-0.5">
            Pemohon: <strong className="text-text-primary">{claim.user_name}</strong> •{' '}
            {new Date(claim.created_at).toLocaleDateString('id-ID', {
              day: 'numeric',
              month: 'short',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </p>
        </div>
        <span className="px-3 py-1 rounded-full bg-accent-gold/15 text-accent-gold font-bold text-xs flex items-center gap-1">
          <Coins className="w-3.5 h-3.5" />
          <span>{claim.coins_redeemed.toLocaleString('id-ID')} Koin</span>
        </span>
      </div>

      {/* Destination info with Copy Button */}
      <div className="p-3 rounded-xl bg-surface border border-border-subtle flex items-center justify-between">
        <div>
          <p className="text-[10px] text-text-secondary uppercase font-bold">
            Tujuan Penukaran / No. Rekening / No. HP
          </p>
          <p className="text-sm font-mono font-bold text-text-primary mt-0.5">
            {claim.target_value}
          </p>
        </div>
        <button
          onClick={handleCopy}
          className="p-2 rounded-lg bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary active:scale-95 transition-all"
          title="Salin Nomor"
        >
          <Copy className="w-4 h-4" />
        </button>
      </div>

      {claim.admin_notes && (
        <p className="text-xs text-text-secondary italic bg-surface-elevated p-2.5 rounded-xl border border-border-subtle">
          Catatan Admin: &quot;{claim.admin_notes}&quot;
        </p>
      )}

      {isPending && (
        <>
          <label
            htmlFor={`claim-note-${claim.id}`}
            className="text-[11px] font-bold text-text-secondary"
          >
            Catatan / No. Ref (opsional)
          </label>
          <input
            id={`claim-note-${claim.id}`}
            type="text"
            placeholder="Contoh: TRF BCA 123456 — 28 Agu 2026"
            value={actionNote || ''}
            onChange={(e) => onNoteChange(e.target.value)}
            className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
          />

          <div className="flex gap-2 pt-1">
            <button
              type="button"
              aria-label={`Tolak pencairan ${claim.target_value}`}
              disabled={processingId === claim.id}
              onClick={() => onProcess(claim.id, 'REJECTED')}
              className="flex-1 py-2.5 rounded-xl bg-surface border border-status-error/20 text-status-error hover:bg-status-error/10 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <XCircle className="w-4 h-4" />
              <span>{processingId === claim.id ? 'Memproses...' : 'Tolak & Refund'}</span>
            </button>

            <button
              type="button"
              aria-label={`Selesaikan pencairan ${claim.target_value}`}
              disabled={processingId === claim.id}
              aria-busy={processingId === claim.id}
              onClick={() => onProcess(claim.id, 'APPROVED')}
              className="flex-1 py-2.5 rounded-xl bg-status-success hover:brightness-110 text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <CheckCircle2 className="w-4 h-4" />
              <span>Sudah Ditransfer</span>
            </button>
          </div>
        </>
      )}

      {isApproved && (
        <div className="p-3 rounded-xl bg-status-success/10 border border-status-success/20 text-xs text-status-success font-bold flex items-center gap-1.5">
          <CheckCircle2 className="w-4 h-4" />
          <span>Klaim telah disetujui & ditransfer</span>
        </div>
      )}

      {isRejected && (
        <div className="p-3 rounded-xl bg-status-error/10 border border-status-error/20 text-xs text-status-error font-bold flex items-center gap-1.5">
          <XCircle className="w-4 h-4" />
          <span>Klaim telah ditolak dan saldo koin dikembalikan</span>
        </div>
      )}
    </div>
  )
}
