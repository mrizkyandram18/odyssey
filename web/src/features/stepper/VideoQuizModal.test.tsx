// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { VideoQuizModal } from './VideoQuizModal'
import { tasksApi } from '../../shared/lib/api'
import type { TaskView } from '../../shared/types'

vi.mock('../../shared/lib/api', () => ({
  tasksApi: {
    submit: vi.fn(),
  },
}))

vi.mock('canvas-confetti', () => ({
  default: vi.fn(),
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('VideoQuizModal Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockQuizTask: TaskView = {
    id: 619,
    title: 'Seberapa Siap Kamu di Dunia Kerja?',
    description: 'Uji kesiapanmu di dunia kerja.',
    task_type: 'QUIZ',
    evaluation_type: 'AUTO',
    step_order: 2,
    reward_coins: 40,
    reward_xp: 100,
    status: 'UNLOCKED',
    is_locked: false,
    config: {
      questions: [
        {
          id: 'q1',
          question: 'Apa langkah paling tepat saat menghadapi dua deadline bersamaan?',
          options: [
            'A. Mengabari atasan lebih awal dan meminta arahan prioritas',
            'B. Membiarkan salah satu tugas tanpa kabar',
          ],
          explanation: 'Komunikasi proaktif membantu atasan menentukan prioritas tugas.',
        },
        {
          id: 'q2',
          question: 'Bagaimana merespon kritik dari atasan?',
          options: [
            'A. Mendengarkan dengan tenang dan mencatat poin perbaikan',
            'B. Langsung membantah dan tersinggung',
          ],
          explanation: 'Menerima kritik dengan tenang menunjukkan kematangan profesional.',
        },
      ],
    },
  }

  it('renders directly into quiz step when no video URL is provided', () => {
    render(
      <VideoQuizModal
        task={mockQuizTask}
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />
    )

    expect(screen.getByText('Seberapa Siap Kamu di Dunia Kerja?')).toBeInTheDocument()
    expect(screen.getByText(/Jawab 2 Pertanyaan Kuis/i)).toBeInTheDocument()
    expect(screen.getByText(/Apa langkah paling tepat saat menghadapi dua deadline bersamaan/i)).toBeInTheDocument()
  })

  it('shows explanation after selecting an option and allows full submission', async () => {
    const onSuccessMock = vi.fn()
    vi.mocked(tasksApi.submit).mockResolvedValueOnce({
      success: true,
      coins_earned: 40,
      xp_earned: 100,
    } as any)

    render(
      <VideoQuizModal
        task={mockQuizTask}
        onClose={vi.fn()}
        onSuccess={onSuccessMock}
      />
    )

    // Initially, explanation is not displayed
    expect(screen.queryByTestId('explanation-q1')).not.toBeInTheDocument()

    // Select option A on Question 1
    const opt1 = screen.getByText(/A. Mengabari atasan lebih awal/i)
    fireEvent.click(opt1)

    // Explanation for Question 1 should now be visible
    expect(screen.getByTestId('explanation-q1')).toBeInTheDocument()
    expect(screen.getByText(/Komunikasi proaktif membantu atasan menentukan prioritas tugas/i)).toBeInTheDocument()

    // Submit should still be disabled because q2 is not yet answered
    const submitBtn = screen.getByRole('button', { name: /Kirim Jawaban/i })
    expect(submitBtn).toBeDisabled()

    // Select option A on Question 2
    const opt2 = screen.getByText(/A. Mendengarkan dengan tenang/i)
    fireEvent.click(opt2)

    // Explanation for Question 2 is now visible
    expect(screen.getByTestId('explanation-q2')).toBeInTheDocument()
    expect(submitBtn).not.toBeDisabled()

    // Click submit
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(tasksApi.submit).toHaveBeenCalledWith(619, {
        answers: {
          q1: 'A',
          q2: 'A',
        },
      })
      expect(screen.getByText(/Hebat! Tugas Selesai/i)).toBeInTheDocument()
      expect(screen.getByText('+40 Koin')).toBeInTheDocument()
    })
  })

  it('displays error message when quiz answers are rejected by API', async () => {
    vi.mocked(tasksApi.submit).mockResolvedValueOnce({
      success: false,
      error: 'Jawaban kuis belum tepat, silakan periksa kembali',
    } as any)

    render(
      <VideoQuizModal
        task={mockQuizTask}
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />
    )

    // Answer both questions
    fireEvent.click(screen.getByText(/A. Mengabari atasan lebih awal/i))
    fireEvent.click(screen.getByText(/B. Langsung membantah dan tersinggung/i))

    const submitBtn = screen.getByRole('button', { name: /Kirim Jawaban/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(screen.getByText(/Jawaban kuis belum tepat, silakan periksa kembali/i)).toBeInTheDocument()
    })
  })
})
