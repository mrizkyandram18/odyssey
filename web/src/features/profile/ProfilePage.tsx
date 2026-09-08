import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../shared/components/atoms/Button'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import { apiClient, crewsApi } from '../../shared/lib/api'
import { Avatar } from '../../shared/components/atoms/Avatar'
import { Shuffle, ArrowLeft, LogOut, Banknote, Flame, ShieldCheck } from 'lucide-react'
import { PushNotificationToggle } from '../../shared/components/molecules/PushNotificationToggle'
import type { Explorer } from '../../shared/types'

import { levelProgress } from '../../shared/lib/level'

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
  const [familyMembers, setFamilyMembers] = useState<Explorer[]>([])

  useEffect(() => {
    crewsApi.members().then((m) => setFamilyMembers(m || [])).catch(() => {})
  }, [profile?.family_id])

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

  const isAdmin = profile.role === 'ADMIN' || profile.role === 'GUIDE' || profile.role === 'BUILDER'
  const roleLabel = isAdmin ? 'Administrator' : 'Anggota'
  const streakDays = profile.streak_days ?? 0

  // Shared Change Password — compact grouping
  const changePasswordForm = (
    <div className="rounded-2xl bg-surface border border-border-subtle p-4 flex flex-col gap-3">
      <h3 className="font-bold text-text-primary text-xs flex items-center gap-2">
        <span>🔒</span> Ubah Kata Sandi
      </h3>
      <form onSubmit={handleChangePassword} className="flex flex-col gap-3">
        <div className="flex flex-col gap-1">
          <label className="text-[11px] font-bold tracking-wide uppercase text-text-secondary">Kata Sandi Saat Ini</label>
          <input
            type="password"
            value={currentPassword}
            onChange={e => setCurrentPassword(e.target.value)}
            placeholder="Masukkan kata sandi saat ini"
            required
            disabled={changing}
            className="w-full px-3.5 py-2.5 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent-magic/30 focus:border-accent-magic transition text-sm"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-[11px] font-bold tracking-wide uppercase text-text-secondary">Kata Sandi Baru</label>
          <input
            type="password"
            value={newPassword}
            onChange={e => setNewPassword(e.target.value)}
            placeholder="Minimal 6 karakter"
            required
            disabled={changing}
            className="w-full px-3.5 py-2.5 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent-magic/30 focus:border-accent-magic transition text-sm"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-[11px] font-bold tracking-wide uppercase text-text-secondary">Konfirmasi Kata Sandi Baru</label>
          <input
            type="password"
            value={confirmPassword}
            onChange={e => setConfirmPassword(e.target.value)}
            placeholder="Ulangi kata sandi baru"
            required
            disabled={changing}
            className="w-full px-3.5 py-2.5 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent-magic/30 focus:border-accent-magic transition text-sm"
          />
        </div>
        {changeError && (
          <div className="px-3 py-2 rounded-xl bg-status-error/10 border border-status-error/20 text-status-error text-xs">{changeError}</div>
        )}
        {changeSuccess && (
          <div className="px-3 py-2 rounded-xl bg-status-success/10 border border-status-success/20 text-status-success text-xs">{changeSuccess}</div>
        )}
        <Button type="submit" variant="primary" size="md" disabled={changing} className="w-full min-h-[42px]">
          {changing ? 'Menyimpan...' : 'Simpan Kata Sandi Baru'}
        </Button>
      </form>
    </div>
  )

  // -------------------------------------------------------------------------
  // 1. ADMIN / GUIDE / BUILDER PROFILE VIEW (Clean, No Gamification, No /shop)
  // -------------------------------------------------------------------------
  if (isAdmin) {
    return (
      <div className="flex flex-col gap-4 max-w-2xl mx-auto">
        <header className="flex items-center gap-2">
          <Link
            to="/admin"
            className="p-2 -ml-2 rounded-xl hover:bg-surface text-text-secondary hover:text-text-primary transition-colors"
            aria-label="Kembali ke Admin Panel"
          >
            <ArrowLeft size={18} />
          </Link>
          <h1 className="text-lg font-bold text-text-primary">Akun Administrator</h1>
        </header>

        {/* Admin Identity — lighter */}
        <div className="rounded-2xl bg-surface border border-border-subtle p-5">
          <div className="flex flex-col items-center text-center gap-3">
            <div className="relative">
              <Avatar
                seed={profile.avatar_seed || profile.uid}
                style={profile.avatar_style || 'adventurer'}
                frame={profile.avatar_frame || 'none'}
                effect={profile.equipped_explorer_effect || 'none'}
                size="xl"
              />
              <button
                onClick={handleRandomizeAvatar}
                disabled={randomizing}
                aria-label="Acak avatar"
                className="absolute -bottom-1 -right-1 p-1.5 bg-accent-magic text-white rounded-full shadow-sm hover:brightness-110 transition-all active:scale-95 disabled:opacity-50"
              >
                <Shuffle size={12} className={randomizing ? 'animate-spin' : ''} />
              </button>
            </div>
            <div>
              <h2 className="text-xl font-bold text-text-primary tracking-tight">{profile.explorer_name}</h2>
              <div className="flex items-center justify-center gap-2 mt-1.5 flex-wrap">
                <span className="text-[11px] font-bold px-2.5 py-1 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/15 inline-flex items-center gap-1">
                  <ShieldCheck size={12} />
                  <span>{roleLabel} ({profile.role})</span>
                </span>
              </div>
            </div>
          </div>
        </div>

        <div className="rounded-2xl bg-surface border border-border-subtle p-4 flex flex-col gap-0 divide-y divide-border-subtle">
          <h3 className="font-bold text-text-primary text-xs flex items-center gap-2 pb-3">
            <span>👤</span> Informasi Akun
          </h3>
          <div className="flex items-center justify-between py-3.5 gap-3">
            <div className="min-w-0">
              <p className="text-sm font-bold text-text-primary">Foto Profil</p>
              <p className="text-xs text-text-secondary">Ubah gaya avatar secara acak</p>
            </div>
            <Button
              variant="secondary"
              size="sm"
              className="flex items-center gap-2 shrink-0"
              onClick={handleRandomizeAvatar}
              disabled={randomizing}
            >
              <Shuffle size={14} className={randomizing ? 'animate-spin' : ''} /> Acak Avatar
            </Button>
          </div>
          <div className="py-3.5">
            <p className="text-sm font-bold text-text-primary">ID Pengguna (UID)</p>
            <p className="text-xs text-text-secondary font-mono break-all mt-0.5">{profile.uid}</p>
          </div>
          <div className="pt-3.5">
            <PushNotificationToggle />
          </div>
        </div>

        {/* Change Password Form */}
        {changePasswordForm}

        {/* Sign Out Button */}
        <Button variant="danger" onClick={logout} className="w-full shadow-md flex items-center justify-center gap-2">
          <LogOut size={16} /> Keluar (Sign Out)
        </Button>
      </div>
    )
  }

  // -------------------------------------------------------------------------
  // 2. MEMBER / SEEKER PROFILE VIEW (Full Gamification & Settings)
  // -------------------------------------------------------------------------
  return (
    <div className="flex flex-col gap-4 max-w-2xl mx-auto">
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
          {/* 1. Adventurer Passport Hero Card */}
          <div className="relative overflow-hidden rounded-3xl bg-surface border border-border-subtle p-6 shadow-sm">
            <div className="flex flex-col items-center text-center gap-3">
              {/* Avatar with equipped Frame & Effect */}
              <div className="relative p-1">
                <Avatar
                  seed={profile.avatar_seed || profile.uid}
                  style={profile.avatar_style || 'adventurer'}
                  frame={profile.avatar_frame || 'none'}
                  effect={profile.equipped_explorer_effect || 'none'}
                  size="xl"
                />
                <button
                  onClick={handleRandomizeAvatar}
                  disabled={randomizing}
                  aria-label="Acak avatar"
                  className="absolute -bottom-1 -right-1 p-2 bg-accent-magic text-white rounded-full shadow-md hover:brightness-110 transition-all active:scale-95 disabled:opacity-50 border-2 border-white"
                >
                  <Shuffle size={13} className={randomizing ? 'animate-spin' : ''} />
                </button>
              </div>

              {/* Explorer Name & Badges */}
              <div>
                <h2 className="text-2xl font-extrabold text-text-primary tracking-tight">{profile.explorer_name}</h2>
                <div className="flex items-center justify-center gap-2 mt-1.5 flex-wrap">
                  <span className="text-[11px] font-extrabold px-3 py-1 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/20">
                    {roleLabel}
                  </span>
                  {streakDays > 0 && (
                    <span className="text-[11px] font-extrabold px-3 py-1 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 inline-flex items-center gap-1 shadow-xs">
                      <Flame size={12} className="fill-amber-500 text-amber-500" />
                      <span>{streakDays} Hari Streak</span>
                    </span>
                  )}
                </div>
              </div>

              {/* Tingkat & Bintang Progression Card */}
              {(() => {
                const progress = levelProgress(profile.xp ?? 0, profile.level ?? 1)
                return (
                  <div className="w-full bg-surface-elevated/70 p-4 rounded-2xl border border-border-subtle mt-2 text-left space-y-2">
                    <div className="flex justify-between items-center text-xs font-bold">
                      <span className="text-accent-magic font-extrabold">Level {profile.level ?? 1}</span>
                      <span className="text-text-secondary">{profile.xp ?? 0} XP</span>
                    </div>

                    <div className="flex justify-between items-center text-xs">
                      <span className="font-extrabold text-text-primary">Tingkat {progress.level}</span>
                      <span className="font-semibold text-accent-magic">{progress.have}/{progress.required} Bintang</span>
                    </div>

                    <div className="h-2.5 w-full bg-surface rounded-full overflow-hidden border border-border-subtle/50">
                      <div
                        className="h-full bg-gradient-to-r from-accent-magic to-sky-400 rounded-full transition-all duration-500"
                        style={{ width: `${progress.percent}%` }}
                      />
                    </div>

                    <p className="text-[11px] text-text-secondary text-center leading-relaxed pt-1">
                      {progress.required - progress.have} Bintang lagi menuju Tingkat {progress.level + 1}
                    </p>

                    <div className="pt-2 border-t border-border-subtle flex items-center justify-between text-xs">
                      <span className="text-text-secondary font-medium">Hiasan profil aktif</span>
                      <Link to="/koleksi" className="font-extrabold text-accent-magic hover:underline inline-flex items-center gap-1">
                        <span>🎨 Koleksi Saya</span>
                        <span>→</span>
                      </Link>
                    </div>
                  </div>
                )
              })()}
            </div>
          </div>

          {/* 2. Coin Balance & Redemption Card */}
          <div className="rounded-3xl bg-surface border border-border-subtle p-5 shadow-sm">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-2xl bg-amber-500/10 text-amber-500 flex items-center justify-center text-2xl shrink-0 border border-amber-500/20">
                  🪙
                </div>
                <div className="min-w-0">
                  <p className="text-[11px] font-black uppercase tracking-wider text-text-secondary">Saldo Koin Petualang</p>
                  <p className="flex items-baseline gap-1.5 mt-0.5">
                    <span className="text-2xl font-black text-text-primary leading-none">{profile.coins ?? 0}</span>
                    <span className="text-xs font-bold text-text-secondary">Koin</span>
                  </p>
                  <p className="text-[11px] text-text-secondary mt-0.5">Tukarkan koin menjadi uang tunai atau saldo digital</p>
                </div>
              </div>

              <Link
                to="/shop"
                className="shrink-0 inline-flex items-center gap-1.5 px-4 py-3 rounded-2xl bg-accent-magic hover:brightness-110 text-white font-black text-xs shadow-md shadow-accent-magic/25 transition-all active:scale-95 min-h-[42px]"
              >
                <Banknote size={15} />
                <span>Pencairan Koin</span>
              </Link>
            </div>
          </div>

          {/* 3. Family Showcase Section */}
          <div className="rounded-3xl bg-surface border border-border-subtle p-5 shadow-sm">
            <div className="flex items-center justify-between mb-3.5">
              <div className="flex items-center gap-2">
                <span className="text-lg">👨‍👩‍👧‍👦</span>
                <h3 className="font-extrabold text-text-primary text-sm tracking-tight">
                  Keluarga Petualang
                </h3>
              </div>
              <span className="text-[11px] font-bold text-text-secondary">
                {familyMembers.length > 0 ? `${familyMembers.length} Anggota` : 'Keluarga'}
              </span>
            </div>

            {familyMembers.length === 0 ? (
              <p className="text-xs text-text-secondary text-center py-4 leading-relaxed">
                Jelajahi misi harian bersama seluruh anggota keluargamu!
              </p>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                {familyMembers.map((member) => {
                  const isCurrent = member.uid === profile.uid
                  return (
                    <div
                      key={member.uid}
                      className={`flex items-center gap-3 p-3 rounded-2xl border transition-all ${
                        isCurrent
                          ? 'bg-accent-magic/[0.06] border-accent-magic/30 shadow-xs ring-1 ring-accent-magic/20'
                          : 'bg-surface-elevated/60 border-border-subtle'
                      }`}
                    >
                      <Avatar
                        seed={member.avatar_seed || member.uid}
                        style={member.avatar_style || 'adventurer'}
                        frame={member.avatar_frame || 'none'}
                        effect={member.equipped_explorer_effect || 'none'}
                        size="md"
                      />
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-1.5">
                          <p className="text-xs font-extrabold text-text-primary truncate">
                            {member.explorer_name}
                          </p>
                          {isCurrent && (
                            <span className="text-[9.5px] font-black px-1.5 py-0.2 rounded-md bg-accent-magic/20 text-accent-magic">
                              Kamu
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-2 mt-0.5 text-[11px] text-text-secondary">
                          <span className="font-bold text-accent-magic">⭐ Tingkat {member.level}</span>
                          {member.streak_days && member.streak_days > 0 ? (
                            <span className="inline-flex items-center gap-0.5 text-amber-600 dark:text-amber-400 font-bold">
                              <Flame size={11} className="fill-amber-500 text-amber-500" />
                              <span>{member.streak_days}h</span>
                            </span>
                          ) : null}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </>
      )}

      {activeView === 'settings' && (
        <div className="flex flex-col gap-4">
          <div className="rounded-2xl bg-surface border border-border-subtle p-4 flex flex-col gap-0 divide-y divide-border-subtle">
            <h3 className="font-bold text-text-primary text-xs flex items-center gap-2 pb-3">
              <span>👤</span> Informasi Akun
            </h3>
            <div className="flex items-center justify-between py-3.5 gap-3">
              <div className="min-w-0">
                <p className="text-sm font-bold text-text-primary">Foto Profil</p>
                <p className="text-xs text-text-secondary">Ubah gaya avatar secara acak</p>
              </div>
              <Button
                variant="secondary"
                size="sm"
                className="flex items-center gap-2 shrink-0"
                onClick={handleRandomizeAvatar}
                disabled={randomizing}
              >
                <Shuffle size={14} className={randomizing ? 'animate-spin' : ''} /> Acak Avatar
              </Button>
            </div>
            <div className="py-3.5">
              <p className="text-sm font-bold text-text-primary">ID Pengguna (UID)</p>
              <p className="text-xs text-text-secondary font-mono break-all mt-0.5">{profile.uid}</p>
            </div>
            <div className="pt-3.5">
              <PushNotificationToggle />
            </div>
          </div>

          {/* Change Password Form */}
          {changePasswordForm}

          <Button variant="danger" onClick={logout} className="w-full shadow-md flex items-center justify-center gap-2">
            <LogOut size={16} /> Keluar (Sign Out)
          </Button>
        </div>
      )}
    </div>
  )
}
