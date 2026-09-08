import React from 'react'
import { Shield, KeyRound, Smartphone } from 'lucide-react'
import type { MemberView } from '../../../../../shared/types'

interface MemberProfileSectionProps {
  member: MemberView
  form: {
    explorer_name: string
    role: 'ADMIN' | 'MEMBER'
    is_active: boolean
    reset_device: boolean
  }
  setForm: React.Dispatch<React.SetStateAction<any>>
  onOpenResetPassword: () => void
}

export const MemberProfileSection: React.FC<MemberProfileSectionProps> = ({
  member,
  form,
  setForm,
  onOpenResetPassword,
}) => {
  void member
  return (
    <div className="space-y-4">
      {/* Nama Lengkap */}
      <div className="space-y-1">
        <label htmlFor="input-member-name" className="text-xs font-bold text-text-secondary">
          Nama Lengkap
        </label>
        <input
          id="input-member-name"
          type="text"
          required
          value={form.explorer_name}
          onChange={(e) => setForm({ ...form, explorer_name: e.target.value })}
          className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
        />
      </div>

      {/* Role */}
      <div className="space-y-1">
        <label htmlFor="input-member-role" className="text-xs font-bold text-text-secondary">
          Peran (Role)
        </label>
        <select
          id="input-member-role"
          value={form.role}
          onChange={(e) =>
            setForm({ ...form, role: e.target.value as 'ADMIN' | 'MEMBER' })
          }
          className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
        >
          <option value="MEMBER">Anggota (MEMBER)</option>
          <option value="ADMIN">Administrator (ADMIN)</option>
        </select>
      </div>

      {/* Status Akun */}
      <div className="space-y-1">
        <label htmlFor="input-member-status" className="text-xs font-bold text-text-secondary">
          Status Akun (Dapat diaktifkan/dinonaktifkan kembali)
        </label>
        <select
          id="input-member-status"
          data-testid="member-status-select"
          value={form.is_active ? 'active' : 'inactive'}
          onChange={(e) =>
            setForm({ ...form, is_active: e.target.value === 'active' })
          }
          className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
        >
          <option value="active">● AKTIF (Bisa login & mengerjakan tugas)</option>
          <option value="inactive">○ BLOKIR (Akses login ditutup sementara)</option>
        </select>
        <p className="text-[11px] text-text-secondary">
          Histori tugas, koin, dan pencairan tetap tersimpan aman saat akun diblokir.
        </p>
      </div>

      {/* Keamanan & Akses Akun */}
      <div className="pt-2 border-t border-border-subtle space-y-3">
        <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
          <Shield className="w-3.5 h-3.5 text-accent-magic" />
          <span>Keamanan & Akses Perangkat</span>
        </h4>

        {/* Reset Device Binding */}
        <div className="p-3 rounded-xl bg-surface-elevated border border-border-subtle flex items-start gap-3">
          <input
            type="checkbox"
            id="reset-device-checkbox"
            checked={Boolean(form.reset_device)}
            onChange={(e) =>
              setForm({ ...form, reset_device: e.target.checked })
            }
            className="mt-0.5 w-4 h-4 rounded text-accent-magic focus:ring-accent-magic cursor-pointer"
          />
          <div className="space-y-0.5">
            <label
              htmlFor="reset-device-checkbox"
              className="text-xs font-bold text-text-primary cursor-pointer flex items-center gap-1.5"
            >
              <Smartphone className="w-3.5 h-3.5 text-text-secondary" />
              <span>Reset Binding Perangkat</span>
            </label>
            <p className="text-[11px] text-text-secondary leading-relaxed">
              Izinkan akun login di HP atau browser baru jika perangkat sebelumnya hilang atau diganti.
            </p>
          </div>
        </div>

        {/* Reset Password */}
        <div className="p-3 rounded-xl bg-surface-elevated border border-border-subtle flex items-center justify-between gap-3">
          <div className="space-y-0.5 min-w-0">
            <span className="text-xs font-bold text-text-primary flex items-center gap-1.5">
              <KeyRound className="w-3.5 h-3.5 text-accent-gold" />
              <span>Reset Password Akun</span>
            </span>
            <p className="text-[11px] text-text-secondary">
              Atur ulang password akun member jika lupa password.
            </p>
          </div>

          <button
            type="button"
            data-testid="reset-password-button"
            onClick={onOpenResetPassword}
            className="px-3 py-1.5 rounded-lg bg-surface border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface-elevated hover:text-accent-magic transition-colors cursor-pointer shrink-0"
          >
            Reset Password
          </button>
        </div>
      </div>
    </div>
  )
}
