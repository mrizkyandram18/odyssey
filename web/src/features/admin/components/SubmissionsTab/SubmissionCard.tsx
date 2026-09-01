import React from 'react'
import {
  CheckCircle2,
  XCircle,
  Clock,
  ExternalLink,
  PenLine,
  Gamepad2,
  FileText,
  Edit3,
  Check,
  X,
} from 'lucide-react'
import type { PendingSubmissionView } from '../../../../shared/types'
import { TaskTypeBadge } from '../shared/TaskTypeBadge'

interface SubmissionCardProps {
  submission: PendingSubmissionView
  processingId: number | null
  actionNote: string
  onNoteChange: (note: string) => void
  actionPenalty: number
  onPenaltyChange: (coins: number) => void
  onVerify: (id: number, status: 'APPROVED' | 'REJECTED') => void
  onOpenEdit: (sub: PendingSubmissionView) => void
  onPreviewImage: (url: string) => void
}

export const SubmissionCard: React.FC<SubmissionCardProps> = ({
  submission,
  processingId,
  actionNote,
  onNoteChange,
  actionPenalty,
  onPenaltyChange,
  onVerify,
  onOpenEdit,
  onPreviewImage,
}) => {
  const isSubApproved = submission.status === 'APPROVED'
  const isSubPending = submission.status === 'PENDING'
  const isSubRejected = submission.status === 'REJECTED'
  const isAutoQuiz = submission.submission_type === 'AUTO_QUIZ'

  const formattedDate = new Date(submission.created_at).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div
      className={`rounded-2xl bg-surface border border-border-subtle shadow-xs overflow-hidden transition-all ${
        isSubApproved
          ? 'border-l-4 border-l-status-success'
          : isSubRejected
          ? 'border-l-4 border-l-status-error'
          : 'border-l-4 border-l-accent-gold'
      }`}
    >
      {/* Top Header Row */}
      <div className="p-3.5 sm:p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 bg-surface-elevated/30 border-b border-border-subtle/60">
        <div className="flex flex-wrap items-center gap-2">
          <TaskTypeBadge type={submission.task_type} />
          <span
            className={`inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md ${
              isSubApproved
                ? 'bg-status-success/15 text-status-success'
                : isSubPending
                ? 'bg-accent-gold/15 text-accent-gold'
                : 'bg-status-error/15 text-status-error'
            }`}
          >
            {isSubApproved ? (
              <>
                <CheckCircle2 className="w-3 h-3" />
                <span>{isAutoQuiz ? 'Disetujui (Otomatis)' : 'Disetujui'}</span>
              </>
            ) : isSubPending ? (
              <>
                <Clock className="w-3 h-3" />
                <span>Menunggu Review</span>
              </>
            ) : (
              <>
                <XCircle className="w-3 h-3" />
                <span>Ditolak</span>
              </>
            )}
          </span>
          <span className="text-xs text-text-secondary">
            Oleh: <strong className="text-text-primary font-bold">{submission.user_name}</strong>
          </span>
          <span className="text-[11px] text-text-secondary font-mono">• {formattedDate}</span>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          <span className="text-xs font-extrabold text-accent-gold px-2 py-0.5 rounded-md bg-accent-gold/10 border border-accent-gold/20">
            +{submission.coins_earned || submission.reward_coins} 🪙
          </span>
        </div>
      </div>

      {/* Main Body */}
      <div className="p-3.5 sm:p-4 space-y-3">
        <div>
          <h4 className="font-heading font-bold text-text-primary text-sm">
            {submission.task_title}
          </h4>
        </div>

        {/* 1. QUIZ / AUTO_QUIZ ANSWERS */}
        {submission.payload &&
          (isAutoQuiz ||
            (typeof submission.payload === 'object' &&
              Object.keys(submission.payload).some((k) => k.startsWith('q') || k === 'answers'))) && (
            <div className="p-2.5 rounded-xl bg-surface-elevated border border-border-subtle space-y-1.5 text-xs">
              <p className="font-bold text-text-secondary text-[11px]">Jawaban Kuis Pengguna:</p>
              <div className="flex flex-wrap gap-1.5">
                {Object.entries(submission.payload.answers || submission.payload).map(([k, v]) => {
                  if (typeof v === 'object' || k === 'submitted_at') return null
                  return (
                    <span
                      key={k}
                      className="px-2 py-0.5 rounded-md bg-surface border border-border-subtle font-mono text-[11px] font-bold text-text-primary flex items-center gap-1"
                    >
                      <span className="text-text-secondary uppercase">{k}:</span>
                      <span className="text-accent-magic">{String(v)}</span>
                    </span>
                  )
                })}
              </div>
            </div>
          )}

        {/* 2. PHOTO or IMAGE */}
        {submission.payload?.file_url &&
          (submission.task_type === 'PHOTO_UPLOAD' ||
            submission.payload.file_url.match(/\.(jpg|jpeg|png|webp)$/i)) && (
            <div className="flex flex-col sm:flex-row items-start gap-3 p-2.5 rounded-xl bg-surface-elevated border border-border-subtle">
              <div
                onClick={() => onPreviewImage(submission.payload.file_url!)}
                className="relative w-full sm:w-48 h-28 rounded-lg overflow-hidden cursor-pointer bg-black shrink-0 group"
              >
                <img
                  src={submission.payload.file_url}
                  alt="Bukti Foto"
                  loading="lazy"
                  className="w-full h-full object-contain group-hover:scale-105 transition-transform"
                />
                <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white text-[11px] font-bold transition-opacity">
                  Perbesar 🔍
                </div>
              </div>
              <div className="flex-1 text-xs space-y-1">
                <p className="font-bold text-text-primary">Bukti Foto Tugas</p>
                {submission.payload.file_size && (
                  <p className="text-[11px] text-text-secondary">
                    Ukuran file: {(submission.payload.file_size / 1024).toFixed(0)} KB
                  </p>
                )}
                <button
                  type="button"
                  onClick={() => onPreviewImage(submission.payload.file_url!)}
                  className="text-xs font-bold text-accent-magic hover:underline flex items-center gap-1 mt-1"
                >
                  Lihat resolusi penuh
                </button>
              </div>
            </div>
          )}

        {/* 3. DOCUMENT */}
        {submission.payload?.file_url &&
          !(
            submission.task_type === 'PHOTO_UPLOAD' ||
            submission.payload.file_url.match(/\.(jpg|jpeg|png|webp)$/i)
          ) && (
            <div className="p-2.5 rounded-xl bg-surface-elevated border border-border-subtle flex items-center justify-between">
              <div className="flex items-center gap-2">
                <FileText className="w-4 h-4 text-accent-magic" />
                <div>
                  <p className="text-xs font-bold text-text-primary line-clamp-1">
                    {submission.payload.file_name || 'Dokumen Bukti'}
                  </p>
                  {submission.payload.file_size && (
                    <p className="text-[10px] text-text-secondary">
                      {(submission.payload.file_size / 1024).toFixed(1)} KB
                    </p>
                  )}
                </div>
              </div>
              <a
                href={submission.payload.file_url}
                target="_blank"
                rel="noreferrer"
                className="px-2.5 py-1 rounded-lg bg-surface border border-border-subtle text-xs font-bold text-accent-magic hover:underline flex items-center gap-1"
              >
                <span>Unduh</span>
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          )}

        {/* 4. TEXT RESPONSE */}
        {submission.payload?.text && (
          <div className="p-2.5 rounded-xl bg-surface-elevated border border-border-subtle space-y-1">
            <p className="text-[10px] text-text-secondary uppercase font-bold flex items-center gap-1">
              <PenLine className="w-3 h-3" />
              <span>Respon Tertulis:</span>
            </p>
            <p className="text-xs text-text-primary whitespace-pre-wrap leading-relaxed">
              {submission.payload.text}
            </p>
          </div>
        )}

        {/* 5. MINI GAME STATS */}
        {submission.payload?.score !== undefined && (
          <div className="p-2.5 rounded-xl bg-surface-elevated border border-border-subtle flex items-center justify-between text-xs">
            <span className="font-bold text-text-secondary flex items-center gap-1">
              <Gamepad2 className="w-3.5 h-3.5 text-accent-magic" />
              <span>Skor Mini Game:</span>
            </span>
            <span className="font-bold text-status-success">
              {submission.payload.score} Poin {submission.payload.moves ? `(${submission.payload.moves} langkah)` : ''}
            </span>
          </div>
        )}

        {submission.payload?.note && (
          <p className="text-xs text-text-secondary italic">
            Catatan Anggota: &quot;{submission.payload.note}&quot;
          </p>
        )}

        {submission.admin_notes && (
          <div className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-secondary">
            <span className="font-bold text-text-primary">Catatan Admin:</span> &quot;{submission.admin_notes}&quot;
          </div>
        )}

        {/* Pending Review Controls */}
        {isSubPending && (
          <div className="pt-2 border-t border-border-subtle space-y-2.5">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <div className="sm:col-span-2 space-y-1">
                <label
                  htmlFor={`note-${submission.id}`}
                  className="text-[11px] font-bold text-text-secondary block"
                >
                  Catatan untuk anggota (opsional, wajib jika menolak)
                </label>
                <input
                  id={`note-${submission.id}`}
                  type="text"
                  placeholder="Contoh: Foto kurang jelas, ulangi dari sudut lain"
                  value={actionNote || ''}
                  onChange={(e) => onNoteChange(e.target.value)}
                  className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic"
                />
              </div>

              <div className="space-y-1">
                <label
                  htmlFor={`penalty-${submission.id}`}
                  className="text-[11px] font-bold text-status-error block"
                >
                  Penalti Koin jika Ditolak:
                </label>
                <div className="flex items-center gap-1.5">
                  <input
                    id={`penalty-${submission.id}`}
                    type="number"
                    min="0"
                    max="10000"
                    placeholder="0"
                    value={actionPenalty || ''}
                    onChange={(e) =>
                      onPenaltyChange(Math.max(0, parseInt(e.target.value, 10) || 0))
                    }
                    className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-status-error focus:outline-none focus:border-status-error"
                  />
                  <span className="text-[10px] text-text-secondary shrink-0">🪙 dipotong</span>
                </div>
              </div>
            </div>

            {/* Actions Bar */}
            <div className="flex items-center justify-end gap-2 pt-1">
              <button
                type="button"
                onClick={() => onOpenEdit(submission)}
                className="px-3 py-2 rounded-xl bg-surface border border-border-subtle hover:bg-surface-elevated text-text-primary font-bold text-xs flex items-center justify-center gap-1.5 transition-colors cursor-pointer"
              >
                <Edit3 className="w-3.5 h-3.5 text-accent-magic" />
                <span>Edit Jawaban</span>
              </button>

              <button
                type="button"
                aria-label={`Tolak verifikasi ${submission.task_title}`}
                disabled={processingId === submission.id}
                onClick={() => onVerify(submission.id, 'REJECTED')}
                className="px-4 py-2 rounded-xl bg-surface border border-status-error/30 text-status-error hover:bg-status-error/10 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                <X className="w-3.5 h-3.5" />
                <span>Tolak</span>
              </button>

              <button
                type="button"
                aria-label={`Setujui verifikasi ${submission.task_title}`}
                disabled={processingId === submission.id}
                onClick={() => onVerify(submission.id, 'APPROVED')}
                className="px-4 py-2 rounded-xl bg-status-success text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-xs hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                <Check className="w-3.5 h-3.5" />
                <span>{processingId === submission.id ? 'Memproses...' : 'Setujui'}</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
