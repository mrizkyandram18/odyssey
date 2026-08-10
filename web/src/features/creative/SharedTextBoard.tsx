import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { boardApi, type BoardPost } from '../../shared/lib/api'
import { ConnectedReactionBar } from '../../shared/components/molecules/ConnectedReactionBar'
import { Button } from '../../shared/components/atoms/Button'
import { Card } from '../../shared/components/atoms/Card'

/**
 * Shared crew text board (Slice 2.3).
 * Multiple append-only text posts per crew — NOT a real-time collaborative editor.
 */
export function SharedTextBoard() {
  const [posts, setPosts] = useState<BoardPost[]>([])
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(true)
  const [posting, setPosting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await boardApi.list()
      setPosts(res.posts ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load board')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return
    setPosting(true)
    setError(null)
    try {
      await boardApi.post(content.trim())
      setContent('')
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to post')
    } finally {
      setPosting(false)
    }
  }

  const fmt = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' })

  return (
    <section className="flex flex-col gap-4" data-testid="shared-text-board">
      <div>
        <h2 className="text-xl font-bold text-text-primary">Shared Text Board</h2>
        <p className="text-sm text-muted-foreground">
          Leave notes for your crew. Everyone can add entries and react — not a live multiplayer editor.
        </p>
      </div>

      <Card className="p-4">
        <form onSubmit={submit} className="flex flex-col gap-3">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Write something the whole crew can see…"
            className="min-h-[96px] w-full resize-none rounded-lg border border-border bg-background p-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            maxLength={2000}
            disabled={posting}
            data-testid="board-compose"
          />
          <div className="flex items-center justify-between">
            <span className="text-xs text-muted-foreground">{content.trim().length}/2000</span>
            <Button
              type="submit"
              size="sm"
              disabled={posting || !content.trim()}
              isLoading={posting}
              data-testid="board-submit"
            >
              Post to board
            </Button>
          </div>
        </form>
      </Card>

      {error && (
        <p className="text-sm text-red-500" data-testid="board-error">
          {error}
        </p>
      )}

      {loading && posts.length === 0 ? (
        <p className="text-sm text-muted-foreground animate-pulse">Loading board…</p>
      ) : posts.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border bg-surface/50 p-8 text-center">
          <p className="text-muted-foreground">No board notes yet.</p>
          <p className="mt-1 text-sm text-muted-foreground">Be the first to leave a message for your crew.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {posts.map((post) => (
            <div
              key={post.id}
              className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4 shadow-sm"
              data-testid={`board-post-${post.id}`}
            >
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-accent-magic/20 text-xs font-bold text-accent-magic">
                  {post.author_uid.substring(0, 2).toUpperCase()}
                </div>
                <div className="flex flex-col">
                  <span className="text-sm font-semibold">{post.author_uid}</span>
                  <span className="text-xs text-muted-foreground">
                    {fmt.format(new Date(post.created_at))}
                  </span>
                </div>
              </div>
              <p className="whitespace-pre-wrap rounded-lg bg-background/50 p-3 text-sm leading-relaxed">
                {post.payload}
              </p>
              <div className="flex items-center justify-between border-t border-border/50 pt-3">
                <span className="text-xs font-medium text-muted-foreground">Shared board</span>
                <ConnectedReactionBar targetType="TEXT_BOARD" targetId={post.id} />
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
