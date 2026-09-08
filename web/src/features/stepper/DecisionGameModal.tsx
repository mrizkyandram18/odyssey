import React, { useState, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import confetti from 'canvas-confetti'
import { Wallet, CheckCircle2, X, Sparkles, Award, AlertCircle, ChevronRight, RotateCcw, TrendingUp, TrendingDown, Minus } from 'lucide-react'
import type { TaskView, DecisionScenario } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { isEarningCapError, EARNING_CAP_MESSAGE } from '../../shared/lib/earning'
import { TaskRevisionBanner } from './TaskRevisionBanner'

interface DecisionGameModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

export function formatRupiah(n: number): string {
  const sign = n < 0 ? '-' : ''
  return `${sign}Rp${Math.abs(n).toLocaleString('id-ID')}`
}

export const DecisionGameModal: React.FC<DecisionGameModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const scenario = (task.config?.scenario || null) as DecisionScenario | null
  const events = useMemo(() => (Array.isArray(scenario?.events) ? scenario!.events : []), [scenario])
  const initialBalance = scenario?.initial_balance ?? 500000

  const isApproved = task.status === 'APPROVED'
  const [eventIndex, setEventIndex] = useState(0)
  const [choices, setChoices] = useState<Record<string, string>>({})
  const [finished, setFinished] = useState(isApproved)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [earnedRewards, setEarnedRewards] = useState<{ coins: number; xp: number } | null>(
    isApproved ? { coins: task.coins_earned || task.reward_coins, xp: task.xp_earned || task.reward_xp } : null
  )

  const currentBalance = useMemo(() => {
    let bal = initialBalance
    for (const ev of events) {
      const optId = choices[ev.id]
      const opt = ev.options?.find((o) => String(o.id) === String(optId))
      if (opt) bal += Number(opt.delta) || 0
    }
    return bal
  }, [choices, events, initialBalance])

  const currentEvent = events[eventIndex]
  const progress = events.length === 0 ? 0 : Math.round((Object.keys(choices).length / events.length) * 100)

  const handleChoose = (optionId: string) => {
    if (!currentEvent || finished) return
    const next = { ...choices, [currentEvent.id]: optionId }
    setChoices(next)
    setErrorMessage(null)
    if (eventIndex + 1 >= events.length) {
      setFinished(true)
    } else {
      setEventIndex(eventIndex + 1)
    }
  }

  const handleRestart = () => {
    setChoices({})
    setEventIndex(0)
    setFinished(false)
    setErrorMessage(null)
  }

  const handleSubmit = async () => {
    setSubmitting(true)
    setErrorMessage(null)
    try {
      const res = await tasksApi.submit(task.id, {
        answers: {
          game: 'DECISION_FINANCE',
          choices,
          // Informational only — server recomputes final_balance authoritatively.
          final_balance: currentBalance,
          score: 100,
        },
      })
      if (res.success) {
        setEarnedRewards({
          coins: res.coins_earned || task.reward_coins,
          xp: res.xp_earned || task.reward_xp,
        })
        confetti({ particleCount: 100, spread: 70, origin: { y: 0.6 } })
      } else {
        setErrorMessage(res.error || 'Hasil permainan belum valid. Coba lagi.')
      }
    } catch (err: any) {
      if (isEarningCapError(err)) {
        setErrorMessage(EARNING_CAP_MESSAGE)
        try { onSuccess() } catch { /* ignore */ }
      } else {
        setErrorMessage(err.message || 'Gagal mengirim hasil. Silakan coba lagi.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  if (!scenario || events.length === 0) {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
        <div className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl p-6 text-center">
          <AlertCircle className="w-8 h-8 mx-auto text-status-error" />
          <p className="text-sm text-text-primary font-bold mt-2">Skenario permainan belum tersedia.</p>
          <button onClick={onClose} className="mt-4 px-4 py-2 rounded-xl bg-surface border border-border-subtle text-text-primary font-bold text-xs">Tutup</button>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.97, opacity: 0, y: 12 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.97, opacity: 0, y: 12 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[92vh] sm:max-h-[90vh]"
      >
        <div className="flex items-center justify-between px-5 py-3.5 border-b border-border-subtle bg-surface shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            <span className="w-7 h-7 rounded-full bg-accent-magic/12 text-accent-magic flex items-center justify-center font-bold text-xs shrink-0">
              #{task.step_order}
            </span>
            <h3 className="font-bold text-text-primary text-[14px] line-clamp-1">{task.title}</h3>
          </div>
          <button onClick={onClose} aria-label="Tutup" className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors shrink-0">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          <AnimatePresence mode="wait">
            {!earnedRewards ? (
              <motion.div key="decision-panel" initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-4">
                {task.status === 'REJECTED' && eventIndex === 0 && (
                  <TaskRevisionBanner adminNotes={task.admin_notes} />
                )}
                <div className="flex items-center justify-between p-3 rounded-2xl bg-surface border border-border-subtle">
                  <div>
                    <span className="text-[10px] uppercase tracking-wider font-extrabold text-accent-magic block">
                      Saldo Simulasi (Uang Virtual)
                    </span>
                    <div className="flex items-center gap-1.5 text-text-primary font-heading font-bold text-base mt-0.5">
                      <Wallet className="w-4 h-4 text-accent-magic" />
                      <span data-testid="current-balance">{formatRupiah(currentBalance)}</span>
                    </div>
                  </div>
                  <div className="text-right">
                    <span className="text-[10px] uppercase tracking-wider font-extrabold text-text-secondary block">
                      Progres
                    </span>
                    <span className="text-xs font-bold text-text-secondary mt-0.5 block" data-testid="progress">
                      {Object.keys(choices).length}/{events.length} situasi
                    </span>
                  </div>
                </div>
                <div className="h-2 w-full bg-surface rounded-full overflow-hidden border border-border-subtle/50">
                  <div className="h-full bg-gradient-to-r from-accent-magic to-sky-400 rounded-full transition-all duration-300" style={{ width: `${progress}%` }} />
                </div>

                {task.description && eventIndex === 0 && !finished && (
                  <p className="text-xs text-text-secondary leading-relaxed bg-surface p-3 rounded-xl border border-border-subtle">{task.description}</p>
                )}

                {!finished && currentEvent ? (
                  <motion.div key={currentEvent.id} initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-3">
                    <div>
                      <p className="text-[11px] font-black uppercase tracking-wider text-accent-magic">Situasi {eventIndex + 1} dari {events.length}</p>
                      <h4 className="font-heading font-bold text-text-primary text-base mt-1">{currentEvent.title}</h4>
                      {currentEvent.description && (
                        <p className="text-xs text-text-secondary mt-1 leading-relaxed">{currentEvent.description}</p>
                      )}
                    </div>
                    <div className="space-y-2">
                      {currentEvent.options?.map((opt) => {
                        const delta = Number(opt.delta) || 0
                        const displayLabel = String(opt.label || '').replace(/\s*\([+-]?Rp[0-9.]+\)\s*$/i, '')
                        return (
                          <button
                            key={opt.id}
                            type="button"
                            data-testid={`option-${currentEvent.id}-${opt.id}`}
                            onClick={() => handleChoose(String(opt.id))}
                            className="w-full text-left p-3.5 rounded-xl border border-border-subtle hover:border-accent-magic/50 bg-surface-elevated text-text-primary transition-all active:scale-[0.99]"
                          >
                            <span className="flex items-center justify-between gap-2">
                              <span className="text-sm font-bold">{displayLabel}</span>
                              <span className={`inline-flex items-center gap-1 text-xs font-extrabold shrink-0 ${delta < 0 ? 'text-status-error' : delta > 0 ? 'text-status-success' : 'text-text-secondary'}`}>
                                {delta < 0 ? <TrendingDown className="w-3.5 h-3.5" /> : delta > 0 ? <TrendingUp className="w-3.5 h-3.5" /> : <Minus className="w-3.5 h-3.5" />}
                                {delta === 0 ? 'Rp0' : `${delta > 0 ? '+' : ''}${formatRupiah(delta)}`}
                              </span>
                            </span>
                            {opt.hint && <span className="block text-[11px] text-text-secondary mt-1">{opt.hint}</span>}
                          </button>
                        )
                      })}
                    </div>
                  </motion.div>
                ) : (
                  <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="p-4 rounded-2xl bg-accent-magic/10 border border-accent-magic/30 text-center space-y-2">
                    <div className="text-2xl">🎉</div>
                    <h4 className="font-heading font-bold text-text-primary text-base">Simulasi Selesai!</h4>
                    <p className="text-xs text-text-secondary">
                      Saldo akhir kamu:{' '}
                      <strong className="text-accent-magic font-extrabold text-sm" data-testid="final-balance">{formatRupiah(currentBalance)}</strong>{' '}
                      (mulai dari {formatRupiah(initialBalance)})
                    </p>
                    <p className="text-[11px] text-text-secondary leading-relaxed">
                      {currentBalance >= initialBalance
                        ? 'Hebat! Kamu berhasil menjaga bahkan menambah uangmu. Pertahankan kebiasaan baik ini.'
                        : currentBalance >= initialBalance * 0.5
                        ? 'Lumayan! Uangmu berkurang tapi masih aman. Coba pikirkan pilihan mana yang bisa lebih hemat.'
                        : 'Wah, uangmu banyak berkurang. Tidak apa-apa — ini simulasi! Pelajarannya: selalu sisakan dana darurat.'}
                    </p>
                  </motion.div>
                )}

                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                    <AlertCircle className="w-4 h-4 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                <div className="flex gap-2 pt-2">
                  <button
                    type="button"
                    onClick={handleRestart}
                    className="p-3 rounded-2xl border border-border-subtle bg-surface text-text-secondary hover:text-text-primary font-bold text-xs flex items-center gap-1.5 transition-colors"
                  >
                    <RotateCcw className="w-4 h-4" />
                    <span>Ulang</span>
                  </button>
                  <button
                    type="button"
                    disabled={!finished || submitting}
                    onClick={handleSubmit}
                    className="flex-1 py-3.5 rounded-2xl bg-accent-magic text-white font-heading font-bold text-sm shadow-lg shadow-accent-magic/30 hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all flex items-center justify-center gap-2"
                  >
                    {submitting ? (
                      <span>Memverifikasi Hasil...</span>
                    ) : (
                      <>
                        <Sparkles className="w-4 h-4" />
                        <span>Klaim Reward (+{task.reward_coins}🪙)</span>
                      </>
                    )}
                  </button>
                </div>
              </motion.div>
            ) : (
              <motion.div key="reward-panel" initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="py-6 text-center space-y-5">
                <div className="w-20 h-20 mx-auto rounded-full bg-status-success/20 text-status-success flex items-center justify-center">
                  <CheckCircle2 className="w-12 h-12" />
                </div>
                <div>
                  <h4 className="font-heading font-bold text-2xl text-text-primary">Luar Biasa! Simulasi Berhasil 🏆</h4>
                  <p className="text-sm text-text-secondary mt-1">Kamu menyelesaikan simulasi keuangan dan mendapatkan reward.</p>
                </div>
                <div className="inline-flex items-center gap-4 px-6 py-3 rounded-2xl bg-accent-gold/15 border border-accent-gold/30">
                  <div className="flex items-center gap-1.5 text-accent-gold font-bold">
                    <Award className="w-5 h-5" />
                    <span>+{earnedRewards.coins} Koin</span>
                  </div>
                  <div className="w-px h-5 bg-border-subtle" />
                  <div className="flex items-center gap-1.5 text-accent-magic font-bold">
                    <Sparkles className="w-5 h-5" />
                    <span>+{earnedRewards.xp} EXP</span>
                  </div>
                </div>
                <button
                  onClick={() => {
                    if (onNextTask) onNextTask()
                    else { onSuccess(); onClose() }
                  }}
                  className="w-full py-4 rounded-2xl bg-accent-magic text-white font-heading font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 active:scale-[0.98] transition-all flex items-center justify-center gap-2"
                >
                  <span>Lanjut ke Tugas Berikutnya</span>
                  <ChevronRight className="w-5 h-5" />
                </button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  )
}
