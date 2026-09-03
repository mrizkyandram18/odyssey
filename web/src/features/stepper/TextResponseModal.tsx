import React, { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { FileEdit, CheckCircle2, X, Clock, Sparkles, AlertCircle, ArrowRight } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'

interface TextResponseModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

export const TextResponseModal: React.FC<TextResponseModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const isAlreadyDone = task.status === 'APPROVED' || task.status === 'PENDING'
  const [response, setResponse] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(isAlreadyDone)

  const minChars = task.config?.minimum_characters || 10
  const maxChars = task.config?.maximum_characters || 5000
  // Hotfix: description holds the real instruction (prompt in DB is just title duplicate for 520/521). Prioritize description
  const prompt = task.description || task.config?.prompt || 'Tuliskan jawaban atau refleksi kamu:'

  const charCount = response.trim().length
  const isValidLength = charCount >= minChars && charCount <= maxChars

  const handleSubmit = async () => {
    if (!isValidLength) {
      if (charCount < minChars) {
        setErrorMessage(`Teks jawaban minimal ${minChars} karakter (saat ini ${charCount})`)
      } else {
        setErrorMessage(`Teks jawaban maksimal ${maxChars} karakter`)
      }
      return
    }

    setSubmitting(true)
    setErrorMessage(null)
    try {
      const res = await tasksApi.submit(task.id, {
        payload: {
          text: response.trim(),
          submitted_at: new Date().toISOString(),
        },
      })

      if (res.success) {
        setSubmitted(true)
      } else {
        setErrorMessage(res.error || 'Gagal mengirimkan respon teks')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengirimkan jawaban. Periksa koneksi Anda.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.97, opacity: 0, y: 12 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.97, opacity: 0, y: 12 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[92vh] sm:max-h-[90vh]"
      >
        {/* Header — compact, unified */}
        <div className="flex items-center justify-between px-5 py-3.5 border-b border-border-subtle bg-surface shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            <span className="w-7 h-7 rounded-full bg-accent-magic/12 text-accent-magic flex items-center justify-center font-bold text-xs shrink-0">
              #{task.step_order}
            </span>
            <h3 className="font-bold text-text-primary text-[14px] line-clamp-1">
              {task.title}
            </h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup"
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors shrink-0"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content Body */}
        <div className="flex-1 overflow-y-auto p-5 space-y-5">
          <AnimatePresence mode="wait">
            {!submitted ? (
              <motion.div
                key="form-step"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-4"
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">
                    ✍️ Tugas Respon Teks
                  </span>
                  <div className="text-xs text-text-secondary font-bold">
                    +{task.reward_coins} 🪙 | +{task.reward_xp} XP
                  </div>
                </div>

                <div className="rounded-xl bg-surface border border-border-subtle p-3.5 space-y-1.5">
                  <div className="flex items-center gap-1.5 text-accent-magic font-bold text-xs">
                    <FileEdit className="w-3.5 h-3.5" />
                    <span>Petunjuk</span>
                  </div>
                  <p className="text-sm text-text-primary leading-relaxed font-medium whitespace-pre-wrap">
                    {prompt}
                  </p>
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs text-text-secondary">
                    <label className="font-bold">Jawaban Kamu:</label>
                    <span className={charCount < minChars ? 'text-accent-gold font-medium' : 'text-status-success font-bold'}>
                      {charCount} / {minChars} - {maxChars} karakter
                    </span>
                  </div>

                  <textarea
                    rows={5}
                    value={response}
                    onChange={(e) => {
                      setResponse(e.target.value)
                      setErrorMessage(null)
                    }}
                    placeholder="Tuliskan respon kamu di sini dengan jelas..."
                    className="w-full p-3.5 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic resize-none leading-relaxed"
                  />
                </div>

                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                    <AlertCircle className="w-4 h-4 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                <button
                  type="button"
                  disabled={!isValidLength || submitting}
                  onClick={handleSubmit}
                  className="w-full py-3.5 rounded-xl bg-accent-magic text-white font-bold shadow-sm hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all flex items-center justify-center gap-2 min-h-[44px]"
                >
                  {submitting ? (
                    <span>Mengirim Jawaban...</span>
                  ) : (
                    <>
                      <Sparkles className="w-4 h-4" />
                      <span>Kirim Jawaban Teks</span>
                    </>
                  )}
                </button>
              </motion.div>
            ) : (
              <motion.div
                key="submitted-step"
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="py-6 text-center space-y-5"
              >
                <div className={`w-20 h-20 mx-auto rounded-full ${task.status === 'APPROVED' ? 'bg-status-success/20 text-status-success' : 'bg-accent-gold/20 text-accent-gold'} flex items-center justify-center`}>
                  {task.status === 'APPROVED' ? <CheckCircle2 className="w-12 h-12" /> : <Clock className="w-12 h-12" />}
                </div>
                <div>
                  <h4 className="font-heading font-bold text-2xl text-text-primary">
                    {task.status === 'APPROVED' ? 'Tugas Selesai (Disetujui)! 🎉' : 'Jawaban Berhasil Terkirim! ✍️'}
                  </h4>
                  <p className="text-sm text-text-secondary mt-1">
                    {task.status === 'APPROVED'
                      ? `Tugas telah disetujui dan kamu mendapatkan +${task.coins_earned || task.reward_coins} Koin & +${task.xp_earned || task.reward_xp} EXP.`
                      : 'Respon teks kamu telah masuk ke antrean verifikasi admin. Koin & EXP akan otomatis diberikan setelah disetujui.'}
                  </p>
                </div>

                <div className="p-4 rounded-2xl bg-surface border border-border-subtle text-left flex items-start gap-3">
                  <CheckCircle2 className="w-5 h-5 text-status-success shrink-0 mt-0.5" />
                  <p className="text-xs text-text-secondary leading-relaxed">
                    <strong className="text-text-primary">Tugas berikutnya sudah terbuka!</strong> Kamu bisa langsung mengerjakan tugas berikutnya tanpa harus menunggu review.
                  </p>
                </div>

                <button
                  onClick={() => {
                    if (onNextTask) {
                      onNextTask()
                    } else {
                      onSuccess()
                      onClose()
                    }
                  }}
                  className="w-full py-3.5 rounded-xl bg-accent-magic text-white font-bold shadow-sm hover:brightness-110 active:scale-[0.98] transition-all flex items-center justify-center gap-2 min-h-[44px]"
                >
                  <span>Lanjut Tugas Berikutnya</span>
                  <ArrowRight className="w-5 h-5" />
                </button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  )
}
