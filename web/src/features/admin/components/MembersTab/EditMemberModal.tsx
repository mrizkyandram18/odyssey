import React, { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Edit3, Smartphone, KeyRound, Copy, Check, ShieldAlert } from 'lucide-react'
import type { MemberView } from '../../../../shared/types'
import { adminMembersApi } from '../../../../shared/lib/api'

interface EditMemberModalProps {
  member: MemberView | null
  form: {
    explorer_name: string
    role: 'ADMIN' | 'MEMBER'
    is_active: boolean
    reset_device: boolean
    monthly_coin_target: number
    payout_frequency: 'THRESHOLD' | 'WEEKLY' | 'MONTHLY'
    minimum_withdrawal_coins: number
    payout_weekday: number
    payout_month_start_day: number
    payout_month_end_day: number
  }
  setForm: React.Dispatch<
    React.SetStateAction<{
      explorer_name: string
      role: 'ADMIN' | 'MEMBER'
      is_active: boolean
      reset_device: boolean
      monthly_coin_target: number
      payout_frequency: 'THRESHOLD' | 'WEEKLY' | 'MONTHLY'
      minimum_withdrawal_coins: number
      payout_weekday: number
      payout_month_start_day: number
      payout_month_end_day: number
    }>
  >
  isSaving: boolean
  onClose: () => void
  onSave: () => void
}

export const EditMemberModal: React.FC<EditMemberModalProps> = ({
  member,
  form,
  setForm,
  isSaving,
  onClose,
  onSave,
}) => {
  const [showResetConfirm, setShowResetConfirm] = useState(false)
  const [isResetting, setIsResetting] = useState(false)
  const [tempPassword, setTempPassword] = useState<string | null>(null)
  const [showResetSuccess, setShowResetSuccess] = useState(false)
  const [copied, setCopied] = useState(false)
  const [resetError, setResetError] = useState<string | null>(null)

  if (!member) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave()
  }

  const handleOpenResetConfirm = () => {
    setResetError(null)
    setShowResetConfirm(true)
  }

  const handleCancelReset = () => {
    setShowResetConfirm(false)
    setResetError(null)
  }

  const handleConfirmReset = async () => {
    setIsResetting(true)
    setResetError(null)
    try {
      const res = await adminMembersApi.resetPassword(member.uid)
      setTempPassword(res.temporary_password)
      setShowResetSuccess(true)
      setShowResetConfirm(false)
      setCopied(false)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal mereset password'
      setResetError(msg)
    } finally {
      setIsResetting(false)
    }
  }

  const handleCopy = async () => {
    if (!tempPassword) return
    try {
      await navigator.clipboard.writeText(tempPassword)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // fallback
      const el = document.createElement('textarea')
      el.value = tempPassword
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleCloseSuccess = () => {
    setShowResetSuccess(false)
    setTempPassword(null)
    setCopied(false)
  }

  const handleCloseModal = () => {
    // Ensure temp password is cleared when main modal closes
    setTempPassword(null)
    setShowResetSuccess(false)
    setShowResetConfirm(false)
    setCopied(false)
    onClose()
  }

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <motion.div
          initial={{ scale: 0.96, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.96, opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="w-full max-w-md bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
            <div>
              <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Edit3 className="w-4 h-4 text-accent-magic" />
                <span>Edit Anggota @{member.username}</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5 font-mono">
                UID: {member.uid}
              </p>
            </div>
            <button
              type="button"
              onClick={handleCloseModal}
              className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors cursor-pointer"
              title="Tutup"
              aria-label="Tutup"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="p-5 space-y-3.5 overflow-y-auto flex-1">
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Nama Lengkap</label>
              <input
                type="text"
                required
                value={form.explorer_name}
                onChange={(e) => setForm({ ...form, explorer_name: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
              />
            </div>

            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Role</label>
              <select
                value={form.role}
                onChange={(e) =>
                  setForm({ ...form, role: e.target.value as 'ADMIN' | 'MEMBER' })
                }
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
              >
                <option value="MEMBER">MEMBER (Anggota biasa)</option>
                <option value="ADMIN">ADMIN (Administrator)</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Status Akun (Reversibel — tidak ada hapus pengguna)</label>
              <select
                data-testid="member-status-select"
                value={form.is_active ? 'active' : 'inactive'}
                onChange={(e) =>
                  setForm({ ...form, is_active: e.target.value === 'active' })
                }
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
              >
                <option value="active">● AKTIF (Bisa Login & Mengerjakan Tugas)</option>
                <option value="inactive">○ BLOKIR (Akses diblokir, riwayat tetap ada)</option>
              </select>
              <p className="text-[11px] text-text-secondary">Blokir bersifat reversibel; histori tugas, koin, dan klaim tetap auditable.</p>
            </div>

            {(member.role === 'MEMBER' || member.role === 'SEEKER') && (
              <>
                <div className="space-y-2 p-3 rounded-xl bg-surface-elevated border border-border-subtle">
                  <label className="text-xs font-bold text-text-secondary">Target Koin Bulanan</label>
                  <input
                    type="number"
                    min={1}
                    max={10000}
                    value={form.monthly_coin_target}
                    onChange={(e) => setForm({ ...form, monthly_coin_target: parseInt(e.target.value || '3200', 10) })}
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                  />
                  {member.monthly_coin_target !== undefined && (
                    <p className="text-[11px] text-text-secondary">Target: {member.monthly_coin_target} • Terpakai bulan ini: {member.earned_this_period ?? 0}</p>
                  )}
                  <p className="text-[11px] text-text-secondary">Sistem akan menghitung pembagian koin otomatis berdasarkan target dan bobot task. Perubahan berlaku bulan ini.</p>
                </div>
                <div className="space-y-2 p-3 rounded-xl bg-violet-50 dark:bg-violet-950/20 border border-violet-200">
                  <label className="text-xs font-bold text-text-secondary">Pengaturan Pencairan Per-User</label>
                  <div className="grid grid-cols-2 gap-2">
                    <div className="space-y-1">
                      <label className="text-[11px] font-bold text-text-secondary">Frekuensi</label>
                      <select value={form.payout_frequency} onChange={(e) => setForm({ ...form, payout_frequency: e.target.value as any })} className="w-full p-2 rounded-lg border border-border-subtle text-xs font-bold">
                        <option value="THRESHOLD">THRESHOLD</option>
                        <option value="WEEKLY">WEEKLY</option>
                        <option value="MONTHLY">MONTHLY</option>
                      </select>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[11px] font-bold text-text-secondary">Min Withdrawal</label>
                      <input type="number" min={1} max={100000} value={form.minimum_withdrawal_coins} onChange={(e) => setForm({ ...form, minimum_withdrawal_coins: parseInt(e.target.value || '500', 10) })} className="w-full p-2 rounded-lg border border-border-subtle text-xs" />
                    </div>
                  </div>
                  {form.payout_frequency === 'WEEKLY' && (
                    <div className="space-y-1">
                      <label className="text-[11px] font-bold text-text-secondary">Hari Payout (0=Sun..6=Sat)</label>
                      <input type="number" min={0} max={6} value={form.payout_weekday} onChange={(e) => setForm({ ...form, payout_weekday: parseInt(e.target.value || '1', 10) })} className="w-full p-2 rounded-lg border border-border-subtle text-xs" />
                    </div>
                  )}
                  {form.payout_frequency === 'MONTHLY' && (
                    <div className="grid grid-cols-2 gap-2">
                      <div className="space-y-1">
                        <label className="text-[11px] font-bold text-text-secondary">Start Day 1-31</label>
                        <input type="number" min={1} max={31} value={form.payout_month_start_day} onChange={(e) => setForm({ ...form, payout_month_start_day: parseInt(e.target.value || '24', 10) })} className="w-full p-2 rounded-lg border border-border-subtle text-xs" />
                      </div>
                      <div className="space-y-1">
                        <label className="text-[11px] font-bold text-text-secondary">End Day 1-31</label>
                        <input type="number" min={1} max={31} value={form.payout_month_end_day} onChange={(e) => setForm({ ...form, payout_month_end_day: parseInt(e.target.value || '26', 10) })} className="w-full p-2 rounded-lg border border-border-subtle text-xs" />
                      </div>
                    </div>
                  )}
                  {member.payout_frequency && (
                    <p className="text-[11px] text-text-secondary">Efektif: {member.payout_frequency} • {member.minimum_withdrawal_coins} koin • source: {member.payout_config_source}</p>
                  )}
                </div>
              </>
            )}

            {/* Reset Password section - separate from device binding */}
            <div className="p-3.5 rounded-xl bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800/40 space-y-2.5">
              <div className="flex items-start gap-2.5">
                <KeyRound className="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-bold text-text-primary">Reset Password</p>
                  <p className="text-[11px] text-text-secondary leading-relaxed">
                    Atur ulang password akun member jika lupa password.
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={handleOpenResetConfirm}
                disabled={isResetting}
                data-testid="reset-password-button"
                className="w-full py-2.5 rounded-xl bg-white dark:bg-surface border border-amber-300 dark:border-amber-700 text-amber-700 dark:text-amber-300 font-bold text-xs hover:bg-amber-50 dark:hover:bg-amber-900/20 active:scale-[0.98] transition-all cursor-pointer flex items-center justify-center gap-1.5 disabled:opacity-50"
              >
                <KeyRound className="w-3.5 h-3.5" />
                Reset Password
              </button>
              {resetError && !showResetConfirm && (
                <p className="text-[11px] text-red-600 dark:text-red-400 font-medium">{resetError}</p>
              )}
            </div>

            {/* Device reset option */}
            <div className="p-3 rounded-xl bg-accent-gold/10 border border-accent-gold/20 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5">
                <Smartphone className="w-4 h-4 text-accent-gold shrink-0" />
                <div>
                  <p className="text-xs font-bold text-text-primary">Reset Binding Perangkat</p>
                  <p className="text-[11px] text-text-secondary">
                    Izinkan akun login di HP/perangkat baru
                  </p>
                </div>
              </div>
              <input
                type="checkbox"
                id="reset-device-checkbox"
                checked={form.reset_device}
                onChange={(e) => setForm({ ...form, reset_device: e.target.checked })}
                className="w-4 h-4 rounded border-border-subtle text-accent-magic focus:ring-accent-magic cursor-pointer"
              />
            </div>

            {/* Sticky Footer */}
            <div className="sticky bottom-0 -mx-5 -mb-5 mt-4 px-5 py-3.5 border-t border-border-subtle bg-surface flex gap-3">
              <button
                type="button"
                onClick={handleCloseModal}
                className="flex-1 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface transition-colors cursor-pointer"
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={isSaving}
                className="flex-1 py-2.5 rounded-xl bg-accent-magic text-white font-bold text-xs shadow-xs hover:brightness-110 disabled:opacity-50 transition-all cursor-pointer"
              >
                {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>

      {/* Confirmation Dialog */}
      <AnimatePresence>
        {showResetConfirm && (
          <div className="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              transition={{ duration: 0.15 }}
              className="w-full max-w-sm bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden"
              role="dialog"
              aria-modal="true"
              aria-labelledby="reset-password-title"
            >
              <div className="p-5 space-y-3">
                <div className="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center mx-auto">
                  <ShieldAlert className="w-5 h-5 text-amber-600 dark:text-amber-400" />
                </div>
                <h4 id="reset-password-title" className="text-sm font-bold text-text-primary text-center">
                  Reset Password?
                </h4>
                <p className="text-xs text-text-secondary text-center leading-relaxed">
                  Password akun <span className="font-bold text-text-primary">{member.explorer_name}</span> akan diganti dengan temporary password baru.
                  <br />
                  <br />
                  Member akan diwajibkan mengganti password saat login berikutnya.
                </p>
                {resetError && (
                  <p className="text-xs text-red-600 dark:text-red-400 text-center font-medium">{resetError}</p>
                )}
              </div>
              <div className="flex gap-3 p-4 border-t border-border-subtle bg-surface">
                <button
                  type="button"
                  onClick={handleCancelReset}
                  disabled={isResetting}
                  data-testid="reset-cancel-button"
                  className="flex-1 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface transition-colors cursor-pointer disabled:opacity-50"
                >
                  Batal
                </button>
                <button
                  type="button"
                  onClick={handleConfirmReset}
                  disabled={isResetting}
                  data-testid="reset-confirm-button"
                  className="flex-1 py-2.5 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-bold text-xs shadow-xs transition-colors cursor-pointer disabled:opacity-50"
                >
                  {isResetting ? 'Memproses...' : 'Reset Password'}
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>

      {/* Success Result */}
      <AnimatePresence>
        {showResetSuccess && tempPassword && (
          <div className="fixed inset-0 z-[70] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              transition={{ duration: 0.15 }}
              className="w-full max-w-sm bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden"
              role="dialog"
              aria-modal="true"
              aria-labelledby="reset-success-title"
            >
              <div className="p-5 space-y-4">
                <div className="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto">
                  <Check className="w-5 h-5 text-green-600 dark:text-green-400" />
                </div>
                <h4 id="reset-success-title" className="text-sm font-bold text-text-primary text-center">
                  Password Berhasil Di-reset
                </h4>

                <div>
                  <p className="text-xs font-semibold text-text-secondary mb-1.5">Temporary password:</p>
                  <div
                    data-testid="temporary-password-display"
                    className="flex items-center gap-2 p-3 rounded-xl bg-surface-elevated border border-border-subtle"
                  >
                    <span className="flex-1 font-mono text-sm font-bold text-text-primary break-all select-all">
                      {tempPassword}
                    </span>
                    <button
                      type="button"
                      onClick={handleCopy}
                      data-testid="copy-password-button"
                      className="shrink-0 px-3 py-1.5 rounded-lg bg-white dark:bg-surface border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface-elevated transition-colors cursor-pointer flex items-center gap-1"
                    >
                      {copied ? (
                        <>
                          <Check className="w-3.5 h-3.5 text-green-600" />
                          <span className="text-green-600">Tersalin</span>
                        </>
                      ) : (
                        <>
                          <Copy className="w-3.5 h-3.5" />
                          Copy
                        </>
                      )}
                    </button>
                  </div>
                  {copied && (
                    <p className="text-[11px] text-green-600 dark:text-green-400 mt-1.5 font-medium flex items-center gap-1">
                      <Check className="w-3 h-3" /> Tersalin
                    </p>
                  )}
                </div>

                <div className="p-3 rounded-xl bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800/30">
                  <p className="text-[11px] text-amber-800 dark:text-amber-300 leading-relaxed">
                    Berikan password ini kepada member. Password hanya ditampilkan sekarang dan member wajib menggantinya saat login. Sampaikan secara aman.
                  </p>
                </div>
              </div>
              <div className="p-4 border-t border-border-subtle bg-surface">
                <button
                  type="button"
                  onClick={handleCloseSuccess}
                  data-testid="close-success-button"
                  className="w-full py-2.5 rounded-xl bg-accent-magic text-white font-bold text-xs hover:brightness-110 transition-all cursor-pointer"
                >
                  Tutup
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </AnimatePresence>
  )
}
