import type { Mission, MissionView } from '../types'

export type TurnCandidate = Pick<Mission | MissionView, 'Mission_type' | 'status' | 'active_challenge_assigned_to'>

export function isMyRelayTurn(Mission: TurnCandidate | null | undefined, uid?: string | null): boolean {
  if (!uid || !Mission) return false
  return (
    Mission.Mission_type === 'RELAY' &&
    Mission.status === 'ACTIVE' &&
    Mission.active_challenge_assigned_to === uid
  )
}
