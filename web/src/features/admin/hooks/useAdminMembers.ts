import { useState, useCallback, useEffect } from 'react'
import { adminMembersApi } from '../../../shared/lib/api'
import type { MemberView, PaginationMeta } from '../../../shared/types'

export function useAdminMembers() {
  const [members, setMembers] = useState<MemberView[]>([])
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, limit: 50, total: 0, has_next: false })
  const [isFetching, setIsFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [processingId, setProcessingId] = useState<string | null>(null)

  // Create member modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [newMember, setNewMember] = useState({
    username: '',
    password: '',
    explorer_name: '',
    role: 'MEMBER' as 'ADMIN' | 'MEMBER',
    monthly_coin_target: 3200,
  })
  const [isCreating, setIsCreating] = useState(false)

  // Edit member modal
  const [selectedMember, setSelectedMember] = useState<MemberView | null>(null)
  const [editMemberForm, setEditMemberForm] = useState({
    explorer_name: '',
    role: 'MEMBER' as 'ADMIN' | 'MEMBER',
    is_active: true,
    reset_device: false,
    monthly_coin_target: 3200,
  })
  const [isSavingEdit, setIsSavingEdit] = useState(false)

  const fetchMembers = useCallback(async (page = 1) => {
    setIsFetching(true)
    setError(null)
    try {
      const res: any = await adminMembersApi.getMembers({ page, limit: 50 })
      const items: MemberView[] = Array.isArray(res) ? res : res?.items || []
      const pag: PaginationMeta = Array.isArray(res)
        ? { page, limit: 50, total: items.length, has_next: false }
        : res?.pagination || { page, limit: 50, total: items.length, has_next: false }
      setMembers(items)
      setPagination(pag)
    } catch (err: any) {
      setError(err?.message || 'Gagal memuat data anggota')
    } finally {
      setIsFetching(false)
    }
  }, [])

  useEffect(() => {
    fetchMembers(1)
  }, [fetchMembers])

  const openCreateModal = () => {
    setNewMember({
      username: '',
      password: '',
      explorer_name: '',
      role: 'MEMBER',
      monthly_coin_target: 3200,
    })
    setIsCreateModalOpen(true)
  }

  const closeCreateModal = () => {
    setIsCreateModalOpen(false)
  }

  const handleCreateMember = async () => {
    if (!newMember.username.trim() || !newMember.password.trim() || !newMember.explorer_name.trim()) {
      alert('Semua field wajib diisi')
      return
    }
    if (newMember.role === 'MEMBER' && (newMember.monthly_coin_target < 1 || newMember.monthly_coin_target > 10000)) {
      alert('Target koin bulanan harus 1..10000')
      return
    }
    setIsCreating(true)
    try {
      await adminMembersApi.createMember({
        username: newMember.username.trim(),
        password: newMember.password.trim(),
        explorer_name: newMember.explorer_name.trim(),
        role: newMember.role,
        monthly_coin_target: newMember.role === 'MEMBER' ? newMember.monthly_coin_target : undefined,
      })
      closeCreateModal()
      await fetchMembers(pagination.page)
    } catch (err: any) {
      alert(`Gagal menambah anggota: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setIsCreating(false)
    }
  }

  const openEditModal = (member: MemberView) => {
    setSelectedMember(member)
    setEditMemberForm({
      explorer_name: member.explorer_name,
      role: (member.role === 'ADMIN' || member.role === 'GUIDE' || member.role === 'BUILDER') ? 'ADMIN' : 'MEMBER',
      is_active: member.is_active,
      reset_device: false,
      monthly_coin_target: member.monthly_coin_target ?? 3200,
    })
  }

  const closeEditModal = () => {
    setSelectedMember(null)
  }

  const handleSaveEditMember = async () => {
    if (!selectedMember) return
    if (!editMemberForm.explorer_name.trim()) {
      alert('Nama anggota tidak boleh kosong')
      return
    }
    if (editMemberForm.monthly_coin_target < 1 || editMemberForm.monthly_coin_target > 10000) {
      alert('Target koin bulanan harus 1..10000')
      return
    }
    setIsSavingEdit(true)
    try {
      await adminMembersApi.updateMember(selectedMember.uid, {
        explorer_name: editMemberForm.explorer_name.trim(),
        role: editMemberForm.role,
        is_active: editMemberForm.is_active,
        reset_device: editMemberForm.reset_device ? true : undefined,
        monthly_coin_target: editMemberForm.monthly_coin_target,
      })
      closeEditModal()
      await fetchMembers(pagination.page)
    } catch (err: any) {
      alert(`Gagal memperbarui anggota: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setIsSavingEdit(false)
    }
  }

  const handleToggleActive = async (member: MemberView) => {
    setProcessingId(member.uid)
    try {
      await adminMembersApi.updateMember(member.uid, {
        is_active: !member.is_active,
      })
      await fetchMembers(pagination.page)
    } catch (err: any) {
      alert(`Gagal mengubah status anggota: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  const handleBlock = async (member: MemberView, reason?: string) => {
    setProcessingId(member.uid)
    try {
      await adminMembersApi.blockMember(member.uid, reason)
      await fetchMembers(pagination.page)
    } catch (err: any) {
      alert(`Gagal memblokir anggota: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  const handleUnblock = async (member: MemberView) => {
    setProcessingId(member.uid)
    try {
      await adminMembersApi.unblockMember(member.uid)
      await fetchMembers(pagination.page)
    } catch (err: any) {
      alert(`Gagal membuka blokir: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  const handleAutoBlock = async () => {
    setIsFetching(true)
    try {
      const res = await adminMembersApi.autoBlock()
      await fetchMembers(pagination.page)
      return res
    } catch (err: any) {
      alert(`Gagal auto-block: ${err?.message || 'Terjadi kesalahan'}`)
      throw err
    } finally {
      setIsFetching(false)
    }
  }

  return {
    members,
    pagination,
    isFetching,
    error,
    processingId,
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
    handleToggleActive,
    handleBlock,
    handleUnblock,
    handleAutoBlock,
  }
}
