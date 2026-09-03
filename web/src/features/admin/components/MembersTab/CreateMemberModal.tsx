import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, UserPlus } from 'lucide-react'

interface CreateMemberModalProps {
  isOpen: boolean
  form: {
    username: string
    password: string
    explorer_name: string
    role: 'ADMIN' | 'MEMBER'
    monthly_coin_target: number
    payout_frequency: 'THRESHOLD' | 'WEEKLY' | 'MONTHLY'
    minimum_withdrawal_coins: number
    payout_weekday: number
    payout_month_start_day: number
    payout_month_end_day: number
  }
  setForm: React.Dispatch<
    React.SetStateAction<{
      username: string
      password: string
      explorer_name: string
      role: 'ADMIN' | 'MEMBER'
      monthly_coin_target: number
      payout_frequency: 'THRESHOLD' | 'WEEKLY' | 'MONTHLY'
      minimum_withdrawal_coins: number
      payout_weekday: number
      payout_month_start_day: number
      payout_month_end_day: number
    }>
  >
  isCreating: boolean
  onClose: () => void
  onSubmit: () => void
}

export const CreateMemberModal: React.FC<CreateMemberModalProps> = ({
  isOpen,
  form,
  setForm,
  isCreating,
  onClose,
  onSubmit,
}) => {
  if (!isOpen) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit()
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
                <UserPlus className="w-4 h-4 text-accent-magic" />
                <span>Tambah Anggota Baru</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5">
                Akun baru dapat langsung login pada perangkat anggota.
              </p>
            </div>
            <button
              type="button"
              onClick={onClose}
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
              <label className="text-xs font-bold text-text-secondary">
                Nama Lengkap / Panggilan <span className="text-status-error">*</span>
              </label>
              <input
                type="text"
                required
                placeholder="Contoh: Andi Wijaya"
                value={form.explorer_name}
                onChange={(e) => setForm({ ...form, explorer_name: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
              />
            </div>

            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Username (untuk Login) <span className="text-status-error">*</span>
              </label>
              <input
                type="text"
                required
                placeholder="Contoh: andiwijaya"
                value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
            </div>

            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Password Awal <span className="text-status-error">*</span>
              </label>
              <input
                type="password"
                required
                minLength={6}
                placeholder="Minimal 6 karakter"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic font-mono"
              />
            </div>

            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Role Akun</label>
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

            {form.role === 'MEMBER' && (
              <>
                <div className="space-y-1">
                  <label className="text-xs font-bold text-text-secondary">Target Koin Bulanan</label>
                  <input
                    type="number"
                    min={1}
                    max={10000}
                    required
                    value={form.monthly_coin_target}
                    onChange={(e) => setForm({ ...form, monthly_coin_target: parseInt(e.target.value || '3200', 10) })}
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                  />
                  <p className="text-[11px] text-text-secondary">Sistem akan menghitung pembagian koin otomatis berdasarkan target dan bobot task. Default 3200.</p>
                </div>
                <div className="space-y-2 p-3 rounded-xl bg-violet-50 dark:bg-violet-950/20 border border-violet-200">
                  <label className="text-xs font-bold text-text-secondary">Pengaturan Pencairan (Per-User Payout Policy)</label>
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
                </div>
              </>
            )}

            {/* Sticky Footer */}
            <div className="sticky bottom-0 -mx-5 -mb-5 mt-4 px-5 py-3.5 border-t border-border-subtle bg-surface flex gap-3">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs hover:bg-surface transition-colors cursor-pointer"
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={isCreating}
                className="flex-1 py-2.5 rounded-xl bg-accent-magic text-white font-bold text-xs shadow-xs hover:brightness-110 disabled:opacity-50 transition-all cursor-pointer"
              >
                {isCreating ? 'Membuat...' : 'Buat Akun'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  )
}
