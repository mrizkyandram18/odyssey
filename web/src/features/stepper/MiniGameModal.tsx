import React, { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import confetti from 'canvas-confetti'
import { Gamepad2, Award, X, Sparkles, CheckCircle2, RotateCcw, Clock, Target, AlertCircle } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'

interface MiniGameModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

interface Card {
  id: number
  icon: string
  isFlipped: boolean
  isMatched: boolean
}

const CARD_ICONS = ['🚀', '🌟', '💎', '🔥', '🛡️', '⚡', '👑', '🎯']

export const MiniGameModal: React.FC<MiniGameModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const isApproved = task.status === 'APPROVED'
  const targetScore = task.config?.target_score || 80
  const difficulty = task.config?.difficulty || 'MEDIUM'

  // Pair count based on difficulty
  const pairCount = difficulty === 'HARD' ? 8 : difficulty === 'EASY' ? 4 : 6

  const [cards, setCards] = useState<Card[]>([])
  const [flippedCards, setFlippedCards] = useState<number[]>([])
  const [moves, setMoves] = useState(0)
  const [, setMatchedPairs] = useState(0)
  const [gameStarted, setGameStarted] = useState(false)
  const [gameFinished, setGameFinished] = useState(isApproved)
  const [score, setScore] = useState(0)
  const [seconds, setSeconds] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [earnedRewards, setEarnedRewards] = useState<{ coins: number; xp: number } | null>(
    isApproved ? { coins: task.coins_earned || task.reward_coins, xp: task.xp_earned || task.reward_xp } : null
  )

  // Initialize deck
  const initializeGame = useCallback(() => {
    const selectedIcons = CARD_ICONS.slice(0, pairCount)
    const deck = [...selectedIcons, ...selectedIcons]
      .sort(() => Math.random() - 0.5)
      .map((icon, idx) => ({
        id: idx,
        icon,
        isFlipped: false,
        isMatched: false,
      }))

    setCards(deck)
    setFlippedCards([])
    setMoves(0)
    setMatchedPairs(0)
    setGameFinished(false)
    setScore(0)
    setSeconds(0)
    setGameStarted(true)
    setErrorMessage(null)
  }, [pairCount])

  useEffect(() => {
    initializeGame()
  }, [initializeGame])

  // Timer tick
  useEffect(() => {
    let interval: any
    if (gameStarted && !gameFinished) {
      interval = setInterval(() => {
        setSeconds((prev) => prev + 1)
      }, 1000)
    }
    return () => clearInterval(interval)
  }, [gameStarted, gameFinished])

  // Handle Card Click
  const handleCardClick = (id: number) => {
    if (flippedCards.length === 2) return
    const clickedCard = cards.find((c) => c.id === id)
    if (!clickedCard || clickedCard.isFlipped || clickedCard.isMatched) return

    const newCards = cards.map((c) => (c.id === id ? { ...c, isFlipped: true } : c))
    setCards(newCards)

    const newFlipped = [...flippedCards, id]
    setFlippedCards(newFlipped)

    if (newFlipped.length === 2) {
      setMoves((m) => m + 1)
      const first = newCards.find((c) => c.id === newFlipped[0])!
      const second = newCards.find((c) => c.id === newFlipped[1])!

      if (first.icon === second.icon) {
        // Matched!
        setTimeout(() => {
          setCards((prev) =>
            prev.map((c) => (c.id === first.id || c.id === second.id ? { ...c, isMatched: true } : c))
          )
          setFlippedCards([])
          setMatchedPairs((mp) => {
            const nextMP = mp + 1
            if (nextMP === pairCount) {
              // Game Won! Calculate bounded score (max 100)
              const timePenalty = Math.min(30, seconds)
              const movesPenalty = Math.max(0, (moves + 1 - pairCount) * 3)
              const finalScore = Math.max(50, Math.min(100, 100 - movesPenalty - Math.floor(timePenalty / 2)))
              setScore(finalScore)
              setGameFinished(true)
            }
            return nextMP
          })
        }, 400)
      } else {
        // Not matched, flip back after brief pause
        setTimeout(() => {
          setCards((prev) =>
            prev.map((c) => (c.id === first.id || c.id === second.id ? { ...c, isFlipped: false } : c))
          )
          setFlippedCards([])
        }, 900)
      }
    }
  }

  // Submit Game Result
  const handleSubmit = async () => {
    setSubmitting(true)
    setErrorMessage(null)
    try {
      const res = await tasksApi.submit(task.id, {
        answers: {
          score,
          moves,
          time_seconds: seconds,
          game: 'MEMORY',
        },
      })

      if (res.success) {
        setEarnedRewards({
          coins: res.coins_earned || task.reward_coins,
          xp: res.xp_earned || task.reward_xp,
        })
        confetti({
          particleCount: 100,
          spread: 70,
          origin: { y: 0.6 },
        })
      } else {
        setErrorMessage(res.error || 'Skor belum mencukupi atau hasil gagal divalidasi.')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengirim skor game. Silakan coba lagi.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.95, opacity: 0, y: 20 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.95, opacity: 0, y: 20 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border-subtle bg-surface">
          <div className="flex items-center gap-2">
            <span className="w-8 h-8 rounded-full bg-accent-magic/20 text-accent-magic flex items-center justify-center font-bold text-sm">
              #{task.step_order}
            </span>
            <h3 className="font-heading font-bold text-text-primary text-base md:text-lg line-clamp-1">
              {task.title}
            </h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup"
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Game Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-5">
          <AnimatePresence mode="wait">
            {!earnedRewards ? (
              <motion.div key="game-panel" initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-4">
                {/* Stats Bar */}
                <div className="flex items-center justify-between p-3 rounded-2xl bg-surface border border-border-subtle text-xs font-bold text-text-secondary">
                  <div className="flex items-center gap-1.5 text-accent-magic">
                    <Gamepad2 className="w-4 h-4" />
                    <span>Langkah: {moves}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-accent-gold">
                    <Clock className="w-4 h-4" />
                    <span>{seconds}s</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-text-primary">
                    <Target className="w-4 h-4" />
                    <span>Target: {targetScore} Poin</span>
                  </div>
                </div>

                {task.description && (
                  <p className="text-xs text-text-secondary leading-relaxed bg-surface p-3 rounded-xl">
                    {task.description}
                  </p>
                )}

                {/* Cards Grid */}
                <div
                  className={`grid gap-2.5 mx-auto max-w-sm ${
                    pairCount === 8 ? 'grid-cols-4' : pairCount === 6 ? 'grid-cols-4' : 'grid-cols-4'
                  }`}
                >
                  {cards.map((card) => {
                    const isRevealed = card.isFlipped || card.isMatched
                    return (
                      <motion.button
                        key={card.id}
                        type="button"
                        whileHover={!isRevealed ? { scale: 1.05 } : {}}
                        whileTap={!isRevealed ? { scale: 0.95 } : {}}
                        onClick={() => handleCardClick(card.id)}
                        disabled={isRevealed || gameFinished}
                        className={`aspect-square rounded-2xl text-2xl flex items-center justify-center font-bold transition-all shadow-sm ${
                          card.isMatched
                            ? 'bg-status-success/20 border-2 border-status-success text-status-success ring-2 ring-status-success/30'
                            : card.isFlipped
                            ? 'bg-accent-magic text-white border-2 border-white dark:border-surface-elevated ring-2 ring-accent-magic/50'
                            : 'bg-surface border border-border-subtle hover:border-accent-magic/50 text-transparent select-none cursor-pointer'
                        }`}
                      >
                        {isRevealed ? card.icon : '❓'}
                      </motion.button>
                    )
                  })}
                </div>

                {/* Game Finished Screen */}
                {gameFinished && (
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="p-4 rounded-2xl bg-accent-magic/10 border border-accent-magic/30 text-center space-y-2"
                  >
                    <div className="text-2xl">🎉</div>
                    <h4 className="font-heading font-bold text-text-primary text-base">
                      Tantangan Selesai!
                    </h4>
                    <p className="text-xs text-text-secondary">
                      Skor Akhir:{' '}
                      <strong className="text-accent-magic font-extrabold text-sm">{score} Poin</strong>{' '}
                      (Selesai dalam {moves} langkah, {seconds} detik)
                    </p>
                  </motion.div>
                )}

                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                    <AlertCircle className="w-4 h-4 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                {/* Controls */}
                <div className="flex gap-2 pt-2">
                  <button
                    type="button"
                    onClick={initializeGame}
                    className="p-3 rounded-2xl border border-border-subtle bg-surface text-text-secondary hover:text-text-primary font-bold text-xs flex items-center gap-1.5 transition-colors"
                  >
                    <RotateCcw className="w-4 h-4" />
                    <span>Ulang</span>
                  </button>

                  <button
                    type="button"
                    disabled={!gameFinished || submitting}
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
              <motion.div
                key="reward-panel"
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="py-6 text-center space-y-5"
              >
                <div className="w-20 h-20 mx-auto rounded-full bg-status-success/20 text-status-success flex items-center justify-center">
                  <CheckCircle2 className="w-12 h-12" />
                </div>
                <div>
                  <h4 className="font-heading font-bold text-2xl text-text-primary">
                    Luar Biasa! Tantangan Berhasil 🏆
                  </h4>
                  <p className="text-sm text-text-secondary mt-1">
                    Kamu berhasil menyelesaikan game memori dan mendapatkan reward koin.
                  </p>
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
                  <CheckCircle2 className="w-5 h-5" />
                </button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  )
}
