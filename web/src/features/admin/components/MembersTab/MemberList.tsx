import React from 'react'
import { Users, UserPlus, Coins, Edit3, ChevronLeft, ChevronRight, RefreshCw, Shield, Sparkles } from 'lucide-react'
import { useAdminMembers } from '../../hooks/useAdminMembers'
import { CreateMemberModal } from './CreateMemberModal'
import { EditMemberModal } from './EditMemberModal'

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
            Kelola data akun, role, status aktif, koin, dan reset binding perangkat.
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
                  <th className="py-3 px-3">Status</th>
                  <th className="py-3 px-3">Koin</th>
                  <th className="py-3 px-3">Level</th>
                  <th className="py-3 px-3">Bergabung</th>
                  <th className="py-3 px-4 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-subtle">
                {members.map((member) => (
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

                    <td className="py-3 px-3">
                      <span
                        className={`inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md ${
                          member.is_active
                            ? 'bg-status-success/15 text-status-success'
                            : 'bg-status-error/15 text-status-error'
                        }`}
                      >
                        <span className={`w-1.5 h-1.5 rounded-full ${member.is_active ? 'bg-status-success' : 'bg-status-error'}`} />
                        {member.is_active ? 'Aktif' : 'Nonaktif'}
                      </span>
                    </td>

                    <td className="py-3 px-3">
                      <span className="font-bold text-accent-gold font-mono flex items-center gap-1">
                        <Coins className="w-3 h-3" />
                        {member.coins.toLocaleString('id-ID')}
                      </span>
                    </td>

                    <td className="py-3 px-3">
                      <span className="font-bold text-accent-magic font-mono flex items-center gap-1">
                        <Sparkles className="w-3 h-3" />
                        Lvl {member.level}
                      </span>
                    </td>

                    <td className="py-3 px-3 text-[11px] text-text-secondary font-mono whitespace-nowrap">
                      {new Date(member.created_at).toLocaleDateString('id-ID', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                      })}
                    </td>

                    <td className="py-3 px-4 text-right">
                      <button
                        type="button"
                        onClick={() => openEditModal(member)}
                        className="p-1.5 rounded-lg bg-surface border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-surface-elevated transition-all cursor-pointer inline-flex items-center justify-center"
                        title={`Edit ${member.explorer_name}`}
                        aria-label={`Edit ${member.explorer_name}`}
                      >
                        <Edit3 className="w-3.5 h-3.5" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile Card View */}
          <div className="sm:hidden space-y-2.5">
            {members.map((member) => (
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
                        className={`text-[10px] font-bold px-1.5 py-0.2 rounded ${
                          member.is_active
                            ? 'bg-status-success/15 text-status-success'
                            : 'bg-status-error/15 text-status-error'
                        }`}
                      >
                        {member.is_active ? 'Aktif' : 'Nonaktif'}
                      </span>
                    </div>
                    <p className="text-[11px] text-text-secondary font-mono mt-0.5">
                      @{member.username} • UID: {member.uid}
                    </p>
                  </div>

                  <button
                    type="button"
                    onClick={() => openEditModal(member)}
                    className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary shrink-0 cursor-pointer"
                    title="Edit Anggota"
                    aria-label={`Edit ${member.explorer_name}`}
                  >
                    <Edit3 className="w-3.5 h-3.5" />
                  </button>
                </div>

                <div className="flex items-center justify-between text-xs pt-1 border-t border-border-subtle/60 text-text-secondary">
                  <div className="flex items-center gap-3">
                    <span className="font-bold text-accent-gold flex items-center gap-1">
                      <Coins className="w-3 h-3" /> {member.coins.toLocaleString('id-ID')}
                    </span>
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
            ))}
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
