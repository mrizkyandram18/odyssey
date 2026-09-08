// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { VideoRecordModal } from './VideoRecordModal'
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
  vi.unstubAllGlobals()
})

const baseTask = {
  id: 701,
  step_order: 4,
  title: 'Kenalin Dirimu dalam 60 Detik',
  description: 'Perkenalkan dirimu.',
  task_type: 'VIDEO',
  status: 'UNLOCKED',
  reward_coins: 50,
  reward_xp: 100,
  config: {
    recording: {
      enabled: true,
      max_duration_seconds: 60,
      camera_facing: 'user',
      instruction: 'Perkenalkan: siapa kamu, apa yang kamu kuasai, apa yang dipelajari, pekerjaan impian.',
    },
  },
} as unknown as TaskView

function renderModal(taskOverrides: Partial<TaskView> = {}) {
  const task = { ...baseTask, ...taskOverrides } as TaskView
  return render(<VideoRecordModal task={task} onClose={vi.fn()} onSuccess={vi.fn()} onNextTask={vi.fn()} />)
}

class FakeRecorder {
  static isTypeSupported = vi.fn(() => true)
  static instances: FakeRecorder[] = []
  state = 'inactive'
  mimeType = 'video/webm;codecs=vp9'
  stream: any
  ondataavailable: ((e: any) => void) | null = null
  onstop: (() => void) | null = null
  onerror: (() => void) | null = null
  constructor(stream: any) {
    this.stream = stream
    FakeRecorder.instances.push(this)
  }
  start() { this.state = 'recording' }
  stop() {
    this.state = 'inactive'
    this.ondataavailable?.({ data: new Blob(['x'], { type: 'video/webm' }) })
    this.onstop?.()
  }
}

function stubCameraOk() {
  const stream = { getTracks: () => [] } as unknown as MediaStream
  vi.stubGlobal('navigator', {
    mediaDevices: { getUserMedia: vi.fn().mockResolvedValue(stream) },
  })
  vi.stubGlobal('MediaRecorder', FakeRecorder as any)
}

describe('VideoRecordModal', () => {
  beforeEach(() => {
    FakeRecorder.instances = []
    if (typeof URL.createObjectURL !== 'function') {
      Object.defineProperty(URL, 'createObjectURL', { value: vi.fn(() => 'blob:fake-video-url'), configurable: true })
    }
    if (typeof URL.revokeObjectURL !== 'function') {
      Object.defineProperty(URL, 'revokeObjectURL', { value: vi.fn(), configurable: true })
    }
  })

  it('renders open-camera button, instruction, and NO file input', () => {
    stubCameraOk()
    const { container } = renderModal()
    expect(screen.getByTestId('open-camera-button')).toBeInTheDocument()
    expect(screen.getByText(/Perkenalkan: siapa kamu/i)).toBeInTheDocument()
    expect(screen.getByText(/maksimal 60 detik/i)).toBeInTheDocument()
    expect(container.querySelector('input[type="file"]')).toBeNull()
  })

  it('shows unsupported error when MediaRecorder is unavailable', () => {
    vi.stubGlobal('navigator', {
      mediaDevices: { getUserMedia: vi.fn().mockResolvedValue({ getTracks: () => [] }) },
    })
    vi.stubGlobal('MediaRecorder', undefined as any)
    renderModal()
    expect(screen.getByTestId('camera-error')).toBeInTheDocument()
  })

  it('shows permission error when getUserMedia is denied', async () => {
    const err = Object.assign(new Error('denied'), { name: 'NotAllowedError' })
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia: vi.fn().mockRejectedValue(err) } })
    vi.stubGlobal('MediaRecorder', FakeRecorder as any)
    renderModal()
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('camera-error')).toHaveTextContent(/Izin kamera/))
  })

  it('full flow: open camera -> record -> stop -> preview -> submit PENDING', async () => {
    stubCameraOk()
    const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
    uploadMock.mockResolvedValue({ file_url: 'https://cdn/x.webm', file_name: 'video-1.webm', file_size: 12345 })
    const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    submitMock.mockResolvedValue({ success: true, status: 'PENDING' })

    renderModal()
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())

    fireEvent.click(screen.getByTestId('start-recording-button'))
    expect(screen.getByTestId('recording-indicator')).toBeInTheDocument()
    expect(FakeRecorder.instances.length).toBe(1)

    fireEvent.click(screen.getByTestId('stop-recording-button'))
    await waitFor(() => expect(screen.getByTestId('record-preview')).toBeInTheDocument())

    // retake returns to camera
    fireEvent.click(screen.getByTestId('retake-button'))
    await waitFor(() => expect(screen.queryByTestId('record-preview')).toBeNull())

    // record again and submit
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('start-recording-button'))
    fireEvent.click(screen.getByTestId('stop-recording-button'))
    await waitFor(() => expect(screen.getByTestId('use-video-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('use-video-button'))

    await waitFor(() => expect(submitMock).toHaveBeenCalledTimes(1))
    const payload = submitMock.mock.calls[0][1].payload
    expect(payload.file_url).toBe('https://cdn/x.webm')
    expect(payload).toHaveProperty('duration_seconds')
    expect(payload).toHaveProperty('captured_at')
    await waitFor(() => expect(screen.getByText(/Video Berhasil Dikirim/)).toBeInTheDocument())
  })

  it('falls back when isTypeSupported reports false for all codecs', async () => {
    stubCameraOk()
    ;(FakeRecorder.isTypeSupported as any).mockReturnValue(false)
    renderModal()
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('start-recording-button'))
    expect(screen.getByTestId('stop-recording-button')).toBeInTheDocument()
  })
})
