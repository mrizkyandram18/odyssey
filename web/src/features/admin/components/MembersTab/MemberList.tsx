import React from 'react'
import { Users, UserPlus, Coins, Edit3, Trash2, ChevronLeft, ChevronRight, RefreshCw, Shield, Sparkles, Calendar, Clock } from 'lucide-react'
import { useAdminMembers } from '../../hooks/useAdminMembers'
import { CreateMemberModal } from './CreateMemberModal'
import { EditMemberModal } from './EditMemberModal'
import { Avatar } from '../../../../shared/components/atoms/Avatar'

function formatCycle(start?: string, end?: string): string {
  if (!start || !end) return '—'
  try {
    const s = new Date(start)
    const e = new Date(end)
    const opts: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'short' }
    return `${s.toLocaleDateString('id-ID', opts)} – ${e.toLocaleDateString('id-ID', opts)}`
  } catch {
    return `${start} – ${end}`
  }
}

function formatLastTask(member: any): { text: string; sub?: string } {
  if (!member.last_completed_date) return { text: '—' }
  try {
    const d = new Date(member.last_completed_date)
    const text = d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
    const sub = member.completed_tasks_current_cycle ? `${member.completed_tasks_current_cycle} task` : undefined
    return { text, sub }
  } catch {
    return { text: member.last_completed_date }
  }
}

function formatInactive(member: any): string {
  if (!member.is_active) return '—'
  if (member.inactivity_status === 'NO_ACTIVITY_THIS_CYCLE') return 'Belum ada aktivitas'
  if (member.inactive_days == null) return '—'
  return `${member.inactive_days} hari`
}

function getStatusConfig(member: any): { label: string; variant: 'success' | 'warning' | 'error' | 'default'; dot?: boolean } {
  if (!member.is_active || member.inactivity_status === 'BLOCKED') {
    return { label: 'DIBLOKIR', variant: 'error' }
  }
  switch (member.inactivity_status) {
    case 'INACTIVE':
      return { label: 'PERLU REVIEW', variant: 'warning' }
    case 'NO_ACTIVITY_THIS_CYCLE':
      return { label: 'BELUM AKTIF', variant: 'default' }
    default:
      return { label: 'AKTIF', variant: 'success', dot: true }
  }
}

function isAdminRole(role: string): boolean {
  return role === 'ADMIN' || role === 'GUIDE' || role === 'BUILDER'
}

function RoleBadge({ role }: { role: string }) {
  const normalized = isAdminRole(role) ? 'ADMIN' : 'MEMBER'
  const isAdmin = normalized === 'ADMIN'
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold tracking-wide ${
        isAdmin
          ? 'bg-accent-magic/15 text-accent-magic border border-accent-magic/20'
          : 'bg-surface-elevated text-text-secondary border border-border-subtle'
      }`}
    >
      <Shield className="h-3 w-3" aria-hidden="true" />
      {normalized}
    </span>
  )
}

function StatusBadge({ member }: { member: any }) {
  const cfg = getStatusConfig(member)
  const variantClasses: Record<string, string> = {
    success: 'bg-status-success/15 text-status-success border border-status-success/20',
    warning: 'bg-accent-gold/15 text-accent-gold border border-accent-gold/20',
    error: 'bg-status-error/15 text-status-error border border-status-error/20',
    default: 'bg-surface-elevated text-text-secondary border border-border-subtle',
  }
  return (
    <span
      data-testid={`member-status-${member.uid}`}
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-bold ${variantClasses[cfg.variant]}`}
    >
      {cfg.variant === 'success' && cfg.dot && <span className="h-1.5 w-1.5 rounded-full bg-status-success" aria-hidden="true" />}
      {cfg.label}
    </span>
  )
}

export interface MemberListProps {
  controller?: ReturnType<typeof useAdminMembers>
}

export const MemberList: React.FC<MemberListProps> = ({ controller }) => {
  const defaultController = useAdminMembers({ enabled: !controller })
  const {
    members,
    pagination,
    isFetching,
    error,
    fetchMembers,
    isCreateModalOpen,
    newMember,
    setNewMember,
    isCreating,
    openCreateModal,
    closeCreateModal,
    handleCreateMember,
    selectedMember,
    editMemberForm,
    setEditMemberForm,
    isSavingEdit,
    openEditModal,
    closeEditModal,
    handleSaveEditMember,
    handleBlock,
    handleUnblock,
    handleDelete,
    processingId,
  } = controller || defaultController

  return (
    <div className="space-y-4">
      {error && (
        <div className="flex items-center justify-between rounded-2xl border border-status-error/30 bg-status-error/10 px-4 py-3 text-xs text-status-error">
          <span>{error}</span>
          <button type="button" onClick={() => fetchMembers(pagination.page)} className="font-bold underline cursor-pointer">
            Coba lagi
          </button>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col gap-3 rounded-2xl border border-border-subtle bg-surface p-4 shadow-xs sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <h3 className="flex items-center gap-2 text-sm font-bold text-text-primary">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-accent-magic/10 text-accent-magic">
              <Users className="h-4 w-4" aria-hidden="true" />
            </span>
            <span>Daftar Anggota</span>
            <span className="rounded-full bg-surface-elevated border border-border-subtle px-2 py-0.5 text-xs font-semibold text-text-secondary">
              {members.length} Pengguna
            </span>
            {isFetching && <RefreshCw className="h-4 w-4 animate-spin text-text-secondary" aria-hidden="true" />}
          </h3>
          <p className="mt-1 text-xs leading-relaxed text-text-secondary">
            Pantau status keaktifan anggota, perolehan koin bulanan, dan kelola hak akses akun.
          </p>
        </div>
        <button
          type="button"
          onClick={openCreateModal}
          className="inline-flex items-center justify-center gap-2 rounded-xl bg-accent-magic px-4 py-2.5 text-xs font-bold text-white shadow-xs transition hover:brightness-110 active:scale-[0.98] cursor-pointer"
        >
          <UserPlus className="h-4 w-4" aria-hidden="true" />
          <span>Tambah Anggota</span>
        </button>
      </div>

      {members.length === 0 && !isFetching ? (
        <div className="rounded-2xl border border-border-subtle bg-surface p-10 text-center space-y-2">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-accent-magic/10 text-accent-magic">
            <Users className="h-6 w-6" aria-hidden="true" />
          </div>
          <p className="text-sm font-bold text-text-primary">Belum Ada Anggota</p>
          <p className="mx-auto max-w-xs text-xs text-text-secondary">
            Klik &quot;Tambah Anggota&quot; untuk mendaftarkan akun anggota keluarga baru.
          </p>
        </div>
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden overflow-hidden rounded-2xl border border-border-subtle bg-surface shadow-xs lg:block">
            <div className="overflow-x-auto">
              <table className="w-full min-w-[860px] text-left text-xs">
                <thead className="bg-surface-elevated/70 border-b border-border-subtle text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
                  <tr>
                    <th className="px-4 py-3 font-semibold">Anggota</th>
                    <th className="px-3 py-3 font-semibold">Role</th>
                    <th className="px-3 py-3 font-semibold">Aktivitas & Siklus</th>
                    <th className="px-3 py-3 font-semibold">Status</th>
                    <th className="px-4 py-3 font-semibold">Perolehan Koin Bulan Ini</th>
                    <th className="px-3 py-3 text-right font-semibold">Saldo & Tingkat</th>
                    <th className="px-4 py-3 text-right font-semibold">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border-subtle/50">
                  {members.map((member) => {
                    const last = formatLastTask(member)
                    const earned = member.earned_this_period ?? 0
                    const capVal = member.monthly_earning_cap || 0
                    const percent = capVal > 0 ? Math.min(100, Math.round((earned / capVal) * 100)) : 0
                    return (
                      <tr key={member.uid} className="transition-colors hover:bg-surface-elevated/40">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2.5 min-w-0">
                            <Avatar
                              seed={member.avatar_seed || member.uid}
                              frame={member.avatar_frame || 'none'}
                              effect={member.avatar_effect || 'none'}
                              size="sm"
                            />
                            <div className="min-w-0">
                              <p className="truncate text-xs font-semibold text-text-primary">{member.explorer_name}</p>
                              <p className="truncate text-[11px] text-text-secondary" title={`@${member.username}`}>
                                @{member.username}
                              </p>
                            </div>
                          </div>
                        </td>
                        <td className="px-3 py-3">
                          <RoleBadge role={member.role} />
                        </td>
                        <td className="px-3 py-3">
                          <div className="space-y-0.5">
                            <div className="flex items-center gap-1.5 text-xs text-text-primary">
                              <span className="font-medium">{last.text}</span>
                              {last.sub && <span className="text-[10px] text-text-secondary">({last.sub})</span>}
                            </div>
                            <div className="flex items-center gap-2 text-[10px] text-text-secondary">
                              <span
                                data-testid={`inactive-days-${member.uid}`}
                                className={member.inactivity_status === 'INACTIVE' ? 'font-bold text-accent-gold' : ''}
                              >
                                {formatInactive(member)}
                              </span>
                              <span>•</span>
                              <span>{formatCycle(member.current_cycle_start, member.current_cycle_end)}</span>
                            </div>
                          </div>
                        </td>
                        <td className="px-3 py-3">
                          <div className="flex flex-col gap-1">
                            <StatusBadge member={member} />
                            {!member.is_active && member.block_reason ? (
                              <span className="max-w-[140px] truncate text-[10px] text-text-secondary" title={member.block_reason}>
                                {member.block_reason}
                              </span>
                            ) : null}
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <div className="space-y-1.5 max-w-[200px]">
                            <div className="flex items-center justify-between text-[11px]">
                              <span className="font-mono font-bold text-text-primary">
                                {earned.toLocaleString('id-ID')} / {capVal > 0 ? `${capVal.toLocaleString('id-ID')} koin` : 'Standar'}
                              </span>
                              <span className={`text-[10px] font-bold px-1.5 py-0.2 rounded-full ${member.earning_locked ? 'bg-accent-reward/15 text-accent-reward' : 'text-text-secondary'}`}>
                                {member.earning_locked ? '🔒 Penuh' : (capVal > 0 ? 'Khusus' : 'Global')}
                              </span>
                            </div>
                            {capVal > 0 && (
                              <div className="w-full h-1.5 rounded-full bg-surface-elevated border border-border-subtle overflow-hidden">
                                <div
                                  className={`h-full transition-all ${member.earning_locked ? 'bg-accent-reward' : 'bg-accent-magic'}`}
                                  style={{ width: `${percent}%` }}
                                />
                              </div>
                            )}
                          </div>
                        </td>
                        <td className="px-3 py-3 text-right">
                          <div className="flex flex-col items-end gap-0.5">
                            <span className="inline-flex items-center gap-1 font-mono text-xs font-bold text-accent-gold">
                              <Coins className="h-3.5 w-3.5" aria-hidden="true" />
                              {member.coins.toLocaleString('id-ID')}
                            </span>
                            <span className="inline-flex items-center gap-1 text-[10px] font-bold text-accent-magic">
                              <Sparkles className="h-3 w-3" aria-hidden="true" />
                              Tingkat {member.level}
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="inline-flex items-center gap-1.5">
                            <button
                              type="button"
                              onClick={() => openEditModal(member)}
                              className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-border-subtle bg-surface text-text-secondary transition hover:bg-surface-elevated hover:text-text-primary cursor-pointer"
                              title={`Edit ${member.explorer_name}`}
                              aria-label={`Edit ${member.explorer_name}`}
                            >
                              <Edit3 className="h-4 w-4" aria-hidden="true" />
                            </button>
                            {member.is_active ? (
                              <button
                                type="button"
                                data-testid={`block-button-${member.uid}`}
                                onClick={() => handleBlock(member)}
                                className="rounded-lg bg-status-error/10 border border-status-error/20 px-3 py-1.5 text-xs font-bold text-status-error transition hover:bg-status-error/20 cursor-pointer"
                                aria-label={`Block ${member.explorer_name}`}
                              >
                                Block
                              </button>
                            ) : (
                              <button
                                type="button"
                                data-testid={`unblock-button-${member.uid}`}
                                onClick={() => handleUnblock(member)}
                                className="rounded-lg bg-status-success/10 border border-status-success/20 px-3 py-1.5 text-xs font-bold text-status-success transition hover:bg-status-success/20 cursor-pointer"
                                aria-label={`Unblock ${member.explorer_name}`}
                              >
                                Unblock
                              </button>
                            )}
                            {!isAdminRole(member.role) && (
                              <button
                                type="button"
                                data-testid={`delete-button-${member.uid}`}
                                onClick={() => handleDelete(member)}
                                disabled={processingId === member.uid}
                                className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-status-error/20 bg-surface text-status-error transition hover:bg-status-error/10 disabled:opacity-40 cursor-pointer"
                                title={`Hapus ${member.explorer_name}`}
                                aria-label={`Hapus ${member.explorer_name}`}
                              >
                                <Trash2 className="h-4 w-4" aria-hidden="true" />
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>

          {/* Mobile cards */}
          <div className="space-y-3 lg:hidden">
            {members.map((member) => {
              const badge = getStatusConfig(member)
              const last = formatLastTask(member)
              return (
                <div key={member.uid} className="rounded-2xl border border-zinc-200 bg-white p-4 shadow-sm">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-bold text-zinc-900">{member.explorer_name}</p>
                      <p className="truncate text-xs text-zinc-500">@{member.username}</p>
                      <div className="mt-2 flex flex-wrap items-center gap-2">
                        <RoleBadge role={member.role} />
                        <span
                          data-testid={`member-status-${member.uid}`}
                          className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-bold ${
                            badge.variant === 'success'
                              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                              : badge.variant === 'warning'
                                ? 'bg-amber-50 text-amber-700 border border-amber-200'
                                : badge.variant === 'error'
                                  ? 'bg-red-50 text-red-700 border border-red-200'
                                  : 'bg-zinc-100 text-zinc-600 border border-zinc-200'
                          }`}
                        >
                          {badge.label}
                        </span>
                      </div>
                    </div>
                    <div className="flex shrink-0 items-center gap-1.5">
                      <button
                        type="button"
                        onClick={() => openEditModal(member)}
                        className="flex h-8 w-8 items-center justify-center rounded-lg border border-zinc-200 bg-white text-zinc-500"
                        aria-label={`Edit ${member.explorer_name}`}
                      >
                        <Edit3 className="h-4 w-4" aria-hidden="true" />
                      </button>
                      {member.is_active ? (
                        <button
                          type="button"
                          data-testid={`block-button-${member.uid}`}
                          onClick={() => handleBlock(member)}
                          className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700"
                          aria-label={`Block ${member.explorer_name}`}
                        >
                          Block
                        </button>
                      ) : (
                        <button
                          type="button"
                          data-testid={`unblock-button-${member.uid}`}
                          onClick={() => handleUnblock(member)}
                          className="rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700"
                          aria-label={`Unblock ${member.explorer_name}`}
                        >
                          Unblock
                        </button>
                      )}
                      {!isAdminRole(member.role) && (
                        <button
                          type="button"
                          data-testid={`delete-button-${member.uid}`}
                          onClick={() => handleDelete(member)}
                          disabled={processingId === member.uid}
                          className="flex h-8 w-8 items-center justify-center rounded-lg border border-zinc-200 bg-white text-red-600 disabled:opacity-40"
                          aria-label={`Hapus ${member.explorer_name}`}
                        >
                          <Trash2 className="h-4 w-4" aria-hidden="true" />
                        </button>
                      )}
                    </div>
                  </div>

                  <div className="mt-3 grid grid-cols-2 gap-3 rounded-xl bg-zinc-50 p-3 text-xs">
                    <div>
                      <p className="text-[11px] font-semibold uppercase tracking-wide text-zinc-500">Siklus</p>
                      <p className="mt-1 inline-flex items-center gap-1.5 font-medium text-zinc-700">
                        <Calendar className="h-3.5 w-3.5 text-zinc-400" aria-hidden="true" />
                        {formatCycle(member.current_cycle_start, member.current_cycle_end)}
                      </p>
                    </div>
                    <div>
                      <p className="text-[11px] font-semibold uppercase tracking-wide text-zinc-500">Aktivitas Terakhir</p>
                      <p className="mt-1 font-medium text-zinc-900">
                        {last.text}
                        {last.sub ? <span className="ml-1 text-zinc-500">{last.sub}</span> : null}
                      </p>
                    </div>
                    <div>
                      <p className="text-[11px] font-semibold uppercase tracking-wide text-zinc-500">Tidak Aktif</p>
                      <p data-testid={`inactive-days-${member.uid}`} className="mt-1 inline-flex items-center gap-1.5 font-medium text-zinc-700">
                        <Clock className="h-3.5 w-3.5 text-zinc-400" aria-hidden="true" />
                        {formatInactive(member)}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-[11px] font-semibold uppercase tracking-wide text-zinc-500">Koin & Level</p>
                      <p className="mt-1 inline-flex items-center gap-3 font-bold">
                        <span className="inline-flex items-center gap-1 text-amber-600">
                          <Coins className="h-3.5 w-3.5" aria-hidden="true" />
                          {member.coins.toLocaleString('id-ID')}
                        </span>
                        <span className="inline-flex items-center gap-1 text-violet-600">
                          <Sparkles className="h-3 w-3" aria-hidden="true" />
                          Lv {member.level}
                        </span>
                      </p>
                      <p className={`mt-1 text-[10px] font-bold inline-flex px-2 py-0.5 rounded-full border ${member.earning_locked ? 'bg-amber-100 text-amber-700 border-amber-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200'}`} title={member.earning_locked ? 'Perolehan koin bulan ini telah mencapai batas' : 'Bisa mendapatkan koin'}>
                        {member.earned_this_period ?? 0}/{member.monthly_earning_cap ? member.monthly_earning_cap : 'Batas Global'} {member.earning_locked ? '🔒 Penuh' : '✓ Aktif'}
                      </p>
                    </div>
                  </div>
                  {!member.is_active && member.block_reason && (
                    <p className="truncate text-xs text-zinc-500" title={member.block_reason}>
                      Alasan: {member.block_reason}
                    </p>
                  )}
                </div>
              )
            })}
          </div>
        </>
      )}

      {/* Pagination */}
      {(pagination.page > 1 || pagination.has_next) && (
        <div className="flex items-center justify-between rounded-2xl border border-zinc-200 bg-white px-4 py-3">
          <button
            type="button"
            disabled={pagination.page <= 1 || isFetching}
            onClick={() => fetchMembers(pagination.page - 1)}
            className="inline-flex items-center gap-1 rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-xs font-bold text-zinc-700 hover:bg-zinc-50 disabled:opacity-40"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
            Sebelumnya
          </button>
          <span className="text-xs font-semibold text-zinc-500">Halaman {pagination.page}</span>
          <button
            type="button"
            disabled={!pagination.has_next || isFetching}
            onClick={() => fetchMembers(pagination.page + 1)}
            className="inline-flex items-center gap-1 rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-xs font-bold text-zinc-700 hover:bg-zinc-50 disabled:opacity-40"
          >
            Selanjutnya
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      )}

      <CreateMemberModal isOpen={isCreateModalOpen} form={newMember} setForm={setNewMember} isCreating={isCreating} onClose={closeCreateModal} onSubmit={handleCreateMember} />
      <EditMemberModal member={selectedMember} form={editMemberForm} setForm={setEditMemberForm} isSaving={isSavingEdit} onClose={closeEditModal} onSave={handleSaveEditMember} />
    </div>
  )
}
