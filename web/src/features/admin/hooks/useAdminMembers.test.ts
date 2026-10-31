// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, cleanup, waitFor } from '@testing-library/react'
import { useAdminMembers } from './useAdminMembers'
import { adminMembersApi } from '../../../shared/lib/api'

vi.mock('../../../shared/lib/api', () => ({
  adminMembersApi: {
    getMembers: vi.fn(),
    createMember: vi.fn(),
    updateMember: vi.fn(),
    blockMember: vi.fn(),
    unblockMember: vi.fn(),
    deleteMember: vi.fn(),
    resetPassword: vi.fn(),
  },
}))

const member = (uid: string, monthly_coin_target: number | null) =>
  ({
    uid,
    username: `user_${uid}`,
    explorer_name: `Member ${uid}`,
    role: 'MEMBER',
    is_active: true,
    monthly_coin_target,
    monthly_earning_cap: 0,
    level: 1,
    xp: 0,
    coins: 0,
  } as any)

beforeEach(() => {
  vi.clearAllMocks()
  window.alert = vi.fn()
  window.confirm = vi.fn()
})

afterEach(() => {
  cleanup()
})

describe('Gate 2-D — NULL target preservation (Q1 guard)', () => {
  it('loads NULL as inherit and omits the key on save (never writes 0)', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.updateMember).mockResolvedValue({} as any)
    const { result } = renderHook(() => useAdminMembers())

    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', null))
    })
    expect(result.current.editMemberForm.monthly_coin_target).toBeNull()

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(adminMembersApi.updateMember).toHaveBeenCalledWith(
      'u1',
      expect.not.objectContaining({ monthly_coin_target: expect.anything() })
    )
    const sent = vi.mocked(adminMembersApi.updateMember).mock.calls[0][1] as Record<string, unknown>
    expect('monthly_coin_target' in sent).toBe(false)
  })

  it('keeps explicit 0 as 0 (ZERO-row intent untouched)', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.updateMember).mockResolvedValue({} as any)
    const { result } = renderHook(() => useAdminMembers())

    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u2', 0))
    })
    expect(result.current.editMemberForm.monthly_coin_target).toBe(0)

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(adminMembersApi.updateMember).toHaveBeenCalledWith(
      'u2',
      expect.objectContaining({ monthly_coin_target: 0 })
    )
    const sent = vi.mocked(adminMembersApi.updateMember).mock.calls[0][1] as Record<string, unknown>
    expect(sent.monthly_coin_target).toBe(0)
  })
})

describe('Phase 1 — Income Configuration Correctness (F-01 / F-02)', () => {
  it('NOT_CONFIGURED forces monthly_earning_cap to 0', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.updateMember).mockResolvedValue({} as any)
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', 3000))
      result.current.setEditMemberForm((prev) => ({
        ...prev,
        income_status: 'NOT_CONFIGURED',
        monthly_earning_cap: 5000, // Even if cap was set higher, NOT_CONFIGURED forces 0
      }))
    })

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(adminMembersApi.updateMember).toHaveBeenCalledWith(
      'u1',
      expect.objectContaining({
        income_status: 'NOT_CONFIGURED',
        monthly_earning_cap: 0,
      })
    )
  })

  it('DISABLED forces monthly_earning_cap to 0', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.updateMember).mockResolvedValue({} as any)
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', 3000))
      result.current.setEditMemberForm((prev) => ({
        ...prev,
        income_status: 'DISABLED',
        monthly_earning_cap: 5000,
      }))
    })

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(adminMembersApi.updateMember).toHaveBeenCalledWith(
      'u1',
      expect.objectContaining({
        income_status: 'DISABLED',
        monthly_earning_cap: 0,
      })
    )
  })

  it('CONFIGURED with valid positive cap sends payload and refreshes members list', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.updateMember).mockResolvedValue({} as any)
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', 3000))
      result.current.setEditMemberForm((prev) => ({
        ...prev,
        income_status: 'CONFIGURED',
        monthly_earning_cap: 3500,
        reason: 'Penyesuaian budget bulanan',
      }))
    })

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(adminMembersApi.updateMember).toHaveBeenCalledWith(
      'u1',
      expect.objectContaining({
        income_status: 'CONFIGURED',
        monthly_earning_cap: 3500,
        reason: 'Penyesuaian budget bulanan',
      })
    )
    expect(result.current.selectedMember).toBeNull()
    expect(adminMembersApi.getMembers).toHaveBeenCalled()
  })

  it('CONFIGURED with 0 or negative cap alerts error and prevents submission', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', 3000))
      result.current.setEditMemberForm((prev) => ({
        ...prev,
        income_status: 'CONFIGURED',
        monthly_earning_cap: 0,
      }))
    })

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(window.alert).toHaveBeenCalledWith(
      expect.stringContaining('harus lebih besar dari 0 untuk status CONFIGURED')
    )
    expect(adminMembersApi.updateMember).not.toHaveBeenCalled()
  })

  it('CONFIGURED with cap > 10000 alerts error and prevents submission', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', 3000))
      result.current.setEditMemberForm((prev) => ({
        ...prev,
        income_status: 'CONFIGURED',
        monthly_earning_cap: 10001,
      }))
    })

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(window.alert).toHaveBeenCalledWith(
      expect.stringContaining('tidak boleh melebihi 10.000')
    )
    expect(adminMembersApi.updateMember).not.toHaveBeenCalled()
  })

  it('failed update does not close modal or report false success', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.updateMember).mockRejectedValue(new Error('RPC Database Error'))
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openEditModal(member('u1', 3000))
      result.current.setEditMemberForm((prev) => ({
        ...prev,
        income_status: 'CONFIGURED',
        monthly_earning_cap: 3500,
      }))
    })

    await act(async () => {
      await result.current.handleSaveEditMember()
    })

    expect(window.alert).toHaveBeenCalledWith(
      expect.stringContaining('RPC Database Error')
    )
    expect(result.current.selectedMember).not.toBeNull()
  })

  it('handleCreateMember sets NOT_CONFIGURED when cap is 0 and CONFIGURED when cap > 0', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue({ items: [], pagination: {} } as any)
    vi.mocked(adminMembersApi.createMember).mockResolvedValue({} as any)
    const { result } = renderHook(() => useAdminMembers())
    await waitFor(() => expect(result.current.isFetching).toBe(false))

    // 1. Cap 0 -> NOT_CONFIGURED
    act(() => {
      result.current.openCreateModal()
      result.current.setNewMember((prev) => ({
        ...prev,
        username: 'new_user',
        password: 'password123',
        explorer_name: 'New Explorer',
        role: 'MEMBER',
        monthly_earning_cap: 0,
      }))
    })

    await act(async () => {
      await result.current.handleCreateMember()
    })

    expect(adminMembersApi.createMember).toHaveBeenCalledWith(
      expect.objectContaining({
        income_status: 'NOT_CONFIGURED',
        monthly_earning_cap: 0,
      })
    )

    // 2. Cap > 0 -> CONFIGURED
    act(() => {
      result.current.openCreateModal()
      result.current.setNewMember((prev) => ({
        ...prev,
        username: 'configured_user',
        password: 'password123',
        explorer_name: 'Configured Explorer',
        role: 'MEMBER',
        monthly_earning_cap: 4000,
      }))
    })

    await act(async () => {
      await result.current.handleCreateMember()
    })

    expect(adminMembersApi.createMember).toHaveBeenCalledWith(
      expect.objectContaining({
        income_status: 'CONFIGURED',
        monthly_earning_cap: 4000,
      })
    )
  })
})
