import { useState, useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { questsApi, realmProgressApi } from '../../shared/lib/api'
import type { QuestView, RealmProgress } from '../../shared/types'
import { YourTurnBadge } from '../../shared/components/molecules/YourTurnBadge'
import { useSession } from '../../shared/hooks/useSession'
import { isMyRelayTurn } from '../../shared/utils/questTurn'
import {
  getMergedRealmProgress,
  getRealmForQuest,
  isRealmUnlocked,
} from '../../shared/utils/realm'

const QuestList = ({
  title,
  list,
  emptyMsg,
  uid,
}: {
  title: string
  list: QuestView[]
  emptyMsg: string
  uid?: string | null
}) => (
  <section className="mb-8">
    <h2 className="font-heading text-2xl text-text-primary mb-4 border-b border-border-subtle pb-2">
      {title}
    </h2>
    {list.length === 0 ? (
      <Card className="text-center py-8 opacity-60 bg-transparent border-dashed">
        <p className="text-text-secondary">{emptyMsg}</p>
      </Card>
    ) : (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {list.map((quest) => (
          <Card key={quest.id} hoverable className="flex flex-col">
            <div className="flex justify-between items-start mb-3">
              <div>
                <h3 className="font-heading text-xl text-text-primary">{quest.title}</h3>
                <p className="text-xs text-accent-magic uppercase tracking-wider">
                  {quest.template_slug.replace(/-/g, ' ')}
                </p>
                {quest.quest_type && (
                  <p className="text-[10px] text-text-secondary uppercase tracking-widest mt-1">
                    {quest.quest_type}
                  </p>
                )}
              </div>
              <div className="flex flex-col items-end gap-1">
                {isMyRelayTurn(quest, uid) && <YourTurnBadge />}
                {quest.status === 'ACTIVE' && (
                  <span className="text-xs font-bold bg-accent-magic/20 text-accent-magic px-2 py-1 rounded">
                    AKTIF
                  </span>
                )}
                {quest.status === 'DONE' && (
                  <span className="text-xs font-bold bg-accent-nature/20 text-accent-nature px-2 py-1 rounded">
                    SELESAI
                  </span>
                )}
                {quest.status === 'PENDING' && (
                  <span className="text-xs font-bold bg-surface border border-border-subtle text-text-secondary px-2 py-1 rounded">
                    MENUNGGU
                  </span>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2 mb-6">
              <div className="h-1.5 flex-1 bg-surface-elevated rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all ${
                    quest.status === 'DONE' ? 'bg-accent-nature' : 'bg-accent-magic'
                  }`}
                  style={{
                    width: `${
                      quest.challenge_count > 0
                        ? (quest.completed_count / quest.challenge_count) * 100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <span className="text-xs text-text-secondary whitespace-nowrap">
                {quest.completed_count} / {quest.challenge_count}
              </span>
            </div>

            <div className="mt-auto">
              <Link to={`/quests/${quest.id}`} className="block">
                <Button
                  variant={quest.status === 'ACTIVE' ? 'primary' : 'secondary'}
                  className="w-full"
                >
                  {quest.status === 'DONE' ? 'Ulas Misi' : 'Lihat Detail'}
                </Button>
              </Link>
            </div>
          </Card>
        ))}
      </div>
    )}
  </section>
)

export function QuestsPage() {
  const { session } = useSession()
  const [searchParams, setSearchParams] = useSearchParams()
  const [quests, setQuests] = useState<QuestView[]>([])
  const [availableQuests, setAvailableQuests] = useState<QuestView[]>([])
  const [realms, setRealms] = useState<RealmProgress[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const selectedRealm = searchParams.get('realm') || 'all'

  useEffect(() => {
    const loadData = async () => {
      try {
        const [questData, realmData, availableData] = await Promise.all([
          questsApi.list().catch(() => []),
          realmProgressApi.list().catch(() => []),
          questsApi.available().catch(() => []),
        ])
        setQuests(questData)
        setRealms(realmData)
        setAvailableQuests(availableData)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to load quests')
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  if (loading) {
    return (
      <div className="flex h-64 w-full items-center justify-center">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">📜</div>
          <p className="text-sm text-text-secondary">Membuka gulungan...</p>
        </div>
      </div>
    )
  }

  const mergedRealms = getMergedRealmProgress(realms)
  const currentRealmInfo = mergedRealms.find((r) => r.slug === selectedRealm)

  const filteredQuests =
    selectedRealm === 'all'
      ? quests
      : quests.filter((q) => getRealmForQuest(q) === selectedRealm)

  const activeQuests = filteredQuests.filter((q) => q.status === 'ACTIVE')
  const completedQuests = filteredQuests.filter((q) => q.status === 'DONE')

  const handleSelectRealm = (slug: string) => {
    if (slug === 'all') {
      searchParams.delete('realm')
      setSearchParams(searchParams)
      return
    }
    const target = mergedRealms.find((r) => r.slug === slug)
    if (target && !isRealmUnlocked(target.status)) return
    setSearchParams({ realm: slug })
  }

  return (
    <div className="max-w-5xl mx-auto flex flex-col gap-6">
      <header className="mb-2">
        <h1 className="font-heading text-4xl text-text-primary mb-2">Misi & Topik</h1>
        <p className="text-text-secondary">
          Tantangan petualangan keluarga di berbagai topik cerita.
        </p>
      </header>

      {/* Realm Filter Tabs */}
      <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-none" data-testid="realm-tabs">
        <button
          onClick={() => handleSelectRealm('all')}
          className={`px-4 py-2 rounded-lg text-sm font-semibold whitespace-nowrap transition-all ${
            selectedRealm === 'all'
              ? 'bg-accent-magic text-black shadow-md'
              : 'bg-surface border border-border-subtle text-text-secondary hover:text-text-primary'
          }`}
        >
          Semua Topik
        </button>

        {mergedRealms.map((r) => {
          const unlocked = isRealmUnlocked(r.status)
          const active = selectedRealm === r.slug

          return (
            <button
              key={r.slug}
              disabled={!unlocked}
              onClick={() => handleSelectRealm(r.slug)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold whitespace-nowrap transition-all ${
                !unlocked
                  ? 'opacity-50 cursor-not-allowed bg-surface-elevated/30 text-text-secondary border border-border-subtle'
                  : active
                  ? 'bg-accent-magic text-black shadow-md'
                  : 'bg-surface border border-border-subtle text-text-secondary hover:text-text-primary'
              }`}
            >
              <span>{r.icon}</span>
              <span>{r.name}</span>
              {!unlocked && <span className="text-xs">🔒</span>}
              {r.status === 'COMPLETE' && <span className="text-xs">✓</span>}
            </button>
          )
        })}
      </div>

      {/* Selected Realm Banner Header (when specific realm selected) */}
      {currentRealmInfo && selectedRealm !== 'all' && (
        <Card className="p-5 border-accent-magic/30 bg-surface-elevated/50 flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className="text-3xl">{currentRealmInfo.icon}</span>
              <div>
                <h2 className="font-heading text-2xl text-text-primary">
                  {currentRealmInfo.name}
                </h2>
                <p className="text-xs text-text-secondary">
                  {currentRealmInfo.description}
                </p>
              </div>
            </div>
            <span
              className={`text-xs font-bold px-3 py-1 rounded-full uppercase ${
                currentRealmInfo.status === 'COMPLETE'
                  ? 'bg-accent-nature/20 text-accent-nature'
                  : 'bg-accent-magic/20 text-accent-magic'
              }`}
            >
              {currentRealmInfo.status === 'COMPLETE' ? 'Selesai' : 'Topik Aktif'}
            </span>
          </div>

          <div className="mt-2">
            <div className="flex justify-between text-xs mb-1 text-text-secondary font-medium">
              <span>Progres Topik Ini</span>
              <span>{currentRealmInfo.progress}%</span>
            </div>
            <ProgressBar
              progress={currentRealmInfo.progress}
              colorClass={
                currentRealmInfo.status === 'COMPLETE'
                  ? 'bg-accent-nature'
                  : 'bg-accent-magic'
              }
            />
          </div>
        </Card>
      )}

      {error && (
        <div className="bg-accent-danger/10 border border-accent-danger/20 p-4 rounded-lg mb-6">
          <p className="text-sm text-accent-danger">{error}</p>
        </div>
      )}

      <QuestList
        title="Petualangan Aktif"
        list={activeQuests}
        emptyMsg="Tidak ada misi aktif untuk topik ini."
        uid={session?.uid}
      />
      <QuestList
        title="Misi Tersedia"
        list={availableQuests}
        emptyMsg="Tidak ada misi baru tersedia di topik ini."
        uid={session?.uid}
      />
      <QuestList
        title="Selesai"
        list={completedQuests}
        emptyMsg="Belum ada misi yang diselesaikan di topik ini."
        uid={session?.uid}
      />
    </div>
  )
}
