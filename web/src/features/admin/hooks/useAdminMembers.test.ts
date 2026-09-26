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
