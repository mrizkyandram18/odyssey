import type { Family, Explorer } from '../../types'

import { Avatar } from '../atoms/Avatar'

export interface FamilyDashboardProps {
  crew: Family
  members: Explorer[]
}

export function FamilyDashboard({ crew, members }: FamilyDashboardProps) {
  void crew
  void members

  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">Dashboard Keluarga</h2>
      <ul className="flex flex-col gap-2">
        {members.map((m) => (
          <li key={m.uid} className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Avatar seed={m.avatar_seed || m.uid} style={m.avatar_style || 'adventurer'} size="sm" />
              <span>{m.explorer_name}</span>
            </div>
            <span className="text-sm text-muted-foreground">
              Level {m.level}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
