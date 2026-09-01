import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Edit3, Smartphone } from 'lucide-react'
import type { MemberView } from '../../../../shared/types'

interface EditMemberModalProps {
  member: MemberView | null
  form: {
    explorer_name: string
    role: 'ADMIN' | 'MEMBER'
    is_active: boolean
    reset_device: boolean
  }
  setForm: React.Dispatch<
    React.SetStateAction<{
      explorer_name: string
      role: 'ADMIN' | 'MEMBER'
      is_active: boolean
      reset_device: boolean
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
  if (!member) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave()
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
              <label className="text-xs font-bold text-text-secondary">Status Akun</label>
              <select
                value={form.is_active ? 'active' : 'inactive'}
                onChange={(e) =>
                  setForm({ ...form, is_active: e.target.value === 'active' })
                }
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
              >
                <option value="active">● Aktif (Bisa Login & Mengerjakan Tugas)</option>
                <option value="inactive">○ Nonaktif (Akses Login Diblokir)</option>
              </select>
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
                onClick={onClose}
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
    </AnimatePresence>
  )
}
