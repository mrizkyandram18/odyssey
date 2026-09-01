import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
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
    <AnimatePresence>
      <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
        <motion.div
          initial={{ opacity: 0, scale: 0.96 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.96 }}
          transition={{ duration: 0.15 }}
          className="w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
            <div>
              <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Edit3 className="w-4 h-4 text-accent-magic" />
                <span>Edit Jawaban Submission</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5">
                {submission.task_title} • {submission.user_name}
              </p>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors cursor-pointer"
              title="Tutup"
              aria-label="Tutup"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-5 space-y-4">
            {/* Text Response editor */}
            {(submission.task_type === 'TEXT_RESPONSE' || formPayload.text !== undefined) && (
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Teks Jawaban / Respon</label>
                <textarea
                  rows={4}
                  value={formPayload.text || ''}
                  onChange={(e) => setFormPayload({ ...formPayload, text: e.target.value })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic leading-relaxed resize-none"
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
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
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
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
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
                      <span className="w-16 px-2 py-1.5 rounded-lg bg-surface-elevated border border-border-subtle text-xs font-mono font-bold text-text-secondary uppercase text-center shrink-0">
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
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
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
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
              />
            </div>

            {/* Sticky Footer */}
            <div className="sticky bottom-0 -mx-5 -mb-5 mt-4 px-5 py-3.5 border-t border-border-subtle bg-surface flex gap-3">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface transition-colors cursor-pointer"
              >
                Batalkan
              </button>
              <button
                type="submit"
                disabled={isSaving}
                className="flex-1 py-2.5 rounded-xl bg-accent-magic text-white font-bold text-xs shadow-xs hover:brightness-110 disabled:opacity-50 transition-all cursor-pointer"
              >
                {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  )
}
