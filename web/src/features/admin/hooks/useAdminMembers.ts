import { useState, useCallback, useEffect } from 'react'
import { adminMembersApi } from '../../../shared/lib/api'
import type { MemberView, PaginationMeta } from '../../../shared/types'

export function useAdminMembers() {
  const [members, setMembers] = useState<MemberView[]>([])
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, limit: 50, total: 0, has_next: false })
  const [isFetching, setIsFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [processingId, setProcessingId] = useState<string | null>(null)

  // Create member modal - default 0 biar configurable, admin set per-user (Selvi 3320 tetap)
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [newMember, setNewMember] = useState({
    username: '',
    password: '',
    explorer_name: '',
    role: 'MEMBER' as 'ADMIN' | 'MEMBER',
    monthly_coin_target: 0,
    monthly_earning_cap: 3320,
    payout_frequency: 'THRESHOLD' as 'THRESHOLD' | 'WEEKLY' | 'MONTHLY',
    minimum_withdrawal_coins: 500,
    payout_weekday: 1,
    payout_month_start_day: 24,
    payout_month_end_day: 26,
  })
  const [isCreating, setIsCreating] = useState(false)

  // Edit member modal
  const [selectedMember, setSelectedMember] = useState<MemberView | null>(null)
  const [editMemberForm, setEditMemberForm] = useState({
    explorer_name: '',
    role: 'MEMBER' as 'ADMIN' | 'MEMBER',
    is_active: true,
    reset_device: false,
    monthly_coin_target: 0,
    monthly_earning_cap: 3320,
    payout_frequency: 'THRESHOLD' as 'THRESHOLD' | 'WEEKLY' | 'MONTHLY',
    minimum_withdrawal_coins: 500,
    payout_weekday: 1,
    payout_month_start_day: 24,
    payout_month_end_day: 26,
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

  // Refetch on window focus / reconnect to avoid stale earning cap / active state
  useEffect(() => {
    const onFocus = () => fetchMembers(pagination.page)
    const onOnline = () => fetchMembers(pagination.page)
    const onVisible = () => { if (document.visibilityState === 'visible') fetchMembers(pagination.page) }
    window.addEventListener('focus', onFocus)
    window.addEventListener('online', onOnline)
    document.addEventListener('visibilitychange', onVisible)
    return () => {
      window.removeEventListener('focus', onFocus)
      window.removeEventListener('online', onOnline)
      document.removeEventListener('visibilitychange', onVisible)
    }
  }, [fetchMembers, pagination.page])

  const openCreateModal = () => {
    setNewMember({
      username: '',
      password: '',
      explorer_name: '',
      role: 'MEMBER',
      monthly_coin_target: 0,
      monthly_earning_cap: 3320,
      payout_frequency: 'THRESHOLD',
      minimum_withdrawal_coins: 500,
      payout_weekday: 1,
      payout_month_start_day: 24,
      payout_month_end_day: 26,
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
    if (newMember.role === 'MEMBER' && (newMember.monthly_coin_target < 0 || newMember.monthly_coin_target > 10000)) {
      alert('Target koin bulanan harus 0..10000')
      return
    }
    if (newMember.role === 'MEMBER' && (newMember.monthly_earning_cap < 0 || newMember.monthly_earning_cap > 10000)) {
      alert('Batas earning bulanan harus 0..10000 (0=unlimited)')
      return
    }
    if (newMember.minimum_withdrawal_coins < 1 || newMember.minimum_withdrawal_coins > 100000) {
      alert('Minimum withdrawal harus 1..100000')
      return
    }
    setIsCreating(true)
    try {
      const payload: any = {
        username: newMember.username.trim(),
        password: newMember.password.trim(),
        explorer_name: newMember.explorer_name.trim(),
        role: newMember.role,
        monthly_coin_target: newMember.role === 'MEMBER' ? newMember.monthly_coin_target : undefined,
        monthly_earning_cap: newMember.role === 'MEMBER' ? newMember.monthly_earning_cap : undefined,
        payout_frequency: newMember.payout_frequency,
        minimum_withdrawal_coins: newMember.minimum_withdrawal_coins,
      }
      if (newMember.payout_frequency === 'WEEKLY') payload.payout_weekday = newMember.payout_weekday
      if (newMember.payout_frequency === 'MONTHLY') {
        payload.payout_month_start_day = newMember.payout_month_start_day
        payload.payout_month_end_day = newMember.payout_month_end_day
      }
      await adminMembersApi.createMember(payload)
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
      monthly_coin_target: member.monthly_coin_target ?? 0,
      monthly_earning_cap: member.monthly_earning_cap ?? 3320,
      payout_frequency: (member.payout_frequency as any) || 'THRESHOLD',
      minimum_withdrawal_coins: member.minimum_withdrawal_coins ?? 500,
      payout_weekday: 1,
      payout_month_start_day: 24,
      payout_month_end_day: 26,
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
    if (editMemberForm.monthly_coin_target < 0 || editMemberForm.monthly_coin_target > 10000) {
      alert('Target koin bulanan harus 0..10000')
      return
    }
    if (editMemberForm.monthly_earning_cap < 0 || editMemberForm.monthly_earning_cap > 10000) {
      alert('Batas earning bulanan harus 0..10000 (0=unlimited)')
      return
    }
    if (editMemberForm.minimum_withdrawal_coins < 1 || editMemberForm.minimum_withdrawal_coins > 100000) {
      alert('Minimum withdrawal harus 1..100000')
      return
    }
    setIsSavingEdit(true)
    try {
      const payload: any = {
        explorer_name: editMemberForm.explorer_name.trim(),
        role: editMemberForm.role,
        is_active: editMemberForm.is_active,
        reset_device: editMemberForm.reset_device ? true : undefined,
        monthly_coin_target: editMemberForm.monthly_coin_target,
        monthly_earning_cap: editMemberForm.monthly_earning_cap,
        payout_frequency: editMemberForm.payout_frequency,
        minimum_withdrawal_coins: editMemberForm.minimum_withdrawal_coins,
      }
      if (editMemberForm.payout_frequency === 'WEEKLY') payload.payout_weekday = editMemberForm.payout_weekday
      if (editMemberForm.payout_frequency === 'MONTHLY') {
        payload.payout_month_start_day = editMemberForm.payout_month_start_day
        payload.payout_month_end_day = editMemberForm.payout_month_end_day
      }
      await adminMembersApi.updateMember(selectedMember.uid, payload)
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
  }
}
