// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { MemberList } from './MemberList'
import { adminMembersApi } from '../../../../shared/lib/api'

vi.mock('../../../../shared/lib/api', () => ({
  adminMembersApi: {
    getMembers: vi.fn(),
    createMember: vi.fn(),
    updateMember: vi.fn(),
    resetPassword: vi.fn(),
    blockMember: vi.fn(),
    unblockMember: vi.fn(),
    deleteMember: vi.fn(),
  },
}))

const member = (uid: string, overrides: any = {}) => ({
  uid,
  family_id: 'fam_1',
  explorer_name: `Explorer ${uid}`,
  username: `user_${uid}`,
  role: 'MEMBER',
  is_active: true,
  level: 1,
  xp: 0,
  coins: 100,
  created_at: new Date().toISOString(),
  ...overrides,
})

const paged = (items: any[]) => ({
  items,
  pagination: { page: 1, limit: 50, total: items.length, has_next: false },
})

describe('MemberList delete action', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.stubGlobal('confirm', vi.fn(() => true))
    vi.stubGlobal('alert', vi.fn())
  })

  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
  })

  it('shows Delete for members but hides it for admin rows', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(
      paged([member('usr_1'), member('adm_1', { role: 'ADMIN', explorer_name: 'Admin' })])
    )
    render(<MemberList />)
    await waitFor(() => {
      expect(screen.getAllByText('Explorer usr_1').length).toBeGreaterThan(0)
    })
    // Desktop table + mobile card each render one button per row.
    expect(screen.getAllByTestId('delete-button-usr_1')).toHaveLength(2)
    expect(screen.queryAllByTestId('delete-button-adm_1')).toHaveLength(0)
  })

  it('requires confirmation before deleting', async () => {
    vi.mocked(window.confirm).mockReturnValue(false)
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([member('usr_1')]))
    render(<MemberList />)
    await waitFor(() => {
      expect(screen.getAllByTestId('delete-button-usr_1')).toHaveLength(2)
    })
    fireEvent.click(screen.getAllByTestId('delete-button-usr_1')[0])
    expect(window.confirm).toHaveBeenCalled()
    expect(adminMembersApi.deleteMember).not.toHaveBeenCalled()
  })

  it('deletes the row member and refreshes the list on success', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(
      paged([member('usr_1'), member('usr_2')])
    )
    vi.mocked(adminMembersApi.deleteMember).mockResolvedValue({ success: true, uid: 'usr_2', is_active: false })
    render(<MemberList />)
    await waitFor(() => {
      expect(screen.getAllByTestId('delete-button-usr_2')).toHaveLength(2)
    })
    const callsBefore = vi.mocked(adminMembersApi.getMembers).mock.calls.length
    fireEvent.click(screen.getAllByTestId('delete-button-usr_2')[0])
    await waitFor(() => {
      // Exactly the clicked member's uid is sent — never a neighbour row.
      expect(adminMembersApi.deleteMember).toHaveBeenCalledWith('usr_2')
    })
    await waitFor(() => {
      expect(vi.mocked(adminMembersApi.getMembers).mock.calls.length).toBeGreaterThan(callsBefore)
    })
  })

  it('shows an error when deletion fails', async () => {
    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([member('usr_1')]))
    vi.mocked(adminMembersApi.deleteMember).mockRejectedValue(new Error('anggota tidak ditemukan'))
    render(<MemberList />)
    await waitFor(() => {
      expect(screen.getAllByTestId('delete-button-usr_1')).toHaveLength(2)
    })
    fireEvent.click(screen.getAllByTestId('delete-button-usr_1')[0])
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith(expect.stringContaining('Gagal menghapus anggota'))
    })
  })
})
