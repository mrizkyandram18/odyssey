import type { Quest, QuestView } from '../types'

export type TurnCandidate = Pick<Quest | QuestView, 'quest_type' | 'status' | 'active_challenge_assigned_to'>

export function isMyRelayTurn(quest: TurnCandidate | null | undefined, uid?: string | null): boolean {
  if (!uid || !quest) return false
  return (
    quest.quest_type === 'RELAY' &&
    quest.status === 'ACTIVE' &&
    quest.active_challenge_assigned_to === uid
  )
}
