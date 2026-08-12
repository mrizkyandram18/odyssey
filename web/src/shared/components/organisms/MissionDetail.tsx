import { useState } from 'react'
import { Link } from 'react-router-dom'
import confetti from 'canvas-confetti'
import type {
  Exercise,
  MissionWithChallenges,
  CompleteChallengeResult,
  CrewMember,
} from '../../types'
import { Card } from '../atoms/Card'
import { Button } from '../atoms/Button'
import { YourTurnBadge } from '../molecules/YourTurnBadge'
import { RelayRotation } from '../molecules/RelayRotation'
import { memberName } from '../../utils/relayRotation'

export interface MissionDetailProps {
  Mission: MissionWithChallenges
  exercises: Exercise[]
  members?: CrewMember[]
  myUID?: string | null
  onStartMission?: () => Promise<void>
  onCompleteChallenge?: (
    exerciseId: number,
    payload?: { answer?: string; content?: string }
  ) => Promise<CompleteChallengeResult | null>
  isMyTurn?: boolean
}

const StepIndicator = ({ activeStep }: { activeStep: number }) => (
  <div className="flex items-center justify-between w-full mb-8 relative px-4">
    <div className="absolute left-4 right-4 top-1/2 -translate-y-1/2 h-1 bg-border-subtle -z-10"></div>
    {[
      { num: 1, label: 'Belajar' },
      { num: 2, label: 'Latihan' },
      { num: 3, label: 'Hasil' },
      { num: 4, label: 'Reward' }
    ].map(s => {
      const isActive = activeStep === s.num;
      const isPast = activeStep > s.num;
      return (
        <div key={s.num} className="flex flex-col items-center gap-2 bg-surface px-2">
          <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm border-2 ${isActive ? 'bg-accent-reward border-accent-reward text-white shadow-md' : isPast ? 'bg-accent-nature border-accent-nature text-white' : 'bg-surface border-border-subtle text-text-secondary'}`}>
            {isPast ? '✓' : s.num}
          </div>
          <span className={`text-xs font-semibold ${isActive ? 'text-accent-reward' : isPast ? 'text-accent-nature' : 'text-text-secondary'}`}>
            {s.label}
          </span>
        </div>
      )
    })}
  </div>
)

export function MissionDetail({
  Mission,
  exercises,
  members,
  myUID,
  onStartMission,
  onCompleteChallenge,
  isMyTurn,
}: MissionDetailProps) {
  const [starting, setStarting] = useState(false)
  const [completingId, setCompletingId] = useState<number | null>(null)
  const [lastResult, setLastResult] = useState<CompleteChallengeResult | null>(null)
  const [inputs, setInputs] = useState<Record<number, string>>({})
  const [activeStep, setActiveStep] = useState<number>(Mission.status === 'DONE' ? 4 : 1)

  // Move to next step
  const handleNextStep = () => {
    setActiveStep(prev => prev + 1)
  }

  const handleStart = async () => {
    if (!onStartMission) return
    setStarting(true)
    try {
      await onStartMission()
    } catch (e) {
      console.error(e)
    } finally {
      setStarting(false)
    }
  }

  const handleComplete = async (exerciseId: number, type?: string) => {
    if (!onCompleteChallenge) return
    setCompletingId(exerciseId)
    const val = inputs[exerciseId]?.trim() || ''

    try {
      const payload = type === 'PUZZLE' ? { answer: val } : type === 'RESEARCH' ? { content: val } : undefined
      const res = await onCompleteChallenge(exerciseId, payload)
      if (res) {
        setLastResult(res)
        confetti({
          particleCount: 100,
          spread: 70,
          origin: { y: 0.6 },
          colors: ['#06b6de', '#10b981', '#f59e0b'],
        })
      }
    } catch (e) {
      console.error(e)
    } finally {
      setCompletingId(null)
    }
  }

  return (
    <div className="flex flex-col gap-6 max-w-4xl mx-auto">
      {/* Mission Header */}
      <Card className="relative overflow-hidden p-8 md:p-12 border-0 bg-surface-elevated/80">
        <div className="absolute inset-0 bg-gradient-to-br from-accent-magic/20 via-transparent to-transparent opacity-50"></div>
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-4">
            <span className="text-3xl">📜</span>
            <span className="text-accent-magic font-bold tracking-widest uppercase text-sm">
              {Mission.template_slug.replace(/-/g, ' ')}
            </span>
          </div>
          <h1 className="font-heading text-4xl md:text-5xl text-text-primary mb-4 leading-tight">
            {Mission.title}
          </h1>

          <div className="flex items-center gap-3 flex-wrap">
            {isMyTurn && <YourTurnBadge />}
            {Mission.status === 'ACTIVE' && (
              <span className="bg-accent-magic/20 text-accent-magic font-bold px-3 py-1 rounded border border-accent-magic/30">
                AKTIF
              </span>
            )}
            {Mission.status === 'DONE' && (
              <span className="bg-accent-nature/20 text-accent-nature font-bold px-3 py-1 rounded border border-accent-nature/30">
                SELESAI
              </span>
            )}
            {Mission.status === 'PENDING' && (
              <span className="bg-surface border border-border-subtle text-text-secondary font-bold px-3 py-1 rounded">
                MENUNGGU
              </span>
            )}
            {Mission.Mission_type && (
              <span className="bg-surface border border-border-subtle text-text-secondary font-bold px-3 py-1 rounded uppercase tracking-wider text-xs">
                {Mission.Mission_type}
              </span>
            )}
          </div>
        </div>
      </Card>

      {/* Step Indicator */}
      {Mission.status !== 'PENDING' && <StepIndicator activeStep={activeStep} />}

      {/* Start Button & Pending State */}
      {Mission.status === 'PENDING' && onStartMission && (
        <Card className="text-center p-8 bg-surface-elevated">
          <p className="text-text-secondary mb-6">Misi ini belum dimulai. Siap belajar?</p>
          <Button
            size="lg"
            isLoading={starting}
            onClick={handleStart}
            className="w-full shadow-lg shadow-accent-reward/20 text-lg"
          >
            Mulai Belajar
          </Button>
        </Card>
      )}

      {/* Relay rotation panel */}
      {Mission.Mission_type === 'RELAY' && (
        <RelayRotation Mission={Mission} exercises={exercises} members={members} myUID={myUID} />
      )}

      {/* Step 1: Learn Section */}
      {activeStep === 1 && Mission.status !== 'PENDING' && (
        <Card className="p-6 bg-surface-elevated flex flex-col gap-4 animate-in fade-in slide-in-from-right-4">
          <div className="flex items-center gap-3 mb-2">
            <h2 className="font-heading text-xl font-bold text-text-primary uppercase">Materi Belajar</h2>
          </div>
          <p className="text-base text-text-primary leading-relaxed whitespace-pre-wrap bg-surface p-5 rounded-xl border border-border-subtle">
            {Mission.learn_text || 'Tidak ada materi spesifik. Silakan lanjut ke latihan.'}
          </p>
          <Button onClick={handleNextStep} size="lg" className="mt-4">
            Lanjut ke Latihan
          </Button>
        </Card>
      )}

      {/* Step 2: Exercises List (Practice) */}
      {activeStep === 2 && (
        <div className="animate-in fade-in slide-in-from-right-4">
          <div className="flex flex-col gap-4">
            <h2 className="font-heading text-xl font-bold text-text-primary uppercase mb-2">Latihan</h2>
        {exercises.length === 0 ? (
          <p className="text-text-secondary italic">
            Belum ada latihan yang tersedia.
          </p>
        ) : (
          <div className="flex flex-col gap-4 relative">
            <div className="absolute left-6 top-6 bottom-6 w-0.5 bg-border-subtle"></div>

            {exercises.map((c, index) => {
              const isDone = c.status === 'DONE'
              const isCompleting = completingId === c.id
              const challengeType = (c as any).type || (c as any).challenge_type || ''

              return (
                <Card
                  key={c.id}
                  className={`flex flex-col gap-4 p-6 relative z-10 transition-all ${
                    isDone
                      ? 'opacity-70 bg-surface-elevated/50'
                      : 'bg-surface border-border-subtle hover:border-accent-magic/50'
                  }`}
                >
                  <div className="flex gap-4 items-start flex-1">
                    <div
                      className={`w-12 h-12 rounded-full flex items-center justify-center shrink-0 border-2 ${
                        isDone
                          ? 'bg-accent-nature border-accent-nature text-bg-app'
                          : 'bg-bg-app border-border-subtle text-text-secondary'
                      }`}
                    >
                      {isDone ? '✓' : index + 1}
                    </div>
                    <div className="flex-1">
                      <p
                        className={`text-base font-medium mb-1 ${
                          isDone ? 'text-text-secondary' : 'text-text-primary'
                        }`}
                      >
                        {c.description}
                      </p>

                      {/* Render question if available */}
                      {c.question && !isDone && (
                        <p className="text-sm font-medium text-text-primary mt-2 mb-3">
                          Q: {c.question}
                        </p>
                      )}

                      {isDone && c.completed_by && (
                        <p className="text-xs text-text-secondary mt-1">
                          Diselesaikan oleh {memberName(members, c.completed_by)}
                        </p>
                      )}

                      {/* Interactive Inputs for PUZZLE / RESEARCH / MCQ / TRUE_FALSE */}
                      {!isDone && Mission.status === 'ACTIVE' && (
                        <div className="mt-3 w-full">
                          {challengeType === 'PUZZLE' && (
                            <input
                              type="text"
                              placeholder="Ketik jawaban teka-teki..."
                              value={inputs[c.id] || ''}
                              onChange={(e) =>
                                setInputs({ ...inputs, [c.id]: e.target.value })
                              }
                              className="w-full text-sm p-3 rounded bg-surface-elevated border border-border-subtle text-text-primary focus:border-accent-magic outline-none"
                            />
                          )}
                          {challengeType === 'RESEARCH' && (
                            <textarea
                              placeholder="Bagikan fakta hasil riset kamu..."
                              value={inputs[c.id] || ''}
                              onChange={(e) =>
                                setInputs({ ...inputs, [c.id]: e.target.value })
                              }
                              rows={3}
                              className="w-full text-sm p-3 rounded bg-surface-elevated border border-border-subtle text-text-primary focus:border-accent-magic outline-none resize-none"
                            />
                          )}
                          {(challengeType === 'MCQ' || challengeType === 'TRUE_FALSE') && c.options && (
                            <div className="flex flex-col gap-2 mt-2">
                              {c.options.map((opt) => (
                                <label
                                  key={opt}
                                  className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                                    inputs[c.id] === opt
                                      ? 'bg-accent-magic/10 border-accent-magic text-text-primary'
                                      : 'bg-surface-elevated border-border-subtle text-text-secondary hover:border-accent-magic/50'
                                  }`}
                                >
                                  <input
                                    type="radio"
                                    name={`challenge-${c.id}`}
                                    value={opt}
                                    checked={inputs[c.id] === opt}
                                    onChange={(e) =>
                                      setInputs({ ...inputs, [c.id]: e.target.value })
                                    }
                                    className="hidden"
                                  />
                                  <div
                                    className={`w-4 h-4 rounded-full border flex items-center justify-center ${
                                      inputs[c.id] === opt
                                        ? 'border-accent-magic bg-accent-magic'
                                        : 'border-text-secondary'
                                    }`}
                                  >
                                    {inputs[c.id] === opt && (
                                      <div className="w-1.5 h-1.5 rounded-full bg-bg-app" />
                                    )}
                                  </div>
                                  <span className="text-sm font-medium">{opt}</span>
                                </label>
                              ))}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </div>

                  {!isDone && Mission.status === 'ACTIVE' && onCompleteChallenge && (
                    <div className="flex justify-end mt-2">
                      <Button
                        variant="secondary"
                        isLoading={isCompleting}
                        onClick={() => handleComplete(c.id, challengeType)}
                        disabled={(challengeType === 'MCQ' || challengeType === 'TRUE_FALSE') && !inputs[c.id]}
                        className="w-full sm:w-auto bg-gradient-to-r from-accent-magic/20 to-transparent border-accent-magic/50 hover:border-accent-magic"
                      >
                        {challengeType === 'MOVEMENT'
                          ? '🚶 Konfirmasi Selesai'
                          : '✓ Kirim Jawaban'}
                      </Button>
                    </div>
                  )}
                </Card>
              )
            })}
          </div>
        )}
        
        {exercises.length > 0 && exercises.every(c => c.status === 'DONE') && (
           <Button onClick={handleNextStep} size="lg" className="w-full mt-6">
             Lihat Hasil
           </Button>
        )}
          </div>
        </div>
      )}


      {/* Step 3: RESULT Section */}
      {activeStep === 3 && (
        <Card className="bg-surface-elevated p-6 animate-in fade-in slide-in-from-right-4">
          <div className="flex items-center gap-3 mb-4">
            <h2 className="font-heading text-xl font-bold text-text-primary uppercase">HASIL & PENJELASAN</h2>
          </div>
          <p className="text-base text-text-primary mb-6 p-5 bg-surface rounded-xl border border-border-subtle">
            {Mission.result_text || 'Kerja bagus! Latihan berhasil diselesaikan.'}
          </p>
          <Button onClick={handleNextStep} size="lg" className="w-full">
             Ambil Reward
          </Button>
        </Card>
      )}

      {/* Step 4: REWARD Section */}
      {activeStep === 4 && (
        <Card className="border-accent-reward bg-accent-reward/5 p-6 text-center shadow-lg animate-in fade-in zoom-in-95">
          <span className="block text-5xl mb-4">✨</span>
          <h3 className="font-heading text-2xl font-bold text-accent-reward mb-2">Petualangan Selesai!</h3>
          <p className="text-text-secondary mb-6">Kamu telah menyelesaikan materi ini dengan baik.</p>
          
          <div className="flex justify-center gap-6 mb-8">
            <div className="text-center bg-surface px-6 py-4 rounded-xl border border-border-subtle shadow-sm">
              <span className="block text-2xl mb-1">⭐</span>
              <span className="font-bold text-text-primary text-lg">+{lastResult?.xp || 100} Poin</span>
            </div>
          </div>

          <div className="flex flex-col items-center gap-4">
            <p className="text-sm font-bold text-text-primary">
              🏆 Cek beranda untuk melanjutkan perjalananmu!
            </p>
            <Link to="/" className="w-full">
              <Button variant="primary" size="lg" className="w-full">Kembali ke Beranda</Button>
            </Link>
          </div>
        </Card>
      )}
    </div>
  )
}
