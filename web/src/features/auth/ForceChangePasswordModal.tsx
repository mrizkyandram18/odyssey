import { useState } from 'react'
import { profileApi } from '../../shared/lib/api'
import { useSessionContext } from '../../app/SessionProvider'

export function ForceChangePasswordModal() {
  const { profile, refreshProfile } = useSessionContext()
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Only show for MEMBER role with must_change_password flag
  if (!profile || profile.role !== 'MEMBER' || !profile.must_change_password) {
    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (newPassword.length < 6) {
      setError('Kata sandi minimal 6 karakter')
      return
    }
    if (newPassword !== confirmPassword) {
      setError('Konfirmasi kata sandi tidak cocok')
      return
    }

    try {
      setLoading(true)
      await profileApi.changePassword(newPassword)
      await refreshProfile()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal mengubah kata sandi'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm px-4">
      <div className="w-full max-w-sm bg-surface-elevated rounded-2xl shadow-2xl p-6 border border-border-subtle">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-amber-500/15 mb-3">
            <svg className="w-7 h-7 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <h2 className="text-lg font-bold text-text-primary">Ubah Kata Sandi</h2>
          <p className="text-sm text-text-secondary mt-1">
            Untuk keamanan akun, harap ubah kata sandi kamu sekarang sebelum melanjutkan.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label className="text-xs font-semibold text-text-secondary uppercase tracking-wide">
              Kata Sandi Baru
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={e => setNewPassword(e.target.value)}
              placeholder="Minimal 6 karakter"
              autoFocus
              required
              disabled={loading}
              className="w-full px-4 py-3 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition"
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-semibold text-text-secondary uppercase tracking-wide">
              Konfirmasi Kata Sandi Baru
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={e => setConfirmPassword(e.target.value)}
              placeholder="Ulangi kata sandi baru"
              required
              disabled={loading}
              className="w-full px-4 py-3 rounded-xl border border-border-subtle bg-bg-app text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition"
            />
          </div>

          {error && (
            <div className="px-4 py-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 rounded-xl bg-primary text-white font-semibold text-sm transition hover:bg-primary/90 active:scale-95 disabled:opacity-60 disabled:cursor-not-allowed"
          >
            {loading ? 'Menyimpan...' : 'Simpan Kata Sandi'}
          </button>
        </form>
      </div>
    </div>
  )
}
