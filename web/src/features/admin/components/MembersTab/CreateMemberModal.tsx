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
    monthly_earning_cap: number
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
      monthly_earning_cap: number
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
                {/* Panduan Edukatif Target vs Batas Koin */}
                <div className="p-3 rounded-xl bg-accent-magic/5 border border-accent-magic/20 space-y-1.5 text-xs">
                  <p className="font-bold text-accent-magic text-[11px] uppercase tracking-wider">
                    Panduan Target vs Batas Koin
                  </p>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[11px] text-text-secondary leading-relaxed">
                    <div className="p-2 rounded-lg bg-surface border border-border-subtle">
                      <strong className="text-text-primary block mb-0.5">🎯 Target Koin</strong>
                      Estimasi reward bulanan. Sistem otomatis membagi koin pada tugas harian agar mendekati angka ini.
                    </div>
                    <div className="p-2 rounded-lg bg-surface border border-border-subtle">
                      <strong className="text-text-primary block mb-0.5">🛡️ Batas Koin (Cap)</strong>
                      Plafon maksimal pengaman budget jika anggota banyak mengerjakan bonus task/streak.
                    </div>
                  </div>
                  <p className="text-[10px] text-text-secondary">
                    💡 <strong>Rekomendasi:</strong> Samakan Batas Koin dengan Target, atau beri toleransi +10%–20%.
                  </p>
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-bold text-text-secondary">Target Koin Bulanan (0 = tanpa payout)</label>
                  <input
                    type="number"
                    min={0}
                    max={10000}
                    required
                    value={form.monthly_coin_target}
                    onChange={(e) => setForm({ ...form, monthly_coin_target: parseInt(e.target.value || '0', 10) })}
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                  />
                  <p className="text-[11px] text-text-secondary">Sistem akan menghitung pembagian koin otomatis berdasarkan target ini. Nilai 0 berarti tanpa payout koin.</p>
                </div>

                <div className="space-y-3 p-3.5 rounded-xl bg-surface-elevated border border-border-subtle">
                  <div className="flex items-center justify-between">
                    <label className="text-xs font-bold text-text-primary">
                      Batas Koin Bulanan (Earning Cap)
                    </label>
                    <span className="text-[10px] font-semibold text-text-secondary">
                      {form.monthly_earning_cap > 0 ? `Batas Khusus (${form.monthly_earning_cap} koin)` : 'Standar Sistem (3.320 koin)'}
                    </span>
                  </div>

                  <div className="space-y-2">
                    <label className="flex items-center gap-2 cursor-pointer text-xs text-text-primary">
                      <input
                        type="radio"
                        name="create_cap_mode"
                        checked={form.monthly_earning_cap === 0}
                        onChange={() => setForm({ ...form, monthly_earning_cap: 0 })}
                        className="text-accent-magic cursor-pointer"
                      />
                      <span>Gunakan batas koin bulanan standar sistem (3.320 koin / bulan)</span>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer text-xs text-text-primary">
                      <input
                        type="radio"
                        name="create_cap_mode"
                        checked={form.monthly_earning_cap > 0}
                        onChange={() => setForm({ ...form, monthly_earning_cap: form.monthly_coin_target > 0 ? form.monthly_coin_target : 3320 })}
                        className="text-accent-magic cursor-pointer"
                      />
                      <span>Atur batas koin khusus untuk anggota ini</span>
                    </label>
                  </div>

                  {form.monthly_earning_cap > 0 && (
                    <div className="pt-2 border-t border-border-subtle/60 space-y-2">
                      <div className="flex items-center justify-between">
                        <label className="text-[11px] font-bold text-text-secondary">
                          Batas Koin Khusus (Per Bulan):
                        </label>
                        <div className="flex items-center gap-1 text-[10px] text-text-secondary">
                          <span>Pilihan:</span>
                          <button
                            type="button"
                            onClick={() => setForm({ ...form, monthly_earning_cap: form.monthly_coin_target > 0 ? form.monthly_coin_target : 3320 })}
                            className="px-2 py-0.5 rounded bg-surface border border-border-subtle hover:text-text-primary"
                          >
                            = Target ({form.monthly_coin_target > 0 ? form.monthly_coin_target : 3320})
                          </button>
                          <button
                            type="button"
                            onClick={() => setForm({ ...form, monthly_earning_cap: Math.round((form.monthly_coin_target > 0 ? form.monthly_coin_target : 3320) * 1.1) })}
                            className="px-2 py-0.5 rounded bg-surface border border-border-subtle hover:text-text-primary"
                          >
                            +10%
                          </button>
                        </div>
                      </div>
                      <div className="relative">
                        <input
                          type="number"
                          min={1}
                          max={10000}
                          value={form.monthly_earning_cap}
                          onChange={(e) => setForm({ ...form, monthly_earning_cap: Math.max(0, parseInt(e.target.value || '0', 10)) })}
                          placeholder="Contoh: 3500"
                          className="w-full p-2.5 pr-24 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                        />
                        <span className="absolute right-3 top-2.5 text-xs text-text-secondary">koin / bulan</span>
                      </div>
                      <p className="text-[10px] text-text-secondary">
                        Batas khusus ini hanya berlaku untuk anggota ini dan menggantikan batas standar sistem.
                      </p>
                    </div>
                  )}

                  <p className="text-[11px] text-text-secondary leading-relaxed">
                    Jika perolehan koin bulan ini mencapai batas, anggota tidak dapat memperoleh koin tambahan dari tugas sampai bulan berikutnya. Saldo yang sudah terkumpul tidak akan berkurang.
                  </p>
                </div>

                <div className="space-y-2 p-3.5 rounded-xl bg-surface-elevated border border-border-subtle">
                  <label className="text-xs font-bold text-text-primary">Ketentuan Pencairan Koin</label>
                  <div className="grid grid-cols-2 gap-2">
                    <div className="space-y-1">
                      <label className="text-[11px] font-bold text-text-secondary">Frekuensi Pencairan</label>
                      <select
                        value={form.payout_frequency === 'MONTHLY' ? 'THRESHOLD' : form.payout_frequency}
                        onChange={(e) => setForm({ ...form, payout_frequency: e.target.value as any })}
                        className="w-full p-2 rounded-lg bg-surface border border-border-subtle text-xs font-bold text-text-primary"
                      >
                        <option value="THRESHOLD">Fleksibel (Kapan Saja)</option>
                        <option value="WEEKLY">Jadwal Mingguan</option>
                      </select>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[11px] font-bold text-text-secondary">Minimal Penarikan (Koin)</label>
                      <input
                        type="number"
                        step={500}
                        min={500}
                        max={100000}
                        value={form.minimum_withdrawal_coins}
                        onChange={(e) => setForm({ ...form, minimum_withdrawal_coins: Math.max(500, parseInt(e.target.value || '500', 10)) })}
                        className="w-full p-2 rounded-lg bg-surface border border-border-subtle text-xs font-mono font-bold"
                      />
                    </div>
                  </div>
                  <p className="text-[10px] text-text-secondary">
                    Penarikan koin dapat diajukan kapan saja setelah mencapai minimal nominal (dalam kelipatan 500 koin).
                  </p>
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
