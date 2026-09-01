import React, { useState } from 'react'
import {
  Wallet,
  Building2,
  Smartphone,
  Coins,
  Copy,
  Check,
  XCircle,
  CheckCircle2,
  Clock,
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
  const [copied, setCopied] = useState(false)
  const isPending = claim.status === 'PENDING'
  const isApproved = claim.status === 'APPROVED'
  const isRejected = claim.status === 'REJECTED'

  const handleCopy = () => {
    navigator.clipboard.writeText(claim.target_value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const formattedDate = new Date(claim.created_at).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div
      className={`rounded-2xl bg-surface border border-border-subtle shadow-xs overflow-hidden transition-all ${
        isApproved
          ? 'border-l-4 border-l-status-success'
          : isRejected
          ? 'border-l-4 border-l-status-error'
          : 'border-l-4 border-l-accent-gold'
      }`}
    >
      {/* Header Row */}
      <div className="p-3.5 sm:p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2 bg-surface-elevated/30 border-b border-border-subtle/60">
        <div className="flex flex-wrap items-center gap-2">
          <div className="flex items-center gap-1.5 font-bold text-xs text-text-primary">
            {claim.target_type === 'EWALLET' ? (
              <Wallet className="w-4 h-4 text-accent-magic" />
            ) : claim.target_type === 'BANK' ? (
              <Building2 className="w-4 h-4 text-status-success" />
            ) : (
              <Smartphone className="w-4 h-4 text-accent-cyan" />
            )}
            <span>Pencairan {claim.target_type}</span>
          </div>

          <span
            className={`inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md ${
              isApproved
                ? 'bg-status-success/15 text-status-success'
                : isPending
                ? 'bg-accent-gold/15 text-accent-gold'
                : 'bg-status-error/15 text-status-error'
            }`}
          >
            {isApproved ? (
              <>
                <CheckCircle2 className="w-3 h-3" />
                <span>Ditransfer</span>
              </>
            ) : isPending ? (
              <>
                <Clock className="w-3 h-3" />
                <span>Menunggu Proses</span>
              </>
            ) : (
              <>
                <XCircle className="w-3 h-3" />
                <span>Ditolak / Refund</span>
              </>
            )}
          </span>

          <span className="text-xs text-text-secondary">
            Pemohon: <strong className="text-text-primary font-bold">{claim.user_name}</strong>
          </span>
          <span className="text-[11px] text-text-secondary font-mono">• {formattedDate}</span>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          <span className="text-xs font-extrabold text-accent-gold px-2.5 py-0.5 rounded-md bg-accent-gold/10 border border-accent-gold/20 flex items-center gap-1">
            <Coins className="w-3 h-3" />
            <span>{claim.coins_redeemed.toLocaleString('id-ID')} Koin</span>
          </span>
        </div>
      </div>

      {/* Main Body */}
      <div className="p-3.5 sm:p-4 space-y-3">
        {/* Destination Box with Inline Copy */}
        <div className="p-2.5 rounded-xl bg-surface-elevated border border-border-subtle flex items-center justify-between gap-3">
          <div className="min-w-0">
            <p className="text-[10px] text-text-secondary uppercase font-bold tracking-wider">
              Tujuan Penukaran / Rekening / No. HP
            </p>
            <p className="text-xs sm:text-sm font-mono font-bold text-text-primary mt-0.5 truncate">
              {claim.target_value}
            </p>
          </div>

          <button
            type="button"
            onClick={handleCopy}
            className="px-2.5 py-1.5 rounded-lg bg-surface border border-border-subtle text-xs font-bold text-text-secondary hover:text-text-primary hover:bg-surface-elevated active:scale-95 transition-all flex items-center gap-1 shrink-0 cursor-pointer"
            title="Salin Nomor"
            aria-label="Salin Nomor Tujuan"
          >
            {copied ? (
              <>
                <Check className="w-3.5 h-3.5 text-status-success" />
                <span className="text-status-success text-[11px]">Tersalin</span>
              </>
            ) : (
              <>
                <Copy className="w-3.5 h-3.5" />
                <span className="text-[11px]">Salin</span>
              </>
            )}
          </button>
        </div>

        {claim.admin_notes && (
          <div className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-secondary">
            <span className="font-bold text-text-primary">Catatan Admin:</span> &quot;{claim.admin_notes}&quot;
          </div>
        )}

        {/* Pending Actions */}
        {isPending && (
          <div className="pt-2 border-t border-border-subtle space-y-2.5">
            <div className="space-y-1">
              <label
                htmlFor={`claim-note-${claim.id}`}
                className="text-[11px] font-bold text-text-secondary block"
              >
                Catatan / No. Ref Transfer (opsional)
              </label>
              <input
                id={`claim-note-${claim.id}`}
                type="text"
                placeholder="Contoh: TRF BCA 123456 — 28 Agu 2026"
                value={actionNote || ''}
                onChange={(e) => onNoteChange(e.target.value)}
                className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic"
              />
            </div>

            <div className="flex items-center justify-end gap-2 pt-1">
              <button
                type="button"
                aria-label={`Tolak pencairan ${claim.target_value}`}
                disabled={processingId === claim.id}
                onClick={() => onProcess(claim.id, 'REJECTED')}
                className="px-4 py-2 rounded-xl bg-surface border border-status-error/30 text-status-error hover:bg-status-error/10 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                <XCircle className="w-3.5 h-3.5" />
                <span>{processingId === claim.id ? 'Memproses...' : 'Tolak & Refund'}</span>
              </button>

              <button
                type="button"
                aria-label={`Selesaikan pencairan ${claim.target_value}`}
                disabled={processingId === claim.id}
                aria-busy={processingId === claim.id}
                onClick={() => onProcess(claim.id, 'APPROVED')}
                className="px-4 py-2 rounded-xl bg-status-success hover:brightness-110 text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-xs transition-all disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                <CheckCircle2 className="w-3.5 h-3.5" />
                <span>Sudah Ditransfer</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
