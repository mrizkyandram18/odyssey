// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { TextResponseModal } from './TextResponseModal'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  tasksApi: { submit: vi.fn() },
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const baseTask = {
  id: 617,
  step_order: 3,
  title: 'Cari Masalah, Jangan Cuma Mengeluh',
  description: 'Tidak perlu mencari masalah besar di dunia.',
  task_type: 'TEXT_RESPONSE',
  status: 'UNLOCKED',
  reward_coins: 50,
  reward_xp: 100,
  config: {
    minimum_characters: 20,
    maximum_characters: 500,
    prompt: '1. Apa masalah yang sering dikeluhkan orang di sekitarmu?\n2. Mengapa itu terjadi?',
  },
} as unknown as TaskView

describe('TextResponseModal', () => {
  it('renders both context and distinct prompt when description and prompt differ', () => {
    render(<TextResponseModal task={baseTask} onClose={vi.fn()} onSuccess={vi.fn()} />)
    expect(screen.getByText(/Konteks Refleksi/i)).toBeInTheDocument()
    expect(screen.getByText(/Tidak perlu mencari masalah besar di dunia/i)).toBeInTheDocument()
    expect(screen.getByText(/Petunjuk \/ Panduan Jawaban/i)).toBeInTheDocument()
    expect(screen.getByText(/1\. Apa masalah yang sering dikeluhkan/i)).toBeInTheDocument()
  })

  it('does not render separate context card when description and prompt are identical', () => {
    const identicalTask = {
      ...baseTask,
      description: 'Tuliskan refleksimu di sini dengan jelas.',
      config: {
        minimum_characters: 10,
        prompt: 'Tuliskan refleksimu di sini dengan jelas.',
      },
    } as TaskView
    render(<TextResponseModal task={identicalTask} onClose={vi.fn()} onSuccess={vi.fn()} />)
    expect(screen.queryByText(/Konteks Refleksi/i)).toBeNull()
    expect(screen.getByText(/Petunjuk \/ Panduan Jawaban/i)).toBeInTheDocument()
  })

  it('validates minimum characters and submits successfully', async () => {
    const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    submitMock.mockResolvedValue({ success: true, status: 'PENDING' })

    render(<TextResponseModal task={baseTask} onClose={vi.fn()} onSuccess={vi.fn()} />)
    const textarea = screen.getByPlaceholderText(/Tuliskan respon kamu di sini dengan jelas/i)
    const submitBtn = screen.getByRole('button', { name: /Kirim Jawaban Teks/i })

    // Initially disabled because charCount < 20
    expect(submitBtn).toBeDisabled()

    // Type less than min
    fireEvent.change(textarea, { target: { value: 'Terlalu pendek' } })
    expect(submitBtn).toBeDisabled()

    // Type valid response >= 20 chars
    fireEvent.change(textarea, { target: { value: 'Masalah antrean warung makan yang lama karena pencatatan manual.' } })
    expect(submitBtn).not.toBeDisabled()

    fireEvent.click(submitBtn)
    await waitFor(() => expect(submitMock).toHaveBeenCalledTimes(1))
    expect(submitMock).toHaveBeenCalledWith(617, expect.objectContaining({
      payload: expect.objectContaining({
        text: 'Masalah antrean warung makan yang lama karena pencatatan manual.',
      }),
    }))
    await waitFor(() => expect(screen.getByText(/Jawaban Berhasil Terkirim!/i)).toBeInTheDocument())
  })

  it('renders TaskRevisionBanner when task status is REJECTED and shows revision CTA', async () => {
    const rejectedTask = {
      ...baseTask,
      status: 'REJECTED',
      admin_notes: 'Jawaban kurang detail, tolong sebutkan contoh konkret.',
    } as unknown as TaskView

    const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    submitMock.mockResolvedValue({ success: true, status: 'PENDING' })

    render(<TextResponseModal task={rejectedTask} onClose={vi.fn()} onSuccess={vi.fn()} />)

    // Revision banner must be visible
    expect(screen.getByTestId('task-revision-banner')).toBeInTheDocument()
    expect(screen.getByText(/Jawaban kurang detail, tolong sebutkan contoh konkret/)).toBeInTheDocument()
    expect(screen.getByText(/Perlu Diperbaiki/)).toBeInTheDocument()

    // Button should be contextual
    const textarea = screen.getByPlaceholderText(/Tuliskan respon kamu di sini dengan jelas/i)
    fireEvent.change(textarea, { target: { value: 'Contoh konkret: warung makan Bu Siti masih pakai nota robek manual.' } })

    const submitBtn = screen.getByRole('button', { name: /Kirim Revisi Jawaban/i })
    expect(submitBtn).not.toBeDisabled()

    fireEvent.click(submitBtn)
    await waitFor(() => expect(submitMock).toHaveBeenCalledTimes(1))
  })

  it('regression: does not render TaskRevisionBanner when task status is UNLOCKED or PENDING', () => {
    render(<TextResponseModal task={baseTask} onClose={vi.fn()} onSuccess={vi.fn()} />)
    expect(screen.queryByTestId('task-revision-banner')).toBeNull()
  })
})
