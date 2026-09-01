import React, { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import confetti from 'canvas-confetti'
import { Play, CheckCircle2, XCircle, Award, Sparkles, ChevronRight, HelpCircle, X } from 'lucide-react'
import type { TaskView, QuizQuestion } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'

interface VideoQuizModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

export const VideoQuizModal: React.FC<VideoQuizModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const rawQuestions = task.config?.questions || []
  const questions: QuizQuestion[] = Array.isArray(rawQuestions) ? rawQuestions : []
  const youtubeUrl = task.config?.youtube_url || ''

  // Format YouTube embed URL if valid
  const getEmbedUrl = (url: string) => {
    if (!url) return ''
    let videoId = ''
    if (url.includes('v=')) {
      videoId = url.split('v=')[1]?.split('&')[0]
    } else if (url.includes('youtu.be/')) {
      videoId = url.split('youtu.be/')[1]?.split('?')[0]
    } else if (url.includes('embed/')) {
      return url
    }
    return videoId ? `https://www.youtube.com/embed/${videoId}?rel=0&autoplay=0` : url
  }

  const embedUrl = getEmbedUrl(youtubeUrl)

  const isApproved = task.status === 'APPROVED'

  const [currentStep, setCurrentStep] = useState<'video' | 'quiz' | 'result'>(
    isApproved ? 'result' : embedUrl ? 'video' : 'quiz'
  )
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [earnedRewards, setEarnedRewards] = useState<{ coins: number; xp: number } | null>(null)

  const handleSelectOption = (questionId: string | number, option: string) => {
    const match = option.match(/^([A-Za-z])[.)]\s*(.*)/)
    const normalizedVal = match ? match[1].toUpperCase() : option
    setAnswers((prev) => ({
      ...prev,
      [String(questionId)]: normalizedVal,
    }))
    setErrorMessage(null)
  }

  const allQuestionsAnswered =
    questions.length > 0 &&
    questions.every((q) => !!answers[String(q.id)])

  const handleSubmit = async () => {
    setSubmitting(true)
    setErrorMessage(null)
    try {
      const res = await tasksApi.submit(task.id, { answers })
      if (res.success) {
        setEarnedRewards({
          coins: res.coins_earned || task.reward_coins,
          xp: res.xp_earned || task.reward_xp,
        })
        setCurrentStep('result')
        confetti({
          particleCount: 80,
          spread: 70,
          origin: { y: 0.6 },
        })
      } else {
        setErrorMessage(res.error || 'Jawaban kuis belum tepat. Periksa kembali jawabanmu.')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengirim jawaban kuis. Coba lagi.')
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
        {/* Header — compact unified */}
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
            {/* Step 1: Video Player */}
            {currentStep === 'video' && (
              <motion.div
                key="video-step"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                className="space-y-4"
              >
                <div className="flex items-center gap-2 text-sm text-text-secondary">
                  <Play className="w-4 h-4 text-accent-magic" />
                  <span>Tonton video panduan di bawah ini sampai selesai:</span>
                </div>

                {embedUrl ? (
                  <div className="relative aspect-video rounded-2xl overflow-hidden shadow-inner border border-border-subtle bg-black">
                    <iframe
                      src={embedUrl}
                      title={task.title}
                      className="w-full h-full"
                      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                      allowFullScreen
                    />
                  </div>
                ) : (
                  <div className="p-8 text-center text-text-secondary bg-surface rounded-2xl border border-dashed border-border-subtle">
                    Tidak ada tautan video. Kamu bisa langsung lanjut ke kuis.
                  </div>
                )}

                {task.description && (
                  <p className="text-sm text-text-secondary leading-relaxed bg-surface p-4 rounded-xl">
                    {task.description}
                  </p>
                )}

                <button
                  onClick={() => setCurrentStep('quiz')}
                  className="w-full py-3.5 rounded-xl bg-accent-magic text-white font-bold shadow-sm hover:brightness-110 active:scale-[0.98] transition-all flex items-center justify-center gap-2 min-h-[44px]"
                >
                  <span>Mulai Kuis Tugas</span>
                  <ChevronRight className="w-5 h-5" />
                </button>
              </motion.div>
            )}

            {/* Step 2: Interactive Quiz */}
            {currentStep === 'quiz' && (
              <motion.div
                key="quiz-step"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                className="space-y-6"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-sm text-accent-magic font-medium">
                    <HelpCircle className="w-4 h-4" />
                    <span>Jawab {questions.length} Pertanyaan Kuis</span>
                  </div>
                  <div className="text-xs bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">
                    +{task.reward_coins} 🪙 | +{task.reward_xp} XP
                  </div>
                </div>

                {questions.length === 0 ? (
                  <div className="text-center py-8 text-text-secondary">
                    Tidak ada pertanyaan kuis untuk tugas ini.
                  </div>
                ) : (
                  questions.map((q, idx) => (
                    <div
                      key={q.id || idx}
                      className="p-3.5 rounded-xl bg-surface border border-border-subtle space-y-3"
                    >
                      <p className="font-heading font-bold text-text-primary text-sm md:text-base">
                        {idx + 1}. {q.question}
                      </p>
                      <div className="space-y-2">
                        {q.options.map((opt) => {
                          const match = opt.match(/^([A-Za-z])[.)]\s*(.*)/)
                          const letter = match ? match[1].toUpperCase() : null
                          const isSelected = answers[String(q.id)] === opt || (letter !== null && answers[String(q.id)] === letter)
                          return (
                            <button
                              key={opt}
                              type="button"
                              onClick={() => handleSelectOption(q.id, opt)}
                              className={`w-full text-left p-3.5 rounded-xl border text-sm font-medium transition-all ${
                                isSelected
                                  ? 'border-accent-magic bg-accent-magic/15 text-text-primary shadow-sm ring-2 ring-accent-magic/30'
                                  : 'border-border-subtle hover:border-accent-magic/40 bg-surface-elevated text-text-secondary hover:text-text-primary'
                              }`}
                            >
                              {opt}
                            </button>
                          )
                        })}
                      </div>
                    </div>
                  ))
                )}

                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2">
                    <XCircle className="w-5 h-5 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                <div className="flex gap-3 pt-2">
                  {embedUrl && (
                    <button
                      type="button"
                      onClick={() => setCurrentStep('video')}
                      className="px-4 py-3.5 rounded-2xl border border-border-subtle text-text-secondary hover:text-text-primary font-bold text-sm"
                    >
                      Tonton Lagi
                    </button>
                  )}
                  <button
                    type="button"
                    disabled={!allQuestionsAnswered || submitting}
                    onClick={handleSubmit}
                    className="flex-1 py-3.5 rounded-xl bg-accent-magic text-white font-bold shadow-sm hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all flex items-center justify-center gap-2 min-h-[44px]"
                  >
                    {submitting ? (
                      <span>Memeriksa Jawaban...</span>
                    ) : (
                      <>
                        <Sparkles className="w-4 h-4" />
                        <span>Kirim Jawaban</span>
                      </>
                    )}
                  </button>
                </div>
              </motion.div>
            )}

            {/* Step 3: Success Result */}
            {currentStep === 'result' && (
              <motion.div
                key="result-step"
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="py-6 text-center space-y-5"
              >
                <div className="w-20 h-20 mx-auto rounded-full bg-status-success/20 text-status-success flex items-center justify-center">
                  <CheckCircle2 className="w-12 h-12" />
                </div>
                <div>
                  <h4 className="font-heading font-bold text-2xl text-text-primary">
                    Hebat! Tugas Selesai 🎉
                  </h4>
                  <p className="text-sm text-text-secondary mt-1">
                    {isApproved
                      ? 'Tugas ini sudah kamu selesaikan dan reward telah diberikan.'
                      : 'Semua jawaban kuis kamu benar dan telah diverifikasi.'}
                  </p>
                </div>

                <div className="inline-flex items-center gap-4 px-6 py-3 rounded-2xl bg-accent-gold/15 border border-accent-gold/30">
                  <div className="flex items-center gap-1.5 text-accent-gold font-bold">
                    <Award className="w-5 h-5" />
                    <span>+{earnedRewards?.coins || task.coins_earned || task.reward_coins} Koin</span>
                  </div>
                  <div className="w-px h-5 bg-border-subtle" />
                  <div className="flex items-center gap-1.5 text-accent-magic font-bold">
                    <Sparkles className="w-5 h-5" />
                    <span>+{earnedRewards?.xp || task.xp_earned || task.reward_xp} EXP</span>
                  </div>
                </div>

                <div className="space-y-2 pt-2">
                  <button
                    onClick={() => {
                      if (onNextTask) {
                        onNextTask()
                      } else {
                        onSuccess()
                        onClose()
                      }
                    }}
                    className="w-full py-4 rounded-2xl bg-accent-magic text-white font-heading font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 active:scale-[0.98] transition-all flex items-center justify-center gap-2"
                  >
                    <span>Lanjut ke Tugas Berikutnya</span>
                    <ChevronRight className="w-5 h-5" />
                  </button>

                  {embedUrl && (
                    <button
                      type="button"
                      onClick={() => setCurrentStep('video')}
                      className="w-full py-2.5 rounded-xl text-text-secondary hover:text-text-primary text-xs font-bold transition-colors"
                    >
                      Tonton Ulang Video
                    </button>
                  )}
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  )
}
