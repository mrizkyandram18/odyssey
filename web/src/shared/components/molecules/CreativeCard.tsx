import type { CreativeSubmission } from '../../types'
import { toSvgDataUri } from '../../utils/svg'
import { ConnectedReactionBar } from './ConnectedReactionBar'

export interface CreativeCardProps {
  submission: CreativeSubmission
}

export function CreativeCard({ submission }: CreativeCardProps) {
  const date = new Date(submission.created_at)
  const formatter = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' })

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/20 text-xs font-bold text-primary">
            {submission.author_uid.substring(0, 2).toUpperCase()}
          </div>
          <div className="flex flex-col">
            <span className="text-sm font-semibold">{submission.author_uid}</span>
            <span className="text-xs text-muted-foreground">
              {formatter.format(date)}
            </span>
          </div>
        </div>
      </div>
      
      <div className="rounded-lg bg-background/50 p-4 text-sm leading-relaxed overflow-hidden">
        {submission.kind === 'DRAWING' ? (
          <img
            src={toSvgDataUri(submission.content)}
            alt={`Drawing by ${submission.author_uid}`}
            className="w-full h-auto bg-white/5 rounded"
          />
        ) : (
          submission.content
        )}
      </div>

      <div className="mt-2 flex items-center justify-between border-t border-border/50 pt-3">
        <span className="text-xs font-medium text-muted-foreground">Quest #{submission.quest_id}</span>
        {/* Live reactions — target_type=JOURNAL maps to CreativeSubmission in backend */}
        <ConnectedReactionBar targetType="JOURNAL" targetId={submission.id} />
      </div>
    </div>
  )
}
