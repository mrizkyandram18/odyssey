export interface ReactionData {
  id: string
  creator_id: string
  target_user_id: string
  emoji_code: string
}

export interface ReactionListProps {
  reactions: ReactionData[]
}

export function ReactionList({ reactions }: ReactionListProps) {
  if (!reactions || reactions.length === 0) return null

  // Group by emoji
  const counts = reactions.reduce((acc, curr) => {
    acc[curr.emoji_code] = (acc[curr.emoji_code] || 0) + 1
    return acc
  }, {} as Record<string, number>)

  return (
    <div className="flex flex-wrap gap-2">
      {Object.entries(counts).map(([emoji, count]) => (
        <div 
          key={emoji} 
          className="flex items-center gap-1 rounded-full border border-border bg-surface px-2 py-1 text-sm shadow-sm"
        >
          <span>{emoji}</span>
          {count > 1 && <span className="text-xs text-muted-foreground">{count}</span>}
        </div>
      ))}
    </div>
  )
}
