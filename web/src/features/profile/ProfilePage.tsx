import { ExplorerIcon } from '../../shared/components/atoms/ExplorerIcon'

export function ProfilePage() {
  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <div className="flex items-center gap-4">
        <ExplorerIcon role="SEEKER" size="lg" />
        <div>
          <h1 className="text-xl font-semibold">Explorer</h1>
          <p className="text-sm text-muted-foreground">Level 1</p>
        </div>
      </div>
    </div>
  )
}
