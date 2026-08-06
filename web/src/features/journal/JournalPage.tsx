export function JournalPage() {
  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <h1 className="text-xl font-semibold">Journal</h1>
      <p className="text-sm text-muted-foreground">
        Your achievements and collected relics will appear here.
      </p>
      <div className="flex flex-col gap-4">
        <section>
          <h2 className="text-sm font-medium text-muted-foreground">Achievements</h2>
          <p className="text-sm text-muted-foreground">No achievements yet.</p>
        </section>
        <section>
          <h2 className="text-sm font-medium text-muted-foreground">Relics</h2>
          <p className="text-sm text-muted-foreground">No relics yet.</p>
        </section>
      </div>
    </div>
  )
}
