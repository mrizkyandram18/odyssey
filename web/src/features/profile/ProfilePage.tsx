import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient } from '../../shared/lib/api'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle, ArrowLeft, LogOut, Banknote, Flame } from 'lucide-react'
import { PushNotificationToggle } from '../../shared/components/molecules/PushNotificationToggle'

export function ProfilePage() {
  const { profile, loading, error, refreshProfile, logout } = useSession()
  const [activeView, setActiveView] = useState<'overview' | 'settings'>('overview')
  const [randomizing, setRandomizing] = useState(false)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [changing, setChanging] = useState(false)
  const [changeError, setChangeError] = useState<string | null>(null)
  const [changeSuccess, setChangeSuccess] = useState<string | null>(null)

  const handleRandomizeAvatar = async () => {
    setRandomizing(true)
    const newSeed = Math.random().toString(36).substring(2, 10)
    try {
      await apiClient.patch('/api/me/avatar', {
        avatar_style: 'adventurer',
        avatar_seed: newSeed,
      })
      await refreshProfile()
    } catch (e) {
      console.error('failed to update avatar', e)
    } finally {
      setRandomizing(false)
    }
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setChangeError(null)
    setChangeSuccess(null)
    if (!currentPassword) {
      setChangeError('Kata sandi saat ini wajib diisi')
      return
    }
    if (newPassword.length < 6) {
      setChangeError('Kata sandi baru minimal 6 karakter')
      return
    }
    if (newPassword !== confirmPassword) {
      setChangeError('Konfirmasi kata sandi tidak cocok')
      return
    }
    if (currentPassword === newPassword) {
      setChangeError('Kata sandi baru tidak boleh sama dengan kata sandi saat ini')
      return
    }
    setChanging(true)
    try {
      await apiClient.post('/api/me/change-password', {
        current_password: currentPassword,
        new_password: newPassword,
        confirm_password: confirmPassword,
      })
      setChangeSuccess('Kata sandi berhasil diubah')
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal mengubah kata sandi'
      setChangeError(msg)
    } finally {
      setChanging(false)
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 w-full items-center justify-center max-w-2xl mx-auto">
        <div className="flex flex-col items-center gap-4 animate-pulse">
          <div className="text-4xl">👤</div>
          <p className="text-sm text-text-secondary">Memuat profil...</p>
        </div>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="flex flex-col gap-6 max-w-2xl mx-auto py-4">
        <header className="flex items-center justify-between">
          <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
            <ArrowLeft size={16} /> Beranda
          </Link>
        </header>
        <Card className="flex flex-col items-center justify-center gap-4 py-16 text-center border-accent-danger/30 bg-accent-danger/5">
          <p className="text-lg font-medium text-text-primary">
            {error || 'Profil tidak ditemukan. Silakan masuk kembali.'}
          </p>
          <Link to="/login" className="text-sm font-bold text-accent-magic hover:underline uppercase tracking-wider">
            Kembali ke Masuk
          </Link>
        </Card>
      </div>
    )
  }

  const roleLabel = profile.role === 'ADMIN' || profile.role === 'GUIDE' ? 'Administrator' : 'Anggota'
  const xpPercent = Math.min(100, (profile.xp ?? 0) % 100)
  const streakDays = profile.streak_days ?? 0

  return (
    <div className="flex flex-col gap-4">
      <header className="flex items-center gap-2">
        {activeView === 'overview' ? (
          <Link to="/" className="p-2 -ml-2 rounded-xl hover:bg-surface text-text-secondary hover:text-text-primary transition-colors" aria-label="Kembali">
            <ArrowLeft size={18} />
          </Link>
        ) : (
          <button onClick={() => setActiveView('overview')} className="p-2 -ml-2 rounded-xl hover:bg-surface text-text-secondary hover:text-text-primary transition-colors" aria-label="Kembali ke profil">
            <ArrowLeft size={18} />
          </button>
        )}
        <h1 className="text-lg font-bold text-text-primary">
          {activeView === 'overview' ? 'Profil' : 'Pengaturan'}
        </h1>
      </header>

      {/* Segmented control for overview/settings */}
      <div className="flex p-1 bg-surface rounded-xl border border-border-subtle gap-1">
        <button onClick={() => setActiveView('overview')} className={`flex-1 py-2 rounded-lg text-xs font-bold transition-colors ${activeView==='overview' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'}`}>Ringkasan</button>
        <button onClick={() => setActiveView('settings')} className={`flex-1 py-2 rounded-lg text-xs font-bold transition-colors ${activeView==='settings' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'}`}>Pengaturan</button>
      </div>

      {activeView === 'overview' && (
        <>
          <Card className="p-5">
            <div className="flex flex-col items-center text-center gap-3">
              <div className="relative">
                <Avatar
                  seed={profile.avatar_seed || profile.uid}
                  style={profile.avatar_style || 'adventurer'}
                  size="xl"
                />
                <button
                  onClick={handleRandomizeAvatar}
                  disabled={randomizing}
                  aria-label="Acak avatar"
                  className="absolute -bottom-1 -right-1 p-2 bg-accent-magic text-white rounded-full shadow-sm hover:brightness-110 transition-all active:scale-95 disabled:opacity-50"
                >
                  <Shuffle size={12} className={randomizing ? 'animate-spin' : ''} />
                </button>
              </div>
              <div>
                <h2 className="text-xl font-bold text-text-primary">{profile.explorer_name}</h2>
                <div className="flex items-center justify-center gap-2 mt-1.5 flex-wrap">
                  <span className="text-[11px] font-bold px-2.5 py-1 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/20">
                    {roleLabel}
                  </span>
                  {streakDays > 0 && (
                    <span className="text-[11px] font-bold px-2.5 py-1 rounded-full bg-accent-danger/10 text-accent-danger border border-accent-danger/20 inline-flex items-center gap-1">
                      <Flame size={12} /> {streakDays} Hari Streak
                    </span>
                  )}
                </div>
              </div>
              <div className="w-full bg-bg-app p-3 rounded-xl border border-border-subtle mt-1">
                <div className="flex justify-between text-xs font-bold mb-2">
                  <span className="text-accent-magic">Level {profile.level ?? 1}</span>
                  <span className="text-text-secondary">{profile.xp ?? 0} XP</span>
                </div>
                <ProgressBar progress={xpPercent} colorClass="bg-accent-magic" />
              </div>
            </div>
          </Card>

          <Card className="p-4 flex flex-col gap-3">
            <div className="flex items-center gap-3">
              <span className="text-2xl p-2.5 bg-accent-gold/15 rounded-xl border border-accent-gold/20">🪙</span>
              <div className="flex-1 min-w-0">
                <p className="flex items-baseline gap-1.5">
                  <span className="text-xl font-bold text-text-primary">{profile.coins ?? 0}</span>
                  <span className="text-xs font-semibold text-text-secondary">Koin</span>
                </p>
                <p className="text-xs text-text-secondary">Tukarkan koin menjadi uang tunai</p>
              </div>
              <Link
                to="/shop"
                className="shrink-0 inline-flex items-center gap-1.5 px-4 py-2.5 rounded-xl bg-accent-magic hover:brightness-110 text-white font-bold text-xs shadow-sm transition-all active:scale-95"
              >
                <Banknote size={14} /> Pencairan Koin
              </Link>
            </div>
          </Card>
        </>
      )}

      {activeView === 'settings' && (
        <div className="flex flex-col gap-4">
          <Card className="p-5 flex flex-col gap-4">
            <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
              <span>👤</span> Informasi Akun
            </h3>
            
            <div className="flex items-center justify-between border-b border-border-subtle/50 pb-4">
              <div>
                <p className="text-sm font-bold text-text-primary">Foto Profil</p>
                <p className="text-xs text-text-secondary">Ubah gaya avatar secara acak</p>
              </div>
              <Button
                variant="secondary"
                size="sm"
                className="flex items-center gap-2"
                onClick={handleRandomizeAvatar}
                disabled={randomizing}
              >
                <Shuffle size={14} className={randomizing ? 'animate-spin' : ''} /> Acak Avatar
              </Button>
            </div>

            <div className="flex items-center justify-between border-b border-border-subtle/50 pb-4">
              <div>
                <p className="text-sm font-bold text-text-primary">ID Pengguna (UID)</p>
                <p className="text-xs text-text-secondary font-mono">{profile.uid}</p>
              </div>
            </div>

            <div className="py-2">
              <PushNotificationToggle />
            </div>
          </Card>

          <Card className="p-5 flex flex-col gap-4">
            <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
              <span>🔒</span> Ubah Kata Sandi
            </h3>
            <form onSubmit={handleChangePassword} className="flex flex-col gap-3">
              <div className="flex flex-col gap-1">
                <label className="text-xs font-semibold text-text-secondary">Kata Sandi Saat Ini</label>
                <input
                  type="password"
                  value={currentPassword}
                  onChange={e => setCurrentPassword(e.target.value)}
                  placeholder="Masukkan kata sandi saat ini"
                  required
                  disabled={changing}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition text-sm"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs font-semibold text-text-secondary">Kata Sandi Baru</label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={e => setNewPassword(e.target.value)}
                  placeholder="Minimal 6 karakter"
                  required
                  disabled={changing}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition text-sm"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs font-semibold text-text-secondary">Konfirmasi Kata Sandi Baru</label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={e => setConfirmPassword(e.target.value)}
                  placeholder="Ulangi kata sandi baru"
                  required
                  disabled={changing}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition text-sm"
                />
              </div>
              {changeError && (
                <div className="px-3 py-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs">{changeError}</div>
              )}
              {changeSuccess && (
                <div className="px-3 py-2 rounded-xl bg-green-500/10 border border-green-500/20 text-green-600 text-xs">{changeSuccess}</div>
              )}
              <Button type="submit" variant="primary" size="sm" disabled={changing} className="w-full">
                {changing ? 'Menyimpan...' : 'Simpan Kata Sandi Baru'}
              </Button>
            </form>
          </Card>

          <Button variant="danger" onClick={logout} className="w-full shadow-md flex items-center justify-center gap-2">
            <LogOut size={16} /> Keluar (Sign Out)
          </Button>
        </div>
      )}
    </div>
  )
}
