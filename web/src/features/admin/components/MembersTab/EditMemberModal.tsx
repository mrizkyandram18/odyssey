import React, { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Edit3, KeyRound, Copy, Check, ShieldAlert, User, Shield, Coins } from 'lucide-react'
import type { MemberView } from '../../../../shared/types'
import { adminMembersApi } from '../../../../shared/lib/api'
import { MemberProfileSection } from './sections/MemberProfileSection'
import { MemberCapSection } from './sections/MemberCapSection'
import { MemberPayoutSection } from './sections/MemberPayoutSection'

type EditTab = 'profile' | 'cap' | 'payout'

interface EditMemberModalProps {
  member: MemberView | null
  form: {
    explorer_name: string
    role: 'ADMIN' | 'MEMBER'
    is_active: boolean
    reset_device: boolean
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
      explorer_name: string
      role: 'ADMIN' | 'MEMBER'
      is_active: boolean
      reset_device: boolean
      monthly_coin_target: number
      monthly_earning_cap: number
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
  const [activeTab, setActiveTab] = useState<EditTab>('profile')
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
    setTempPassword(null)
    setShowResetSuccess(false)
    setShowResetConfirm(false)
    setCopied(false)
    onClose()
  }

  const isRegularMember = member.role === 'MEMBER' || (member.role as string) === 'SEEKER'

  return (
    <AnimatePresence>
      <div key="edit-member-modal-backdrop" className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <motion.div
          initial={{ scale: 0.96, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.96, opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-3.5 border-b border-border-subtle bg-surface">
            <div>
              <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Edit3 className="w-4 h-4 text-accent-magic" />
                <span>Edit Anggota @{member.username}</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5">
                Profil & Pengaturan Akun
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

          {/* Tab Navigation */}
          {isRegularMember && (
            <div className="flex items-center border-b border-border-subtle bg-surface-elevated/40 px-5 pt-2 gap-2 text-xs">
              <button
                type="button"
                onClick={() => setActiveTab('profile')}
                className={`pb-2 px-2 font-bold transition-all border-b-2 cursor-pointer flex items-center gap-1.5 ${
                  activeTab === 'profile'
                    ? 'border-accent-magic text-accent-magic'
                    : 'border-transparent text-text-secondary hover:text-text-primary'
                }`}
              >
                <User className="w-3.5 h-3.5" />
                <span>Profil & Akun</span>
              </button>
              <button
                type="button"
                onClick={() => setActiveTab('cap')}
                className={`pb-2 px-2 font-bold transition-all border-b-2 cursor-pointer flex items-center gap-1.5 ${
                  activeTab === 'cap'
                    ? 'border-accent-magic text-accent-magic'
                    : 'border-transparent text-text-secondary hover:text-text-primary'
                }`}
              >
                <Shield className="w-3.5 h-3.5" />
                <span>Batas Koin</span>
              </button>
              <button
                type="button"
                onClick={() => setActiveTab('payout')}
                className={`pb-2 px-2 font-bold transition-all border-b-2 cursor-pointer flex items-center gap-1.5 ${
                  activeTab === 'payout'
                    ? 'border-accent-magic text-accent-magic'
                    : 'border-transparent text-text-secondary hover:text-text-primary'
                }`}
              >
                <Coins className="w-3.5 h-3.5" />
                <span>Pencairan</span>
              </button>
            </div>
          )}

          {/* Form Content */}
          <form onSubmit={handleSubmit} className="p-5 overflow-y-auto flex-1 space-y-4">
            {(!isRegularMember || activeTab === 'profile') && (
              <MemberProfileSection
                member={member}
                form={form}
                setForm={setForm}
                onOpenResetPassword={handleOpenResetConfirm}
              />
            )}

            {isRegularMember && activeTab === 'cap' && (
              <MemberCapSection
                member={member}
                form={form}
                setForm={setForm}
              />
            )}

            {isRegularMember && activeTab === 'payout' && (
              <MemberPayoutSection
                form={form}
                setForm={setForm}
              />
            )}

            {/* Sticky Action Buttons */}
            <div className="pt-4 border-t border-border-subtle flex items-center justify-end gap-2.5">
              <button
                type="button"
                onClick={handleCloseModal}
                className="px-4 py-2.5 rounded-xl border border-border-subtle bg-surface-elevated text-xs font-bold text-text-secondary hover:text-text-primary hover:bg-surface transition-colors cursor-pointer"
              >
                Batal
              </button>
              <button
                type="submit"
                disabled={isSaving}
                className="px-5 py-2.5 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 active:scale-[0.98] shadow-xs transition-all disabled:opacity-50 cursor-pointer"
              >
                {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>

      {/* Confirmation Modal for Reset Password */}
      {showResetConfirm && (
        <div key="reset-confirm-backdrop" className="fixed inset-0 z-60 flex items-center justify-center p-4 bg-black/70 backdrop-blur-xs">
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            className="w-full max-w-sm bg-surface border border-border-subtle rounded-2xl p-5 shadow-2xl space-y-4"
          >
            <div className="flex items-center gap-3 text-status-error">
              <div className="w-10 h-10 rounded-xl bg-status-error/10 flex items-center justify-center shrink-0">
                <ShieldAlert className="w-5 h-5 text-status-error" />
              </div>
              <div>
                <h4 className="font-bold text-sm text-text-primary">Reset Password?</h4>
                <p className="text-xs text-text-secondary">Tindakan ini memerlukan perhatian.</p>
              </div>
            </div>

            <p className="text-xs text-text-secondary leading-relaxed">
              Password akun <strong>{member.explorer_name}</strong> (@{member.username}) akan direset ke password sementara otomatis. Anda harus memberikan password baru tersebut kepada pengguna.
            </p>

            {resetError && (
              <p className="text-xs text-status-error p-2.5 rounded-lg bg-status-error/10 border border-status-error/20">
                {resetError}
              </p>
            )}

            <div className="flex items-center justify-end gap-2 pt-2">
              <button
                type="button"
                data-testid="reset-cancel-button"
                disabled={isResetting}
                onClick={handleCancelReset}
                className="px-3.5 py-2 rounded-xl text-xs font-bold text-text-secondary hover:bg-surface-elevated transition-colors cursor-pointer"
              >
                Batal
              </button>
              <button
                type="button"
                data-testid="reset-confirm-button"
                disabled={isResetting}
                onClick={handleConfirmReset}
                className="px-4 py-2 rounded-xl bg-status-error text-white text-xs font-bold hover:brightness-110 active:scale-95 transition-all disabled:opacity-50 cursor-pointer flex items-center gap-1.5"
              >
                <KeyRound className="w-3.5 h-3.5" />
                <span>{isResetting ? 'Mereset...' : 'Ya, Reset Password'}</span>
              </button>
            </div>
          </motion.div>
        </div>
      )}

      {/* Success Modal Showing Temporary Password */}
      {showResetSuccess && tempPassword && (
        <div key="reset-success-backdrop" className="fixed inset-0 z-60 flex items-center justify-center p-4 bg-black/70 backdrop-blur-xs">
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            className="w-full max-w-sm bg-surface border border-border-subtle rounded-2xl p-5 shadow-2xl space-y-4"
          >
            <div className="flex items-center gap-2.5 text-status-success">
              <div className="w-9 h-9 rounded-xl bg-status-success/15 flex items-center justify-center shrink-0">
                <Check className="w-5 h-5 text-status-success" />
              </div>
              <h4 className="font-bold text-sm text-text-primary">Password Berhasil Di-reset</h4>
            </div>

            <p className="text-xs text-text-secondary leading-relaxed">
              Berikan password sementara berikut kepada anggota. Pengguna akan diminta mengganti password pada saat login berikutnya.
            </p>

            <div className="p-3 rounded-xl bg-surface-elevated border border-border-subtle flex items-center justify-between gap-2">
              <span
                data-testid="temporary-password-display"
                className="font-mono text-sm font-bold text-text-primary tracking-wider"
              >
                {tempPassword}
              </span>
              <button
                type="button"
                data-testid="copy-password-button"
                onClick={handleCopy}
                className="px-3 py-1.5 rounded-lg bg-surface border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface-elevated transition-all flex items-center gap-1 cursor-pointer"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-status-success" />
                    <span className="text-status-success">Tersalin</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Salin</span>
                  </>
                )}
              </button>
            </div>

            <div className="pt-2 flex justify-end">
              <button
                type="button"
                data-testid="close-success-button"
                onClick={handleCloseSuccess}
                className="w-full py-2.5 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 transition-all cursor-pointer"
              >
                Selesai & Tutup
              </button>
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
