import type { Challenge, CrewMember, Quest, QuestWithChallenges } from '../../types'
import { deriveRelayLegs, memberName } from '../../utils/relayRotation'

export interface RelayRotationProps {
  quest: Quest | QuestWithChallenges
  challenges: Challenge[]
  members?: CrewMember[]
  myUID?: string | null
}

const legStyles = {
  done: {
    circle: 'bg-accent-nature border-accent-nature text-bg-app',
    ring: '',
    label: 'text-accent-nature',
  },
  active: {
    circle: 'bg-accent-magic border-accent-magic text-bg-app shadow-[0_0_12px_rgba(6,182,222,0.45)]',
    ring: 'ring-2 ring-accent-magic/30',
    label: 'text-accent-magic',
  },
  next: {
    circle: 'bg-bg-app border-accent-reward text-accent-reward',
    ring: '',
    label: 'text-accent-reward',
  },
  open: {
    circle: 'bg-bg-app border-border-subtle text-text-secondary',
    ring: '',
    label: 'text-text-secondary',
  },
} as const

export function RelayRotation({ quest, challenges, members, myUID }: RelayRotationProps) {
  if (quest.quest_type !== 'RELAY') return null

  const legs = deriveRelayLegs(challenges, quest.active_challenge_assigned_to, myUID)
  const active = legs.find((l) => l.state === 'active')
  const next = legs.find((l) => l.state === 'next')
  const doneCount = legs.filter((l) => l.state === 'done').length

  return (
    <section
      aria-label="Relay rotation"
      className="rounded-xl border border-border-subtle bg-surface p-5 md:p-6"
    >
      <div className="mb-4">
        <h2 className="font-heading text-xl text-text-primary">Giliran Keluarga</h2>
        <p className="text-xs text-text-secondary mt-0.5">
          Lanjutkan petualangan — setiap misi dikerjakan bergantian.
        </p>
      </div>

      {legs.length === 0 ? (
        <p className="text-sm text-text-secondary italic">No legs to show yet.</p>
      ) : (
        <ol className="flex flex-col gap-1 relative">
          <div className="absolute left-4 top-4 bottom-4 w-0.5 bg-border-subtle" aria-hidden="true" />
          {legs.map((leg, index) => {
            const style = legStyles[leg.state]
            const label = {
              done: `Done by ${memberName(members, leg.challenge.completed_by)}`,
              active: `${memberName(members, leg.challenge.assigned_to)}'s turn`,
              next: 'Up next',
              open: 'Open',
            }[leg.state]

            return (
              <li key={leg.challenge.id} className="flex gap-3 py-2.5 relative z-10">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 border-2 text-sm font-bold ${style.circle} ${style.ring}`}
                  aria-hidden="true"
                >
                  {leg.state === 'done' ? '✓' : index + 1}
                </div>
                <div className="min-w-0 flex-1 pt-0.5">
                  <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <span className={`text-[10px] font-bold uppercase tracking-widest ${style.label}`}>
                      {label}
                    </span>
                    {leg.state === 'active' && leg.isMyTurn && (
                      <span className="animate-pulse rounded-full border border-yellow-500/30 bg-yellow-500/10 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-yellow-600 dark:text-yellow-400">
                        Your turn
                      </span>
                    )}
                  </div>
                  <p className={`text-sm mt-0.5 ${leg.state === 'done' ? 'text-text-secondary' : 'text-text-primary'}`}>
                    {leg.challenge.description}
                  </p>
                </div>
              </li>
            )
          })}
        </ol>
      )}

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-border-subtle pt-3 text-[11px] text-text-secondary">
        <span>
          <span className="font-bold text-accent-nature">{doneCount}</span> done
        </span>
        {active && (
          <span>
          <span className="font-bold text-accent-magic">{memberName(members, active.challenge.assigned_to) || 'Siapa saja'}</span> sekarang
          </span>
        )}
        {next && <span>selanjutnya: {memberName(members, next.challenge.assigned_to) || 'siapa saja'}</span>}
      </div>
    </section>
  )
}
