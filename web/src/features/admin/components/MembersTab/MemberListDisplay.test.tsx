// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor } from '@testing-library/react'
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

const paged = (items: any[]) => ({
  items,
  pagination: { page: 1, limit: 50, total: items.length, has_next: false },
})

describe('MemberList Display & Unicode Regression', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders Member with normal target and clean Unicode characters', async () => {
    const normalMember = {
      uid: 'usr_normal',
      family_id: 'fam_1',
      explorer_name: 'Budi Santoso',
      username: 'budi',
      role: 'MEMBER',
      is_active: true,
      income_status: 'CONFIGURED',
      level: 3,
      xp: 400,
      coins: 850,
      monthly_coin_target: 3000,
      monthly_earning_cap: 3000,
      effective_earning_cap: 3000,
      earned_this_period: 1500,
      earning_locked: false,
      current_cycle_start: '2026-11-01',
      current_cycle_end: '2026-11-30',
      last_completed_date: '2026-11-01',
      inactive_days: 0,
      inactivity_status: 'ACTIVE',
      created_at: new Date().toISOString(),
    }

    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([normalMember]))
    render(<MemberList />)

    await waitFor(() => {
      expect(screen.getAllByText('Budi Santoso').length).toBeGreaterThan(0)
    })

    // Verify target & progress display: 1.500 / 3.000
    expect(screen.getAllByText('1.500 / 3.000').length).toBeGreaterThan(0)
    expect(screen.getAllByText(/Batas: 3\.000 koin/).length).toBeGreaterThan(0)
    // Bullet separator should be clean '•', never 'â€¢'
    expect(screen.getAllByText('•').length).toBeGreaterThan(0)
  })

  it('renders Member with monthly adjustment correctly (e.g. Selvi: Base 3.320, Adj 1.437 -> Eff 1.883)', async () => {
    const selviMember = {
      uid: 'usr_selvi',
      family_id: 'fam_1',
      explorer_name: 'Selvi',
      username: 'selvi',
      role: 'MEMBER',
      is_active: true,
      income_status: 'CONFIGURED',
      level: 5,
      xp: 1200,
      coins: 1324, // Wallet balance
      monthly_coin_target: 3320, // Base target
      base_earning_cap: 3320,
      monthly_adjustment: 1437, // Scheduled adjustment
      effective_earning_cap: 1883, // Effective November cap
      earned_this_period: 0, // Settled November earnings
      earning_locked: false,
      current_cycle_start: '2026-11-01',
      current_cycle_end: '2026-11-30',
      last_completed_date: null,
      inactive_days: 0,
      inactivity_status: 'NO_ACTIVITY_THIS_CYCLE',
      created_at: new Date().toISOString(),
    }

    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([selviMember]))
    render(<MemberList />)

    await waitFor(() => {
      expect(screen.getAllByText('Selvi').length).toBeGreaterThan(0)
    })

    // Progress tracks against effective target: 0 / 1.883
    expect(screen.getAllByText('0 / 1.883').length).toBeGreaterThan(0)
    // Detailed breakdown in Batas
    expect(screen.getAllByText(/Batas:/).length).toBeGreaterThan(0)
    expect(screen.getAllByText('1.883 koin').length).toBeGreaterThan(0)
    // Wallet balance is clearly 1.324 in the saldo column
    expect(screen.getAllByText('1.324').length).toBeGreaterThan(0)
    // Placeholder em dash for missing activity should be '—', never 'â€”'
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })

  it('renders Member with earning locked (Cap Tercapai) with clean lock emoji 🔒', async () => {
    const cappedMember = {
      uid: 'usr_capped',
      family_id: 'fam_1',
      explorer_name: 'Capped User',
      username: 'capped',
      role: 'MEMBER',
      is_active: true,
      income_status: 'CONFIGURED',
      level: 10,
      xp: 5000,
      coins: 2500,
      monthly_coin_target: 1883,
      effective_earning_cap: 1883,
      earned_this_period: 1883,
      earning_locked: true,
      current_cycle_start: '2026-11-01',
      current_cycle_end: '2026-11-30',
      last_completed_date: '2026-11-01',
      inactive_days: 0,
      inactivity_status: 'ACTIVE',
      created_at: new Date().toISOString(),
    }

    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([cappedMember]))
    render(<MemberList />)

    await waitFor(() => {
      expect(screen.getAllByText('Capped User').length).toBeGreaterThan(0)
    })

    // Displays clean lock emoji 🔒 Penuh, never mojibake ðŸ”’
    expect(screen.getAllByText('🔒 Penuh').length).toBeGreaterThan(0)
  })

  it('renders Admin account distinctly without confusing progress bar or earning target', async () => {
    const adminUser = {
      uid: 'usr_admin',
      family_id: 'fam_1',
      explorer_name: 'Papa Admin',
      username: 'papa',
      role: 'ADMIN',
      is_active: true,
      income_status: 'NOT_CONFIGURED',
      level: 1,
      xp: 0,
      coins: 0,
      monthly_coin_target: null,
      effective_earning_cap: 0,
      earned_this_period: 0,
      earning_locked: false,
      created_at: new Date().toISOString(),
    }

    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([adminUser]))
    render(<MemberList />)

    await waitFor(() => {
      expect(screen.getAllByText('Papa Admin').length).toBeGreaterThan(0)
    })

    // Admin should show "Akun Pengelola (Admin)" or "Admin", NOT 0 / 3.320
    expect(screen.queryByText('0 / 3.320')).toBeNull()
    expect(screen.getAllByText(/Akun Pengelola \(Admin\)|Admin/).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/Tidak ada target kompensasi koin|Tanpa target koin/).length).toBeGreaterThan(0)
  })

  it('renders Member with earning DISABLED with clear status and Rp0', async () => {
    const disabledMember = {
      uid: 'usr_disabled',
      family_id: 'fam_1',
      explorer_name: 'Disabled User',
      username: 'disabled_user',
      role: 'MEMBER',
      is_active: true,
      income_status: 'DISABLED',
      level: 1,
      xp: 50,
      coins: 200,
      monthly_coin_target: 0,
      effective_earning_cap: 0,
      earned_this_period: 0,
      earning_locked: false,
      created_at: new Date().toISOString(),
    }

    vi.mocked(adminMembersApi.getMembers).mockResolvedValue(paged([disabledMember]))
    render(<MemberList />)

    await waitFor(() => {
      expect(screen.getAllByText('Disabled User').length).toBeGreaterThan(0)
    })

    expect(screen.getAllByText(/Dinonaktifkan \(Rp0\)/).length).toBeGreaterThan(0)
  })
})
