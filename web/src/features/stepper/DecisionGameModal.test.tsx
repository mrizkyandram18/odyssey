// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { DecisionGameModal, formatRupiah } from './DecisionGameModal'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  tasksApi: { submit: vi.fn() },
}))

vi.mock('canvas-confetti', () => ({ default: vi.fn() }))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const scenario = {
  initial_balance: 500000,
  events: [
    {
      id: 'day_1',
      title: 'Teman mengajak nongkrong',
      description: 'Sepulang sekolah teman mengajak nongkrong.',
      options: [
        { id: 'a', label: 'Ikut nongkrong', delta: -75000 },
        { id: 'b', label: 'Menolak dan menabung', delta: 0 },
        { id: 'c', label: 'Ikut tapi batasi jajan', delta: -30000 },
      ],
    },
    {
      id: 'day_2',
      title: 'Kuota internet habis',
      options: [
        { id: 'a', label: 'Beli paket Rp50.000', delta: -50000 },
        { id: 'b', label: 'Cari WiFi gratis', delta: 0 },
      ],
    },
  ],
}

const baseTask = {
  id: 700,
  step_order: 3,
  title: 'Kamu Jadi Bos Keuanganmu',
  description: 'Simulasi keuangan 500 ribu.',
  task_type: 'MINI_GAME',
  status: 'UNLOCKED',
  reward_coins: 50,
  reward_xp: 100,
  config: { game: 'DECISION_FINANCE', target_score: 100, scenario },
} as unknown as TaskView

function renderModal(taskOverrides: Partial<TaskView> = {}) {
  const task = { ...baseTask, config: { ...(baseTask.config as object), ...((taskOverrides as any).config || {}) }, ...taskOverrides } as TaskView
  const props = { task, onClose: vi.fn(), onSuccess: vi.fn(), onNextTask: vi.fn() }
  render(<DecisionGameModal {...props} />)
  return props
}

describe('DecisionGameModal', () => {
  it('starts at initial balance and shows first event', () => {
    renderModal()
    expect(screen.getByTestId('current-balance')).toHaveTextContent('Rp500.000')
    expect(screen.getByText('Teman mengajak nongkrong')).toBeInTheDocument()
    expect(screen.getByText('Ikut nongkrong')).toBeInTheDocument()
  })

  it('updates balance after choosing and advances to next event', () => {
    renderModal()
    fireEvent.click(screen.getByTestId('option-day_1-a'))
    expect(screen.getByTestId('current-balance')).toHaveTextContent('Rp425.000')
    expect(screen.getByText('Kuota internet habis')).toBeInTheDocument()
  })

  it('shows final balance summary after last event', () => {
    renderModal()
    fireEvent.click(screen.getByTestId('option-day_1-b'))
    fireEvent.click(screen.getByTestId('option-day_2-b'))
    expect(screen.getByTestId('final-balance')).toHaveTextContent('Rp500.000')
    expect(screen.getByText('Simulasi Selesai!')).toBeInTheDocument()
  })

  it('submit sends choices and shows reward panel on success', async () => {
    const mocked = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    mocked.mockResolvedValue({ success: true, coins_earned: 50, xp_earned: 100 })
    renderModal()
    fireEvent.click(screen.getByTestId('option-day_1-c'))
    fireEvent.click(screen.getByTestId('option-day_2-a'))
    fireEvent.click(screen.getByText(/Klaim Reward/))
    await waitFor(() => expect(mocked).toHaveBeenCalledTimes(1))
    const payload = mocked.mock.calls[0][1]
    expect(payload.answers.game).toBe('DECISION_FINANCE')
    expect(payload.answers.choices).toEqual({ day_1: 'c', day_2: 'a' })
    await waitFor(() => expect(screen.getByText(/Simulasi Berhasil/)).toBeInTheDocument())
  })

  it('shows error when server rejects invalid result', async () => {
    const mocked = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    mocked.mockRejectedValue(new Error('hasil permainan tidak valid: pilihan tidak valid untuk event day_1'))
    renderModal()
    fireEvent.click(screen.getByTestId('option-day_1-a'))
    fireEvent.click(screen.getByTestId('option-day_2-a'))
    fireEvent.click(screen.getByText(/Klaim Reward/))
    await waitFor(() => expect(screen.getByText(/hasil permainan tidak valid/)).toBeInTheDocument())
  })

  it('renders fallback when scenario is missing (memory game untouched)', () => {
    const { container } = render(
      <DecisionGameModal
        task={{ ...baseTask, config: { game: 'MEMORY' } } as TaskView}
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />
    )
    expect(screen.getByText(/Skenario permainan belum tersedia/)).toBeInTheDocument()
    expect(container.querySelector('[data-testid="current-balance"]')).toBeNull()
  })

  it('formatRupiah handles negative and zero', () => {
    expect(formatRupiah(500000)).toBe('Rp500.000')
    expect(formatRupiah(-75000)).toBe('-Rp75.000')
    expect(formatRupiah(0)).toBe('Rp0')
  })
})
