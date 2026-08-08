import { Link } from 'react-router-dom'
import { QuestDetail } from '../../shared/components/organisms/QuestDetail'
import { useQuest } from '../../shared/hooks/useQuest'

export function QuestView({ questId }: { questId: number }) {
  const { quest, challenges, loading, error, startQuest, completeChallenge } = useQuest(questId)

  if (loading) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <Link to="/" className="text-sm text-muted-foreground">← Back to Home</Link>
        <p className="text-sm text-muted-foreground">Loading quest details...</p>
      </div>
    )
  }

  if (error || !quest) {
    return (
      <div className="flex flex-col gap-4 p-4 pb-safe">
        <Link to="/" className="text-sm text-muted-foreground">← Back to Home</Link>
        <p className="text-sm text-error">{error || 'Quest not found.'}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 p-4 pb-safe">
      <Link to="/" className="text-sm text-muted-foreground hover:text-primary transition-colors">
        ← Back to Home
      </Link>
      <QuestDetail
        quest={quest}
        challenges={challenges}
        onStartQuest={startQuest}
        onCompleteChallenge={completeChallenge}
      />
    </div>
  )
}
