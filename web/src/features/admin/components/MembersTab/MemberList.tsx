import React from 'react'
import { Users, UserPlus, Coins, Edit3, ChevronLeft, ChevronRight, RefreshCw, Shield, Sparkles, Calendar, Clock } from 'lucide-react'
import { useAdminMembers } from '../../hooks/useAdminMembers'
import { CreateMemberModal } from './CreateMemberModal'
import { EditMemberModal } from './EditMemberModal'

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

function RoleBadge({ role }: { role: string }) {
  const normalized = role === 'ADMIN' || role === 'GUIDE' || role === 'BUILDER' ? 'ADMIN' : 'MEMBER'
  const isAdmin = normalized === 'ADMIN'
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-bold tracking-wide ${
        isAdmin ? 'bg-violet-100 text-violet-700 border border-violet-200' : 'bg-zinc-100 text-zinc-600 border border-zinc-200'
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
    success: 'bg-emerald-50 text-emerald-700 border border-emerald-200',
    warning: 'bg-amber-50 text-amber-700 border border-amber-200',
    error: 'bg-red-50 text-red-700 border border-red-200',
    default: 'bg-zinc-100 text-zinc-600 border border-zinc-200',
  }
  return (
    <span
      data-testid={`member-status-${member.uid}`}
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-bold ${variantClasses[cfg.variant]}`}
    >
      {cfg.variant === 'success' && cfg.dot && <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" aria-hidden="true" />}
      {cfg.label}
    </span>
  )
}

export const MemberList: React.FC = () => {
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
  } = useAdminMembers()

  return (
    <div className="space-y-4">
      {error && (
        <div className="flex items-center justify-between rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-xs text-red-700">
          <span>{error}</span>
          <button type="button" onClick={() => fetchMembers(pagination.page)} className="font-bold underline">
            Coba lagi
          </button>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col gap-3 rounded-2xl border border-zinc-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <h3 className="flex items-center gap-2 text-sm font-bold text-zinc-900">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-violet-100 text-violet-600">
              <Users className="h-4 w-4" aria-hidden="true" />
            </span>
            Daftar Anggota
            <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-semibold text-zinc-600">{members.length} Pengguna</span>
            {isFetching && <RefreshCw className="h-4 w-4 animate-spin text-zinc-400" aria-hidden="true" />}
          </h3>
          <p className="mt-1 text-xs leading-relaxed text-zinc-500">Pantau aktivitas anggota dalam siklus berjalan dan lakukan Block/Unblock secara manual.</p>
        </div>
        <button
          type="button"
          onClick={openCreateModal}
          className="inline-flex items-center justify-center gap-2 rounded-xl bg-violet-600 px-4 py-2.5 text-xs font-bold text-white shadow-sm transition hover:bg-violet-700 active:scale-[0.98]"
        >
          <UserPlus className="h-4 w-4" aria-hidden="true" />
          Tambah Anggota
        </button>
      </div>

      {members.length === 0 && !isFetching ? (
        <div className="rounded-2xl border border-zinc-200 bg-white p-10 text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-violet-50 text-violet-600">
            <Users className="h-6 w-6" aria-hidden="true" />
          </div>
          <p className="mt-3 text-sm font-bold text-zinc-900">Belum Ada Anggota</p>
          <p className="mx-auto mt-1 max-w-xs text-xs text-zinc-500">Klik “Tambah Anggota” untuk mendaftarkan anggota baru.</p>
        </div>
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm lg:block">
            <div className="overflow-x-auto">
              <table className="w-full min-w-[860px] text-left text-xs">
                <thead className="bg-zinc-50 text-[11px] font-semibold uppercase tracking-wider text-zinc-500">
                  <tr>
                    <th className="px-4 py-3 font-semibold">Anggota</th>
                    <th className="px-3 py-3 font-semibold">Role</th>
                    <th className="px-3 py-3 font-semibold">Siklus</th>
                    <th className="px-3 py-3 font-semibold">Aktivitas Terakhir</th>
                    <th className="px-3 py-3 font-semibold">Tidak Aktif</th>
                    <th className="px-3 py-3 font-semibold">Status</th>
                    <th className="px-3 py-3 text-right font-semibold">Koin</th>
                    <th className="px-4 py-3 text-right font-semibold">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-100">
                  {members.map((member) => {
                    const last = formatLastTask(member)
                    return (
                      <tr key={member.uid} className="transition-colors hover:bg-zinc-50/60">
                        <td className="px-4 py-3">
                          <div className="min-w-0">
                            <p className="truncate text-xs font-semibold text-zinc-900">{member.explorer_name}</p>
                            <p className="truncate text-[11px] text-zinc-500" title={`@${member.username}`}>
                              @{member.username}
                            </p>
                          </div>
                        </td>
                        <td className="px-3 py-3">
                          <RoleBadge role={member.role} />
                        </td>
                        <td className="px-3 py-3">
                          <span className="inline-flex items-center gap-1.5 whitespace-nowrap text-xs text-zinc-600">
                            <Calendar className="h-3.5 w-3.5 text-zinc-400" aria-hidden="true" />
                            {formatCycle(member.current_cycle_start, member.current_cycle_end)}
                          </span>
                        </td>
                        <td className="px-3 py-3">
                          <div className="flex flex-col">
                            <span className="font-medium text-zinc-900">{last.text}</span>
                            {last.sub && <span className="text-[11px] text-zinc-500">{last.sub}</span>}
                          </div>
                        </td>
                        <td className="px-3 py-3">
                          <span
                            data-testid={`inactive-days-${member.uid}`}
                            className={`inline-flex items-center gap-1.5 text-xs ${member.inactivity_status === 'INACTIVE' ? 'font-semibold text-amber-700' : 'text-zinc-600'}`}
                          >
                            <Clock className="h-3.5 w-3.5 text-zinc-400" aria-hidden="true" />
                            {formatInactive(member)}
                          </span>
                        </td>
                        <td className="px-3 py-3">
                          <div className="flex flex-col gap-1">
                            <StatusBadge member={member} />
                            {!member.is_active && member.block_reason ? (
                              <span className="max-w-[140px] truncate text-[11px] text-zinc-500" title={member.block_reason}>
                                {member.block_reason}
                              </span>
                            ) : null}
                          </div>
                        </td>
                        <td className="px-3 py-3 text-right">
                          <div className="flex flex-col items-end gap-0.5">
                            <span className="inline-flex items-center gap-1 font-mono text-xs font-bold text-amber-600">
                              <Coins className="h-3.5 w-3.5" aria-hidden="true" />
                              {member.coins.toLocaleString('id-ID')}
                            </span>
                            <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-violet-600">
                              <Sparkles className="h-3 w-3" aria-hidden="true" />
                              Lv {member.level}
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="inline-flex items-center gap-1.5">
                            <button
                              type="button"
                              onClick={() => openEditModal(member)}
                              className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-zinc-200 bg-white text-zinc-500 transition hover:bg-zinc-50 hover:text-zinc-900"
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
                                className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 transition hover:bg-red-100"
                                aria-label={`Block ${member.explorer_name}`}
                              >
                                Block
                              </button>
                            ) : (
                              <button
                                type="button"
                                data-testid={`unblock-button-${member.uid}`}
                                onClick={() => handleUnblock(member)}
                                className="rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700 transition hover:bg-emerald-100"
                                aria-label={`Unblock ${member.explorer_name}`}
                              >
                                Unblock
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
