import { useEffect, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { useSession } from '../../shared/hooks/useSession'
import { adminApi, type AdminStats, type MissionStat, type ActivityStat } from '../../shared/lib/api'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'

export function AdminPage() {
  const { session, loading } = useSession()

  const [stats, setStats] = useState<AdminStats | null>(null)
  const [missions, setMissions] = useState<MissionStat[]>([])
  const [activities, setActivities] = useState<ActivityStat[]>([])
  const [error, setError] = useState<string | null>(null)
  const [isFetching, setIsFetching] = useState(true)

  const fetchData = async () => {
    try {
      setIsFetching(true)
      const [s, q, a] = await Promise.all([
        adminApi.getStats(),
        adminApi.getMissions(),
        adminApi.getDailyActivities(),
      ])
      setStats(s)
      setMissions(q)
      setActivities(a)
    } catch (err: any) {
      setError(err.message || 'Gagal mengambil data admin')
    } finally {
      setIsFetching(false)
    }
  }

  useEffect(() => {
    if (session?.role !== 'GUIDE') return
    fetchData()
  }, [session])

  const handleToggleMission = async (slug: string) => {
    try {
      await adminApi.toggleMission(slug)
      setMissions((prev) =>
        prev.map((q) => (q.slug === slug ? { ...q, published: !q.published } : q))
      )
    } catch (err: any) {
      alert(err.message)
    }
  }

  const handleToggleActivity = async (id: number) => {
    try {
      await adminApi.toggleActivity(id)
      setActivities((prev) =>
        prev.map((a) => (a.id === id ? { ...a, active: !a.active } : a))
      )
    } catch (err: any) {
      alert(err.message)
    }
  }

  if (loading) return <div>Loading...</div>

  // UX Guard - actual security is enforced in the backend
  if (session?.role !== 'GUIDE') {
    return <Navigate to="/" replace />
  }

  return (
    <div className="p-4 md:p-8 space-y-8 text-text-primary">
      <h1 className="text-3xl font-bold font-heading text-accent-magic">
        Admin Dashboard
      </h1>

      {error && <div className="text-status-error bg-status-error/10 p-4 rounded-lg">{error}</div>}

      {isFetching && !stats ? (
        <div className="animate-pulse flex space-x-4">
          <div className="h-10 bg-surface-elevated rounded w-1/4"></div>
          <div className="h-10 bg-surface-elevated rounded w-1/4"></div>
        </div>
      ) : stats ? (
        <section className="space-y-4">
          <h2 className="text-2xl font-semibold">Ringkasan</h2>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <Card className="p-4 flex flex-col items-center">
              <span className="text-sm text-text-secondary">Total Pengguna</span>
              <span className="text-2xl font-bold">{stats.total_users}</span>
            </Card>
            <Card className="p-4 flex flex-col items-center">
              <span className="text-sm text-text-secondary">Aktif 7 Hari</span>
              <span className="text-2xl font-bold text-status-success">{stats.active_users_7d}</span>
            </Card>
            <Card className="p-4 flex flex-col items-center">
              <span className="text-sm text-text-secondary">Aktif 30 Hari</span>
              <span className="text-2xl font-bold">{stats.active_users_30d}</span>
            </Card>
            <Card className="p-4 flex flex-col items-center">
              <span className="text-sm text-text-secondary">Misi Selesai</span>
              <span className="text-2xl font-bold text-accent-magic">{stats.Mission_completions}</span>
            </Card>
            <Card className="p-4 flex flex-col items-center">
              <span className="text-sm text-text-secondary">Aktivitas Hari Ini</span>
              <span className="text-2xl font-bold text-status-warning">{stats.daily_activity_completions_today}</span>
            </Card>
          </div>
        </section>
      ) : null}

      <section className="space-y-4">
        <h2 className="text-2xl font-semibold">Misi</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-left bg-surface-elevated rounded-lg overflow-hidden">
            <thead className="bg-surface-elevated/50 text-text-secondary">
              <tr>
                <th className="p-3">Judul</th>
                <th className="p-3">Selesai</th>
                <th className="p-3">Status</th>
                <th className="p-3">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-base">
              {missions.map((q) => (
                <tr key={q.slug}>
                  <td className="p-3">
                    <div className="font-medium">{q.title}</div>
                    <div className="text-xs text-text-secondary">{q.slug}</div>
                  </td>
                  <td className="p-3">{q.completion_count}</td>
                  <td className="p-3">
                    {q.published ? (
                      <span className="text-status-success text-sm font-semibold">Published</span>
                    ) : (
                      <span className="text-text-secondary text-sm font-semibold">Draft</span>
                    )}
                  </td>
                  <td className="p-3">
                    <Button 
                      variant={q.published ? 'ghost' : 'primary'}
                      size="sm"
                      onClick={() => handleToggleMission(q.slug)}
                    >
                      {q.published ? 'Hide' : 'Publish'}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="space-y-4">
        <h2 className="text-2xl font-semibold">Aktivitas Hari Ini</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-left bg-surface-elevated rounded-lg overflow-hidden">
            <thead className="bg-surface-elevated/50 text-text-secondary">
              <tr>
                <th className="p-3">Judul</th>
                <th className="p-3">Selesai</th>
                <th className="p-3">Status</th>
                <th className="p-3">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-base">
              {activities.map((a) => (
                <tr key={a.id}>
                  <td className="p-3">
                    <div className="font-medium">{a.title}</div>
                    <div className="text-xs text-text-secondary">{a.slug}</div>
                  </td>
                  <td className="p-3">{a.completion_count}</td>
                  <td className="p-3">
                    {a.active ? (
                      <span className="text-status-success text-sm font-semibold">Active</span>
                    ) : (
                      <span className="text-status-error text-sm font-semibold">Inactive</span>
                    )}
                  </td>
                  <td className="p-3">
                    <Button 
                      variant={a.active ? 'ghost' : 'primary'}
                      size="sm"
                      onClick={() => handleToggleActivity(a.id)}
                    >
                      {a.active ? 'Disable' : 'Enable'}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  )
}
