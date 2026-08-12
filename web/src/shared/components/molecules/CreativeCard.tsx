import type { CreativeSubmission } from '../../types'
import { Link } from 'react-router-dom'
import { toSvgDataUri } from '../../utils/svg'
import { parseComicPayload } from '../../utils/comic'
import { parsePhotoPayload, parseVideoPayload } from '../../utils/media'
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
              {submission.kind === 'COMIC' ? ' · Comic' : submission.kind === 'DRAWING' ? ' · Drawing' : submission.kind === 'PHOTO' ? ' · Photo' : submission.kind === 'VIDEO' ? ' · Video' : ''}
            </span>
          </div>
        </div>
      </div>

      <div className="rounded-lg bg-background/50 p-4 text-sm leading-relaxed overflow-hidden">
        <CreativeBody submission={submission} />
      </div>

      <div className="mt-2 flex items-center justify-between border-t border-border/50 pt-3">
        <div className="flex items-center gap-3">
          <span className="text-xs font-medium text-muted-foreground">Misi #{submission.mission_id}</span>
          {submission.kind === 'DRAWING' && (
            <Link to={`/stories/${submission.id}`} className="text-xs text-primary underline">
              Lihat Detail
            </Link>
          )}
        </div>
        {/* Live reactions — target_type=JOURNAL maps to CreativeSubmission in backend */}
        <ConnectedReactionBar targetType="JOURNAL" targetId={submission.id} />
      </div>
    </div>
  )
}

function CreativeBody({ submission }: { submission: CreativeSubmission }) {
  if (submission.kind === 'DRAWING') {
    return (
      <img
        src={toSvgDataUri(submission.content)}
        alt={`Drawing by ${submission.author_uid}`}
        className="w-full h-auto bg-white/5 rounded"
      />
    )
  }

  if (submission.kind === 'COMIC') {
    const comic = parseComicPayload(submission.content)
    if (!comic) {
      return <p className="text-muted-foreground text-xs">Comic could not be displayed.</p>
    }
    return (
      <div className="flex flex-col gap-3">
        {comic.panels.map((panel, i) => (
          <div
            key={i}
            className="rounded-md border border-border/60 bg-background/40 overflow-hidden"
          >
            {panel.svg ? (
              <img
                src={toSvgDataUri(panel.svg)}
                alt={`Comic panel ${i + 1}`}
                className="w-full h-auto bg-white/5"
              />
            ) : null}
            {panel.caption ? (
              <p className="px-3 py-2 text-sm leading-snug">{panel.caption}</p>
            ) : null}
          </div>
        ))}
      </div>
    )
  }

  if (submission.kind === 'PHOTO') {
    const photo = parsePhotoPayload(submission.content)
    if (!photo) {
      return <p className="text-muted-foreground text-xs">Photo could not be displayed.</p>
    }
    return (
      <div className="flex flex-col gap-2">
        <img src={photo.photo} alt="Submitted photo" className="w-full h-auto rounded-md object-cover" />
        {photo.caption ? <p className="text-sm leading-snug">{photo.caption}</p> : null}
      </div>
    )
  }

  if (submission.kind === 'VIDEO') {
    const video = parseVideoPayload(submission.content)
    if (!video) {
      return <p className="text-muted-foreground text-xs">Video could not be displayed.</p>
    }
    return (
      <div className="flex flex-col gap-2">
        <video src={video.video} controls className="w-full h-auto rounded-md object-cover" />
        {video.caption ? <p className="text-sm leading-snug">{video.caption}</p> : null}
      </div>
    )
  }

  return <>{submission.content}</>
}
