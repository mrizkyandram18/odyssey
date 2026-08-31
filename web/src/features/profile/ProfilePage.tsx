import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../shared/components/atoms/Button'
import { ProgressBar } from '../../shared/components/atoms/ProgressBar'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient } from '../../shared/lib/api'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle, ArrowLeft, LogOut, ShoppingBag, Flame } from 'lucide-react'
import { PushNotificationToggle } from '../../shared/components/molecules/PushNotificationToggle'

export function ProfilePage() {
  const { profile, loading, error, refreshProfile, logout } = useSession()
  const [activeView, setActiveView] = useState<'overview' | 'settings'>('overview')
  const [randomizing, setRandomizing] = useState(false)

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

  const roleLabel = profile.role === 'GUIDE' ? 'Admin Keluarga' : 'Anggota Keluarga'
  const xpPercent = Math.min(100, (profile.xp ?? 0) % 100)
  const streakDays = profile.streak_days ?? 0

  return (
    <div className="flex flex-col gap-6 max-w-2xl mx-auto py-4 animate-in fade-in duration-500">
      <header className="flex items-center justify-between mb-2">
        {activeView === 'overview' ? (
          <Link to="/" className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2">
            <ArrowLeft size={16} /> Beranda
          </Link>
        ) : (
          <button onClick={() => setActiveView('overview')} className="text-sm font-medium text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 cursor-pointer">
            <ArrowLeft size={16} /> Kembali ke Profil
          </button>
        )}
        <h1 className="font-heading text-2xl md:text-3xl text-text-primary">
          {activeView === 'overview' ? 'Profil Keluarga' : 'Pengaturan Akun'}
        </h1>
        {activeView === 'overview' ? (
          <button 
            onClick={() => setActiveView('settings')}
            className="text-xs font-bold uppercase tracking-wider px-3 py-1.5 rounded-lg bg-surface-elevated border border-border-subtle hover:border-accent-magic text-text-secondary hover:text-text-primary transition-colors"
          >
            Pengaturan
          </button>
        ) : null}
      </header>

      {activeView === 'overview' && (
        <>
          {/* Hero Profile Identity */}
          <Card className="relative overflow-hidden p-6 border-border-subtle bg-surface-elevated/80 shadow-md">
            <div className="relative z-10 flex flex-col items-center text-center gap-4">
              <div className="relative group">
                <Avatar
                  seed={profile.avatar_seed || profile.uid}
                  style={profile.avatar_style || 'adventurer'}
                  size="xl"
                />
                <button
                  onClick={handleRandomizeAvatar}
                  disabled={randomizing}
                  title="Acak Avatar Baru"
                  className="absolute bottom-0 right-0 p-2 bg-accent-magic text-white rounded-full shadow-md hover:bg-accent-magic/80 transition-transform active:scale-95"
                >
                  <Shuffle size={14} className={randomizing ? 'animate-spin' : ''} />
                </button>
              </div>

              <div>
                <h2 className="font-heading text-2xl md:text-3xl text-text-primary font-bold">{profile.explorer_name}</h2>
                <div className="flex items-center justify-center gap-2 mt-1">
                  <span className="text-xs font-bold uppercase px-2.5 py-0.5 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/20">
                    {roleLabel}
                  </span>
                  {streakDays > 0 && (
                    <span className="text-xs font-bold uppercase px-2.5 py-0.5 rounded-full bg-orange-500/10 text-orange-400 border border-orange-500/20 inline-flex items-center gap-1">
                      <Flame size={12} className="text-orange-400" /> {streakDays} Hari Streak
                    </span>
                  )}
                </div>
              </div>

              <div className="w-full bg-surface p-4 rounded-xl border border-border-subtle mt-2">
                <div className="flex justify-between text-xs font-bold uppercase tracking-wider mb-2">
                  <span className="text-accent-reward font-extrabold">Level {profile.level ?? 1}</span>
                  <span className="text-text-secondary">{profile.xp ?? 0} Total EXP</span>
                </div>
                <ProgressBar progress={xpPercent} colorClass="bg-accent-reward" />
              </div>
            </div>
          </Card>

          {/* Coin Balance & Shop Quick Access */}
          <Card className="p-6 flex flex-col sm:flex-row items-center justify-between gap-4 bg-gradient-to-r from-accent-reward/10 to-amber-500/5 border-accent-reward/30 shadow-md">
            <div className="flex items-center gap-4 text-center sm:text-left">
              <span className="text-4xl p-3 bg-amber-400/20 rounded-2xl border border-amber-400/30">🪙</span>
              <div>
                <div className="flex items-baseline gap-1">
                  <span className="font-heading text-3xl font-black text-accent-reward">{profile.coins ?? 0}</span>
                  <span className="text-xs font-bold text-text-secondary uppercase">Koin</span>
                </div>
                <p className="text-xs text-text-secondary mt-0.5">Dapat ditukarkan dengan Pulsa, E-Wallet, atau Tunai</p>
              </div>
            </div>

            <Link
              to="/shop"
              className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-accent-reward hover:bg-accent-reward/90 text-surface-ground font-bold text-sm shadow-md transition-transform active:scale-95"
            >
              <ShoppingBag size={16} /> Tukar Hadiah
            </Link>
          </Card>
        </>
      )}

      {activeView === 'settings' && (
        <div className="flex flex-col gap-6 animate-in fade-in slide-in-from-right-4">
          <Card className="p-6 flex flex-col gap-4">
            <h3 className="font-heading text-xl text-text-primary mb-2 flex items-center gap-2">
              <span>👤</span> Informasi Akun
            </h3>
            
            <div className="flex items-center justify-between border-b border-border-subtle/50 pb-4">
              <div>
                <p className="text-sm font-bold text-text-primary">Avatar Karakter</p>
                <p className="text-xs text-text-secondary">Ubah gaya wajah karakter secara acak</p>
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

          <Button variant="danger" onClick={logout} className="w-full shadow-md flex items-center justify-center gap-2">
            <LogOut size={16} /> Keluar (Sign Out)
          </Button>
        </div>
      )}
    </div>
  )
}
