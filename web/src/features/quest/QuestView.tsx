import { Link } from 'react-router-dom'
import { QuestDetail } from '../../shared/components/organisms/QuestDetail'
import { useQuest } from '../../shared/hooks/useQuest'
import { useSession } from '../../shared/hooks/useSession'
import { isMyRelayTurn } from '../../shared/utils/questTurn'


export function QuestView({ questId }: { questId: number }) {
  const { quest, challenges, loading, error, startQuest, completeChallenge, selectBranch } = useQuest(questId)
  const { session } = useSession()

  if (loading) {
    return (
      <div className="flex flex-col max-w-4xl mx-auto py-8">
        <Link to="/quests" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors mb-8 inline-flex items-center gap-2">
          <span>←</span> Back to Quests
        </Link>
        <div className="flex h-64 w-full items-center justify-center">
          <div className="flex flex-col items-center gap-4 animate-pulse">
            <div className="text-4xl">📜</div>
            <p className="text-sm text-text-secondary">Unrolling the scroll...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error || !quest) {
    return (
      <div className="flex flex-col max-w-4xl mx-auto py-8">
        <Link to="/quests" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors mb-8 inline-flex items-center gap-2">
          <span>←</span> Back to Quests
        </Link>
        <div className="bg-accent-danger/10 border border-accent-danger/30 p-6 rounded-lg text-center">
          <p className="text-lg font-medium text-accent-danger">{error || 'Quest not found.'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col max-w-4xl mx-auto py-4">
      <Link to="/quests" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors mb-6 inline-flex items-center gap-2">
        <span>←</span> Back to Quests
      </Link>
      <QuestDetail
        quest={quest}
        challenges={challenges}
        members={quest.members}
        myUID={session?.uid}
        onStartQuest={startQuest}
        onCompleteChallenge={completeChallenge}
        onSelectBranch={selectBranch}
        isMyTurn={isMyRelayTurn(quest, session?.uid)}
      />
    </div>
  )
}
