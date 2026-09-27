// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { CameraCaptureModal } from './CameraCaptureModal'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import * as compressModule from '../../shared/lib/compress'

vi.mock('../../shared/lib/compress', () => ({
  compressImage: vi.fn(),
  uploadTaskProof: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  tasksApi: { submit: vi.fn() },
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const baseTask = {
  id: 702,
  step_order: 2,
  title: 'Cari Sesuatu yang Bisa Dibikin Lebih Baik',
  description: 'Temukan satu hal kecil yang bisa dibuat lebih baik.',
  task_type: 'PHOTO_UPLOAD',
  status: 'UNLOCKED',
  reward_coins: 40,
  reward_xp: 100,
  config: {
    max_files: 1,
    camera_only: false,
    instruction: 'Ambil 1 foto dan tulis penjelasan singkat di kolom catatan.',
  },
} as unknown as TaskView

function renderModal(taskOverrides: Partial<TaskView> = {}) {
  const task = { ...baseTask, ...taskOverrides } as TaskView
  const props = { task, onClose: vi.fn(), onSuccess: vi.fn(), onNextTask: vi.fn() }
  render(<CameraCaptureModal {...props} />)
  return props
}

describe('CameraCaptureModal photo explanation note', () => {
  it('renders the short-explanation textarea', () => {
    renderModal()
    expect(screen.getByPlaceholderText(/penjelasan singkat tentang fotomu/i)).toBeInTheDocument()
  })

  it('submits file_url plus trimmed note in payload (DocUploadModal convention)', async () => {
    const compressMock = compressModule.compressImage as unknown as ReturnType<typeof vi.fn>
    const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
    const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    const file = new File(['fake-image'], 'kondisi.jpg', { type: 'image/jpeg' })
    compressMock.mockResolvedValue({ file, dataUrl: 'data:image/jpeg;base64,xxx' })
    uploadMock.mockResolvedValue({ file_url: 'https://cdn/foto.jpg', file_name: 'kondisi.jpg', file_size: 12345 })
    submitMock.mockResolvedValue({ success: true })

    const { container } = render(
      <CameraCaptureModal task={baseTask} onClose={vi.fn()} onSuccess={vi.fn()} onNextTask={vi.fn()} />
    )
    const input = container.querySelector('input[type="file"]') as HTMLInputElement
    expect(input).not.toBeNull()
    fireEvent.change(input, { target: { files: [file] } })
    await waitFor(() => expect(screen.getByAltText('Preview Foto')).toBeInTheDocument())

    fireEvent.change(screen.getByPlaceholderText(/penjelasan singkat tentang fotomu/i), {
      target: { value: '  Meja belajarku berantakan, kabel perlu dirapikan agar nyaman.  ' },
    })
    fireEvent.click(screen.getByText(/Kirim Foto Bukti/))

    await waitFor(() => expect(submitMock).toHaveBeenCalledTimes(1))
    const payload = submitMock.mock.calls[0][1].payload
    expect(payload.file_url).toBe('https://cdn/foto.jpg')
    expect(payload.note).toBe('Meja belajarku berantakan, kabel perlu dirapikan agar nyaman.')
    await waitFor(() => expect(screen.getByText(/Foto Berhasil Dikirim/)).toBeInTheDocument())
  })

  it('submits empty note when member writes nothing (admin review enforces completeness)', async () => {
    const compressMock = compressModule.compressImage as unknown as ReturnType<typeof vi.fn>
    const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
    const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    const file = new File(['fake-image'], 'kondisi.jpg', { type: 'image/jpeg' })
    compressMock.mockResolvedValue({ file, dataUrl: 'data:image/jpeg;base64,xxx' })
    uploadMock.mockResolvedValue({ file_url: 'https://cdn/foto.jpg', file_name: 'kondisi.jpg', file_size: 12345 })
    submitMock.mockResolvedValue({ success: true })

    const { container } = render(
      <CameraCaptureModal task={baseTask} onClose={vi.fn()} onSuccess={vi.fn()} onNextTask={vi.fn()} />
    )
    const input = container.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [file] } })
    await waitFor(() => expect(screen.getByAltText('Preview Foto')).toBeInTheDocument())

    fireEvent.click(screen.getByText(/Kirim Foto Bukti/))
    await waitFor(() => expect(submitMock).toHaveBeenCalledTimes(1))
    expect(submitMock.mock.calls[0][1].payload.note).toBe('')
  })
})
