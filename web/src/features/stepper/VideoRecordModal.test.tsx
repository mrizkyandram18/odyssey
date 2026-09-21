// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor, act } from '@testing-library/react'
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
  vi.useRealTimers()
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

function renderModal(taskOverrides: Partial<TaskView> = {}, extraProps: Partial<React.ComponentProps<typeof VideoRecordModal>> = {}) {
  const task = { ...baseTask, ...taskOverrides } as TaskView
  return render(<VideoRecordModal task={task} onClose={vi.fn()} onSuccess={vi.fn()} onNextTask={vi.fn()} {...extraProps} />)
}

class FakeRecorder {
  static isTypeSupported = vi.fn(() => true)
  static instances: FakeRecorder[] = []
  static defaultBlob: Blob = new Blob(['x'], { type: 'video/webm' })
  state = 'inactive'
  mimeType = 'video/webm;codecs=vp9'
  stream: any
  options?: any
  ondataavailable: ((e: any) => void) | null = null
  onstop: (() => void) | null = null
  onerror: (() => void) | null = null
  constructor(stream: any, options?: any) {
    this.stream = stream
    this.options = options
    FakeRecorder.instances.push(this)
  }
  start() { this.state = 'recording' }
  stop() {
    this.state = 'inactive'
    this.ondataavailable?.({ data: FakeRecorder.defaultBlob })
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

  it('countdown 3 -> 2 -> 1 before recording starts, recorder not started before countdown', async () => {
    stubCameraOk()
    renderModal({}, { countdownSeconds: 3 })
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())

    vi.useFakeTimers()
    try {
      fireEvent.click(screen.getByTestId('start-recording-button'))
      // Overlay shows 3, recorder not started
      expect(screen.getByTestId('countdown-number')).toHaveTextContent('3')
      expect(FakeRecorder.instances.length).toBe(0)
      expect(screen.queryByTestId('recording-indicator')).toBeNull()

      // Advance 1s -> 2
      await act(async () => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.getByTestId('countdown-number')).toHaveTextContent('2')
      expect(FakeRecorder.instances.length).toBe(0)

      // Advance 1s -> 1
      await act(async () => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.getByTestId('countdown-number')).toHaveTextContent('1')
      expect(FakeRecorder.instances.length).toBe(0)

      // Advance 1s -> recording starts
      await act(async () => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.queryByTestId('countdown-overlay')).toBeNull()
      expect(FakeRecorder.instances.length).toBe(1)
      expect(screen.getByTestId('recording-indicator')).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('cleans up countdown when cancelled before countdown finishes', async () => {
    stubCameraOk()
    renderModal({}, { countdownSeconds: 3 })
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())

    vi.useFakeTimers()
    try {
      fireEvent.click(screen.getByTestId('start-recording-button'))
      expect(screen.getByTestId('countdown-overlay')).toBeInTheDocument()

      // Click cancel
      fireEvent.click(screen.getByTestId('cancel-countdown-button'))
      expect(screen.queryByTestId('countdown-overlay')).toBeNull()
      expect(FakeRecorder.instances.length).toBe(0)

      // Advance time to ensure timer was truly cleared
      await vi.advanceTimersByTimeAsync(5000)
      expect(FakeRecorder.instances.length).toBe(0)
    } finally {
      vi.useRealTimers()
    }
  })

  it('full flow: open camera -> record -> stop -> preview -> submit PENDING', async () => {
    stubCameraOk()
    const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
    uploadMock.mockResolvedValue({ file_url: 'https://cdn/x.webm', file_name: 'video-1.webm', file_size: 12345 })
    const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
    submitMock.mockResolvedValue({ success: true, status: 'PENDING' })

    renderModal({}, { countdownSeconds: 0 })
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
    renderModal({}, { countdownSeconds: 0 })
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('start-recording-button'))
    expect(screen.getByTestId('stop-recording-button')).toBeInTheDocument()
  })

  it('configures MediaRecorder with low bitrate options (350kbps video, 48kbps audio)', async () => {
    stubCameraOk()
    renderModal({}, { countdownSeconds: 0 })
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('start-recording-button'))
    expect(FakeRecorder.instances.length).toBe(1)
    const opts = FakeRecorder.instances[0].options
    expect(opts).toBeDefined()
    expect(opts.videoBitsPerSecond).toBe(350_000)
    expect(opts.audioBitsPerSecond).toBe(48_000)
  })

  it('blocks upload and displays friendly error when recorded file exceeds 4.2 MB', async () => {
    stubCameraOk()
    // Create an oversized blob (5 MB)
    const oversizedBlob = new Blob([new Uint8Array(5 * 1024 * 1024)], { type: 'video/webm' })
    FakeRecorder.defaultBlob = oversizedBlob

    const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
    renderModal({}, { countdownSeconds: 0 })
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('start-recording-button'))
    fireEvent.click(screen.getByTestId('stop-recording-button'))

    await waitFor(() => expect(screen.getByTestId('use-video-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('use-video-button'))

    // Upload must NOT be called
    expect(uploadMock).not.toHaveBeenCalled()
    // Error message should be displayed
    await waitFor(() => {
      expect(screen.getByTestId('camera-error')).toHaveTextContent(/melebihi batas upload/i)
    })

    // Reset defaultBlob
    FakeRecorder.defaultBlob = new Blob(['x'], { type: 'video/webm' })
  })

  it('shows friendly Indonesian error message when upload fails with Vercel FUNCTION_PAYLOAD_TOO_LARGE', async () => {
    stubCameraOk()
    const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
    uploadMock.mockRejectedValue(
      new Error('Ukuran file melebihi batas server (maksimal 4 MB). Silakan perkecil atau rekam ulang file bukti.')
    )

    renderModal({}, { countdownSeconds: 0 })
    fireEvent.click(screen.getByTestId('open-camera-button'))
    await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('start-recording-button'))
    fireEvent.click(screen.getByTestId('stop-recording-button'))

    await waitFor(() => expect(screen.getByTestId('use-video-button')).toBeInTheDocument())
    fireEvent.click(screen.getByTestId('use-video-button'))

    await waitFor(() => {
      expect(screen.getByTestId('camera-error')).toHaveTextContent(/maksimal 4 MB/i)
    })
  })

  describe('camera switch front/rear (hotfix)', () => {
    const okStream = () => ({ getTracks: () => [] }) as unknown as MediaStream

    function stubCameraWith(impl: (constraints?: any) => Promise<MediaStream>) {
      const getUserMedia = vi.fn(impl)
      vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
      vi.stubGlobal('MediaRecorder', FakeRecorder as any)
      return getUserMedia
    }

    async function openCameraAndWait() {
      fireEvent.click(screen.getByTestId('open-camera-button'))
      await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
    }

    function envTask() {
      return {
        config: {
          recording: {
            enabled: true,
            max_duration_seconds: 60,
            camera_facing: 'environment' as const,
            instruction: 'instruksi',
          },
        },
      } as Partial<TaskView>
    }

    it('requests front camera initially when camera_facing is user', async () => {
      const gum = stubCameraWith(async () => okStream())
      renderModal()
      await openCameraAndWait()
      expect(gum).toHaveBeenCalledTimes(1)
      expect(gum.mock.calls[0][0]).toEqual({
        video: { facingMode: { ideal: 'user' }, width: { ideal: 640, max: 720 }, height: { ideal: 480, max: 1280 } },
        audio: true,
      })
    })

    it('requests rear camera initially when camera_facing is environment', async () => {
      const gum = stubCameraWith(async () => okStream())
      renderModal(envTask())
      await openCameraAndWait()
      expect(gum).toHaveBeenCalledTimes(1)
      expect(gum.mock.calls[0][0].video.facingMode).toEqual({ ideal: 'environment' })
      expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Belakang')
    })

    it('switches user -> environment and updates UI', async () => {
      const gum = stubCameraWith(async () => okStream())
      renderModal()
      await openCameraAndWait()
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() =>
        expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Belakang')
      )
      expect(gum).toHaveBeenCalledTimes(2)
      expect(gum.mock.calls[1][0]).toEqual({ video: { facingMode: { ideal: 'environment' } }, audio: true })
    })

    it('switches environment -> user and updates UI', async () => {
      const gum = stubCameraWith(async () => okStream())
      renderModal(envTask())
      await openCameraAndWait()
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() =>
        expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Depan')
      )
      expect(gum).toHaveBeenCalledTimes(2)
      expect(gum.mock.calls[1][0]).toEqual({ video: { facingMode: { ideal: 'user' } }, audio: true })
    })

    it('restores previous front stream when rear switch fails', async () => {
      const userStream = okStream()
      const rearErr = Object.assign(new Error('rear unavailable'), { name: 'OverconstrainedError' })
      let calls = 0
      const gum = stubCameraWith(async () => {
        calls += 1
        if (calls === 2) throw rearErr
        return userStream
      })
      renderModal()
      await openCameraAndWait()
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() =>
        expect(screen.getByTestId('camera-error')).toHaveTextContent(/Kamera belakang tidak tersedia/)
      )
      // failed switch + fallback restore of the previous front stream
      await waitFor(() => expect(gum).toHaveBeenCalledTimes(3))
      expect(gum.mock.calls[2][0]).toEqual({ video: { facingMode: { ideal: 'user' } }, audio: true })
      // still on front camera and able to continue
      expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Depan')
      expect(screen.getByTestId('start-recording-button')).toBeInTheDocument()
    })

    it('blocks switch while recording', async () => {
      stubCameraOk()
      renderModal({}, { countdownSeconds: 0 })
      await openCameraAndWait()
      fireEvent.click(screen.getByTestId('start-recording-button'))
      expect(screen.getByTestId('recording-indicator')).toBeInTheDocument()
      const gum = navigator.mediaDevices.getUserMedia as unknown as ReturnType<typeof vi.fn>
      const before = gum.mock.calls.length
      expect(screen.getByTestId('switch-camera-overlay-button')).toBeDisabled()
      fireEvent.click(screen.getByTestId('switch-camera-overlay-button'))
      expect(gum.mock.calls.length).toBe(before)
      expect(screen.getByTestId('recording-indicator')).toBeInTheDocument()
    })

    it('retake keeps the switched rear camera', async () => {
      const gum = stubCameraWith(async () => okStream())
      renderModal({}, { countdownSeconds: 0 })
      await openCameraAndWait()
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() =>
        expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Belakang')
      )
      fireEvent.click(screen.getByTestId('start-recording-button'))
      fireEvent.click(screen.getByTestId('stop-recording-button'))
      await waitFor(() => expect(screen.getByTestId('retake-button')).toBeInTheDocument())
      fireEvent.click(screen.getByTestId('retake-button'))
      await waitFor(() => expect(screen.getByTestId('start-recording-button')).toBeInTheDocument())
      const lastCall = gum.mock.calls[gum.mock.calls.length - 1][0]
      expect(lastCall.video.facingMode).toEqual({ ideal: 'environment' })
      expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Belakang')
    })

    it('submits payload.camera_facing reflecting the switched camera', async () => {
      stubCameraOk()
      const uploadMock = compressModule.uploadTaskProof as unknown as ReturnType<typeof vi.fn>
      uploadMock.mockResolvedValue({ file_url: 'https://cdn/x.webm', file_name: 'video-1.webm', file_size: 12345 })
      const submitMock = tasksApi.submit as unknown as ReturnType<typeof vi.fn>
      submitMock.mockResolvedValue({ success: true, status: 'PENDING' })
      renderModal({}, { countdownSeconds: 0 })
      await openCameraAndWait()
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() =>
        expect(screen.getByTestId('switch-camera-overlay-button')).toHaveTextContent('Kamera Belakang')
      )
      fireEvent.click(screen.getByTestId('start-recording-button'))
      fireEvent.click(screen.getByTestId('stop-recording-button'))
      await waitFor(() => expect(screen.getByTestId('use-video-button')).toBeInTheDocument())
      fireEvent.click(screen.getByTestId('use-video-button'))
      await waitFor(() => expect(submitMock).toHaveBeenCalledTimes(1))
      expect(submitMock.mock.calls[0][1].payload.camera_facing).toBe('environment')
    })

    it('switches back and forth repeatedly before recording', async () => {
      const gum = stubCameraWith(async () => okStream())
      renderModal()
      await openCameraAndWait()
      const overlay = () => screen.getByTestId('switch-camera-overlay-button')
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() => expect(overlay()).toHaveTextContent('Kamera Belakang'))
      fireEvent.click(screen.getByTestId('switch-camera-button'))
      await waitFor(() => expect(overlay()).toHaveTextContent('Kamera Depan'))
      fireEvent.click(screen.getByTestId('switch-camera-overlay-button'))
      await waitFor(() => expect(overlay()).toHaveTextContent('Kamera Belakang'))
      expect(gum).toHaveBeenCalledTimes(4)
      expect(gum.mock.calls[1][0].video.facingMode).toEqual({ ideal: 'environment' })
      expect(gum.mock.calls[2][0].video.facingMode).toEqual({ ideal: 'user' })
      expect(gum.mock.calls[3][0].video.facingMode).toEqual({ ideal: 'environment' })
      // camera still usable for normal recording afterwards
      expect(screen.getByTestId('start-recording-button')).toBeInTheDocument()
    })
  })
})
