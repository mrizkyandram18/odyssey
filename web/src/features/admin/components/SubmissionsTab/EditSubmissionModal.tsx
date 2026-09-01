import React from 'react'
import { motion } from 'framer-motion'
import { X, Edit3 } from 'lucide-react'
import type { PendingSubmissionView } from '../../../../shared/types'

interface EditSubmissionModalProps {
  submission: PendingSubmissionView | null
  formPayload: Record<string, any>
  setFormPayload: React.Dispatch<React.SetStateAction<Record<string, any>>>
  adminNotes: string
  setAdminNotes: (notes: string) => void
  isSaving: boolean
  onClose: () => void
  onSave: () => void
}

export const EditSubmissionModal: React.FC<EditSubmissionModalProps> = ({
  submission,
  formPayload,
  setFormPayload,
  adminNotes,
  setAdminNotes,
  isSaving,
  onClose,
  onSave,
}) => {
  if (!submission) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave()
  }

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        <div className="p-5 border-b border-border-subtle flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Edit3 className="w-5 h-5 text-accent-magic" />
            <h3 className="font-heading font-bold text-text-primary text-base">
              Edit Jawaban Submission
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded-lg text-text-secondary hover:text-text-primary transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4 overflow-y-auto">
          {/* Text Response editor */}
          {(submission.task_type === 'TEXT_RESPONSE' || formPayload.text !== undefined) && (
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Teks Jawaban / Respon</label>
              <textarea
                rows={4}
                value={formPayload.text || ''}
                onChange={(e) => setFormPayload({ ...formPayload, text: e.target.value })}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic leading-relaxed"
                placeholder="Masukkan teks jawaban..."
              />
            </div>
          )}

          {/* Mini Game editor */}
          {(submission.task_type === 'MINI_GAME' || formPayload.score !== undefined) && (
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Skor Game</label>
                <input
                  type="number"
                  min="0"
                  max="1000000"
                  value={formPayload.score !== undefined ? formPayload.score : 0}
                  onChange={(e) =>
                    setFormPayload({ ...formPayload, score: parseInt(e.target.value, 10) || 0 })
                  }
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Jumlah Langkah (Moves)</label>
                <input
                  type="number"
                  min="0"
                  value={formPayload.moves !== undefined ? formPayload.moves : 0}
                  onChange={(e) =>
                    setFormPayload({ ...formPayload, moves: parseInt(e.target.value, 10) || 0 })
                  }
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
            </div>
          )}

          {/* Quiz / Auto Quiz Answers editor */}
          {formPayload.answers && typeof formPayload.answers === 'object' && (
            <div className="space-y-2">
              <label className="text-xs font-bold text-text-secondary">Jawaban Kuis per Soal</label>
              <div className="space-y-2">
                {Object.entries(formPayload.answers).map(([qKey, qAns]) => (
                  <div key={qKey} className="flex items-center gap-2">
                    <span className="w-16 px-2 py-1.5 rounded-lg bg-surface border border-border-subtle text-xs font-mono font-bold text-text-secondary uppercase text-center shrink-0">
                      {qKey}
                    </span>
                    <input
                      type="text"
                      value={String(qAns || '')}
                      onChange={(e) =>
                        setFormPayload({
                          ...formPayload,
                          answers: {
                            ...formPayload.answers,
                            [qKey]: e.target.value,
                          },
                        })
                      }
                      className="flex-1 p-2 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Attachment / File note editor */}
          <div className="space-y-1">
            <label className="text-xs font-bold text-text-secondary">Catatan Bukti / Note Pengguna</label>
            <input
              type="text"
              value={formPayload.note || ''}
              onChange={(e) => setFormPayload({ ...formPayload, note: e.target.value })}
              placeholder="Catatan pengguna..."
              className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
            />
          </div>

          {/* Admin Note editor */}
          <div className="space-y-1">
            <label className="text-xs font-bold text-text-secondary">Catatan Admin</label>
            <input
              type="text"
              value={adminNotes}
              onChange={(e) => setAdminNotes(e.target.value)}
              placeholder="Catatan review admin..."
              className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
            />
          </div>

          <div className="pt-2 flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-md hover:brightness-110 disabled:opacity-50"
            >
              {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  )
}
