import { useState } from 'react'
import { Link } from 'react-router-dom'
import confetti from 'canvas-confetti'
import type {
  Challenge,
  QuestWithChallenges,
  CompleteChallengeResult,
  CrewMember,
  BranchOption,
} from '../../types'
import { Card } from '../atoms/Card'
import { Button } from '../atoms/Button'
import { YourTurnBadge } from '../molecules/YourTurnBadge'
import { RelayRotation } from '../molecules/RelayRotation'
import { memberName } from '../../utils/relayRotation'
import { SubmissionForm } from '../../../features/creative/SubmissionForm'

export interface QuestDetailProps {
  quest: QuestWithChallenges
  challenges: Challenge[]
  members?: CrewMember[]
  myUID?: string | null
  onStartQuest?: () => Promise<void>
  onCompleteChallenge?: (
    challengeId: number,
    payload?: { answer?: string; content?: string }
  ) => Promise<CompleteChallengeResult | null>
  onSelectBranch?: (branchSlug: string) => Promise<unknown>
  isMyTurn?: boolean
}

export function QuestDetail({
  quest,
  challenges,
  members,
  myUID,
  onStartQuest,
  onCompleteChallenge,
  onSelectBranch,
  isMyTurn,
}: QuestDetailProps) {
  const [starting, setStarting] = useState(false)
  const [completingId, setCompletingId] = useState<number | null>(null)
  const [selectingBranch, setSelectingBranch] = useState<string | null>(null)
  const [lastResult, setLastResult] = useState<CompleteChallengeResult | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [inputs, setInputs] = useState<Record<number, string>>({})

  const handleStart = async () => {
    if (!onStartQuest) return
    setStarting(true)
    setActionError(null)
    try {
      await onStartQuest()
    } catch (e) {
      setActionError(e instanceof Error ? e.message : 'Failed to start quest')
    } finally {
      setStarting(false)
    }
  }

  const handleComplete = async (challengeId: number, type?: string) => {
    if (!onCompleteChallenge) return
    setCompletingId(challengeId)
    setActionError(null)
    const val = inputs[challengeId]?.trim() || ''

    try {
      const payload = type === 'PUZZLE' ? { answer: val } : type === 'RESEARCH' ? { content: val } : undefined
      const res = await onCompleteChallenge(challengeId, payload)
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
      setActionError(e instanceof Error ? e.message : 'Failed to complete challenge')
    } finally {
      setCompletingId(null)
    }
  }

  const handleBranchSelect = async (branchSlug: string) => {
    if (!onSelectBranch) return
    setSelectingBranch(branchSlug)
    setActionError(null)
    try {
      await onSelectBranch(branchSlug)
    } catch (e) {
      setActionError(e instanceof Error ? e.message : 'Failed to select branch')
    } finally {
      setSelectingBranch(null)
    }
  }

  const branchOptions: BranchOption[] = quest.branch_options || []

  return (
    <div className="flex flex-col gap-6 max-w-4xl mx-auto">
      {/* Quest Header */}
      <Card className="relative overflow-hidden p-8 md:p-12 border-0 bg-surface-elevated/80">
        <div className="absolute inset-0 bg-gradient-to-br from-accent-magic/20 via-transparent to-transparent opacity-50"></div>
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-4">
            <span className="text-3xl">📜</span>
            <span className="text-accent-magic font-bold tracking-widest uppercase text-sm">
              {quest.template_slug.replace(/-/g, ' ')}
            </span>
          </div>
          <h1 className="font-heading text-4xl md:text-5xl text-text-primary mb-4 leading-tight">
            {quest.title}
          </h1>

          <div className="flex items-center gap-3 flex-wrap">
            {isMyTurn && <YourTurnBadge />}
            {quest.status === 'ACTIVE' && (
              <span className="bg-accent-magic/20 text-accent-magic font-bold px-3 py-1 rounded border border-accent-magic/30">
                AKTIF
              </span>
            )}
            {quest.status === 'DONE' && (
              <span className="bg-accent-nature/20 text-accent-nature font-bold px-3 py-1 rounded border border-accent-nature/30">
                SELESAI
              </span>
            )}
            {quest.status === 'PENDING' && (
              <span className="bg-surface border border-border-subtle text-text-secondary font-bold px-3 py-1 rounded">
                MENUNGGU
              </span>
            )}
            {quest.quest_type && (
              <span className="bg-surface border border-border-subtle text-text-secondary font-bold px-3 py-1 rounded uppercase tracking-wider text-xs">
                {quest.quest_type}
              </span>
            )}
          </div>
        </div>
      </Card>

      {actionError && (
        <div className="bg-accent-danger/10 border border-accent-danger/30 p-4 rounded-lg">
          <p className="text-sm font-medium text-accent-danger">{actionError}</p>
        </div>
      )}

      {/* Narrative Branch Options Section */}
      {branchOptions.length > 0 && (
        <Card className="p-6 border-accent-magic/40 bg-surface flex flex-col gap-4">
          <div className="flex items-center gap-2">
            <span className="text-2xl">🌿</span>
            <div>
              <h2 className="font-heading text-xl text-text-primary">
                Cabang Narasi Cerita
              </h2>
              <p className="text-xs text-text-secondary">
                Pilih jalur keputusan cerita keluarga untuk menentukan alur petualangan ranah.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-2">
            {branchOptions.map((opt) => {
              const isSelecting = selectingBranch === opt.slug

              return (
                <div
                  key={opt.slug}
                  className="flex flex-col justify-between p-4 rounded-lg bg-surface-elevated border border-border-subtle hover:border-accent-magic/50 transition-all"
                >
                  <div className="mb-3">
                    <h3 className="font-medium text-text-primary text-base mb-1">
                      {opt.title}
                    </h3>
                    <p className="text-xs text-text-secondary italic">{opt.description}</p>
                  </div>
                  {onSelectBranch && quest.status === 'ACTIVE' && (
                    <Button
                      size="sm"
                      variant="secondary"
                      isLoading={isSelecting}
                      onClick={() => handleBranchSelect(opt.slug)}
                      className="w-full mt-2"
                    >
                      Pilih Jalur Ini
                    </Button>
                  )}
                </div>
              )
            })}
          </div>
        </Card>
      )}

      {/* Rewards / Result Banner */}
      {lastResult && lastResult.next_action !== 'CREATE_MEMORY' && (
        <Card className="border-accent-reward bg-accent-reward/5 p-6 text-center shadow-[0_0_20px_rgba(245,158,11,0.1)]">
          <h3 className="font-heading text-2xl text-accent-reward mb-4">Kemenangan!</h3>
          <div className="flex justify-center gap-6 mb-4">
            <div className="text-center">
              <span className="block text-2xl mb-1">✨</span>
              <span className="font-bold text-text-primary">+{lastResult.xp} XP</span>
            </div>
            <div className="text-center">
              <span className="block text-2xl mb-1">🪙</span>
              <span className="font-bold text-text-primary">+5 Koin</span>
            </div>
          </div>

          {lastResult.level_up && (
            <p className="text-sm font-bold text-accent-magic mt-4 bg-accent-magic/10 py-2 rounded">
              🎉 Naik Level! Kamu sekarang Level {lastResult.new_level}!
            </p>
          )}
          {lastResult.quest_completed && (
            <div className="flex flex-col items-center gap-4 mt-4">
              <p className="text-sm font-bold text-accent-nature">
                🏆 Misi Selesai! Cek beranda untuk mengambil peti hadiahmu.
              </p>
              <Link to="/">
                <Button variant="primary">Ambil Hadiah</Button>
              </Link>
            </div>
          )}
        </Card>
      )}

      {/* Story Memory Form */}
      {lastResult && lastResult.next_action === 'CREATE_MEMORY' && (
        <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
          <SubmissionForm
            questId={quest.id}
            challengeId={challenges[challenges.length - 1]?.id || 1}
            onComplete={() => setLastResult({ ...lastResult, next_action: undefined })}
            onSkip={() => setLastResult({ ...lastResult, next_action: undefined })}
          />
        </div>
      )}

      {/* Start Button */}
      {quest.status === 'PENDING' && onStartQuest && (
        <Button
          size="lg"
          isLoading={starting}
          onClick={handleStart}
          className="w-full shadow-lg shadow-accent-magic/20 text-lg"
        >
          Mulai Petualangan
        </Button>
      )}

      {/* Relay rotation panel */}
      {quest.quest_type === 'RELAY' && (
        <RelayRotation quest={quest} challenges={challenges} members={members} myUID={myUID} />
      )}

      {/* Learn Section */}
      {quest.status === 'ACTIVE' && quest.learn_text && !starting && (
        <Card className="p-6 border-accent-magic/40 bg-surface-elevated flex flex-col gap-4 animate-in fade-in slide-in-from-bottom-4">
          <div className="flex items-center gap-3">
            <span className="text-3xl">📖</span>
            <h2 className="font-heading text-2xl text-text-primary">BELAJAR</h2>
          </div>
          <p className="text-base text-text-primary leading-relaxed whitespace-pre-wrap bg-surface p-4 rounded-lg border border-border-subtle">
            {quest.learn_text}
          </p>
        </Card>
      )}

      {/* Challenges List (Practice) */}
      <div className="mt-4">
        <h2 className="font-heading text-2xl text-text-primary mb-6">
          {quest.status === 'ACTIVE' ? 'PRAKTIK' : 'Tujuan Misi'}
        </h2>
        {challenges.length === 0 ? (
          <p className="text-text-secondary italic">
            Jalan ke depan dipenuhi misteri. Belum ada tujuan.
          </p>
        ) : (
          <div className="flex flex-col gap-4 relative">
            <div className="absolute left-6 top-6 bottom-6 w-0.5 bg-border-subtle"></div>

            {challenges.map((c, index) => {
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

                      {/* Render Question if available */}
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
                      {!isDone && quest.status === 'ACTIVE' && (
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

                  {!isDone && quest.status === 'ACTIVE' && onCompleteChallenge && (
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
      </div>

      {/* RESULT & REWARD Section (Shows at the bottom when quest is DONE or last challenge completes) */}
      {lastResult && lastResult.quest_completed && quest.result_text && (
        <Card className="mt-4 border-accent-magic/50 bg-surface-elevated p-6 animate-in fade-in slide-in-from-bottom-4">
          <div className="flex items-center gap-3 mb-4">
            <span className="text-3xl">🎯</span>
            <h2 className="font-heading text-2xl text-text-primary">HASIL</h2>
          </div>
          <p className="text-base text-text-primary mb-6 p-4 bg-surface rounded-lg border border-border-subtle">
            {quest.result_text}
          </p>
          <div className="w-full h-px bg-border-subtle mb-6"></div>
          {/* Rewards are rendered above, but wait, I can just show the result text here */}
        </Card>
      )}
    </div>
  )
}
