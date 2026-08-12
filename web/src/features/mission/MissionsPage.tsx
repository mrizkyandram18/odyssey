import { useState, useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { MissionsApi, JourneyProgressApi } from '../../shared/lib/api'
import type { MissionView, JourneyProgress } from '../../shared/types'
import { YourTurnBadge } from '../../shared/components/molecules/YourTurnBadge'
import { WorldMap } from '../../shared/components/organisms/WorldMap'
import { useSession } from '../../shared/hooks/useSession'
import { isMyRelayTurn } from '../../shared/utils/missionTurn'
import {
  getMergedJourneyProgress,
  getRealmForMission,
  isRealmUnlocked,
} from '../../shared/utils/journey'

const MissionList = ({
  title,
  list,
  emptyMsg,
  uid,
}: {
  title: string
  list: MissionView[]
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
        {list.map((Mission) => (
          <Card key={Mission.id} hoverable className="flex flex-col">
            <div className="flex justify-between items-start mb-3">
              <div>
                <h3 className="font-heading text-xl text-text-primary">{Mission.title}</h3>
                <p className="text-xs text-accent-magic uppercase tracking-wider">
                  {Mission.template_slug.replace(/-/g, ' ')}
                </p>
                {Mission.Mission_type && (
                  <p className="text-[10px] text-text-secondary uppercase tracking-widest mt-1">
                    {Mission.Mission_type}
                  </p>
                )}
              </div>
              <div className="flex flex-col items-end gap-1">
                {isMyRelayTurn(Mission, uid) && <YourTurnBadge />}
                {Mission.status === 'ACTIVE' && (
                  <span className="text-xs font-bold bg-accent-magic/20 text-accent-magic px-2 py-1 rounded">
                    AKTIF
                  </span>
                )}
                {Mission.status === 'DONE' && (
                  <span className="text-xs font-bold bg-accent-nature/20 text-accent-nature px-2 py-1 rounded">
                    SELESAI
                  </span>
                )}
                {Mission.status === 'PENDING' && (
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
                    Mission.status === 'DONE' ? 'bg-accent-nature' : 'bg-accent-magic'
                  }`}
                  style={{
                    width: `${
                      Mission.challenge_count > 0
                        ? (Mission.completed_count / Mission.challenge_count) * 100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <span className="text-xs text-text-secondary whitespace-nowrap">
                {Mission.completed_count} / {Mission.challenge_count}
              </span>
            </div>

            <div className="mt-auto">
              <Link to={`/missions/${Mission.id}`} className="block">
                <Button
                  variant={Mission.status === 'ACTIVE' ? 'primary' : 'secondary'}
                  className="w-full"
                >
                  {Mission.status === 'DONE' ? 'Ulas Misi' : 'Lihat Detail'}
                </Button>
              </Link>
            </div>
          </Card>
        ))}
      </div>
    )}
  </section>
)

export function MissionsPage() {
  const { session } = useSession()
  const [searchParams, setSearchParams] = useSearchParams()
  const [missions, setMissions] = useState<MissionView[]>([])
  const [availableMissions, setAvailableMissions] = useState<MissionView[]>([])
  const [realms, setRealms] = useState<JourneyProgress[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const selectedRealm = searchParams.get('journey') || 'all'

  useEffect(() => {
    const loadData = async () => {
      try {
        const [MissionData, realmData, availableData] = await Promise.all([
          MissionsApi.list().catch(() => []),
          JourneyProgressApi.list().catch(() => []),
          MissionsApi.available().catch(() => []),
        ])
        setMissions(MissionData)
        setRealms(realmData)
        setAvailableMissions(availableData)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to load missions')
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

  const mergedRealms = getMergedJourneyProgress(realms)
  const currentRealmInfo = mergedRealms.find((r) => r.slug === selectedRealm)

  const filteredMissions =
    selectedRealm === 'all'
      ? missions
      : missions.filter((q) => getRealmForMission(q) === selectedRealm)

  const activeMissions = filteredMissions.filter((q) => q.status === 'ACTIVE')
  const completedMissions = filteredMissions.filter((q) => q.status === 'DONE')
  const pendingMissions = filteredMissions.filter(
    (q) =>
      q.status === 'PENDING' &&
      availableMissions.some((aq) => aq.template_slug === q.template_slug)
  )

  const handleSelectRealm = (slug: string) => {
    if (slug === 'all') {
      searchParams.delete('journey')
      setSearchParams(searchParams)
      return
    }
    const target = mergedRealms.find((r) => r.slug === slug)
    if (target && !isRealmUnlocked(target.status)) return
    setSearchParams({ journey: slug })
  }

  return (
    <div className="max-w-5xl mx-auto flex flex-col gap-6">
      <header className="mb-2">
        <h1 className="font-heading text-4xl text-text-primary mb-2">Misi & Topik</h1>
        <p className="text-text-secondary">
          Tantangan petualangan keluarga di berbagai topik cerita.
        </p>
      </header>

      {/* World Map Section */}
      <section className="mb-8">
        <h2 className="font-heading text-2xl text-text-primary mb-4 border-b border-border-subtle pb-2">
          Mengenal Topik
        </h2>
        <WorldMap 
          realms={realms}
          onRealmSelect={handleSelectRealm}
          selectedRealm={selectedRealm}
        />
      </section>

      {/* Selected Journey Banner Header (when specific journey selected) */}
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

      <MissionList
        title="Misi Aktif"
        list={activeMissions}
        emptyMsg="Tidak ada misi aktif untuk topik ini."
        uid={session?.uid}
      />
      <MissionList
        title="Misi Tersedia"
        list={pendingMissions}
        emptyMsg="Tidak ada misi baru tersedia di topik ini."
        uid={session?.uid}
      />
      <MissionList
        title="Selesai"
        list={completedMissions}
        emptyMsg="Belum ada misi yang diselesaikan di topik ini."
        uid={session?.uid}
      />
    </div>
  )
}
