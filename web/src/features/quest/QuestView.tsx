import { QuestDetail } from '../../shared/components/organisms/QuestDetail'
import { useQuest } from '../../shared/hooks/useQuest'
import type { Challenge, Quest } from '../../shared/types'

export function QuestView({ questId }: { questId: number }) {
  const { quest, challenges } = useQuest(questId)

  if (!quest) {
    return (
      <div className="p-4">
        <p className="text-sm text-muted-foreground">Quest not found.</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 p-4">
      <QuestDetail quest={quest as Quest} challenges={challenges as Challenge[]} />
    </div>
  )
}
