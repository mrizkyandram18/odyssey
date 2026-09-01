import React from 'react'
import { Users, UserPlus, Coins, Edit3, ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react'
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
    <div className="space-y-4">
      {error && (
        <div className="p-4 rounded-2xl bg-status-error/10 border border-status-error/20 text-status-error text-xs flex items-center justify-between">
          <span>{error}</span>
          <button
            type="button"
            onClick={() => fetchMembers(pagination.page)}
            className="font-bold underline ml-2"
          >
            Coba lagi
          </button>
        </div>
      )}

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
            <Users className="w-4 h-4 text-accent-magic" />
            <span>Daftar Anggota</span>
            {isFetching && <RefreshCw className="w-3.5 h-3.5 animate-spin text-text-secondary" />}
          </h3>
          <p className="text-xs text-text-secondary mt-0.5">
            Kelola profil anggota, role, status aktif, dan buat akun baru.
          </p>
        </div>
        <button
          onClick={openCreateModal}
          className="px-3.5 py-2 rounded-xl bg-accent-magic text-white font-heading font-bold text-xs flex items-center gap-1.5 shadow-sm shadow-accent-magic/30 hover:brightness-110 active:scale-95 transition-all shrink-0 self-start sm:self-auto"
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
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {members.map((member) => (
            <div
              key={member.uid}
              className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm flex items-start justify-between gap-3"
            >
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <h4 className="font-heading font-bold text-text-primary text-sm">
                    {member.explorer_name}
                  </h4>
                  <span
                    className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                      member.role === 'ADMIN' || member.role === 'GUIDE'
                        ? 'bg-accent-magic/15 text-accent-magic'
                        : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                    }`}
                  >
                    {member.role}
                  </span>
                  <span
                    className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                      member.is_active
                        ? 'bg-status-success/15 text-status-success'
                        : 'bg-status-error/15 text-status-error'
                    }`}
                  >
                    {member.is_active ? 'Aktif' : 'Nonaktif'}
                  </span>
                </div>

                <p className="text-xs text-text-secondary font-mono">
                  @{member.username} • <span className="text-[10px] opacity-70">UID: {member.uid}</span>
                </p>

                <div className="flex items-center gap-3 text-xs pt-1">
                  <span className="font-bold text-accent-gold flex items-center gap-1">
                    <Coins className="w-3.5 h-3.5" /> {member.coins} Koin
                  </span>
                  <span className="font-bold text-accent-magic">
                    Level {member.level}
                  </span>
                </div>

                <p className="text-[10px] text-text-secondary pt-0.5">
                  Bergabung:{' '}
                  {new Date(member.created_at).toLocaleDateString('id-ID', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                  })}
                </p>
              </div>

              <button
                onClick={() => openEditModal(member)}
                className="p-2 rounded-xl bg-surface border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-surface-elevated transition-all shrink-0"
                title="Edit Anggota"
              >
                <Edit3 className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Pagination Controls */}
      {(pagination.page > 1 || pagination.has_next) && (
        <div className="flex items-center justify-between p-3 rounded-2xl bg-surface border border-border-subtle">
          <button
            type="button"
            disabled={pagination.page <= 1 || isFetching}
            onClick={() => fetchMembers(pagination.page - 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Sebelumnya</span>
          </button>

          <span className="text-xs font-bold text-text-secondary">
            Halaman {pagination.page}
          </span>

          <button
            type="button"
            disabled={!pagination.has_next || isFetching}
            onClick={() => fetchMembers(pagination.page + 1)}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1"
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
