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

  if (loading) return (
    <div className="flex h-64 w-full items-center justify-center">
      <div className="flex flex-col items-center gap-4 animate-pulse">
        <div className="text-4xl text-accent-magic">⚙️</div>
        <p className="text-sm text-text-secondary">Memuat Admin Dashboard...</p>
      </div>
    </div>
  )

  // UX Guard - actual security is enforced in the backend
  if (session?.role !== 'GUIDE') {
    return <Navigate to="/" replace />
  }

  return (
    <div className="max-w-6xl mx-auto flex flex-col gap-8 p-4 md:p-8">
      <header className="flex flex-col md:flex-row justify-between items-start md:items-end gap-4 border-b border-border-subtle pb-6">
        <div>
          <h1 className="text-4xl font-bold font-heading text-text-primary mb-2">
            Dashboard Pengelola
          </h1>
          <p className="text-text-secondary">
            Pantau statistik, kelola kurikulum modul, dan atur aktivitas harian pengguna.
          </p>
        </div>
        <Button onClick={fetchData} variant="secondary" className="gap-2">
          <span>🔄</span> Segarkan Data
        </Button>
      </header>

      {error && (
        <div className="bg-status-error/10 border border-status-error/20 p-4 rounded-lg flex items-center gap-3 text-status-error">
          <span className="text-xl">⚠️</span>
          <p className="text-sm">{error}</p>
        </div>
      )}

      {isFetching && !stats ? (
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 animate-pulse">
          {[1, 2, 3, 4, 5].map(i => (
            <div key={i} className="h-28 bg-surface-elevated rounded-xl"></div>
          ))}
        </div>
      ) : stats ? (
        <section className="space-y-4">
          <h2 className="text-xl font-semibold text-text-primary font-heading">Ringkasan Utama</h2>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <Card className="p-5 flex flex-col justify-center border-l-4 border-l-surface-elevated">
              <span className="text-xs text-text-secondary uppercase tracking-wider mb-1">Total Pengguna</span>
              <span className="text-3xl font-bold text-text-primary">{stats.total_users}</span>
            </Card>
            <Card className="p-5 flex flex-col justify-center border-l-4 border-l-status-success">
              <span className="text-xs text-text-secondary uppercase tracking-wider mb-1">Aktif (7 Hari)</span>
              <span className="text-3xl font-bold text-status-success">{stats.active_users_7d}</span>
            </Card>
            <Card className="p-5 flex flex-col justify-center border-l-4 border-l-accent-primary">
              <span className="text-xs text-text-secondary uppercase tracking-wider mb-1">Aktif (30 Hari)</span>
              <span className="text-3xl font-bold text-text-primary">{stats.active_users_30d}</span>
            </Card>
            <Card className="p-5 flex flex-col justify-center border-l-4 border-l-accent-magic">
              <span className="text-xs text-text-secondary uppercase tracking-wider mb-1">Modul Selesai</span>
              <span className="text-3xl font-bold text-accent-magic">{stats.Mission_completions}</span>
            </Card>
            <Card className="p-5 flex flex-col justify-center border-l-4 border-l-status-warning">
              <span className="text-xs text-text-secondary uppercase tracking-wider mb-1">Aktivitas Hari Ini</span>
              <span className="text-3xl font-bold text-status-warning">{stats.daily_activity_completions_today}</span>
            </Card>
          </div>
        </section>
      ) : null}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <section className="space-y-4">
          <h2 className="text-xl font-semibold text-text-primary font-heading">Manajemen Modul Belajar</h2>
          <Card className="overflow-hidden border border-border-subtle p-0">
            <div className="overflow-x-auto max-h-[600px]">
              <table className="w-full text-left">
                <thead className="bg-surface-elevated sticky top-0 z-10 shadow-sm">
                  <tr>
                    <th className="p-4 text-xs font-semibold text-text-secondary uppercase tracking-wider">Judul Modul</th>
                    <th className="p-4 text-xs font-semibold text-text-secondary uppercase tracking-wider text-center">Selesai</th>
                    <th className="p-4 text-xs font-semibold text-text-secondary uppercase tracking-wider text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border-subtle">
                  {missions.map((q) => (
                    <tr key={q.slug} className="hover:bg-surface-elevated/30 transition-colors">
                      <td className="p-4">
                        <div className="font-medium text-text-primary mb-1">{q.title}</div>
                        <div className="flex items-center gap-2">
                          <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold ${
                            q.published ? 'bg-status-success/20 text-status-success' : 'bg-surface-elevated text-text-secondary'
                          }`}>
                            {q.published ? 'DIPUBLIKASIKAN' : 'DRAFT'}
                          </span>
                        </div>
                      </td>
                      <td className="p-4 text-center font-medium text-text-primary">
                        {q.completion_count} <span className="text-xs font-normal text-text-secondary">kali</span>
                      </td>
                      <td className="p-4 text-right">
                        <Button 
                          variant={q.published ? 'secondary' : 'primary'}
                          size="sm"
                          className="text-xs"
                          onClick={() => handleToggleMission(q.slug)}
                        >
                          {q.published ? 'Sembunyikan' : 'Publikasikan'}
                        </Button>
                      </td>
                    </tr>
                  ))}
                  {missions.length === 0 && !isFetching && (
                    <tr>
                      <td colSpan={3} className="p-8 text-center text-text-secondary text-sm italic">
                        Tidak ada data modul.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </Card>
        </section>

        <section className="space-y-4">
          <h2 className="text-xl font-semibold text-text-primary font-heading">Tugas Harian (Habits)</h2>
          <Card className="overflow-hidden border border-border-subtle p-0">
            <div className="overflow-x-auto max-h-[600px]">
              <table className="w-full text-left">
                <thead className="bg-surface-elevated sticky top-0 z-10 shadow-sm">
                  <tr>
                    <th className="p-4 text-xs font-semibold text-text-secondary uppercase tracking-wider">Tugas Rutin</th>
                    <th className="p-4 text-xs font-semibold text-text-secondary uppercase tracking-wider text-center">Selesai</th>
                    <th className="p-4 text-xs font-semibold text-text-secondary uppercase tracking-wider text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border-subtle">
                  {activities.map((a) => (
                    <tr key={a.id} className="hover:bg-surface-elevated/30 transition-colors">
                      <td className="p-4">
                        <div className="font-medium text-text-primary mb-1">{a.title}</div>
                        <div className="flex items-center gap-2">
                          <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold ${
                            a.active ? 'bg-accent-magic/20 text-accent-magic' : 'bg-status-error/20 text-status-error'
                          }`}>
                            {a.active ? 'AKTIF' : 'TIDAK AKTIF'}
                          </span>
                        </div>
                      </td>
                      <td className="p-4 text-center font-medium text-text-primary">
                        {a.completion_count} <span className="text-xs font-normal text-text-secondary">kali</span>
                      </td>
                      <td className="p-4 text-right">
                        <Button 
                          variant={a.active ? 'secondary' : 'primary'}
                          size="sm"
                          className="text-xs"
                          onClick={() => handleToggleActivity(a.id)}
                        >
                          {a.active ? 'Nonaktifkan' : 'Aktifkan'}
                        </Button>
                      </td>
                    </tr>
                  ))}
                  {activities.length === 0 && !isFetching && (
                    <tr>
                      <td colSpan={3} className="p-8 text-center text-text-secondary text-sm italic">
                        Tidak ada data tugas rutin.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </Card>
        </section>
      </div>
    </div>
  )
}
