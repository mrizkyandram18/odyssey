import React from 'react'
import { Users, UserPlus, Coins, Edit3, ChevronLeft, ChevronRight, RefreshCw, Shield, Sparkles, Ban, CheckCircle, Calendar, Clock, Activity } from 'lucide-react'
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
  } catch { return `${start} – ${end}` }
}
function formatInactive(member: any): string {
  if (!member.is_active) return '—'
  if (member.inactivity_status === 'NO_ACTIVITY_THIS_CYCLE') return 'Belum ada aktivitas di siklus ini'
  if (member.inactive_days == null || member.inactive_days === undefined) return '—'
  return `${member.inactive_days} hari`
}
function inactivityBadge(member: any): { label: string; className: string } {
  if (!member.is_active || member.inactivity_status === 'BLOCKED') return { label: 'BLOKIR', className: 'bg-status-error/15 text-status-error' }
  switch (member.inactivity_status) {
    case 'INACTIVE': return { label: 'PERLU REVIEW', className: 'bg-amber-100 text-amber-700 border border-amber-200' }
    case 'NO_ACTIVITY_THIS_CYCLE': return { label: 'NO ACTIVITY', className: 'bg-surface-elevated text-text-secondary border border-border-subtle' }
    default: return { label: 'AKTIF', className: 'bg-status-success/15 text-status-success' }
  }
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
    <div className="space-y-3.5">
      {error && (
        <div className="p-3.5 rounded-2xl bg-status-error/10 border border-status-error/20 text-status-error text-xs flex items-center justify-between">
          <span>{error}</span>
          <button
            type="button"
            onClick={() => fetchMembers(pagination.page)}
            className="font-bold underline ml-2 cursor-pointer"
          >
            Coba lagi
          </button>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-2xl bg-surface border border-border-subtle shadow-xs">
        <div>
          <h3 className="font-bold text-text-primary text-xs flex items-center gap-2">
            <Users className="w-4 h-4 text-accent-magic" />
            <span>Daftar Anggota ({members.length} Pengguna)</span>
            {isFetching && <RefreshCw className="w-3.5 h-3.5 animate-spin text-text-secondary" />}
          </h3>
          <p className="text-[11px] text-text-secondary mt-0.5">
            Pelacakan inaktivitas siklus berjalan — admin memutuskan manual Block (otomatis diblokir dimatikan).
          </p>
        </div>

        <button
          type="button"
          onClick={openCreateModal}
          className="px-4 py-2 rounded-xl bg-accent-magic text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-xs hover:brightness-110 active:scale-95 transition-all self-stretch sm:self-auto cursor-pointer"
        >
          <UserPlus className="w-4 h-4" />
          <span>+ Tambah Anggota</span>
        </button>
      </div>

      {members.length === 0 && !isFetching ? (
        <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
          <div className="w-10 h-10 mx-auto rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center">
            <Users className="w-5 h-5" />
          </div>
          <p className="font-bold text-text-primary text-sm">Belum Ada Anggota</p>
          <p className="text-xs text-text-secondary max-w-xs mx-auto">
            Klik &quot;+ Tambah Anggota&quot; untuk mendaftarkan anggota baru.
          </p>
        </div>
      ) : (
        <>
          {/* Desktop & Tablet Table View */}
          <div className="hidden sm:block overflow-x-auto bg-surface border border-border-subtle rounded-2xl shadow-xs">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-border-subtle bg-surface-elevated/50 text-[11px] font-bold text-text-secondary uppercase tracking-wider">
                  <th className="py-3 px-4">Anggota</th>
                  <th className="py-3 px-3">Role</th>
                  <th className="py-3 px-3">Siklus</th>
                  <th className="py-3 px-3">Last Task</th>
                  <th className="py-3 px-3">Tidak Aktif</th>
                  <th className="py-3 px-3">Status</th>
                  <th className="py-3 px-3">Koin</th>
                  <th className="py-3 px-4 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-subtle">
                {members.map((member) => {
                  const badge = inactivityBadge(member)
                  return (
                  <tr key={member.uid} className="hover:bg-surface-elevated/40 transition-colors">
                    <td className="py-3 px-4">
                      <div>
                        <p className="font-bold text-text-primary text-xs">{member.explorer_name}</p>
                        <p className="text-[11px] text-text-secondary font-mono">
                          @{member.username} <span className="opacity-50">• {member.uid}</span>
                        </p>
                      </div>
                    </td>

                    <td className="py-3 px-3">
                      <span
                        className={`inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md ${
                          member.role === 'ADMIN' || member.role === 'GUIDE' || member.role === 'BUILDER'
                            ? 'bg-accent-magic/15 text-accent-magic border border-accent-magic/20'
                            : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                        }`}
                      >
                        <Shield className="w-3 h-3" />
                        {member.role}
                      </span>
                    </td>

                    <td className="py-3 px-3 text-[11px] font-mono">
                      <span className="inline-flex items-center gap-1 text-text-secondary"><Calendar className="w-3 h-3" />{formatCycle(member.current_cycle_start, member.current_cycle_end)}</span>
                    </td>

                    <td className="py-3 px-3 text-[11px] font-mono">
                      {member.last_completed_date ? (
                        <span>{new Date(member.last_completed_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })}</span>
                      ) : (
                        <span className="text-text-secondary">—</span>
                      )}
                      {member.completed_tasks_current_cycle !== undefined && member.completed_tasks_current_cycle > 0 && (
                        <span className="ml-1 text-[10px] text-text-secondary">({member.completed_tasks_current_cycle})</span>
                      )}
                    </td>

                    <td className="py-3 px-3 text-[11px]">
                      <span data-testid={`inactive-days-${member.uid}`} className="inline-flex items-center gap-1"><Clock className="w-3 h-3 text-text-secondary" />{formatInactive(member)}</span>
                    </td>

                    <td className="py-3 px-3">
                      <span
                        data-testid={`member-status-${member.uid}`}
                        className={`inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md ${badge.className}`}
                      >
                        <Activity className="w-3 h-3" />
                        {badge.label}
                      </span>
                      {!member.is_active && member.block_reason && (
                        <span className="block text-[10px] text-text-secondary mt-0.5 truncate max-w-[120px]" title={member.block_reason}>{member.block_reason}</span>
                      )}
                      {member.is_active && member.inactivity_status === 'INACTIVE' && (
                        <span className="block text-[10px] text-amber-600 mt-0.5">≥5 hari</span>
                      )}
                    </td>

                    <td className="py-3 px-3">
                      <div className="flex flex-col gap-0.5">
                        <span className="font-bold text-accent-gold font-mono flex items-center gap-1">
                          <Coins className="w-3 h-3" />
                          {member.coins.toLocaleString('id-ID')}
                        </span>
                        <span className="font-bold text-accent-magic font-mono flex items-center gap-1 text-[10px]">
                          <Sparkles className="w-3 h-3" />
                          Lvl {member.level}
                        </span>
                      </div>
                    </td>

                    <td className="py-3 px-4 text-right">
                      <div className="inline-flex items-center gap-1">
                        <button
                          type="button"
                          onClick={() => openEditModal(member)}
                          className="p-1.5 rounded-lg bg-surface border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-surface-elevated transition-all cursor-pointer inline-flex items-center justify-center"
                          title={`Edit ${member.explorer_name}`}
                          aria-label={`Edit ${member.explorer_name}`}
                        >
                          <Edit3 className="w-3.5 h-3.5" />
                        </button>
                        {member.is_active ? (
                          <button
                            type="button"
                            data-testid={`block-button-${member.uid}`}
                            onClick={() => handleBlock(member)}
                            className="p-1.5 rounded-lg bg-status-error/10 border border-status-error/20 text-status-error hover:bg-status-error/20 transition-all cursor-pointer inline-flex items-center justify-center"
                            title="Blokir anggota"
                            aria-label={`Block ${member.explorer_name}`}
                          >
                            <Ban className="w-3.5 h-3.5" />
                          </button>
                        ) : (
                          <button
                            type="button"
                            data-testid={`unblock-button-${member.uid}`}
                            onClick={() => handleUnblock(member)}
                            className="p-1.5 rounded-lg bg-status-success/10 border border-status-success/20 text-status-success hover:bg-status-success/20 transition-all cursor-pointer inline-flex items-center justify-center"
                            title="Buka blokir"
                            aria-label={`Unblock ${member.explorer_name}`}
                          >
                            <CheckCircle className="w-3.5 h-3.5" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                )})}
              </tbody>
            </table>
          </div>

          {/* Mobile Card View */}
          <div className="sm:hidden space-y-2.5">
            {members.map((member) => {
              const badge = inactivityBadge(member)
              return (
              <div
                key={member.uid}
                className="p-3.5 rounded-2xl bg-surface border border-border-subtle shadow-xs space-y-2.5"
              >
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <div className="flex items-center gap-1.5 flex-wrap">
                      <h4 className="font-bold text-text-primary text-xs">{member.explorer_name}</h4>
                      <span
                        className={`text-[10px] font-bold px-1.5 py-0.2 rounded ${
                          member.role === 'ADMIN' || member.role === 'GUIDE'
                            ? 'bg-accent-magic/15 text-accent-magic'
                            : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                        }`}
                      >
                        {member.role}
                      </span>
                      <span
                        data-testid={`member-status-${member.uid}`}
                        className={`text-[10px] font-bold px-1.5 py-0.2 rounded ${badge.className}`}
                      >
                        {badge.label}
                      </span>
                    </div>
                    <p className="text-[11px] text-text-secondary font-mono mt-0.5">
                      @{member.username} • UID: {member.uid}
                    </p>
                    <p className="text-[11px] text-text-secondary mt-1 flex items-center gap-1">
                      <Calendar className="w-3 h-3" /> Siklus {formatCycle(member.current_cycle_start, member.current_cycle_end)}
                    </p>
                    <p className="text-[11px] text-text-secondary flex items-center gap-1">
                      <Clock className="w-3 h-3" /> Last: {member.last_completed_date ? new Date(member.last_completed_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }) : '—'} • {formatInactive(member)}
                    </p>
                  </div>

                  <div className="flex items-center gap-1 shrink-0">
                    <button
                      type="button"
                      onClick={() => openEditModal(member)}
                      className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary cursor-pointer"
                      title="Edit Anggota"
                      aria-label={`Edit ${member.explorer_name}`}
                    >
                      <Edit3 className="w-3.5 h-3.5" />
                    </button>
                    {member.is_active ? (
                      <button
                        type="button"
                        data-testid={`block-button-${member.uid}`}
                        onClick={() => handleBlock(member)}
                        className="p-2 rounded-xl bg-status-error/10 border border-status-error/20 text-status-error cursor-pointer"
                        aria-label={`Block ${member.explorer_name}`}
                      >
                        <Ban className="w-3.5 h-3.5" />
                      </button>
                    ) : (
                      <button
                        type="button"
                        data-testid={`unblock-button-${member.uid}`}
                        onClick={() => handleUnblock(member)}
                        className="p-2 rounded-xl bg-status-success/10 border border-status-success/20 text-status-success cursor-pointer"
                        aria-label={`Unblock ${member.explorer_name}`}
                      >
                        <CheckCircle className="w-3.5 h-3.5" />
                      </button>
                    )}
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs pt-1 border-t border-border-subtle/60 text-text-secondary">
                  <div className="flex items-center gap-3">
                    <span className="font-bold text-accent-gold flex items-center gap-1">
                      <Coins className="w-3 h-3" /> {member.coins.toLocaleString('id-ID')}
                    </span>
                    {member.monthly_coin_target !== undefined && (
                      <span className="text-[11px] font-mono">Target {member.monthly_coin_target}</span>
                    )}
                    <span className="font-bold text-accent-magic">
                      Level {member.level}
                    </span>
                  </div>
                  <span className="text-[10px] font-mono">
                    {new Date(member.created_at).toLocaleDateString('id-ID', {
                      year: 'numeric',
                      month: 'short',
                      day: 'numeric',
                    })}
                  </span>
                </div>
              </div>
            )})}
          </div>
        </>
      )}

      {/* Pagination Controls */}
      {(pagination.page > 1 || pagination.has_next) && (
        <div className="flex items-center justify-between p-3 rounded-2xl bg-surface border border-border-subtle">
          <button
            type="button"
            disabled={pagination.page <= 1 || isFetching}
            onClick={() => fetchMembers(pagination.page - 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1 cursor-pointer"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Sebelumnya</span>
          </button>

          <span className="text-xs font-bold text-text-secondary font-mono">
            Halaman {pagination.page}
          </span>

          <button
            type="button"
            disabled={!pagination.has_next || isFetching}
            onClick={() => fetchMembers(pagination.page + 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1 cursor-pointer"
          >
            <span>Selanjutnya</span>
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* Modals */}
      <CreateMemberModal
        isOpen={isCreateModalOpen}
        form={newMember}
        setForm={setNewMember}
        isCreating={isCreating}
        onClose={closeCreateModal}
        onSubmit={handleCreateMember}
      />

      <EditMemberModal
        member={selectedMember}
        form={editMemberForm}
        setForm={setEditMemberForm}
        isSaving={isSavingEdit}
        onClose={closeEditModal}
        onSave={handleSaveEditMember}
      />
    </div>
  )
}
