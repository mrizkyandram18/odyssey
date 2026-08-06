import type { Crew, Explorer } from '../../types'

export interface CrewDashboardProps {
  crew: Crew
  members: Explorer[]
}

export function CrewDashboard({ crew, members }: CrewDashboardProps) {
  void crew
  void members

  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">Crew Dashboard</h2>
      <ul className="flex flex-col gap-2">
        {members.map((m) => (
          <li key={m.uid} className="flex items-center justify-between">
            <span>{m.explorer_name}</span>
            <span className="text-sm text-muted-foreground">
              Level {m.level}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
