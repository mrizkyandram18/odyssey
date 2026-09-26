// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { SubmissionCard } from './SubmissionCard'
import type { PendingSubmissionView } from '../../../../shared/types'

const base: PendingSubmissionView = {
  id: 193,
  task_id: 560,
  task_title: 'Rencana Karir 3 Bulan',
  task_type: 'TEXT_RESPONSE',
  user_uid: 'u1',
  user_name: 'Selvi',
  submission_type: 'MANUAL_VERIFY',
  status: 'APPROVED',
  payload: {},
  created_at: '2026-09-25T13:05:36Z',
  reward_coins: 120,
  reward_xp: 45,
}

const props = {
  processingId: null,
  actionNote: '',
  onNoteChange: vi.fn(),
  actionPenalty: 0,
  onPenaltyChange: vi.fn(),
  onVerify: vi.fn(),
  onOpenEdit: vi.fn(),
  onPreviewImage: vi.fn(),
}

describe('SubmissionCard reward badge', () => {
  afterEach(() => cleanup())

  it('approved zero-coin submission stays zero (never nominal reward)', () => {
    render(<SubmissionCard submission={{ ...base, status: 'APPROVED', coins_earned: 0 }} {...props} />)
    expect(screen.getByText('+0 🪙')).toBeInTheDocument()
  })

  it('approved rewarded submission shows actual credited amount', () => {
    render(<SubmissionCard submission={{ ...base, status: 'APPROVED', coins_earned: 55 }} {...props} />)
    expect(screen.getByText('+55 🪙')).toBeInTheDocument()
  })

  it('pending submission shows nominal reward (not yet calculated)', () => {
    render(<SubmissionCard submission={{ ...base, status: 'PENDING' }} {...props} />)
    expect(screen.getByText('+120 🪙')).toBeInTheDocument()
  })
})
