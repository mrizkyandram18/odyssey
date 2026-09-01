import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Edit3 } from 'lucide-react'
import type { TaskView } from '../../../../shared/types'

interface EditTaskModalProps {
  task: TaskView | null
  form: {
    title: string
    description: string
    reward_coins: number
    reward_xp: number
    video_url: string
  }
  setForm: React.Dispatch<
    React.SetStateAction<{
      title: string
      description: string
      reward_coins: number
      reward_xp: number
      video_url: string
    }>
  >
  isSaving: boolean
  onClose: () => void
  onSave: () => void
}

export const EditTaskModal: React.FC<EditTaskModalProps> = ({
  task,
  form,
  setForm,
  isSaving,
  onClose,
  onSave,
}) => {
  if (!task) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave()
  }

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <motion.div
          initial={{ scale: 0.96, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.96, opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
            <div>
              <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Edit3 className="w-4 h-4 text-accent-magic" />
                <span>Edit Tugas</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5">
                #{task.step_order} — {task.task_type}
              </p>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle flex items-center justify-center text-text-secondary hover:text-text-primary transition-colors cursor-pointer"
              title="Tutup"
              aria-label="Tutup"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-5 space-y-4">
            {/* Judul */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Judul Tugas <span className="text-status-error">*</span>
              </label>
              <input
                type="text"
                required
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
              />
            </div>

            {/* Deskripsi */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Deskripsi
              </label>
              <textarea
                rows={3}
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic resize-none"
              />
            </div>

            {/* YouTube URL */}
            {(task.task_type === 'VIDEO' || task.task_type === 'QUIZ') && (
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">
                  Link YouTube (Opsional)
                </label>
                <input
                  type="url"
                  value={form.video_url}
                  onChange={(e) => setForm({ ...form, video_url: e.target.value })}
                  placeholder="https://www.youtube.com/watch?v=..."
                  className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
            )}

            {/* Rewards */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Hadiah Koin (🪙)
              </label>
              <input
                type="number"
                min={0}
                value={form.reward_coins}
                onChange={(e) => setForm({ ...form, reward_coins: Number(e.target.value) })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic font-mono font-bold"
              />
            </div>

            {/* Sticky Footer */}
            <div className="sticky bottom-0 -mx-5 -mb-5 mt-4 px-5 py-3.5 border-t border-border-subtle bg-surface flex gap-3">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface transition-colors cursor-pointer"
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={isSaving}
                className="flex-1 py-2.5 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 disabled:opacity-60 transition-all shadow-xs cursor-pointer"
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
