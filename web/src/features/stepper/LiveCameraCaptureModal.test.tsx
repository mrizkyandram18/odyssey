// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { LiveCameraCaptureModal } from './LiveCameraCaptureModal'
import { CameraCaptureModal } from './CameraCaptureModal'
import { isLegacyCameraOnlyTask } from './LinearPath'
import type { TaskView } from '../../shared/types'
import * as compressModule from '../../shared/lib/compress'
import { tasksApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/compress', () => ({
  compressImage: vi.fn(),
  uploadTaskProof: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  tasksApi: {
    submit: vi.fn(),
  },
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const baseTask = {
  id: 9,
  step_order: 1,
  title: 'Bukti Diskusi',
  description: 'Tugas diskusi bersama teman.',
  task_type: 'PHOTO_UPLOAD',
  status: 'UNLOCKED',
  reward_coins: 50,
  reward_xp: 100,
} as unknown as TaskView

function renderLiveModal(taskOverrides: Partial<TaskView> = {}) {
  const task = { ...baseTask, ...taskOverrides } as TaskView
  return render(
    <LiveCameraCaptureModal
      task={task}
      onClose={vi.fn()}
      onSuccess={vi.fn()}
      onNextTask={vi.fn()}
    />
  )
}

describe('Test 4 & AC1–AC2: LiveCameraCaptureModal (camera_only=true)', () => {
  it('renders open-camera button and contains NO file/gallery input', () => {
    const { container } = renderLiveModal({ config: { camera_only: true } })
    expect(screen.getByTestId('open-camera-button')).toBeInTheDocument()
    // Must NOT have <input type="file">
    expect(container.querySelector('input[type="file"]')).toBeNull()
  })

  it('shows discussion-specific instruction when task title is Bukti Diskusi and no custom instruction set', () => {
    renderLiveModal({ title: 'Bukti Diskusi', config: { camera_only: true } })
    expect(screen.getByText(/Selesaikan sesi diskusi bersama teman\/rekanmu/i)).toBeInTheDocument()
    expect(screen.getByText(/tidak ada pilihan galeri/i)).toBeInTheDocument()
  })

  it('shows CV-specific instruction when task title is Foto Langsung Kesiapan Profil CV', () => {
    renderLiveModal({
      title: 'Foto Langsung Kesiapan Profil CV (Rapi & Profesional)',
      config: { camera_only: true },
    })
    expect(screen.getByText(/Gunakan kemeja\/pakaian berkerah rapi/i)).toBeInTheDocument()
    expect(screen.getByText(/Siap Kerja/i)).toBeInTheDocument()
  })

  it('shows generic camera instruction for other camera-only tasks without custom instruction', () => {
    renderLiveModal({
      title: 'Aktivitas Olahraga Pagi',
      config: { camera_only: true },
    })
    expect(screen.getByText(/Ambil foto bukti aktivitas langsung menggunakan kamera perangkat/i)).toBeInTheDocument()
  })

  it('shows custom instruction when camera_instruction is configured (Priority 1)', () => {
    renderLiveModal({
      title: 'Bukti Diskusi',
      config: {
        camera_only: true,
        camera_instruction: 'Instruksi Khusus: Foto berdua sambil memegang catatan modul 4.',
      },
    })
    expect(screen.getByText('Instruksi Khusus: Foto berdua sambil memegang catatan modul 4.')).toBeInTheDocument()
    expect(screen.queryByText(/Selesaikan sesi diskusi/i)).toBeNull()
  })
})

describe('Test 5 & AC6: Normal Photo Upload (camera_only=false/undefined)', () => {
  it('CameraCaptureModal contains file input allowing gallery/file selection', () => {
    const task = {
      ...baseTask,
      title: 'ANTIGRAVITY: Bikin Poster Diri di Canva (HP)',
      config: { camera_only: false },
    } as TaskView
    const { container } = render(
      <CameraCaptureModal
        task={task}
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />
    )
    const fileInput = container.querySelector('input[type="file"]')
    expect(fileInput).not.toBeNull()
    expect(fileInput).toHaveAttribute('accept', 'image/*')
  })

  it('isLegacyCameraOnlyTask correctly identifies legacy camera tasks', () => {
    expect(isLegacyCameraOnlyTask({ title: 'Bukti Diskusi' } as TaskView)).toBe(true)
    expect(isLegacyCameraOnlyTask({ title: 'Foto Langsung Kesiapan Profil CV (Rapi & Profesional)' } as TaskView)).toBe(true)
    expect(isLegacyCameraOnlyTask({ title: 'ANTIGRAVITY: Bikin Poster Diri di Canva (HP)' } as TaskView)).toBe(false)
    expect(isLegacyCameraOnlyTask({ title: 'Rapikan Nama File di HP' } as TaskView)).toBe(false)
    expect(isLegacyCameraOnlyTask({ title: 'Bikin Anggaran di Google Sheets' } as TaskView)).toBe(false)
  })
})

describe('Camera Facing Configuration', () => {
  it('requests environment (rear) camera when camera_facing is environment', async () => {
    const getUserMedia = vi.fn().mockResolvedValue({ getTracks: () => [] })
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
    vi.spyOn(HTMLVideoElement.prototype, 'play').mockResolvedValue(undefined as any)

    renderLiveModal({ config: { camera_only: true, camera_facing: 'environment' } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    await screen.findByTestId('capture-button')
    // Uses { ideal } so the browser gracefully picks the right camera without OverconstrainedError
    expect(getUserMedia).toHaveBeenCalledWith({ video: { facingMode: { ideal: 'environment' } } })
  })

  it('requests user (front) camera when camera_facing is user', async () => {
    const getUserMedia = vi.fn().mockResolvedValue({ getTracks: () => [] })
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
    vi.spyOn(HTMLVideoElement.prototype, 'play').mockResolvedValue(undefined as any)

    renderLiveModal({ config: { camera_only: true, camera_facing: 'user' } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    await screen.findByTestId('capture-button')
    expect(getUserMedia).toHaveBeenCalledWith({ video: { facingMode: { ideal: 'user' } } })
  })

  it('defaults to rear camera (environment) when camera_facing is not configured', async () => {
    const getUserMedia = vi.fn().mockResolvedValue({ getTracks: () => [] })
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
    vi.spyOn(HTMLVideoElement.prototype, 'play').mockResolvedValue(undefined as any)

    // No camera_facing in config — should default to 'environment'
    renderLiveModal({ config: { camera_only: true } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    await screen.findByTestId('capture-button')
    expect(getUserMedia).toHaveBeenCalledWith({ video: { facingMode: { ideal: 'environment' } } })
  })

  it('renders switch camera buttons and toggles between rear and front camera', async () => {
    const stopTrack = vi.fn()
    const getUserMedia = vi.fn().mockResolvedValue({ getTracks: () => [{ stop: stopTrack }] })
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
    vi.spyOn(HTMLVideoElement.prototype, 'play').mockResolvedValue(undefined as any)

    renderLiveModal({ config: { camera_only: true, camera_facing: 'environment' } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    await screen.findByTestId('capture-button')
    expect(screen.getByTestId('switch-camera-button')).toBeInTheDocument()
    expect(screen.getByTestId('switch-camera-overlay-button')).toBeInTheDocument()
    expect(getUserMedia).toHaveBeenCalledWith({ video: { facingMode: { ideal: 'environment' } } })

    // Click switch camera -> flips to front ('user')
    fireEvent.click(screen.getByTestId('switch-camera-button'))
    await waitFor(() => {
      expect(getUserMedia).toHaveBeenCalledWith({ video: { facingMode: { ideal: 'user' } } })
    })

    // Click switch camera again -> flips back to rear ('environment')
    fireEvent.click(screen.getByTestId('switch-camera-overlay-button'))
    await waitFor(() => {
      expect(getUserMedia).toHaveBeenCalledWith({ video: { facingMode: { ideal: 'environment' } } })
    })
  })
})


describe('Camera Flow: Open -> Take Photo -> Preview -> Retake -> Take Photo -> Use Photo -> Upload -> PENDING', () => {
  beforeEach(() => {
    const mockTrack = { stop: vi.fn() }
    const getUserMedia = vi.fn().mockResolvedValue({
      getTracks: () => [mockTrack],
    })
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })
    vi.spyOn(HTMLVideoElement.prototype, 'play').mockResolvedValue(undefined as any)

    // Mock Canvas context & toBlob
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({
      drawImage: vi.fn(),
    } as any)
    vi.spyOn(HTMLCanvasElement.prototype, 'toBlob').mockImplementation((cb: any) => {
      cb(new Blob(['fake-jpeg-bytes'], { type: 'image/jpeg' }))
    })

    // Mock compressImage
    vi.mocked(compressModule.compressImage).mockResolvedValue({
      file: new File(['compressed'], 'capture-123.jpg', { type: 'image/jpeg' }),
      dataUrl: 'data:image/jpeg;base64,fakeDataUrl',
      width: 1280,
      height: 960,
    })

    // Mock uploadTaskProof
    vi.mocked(compressModule.uploadTaskProof).mockResolvedValue({
      file_url: 'https://storage.supabase.co/task-proofs/fake.jpg',
      file_name: 'capture-123.jpg',
      file_size: 204800,
    })

    // Mock tasksApi.submit
    vi.mocked(tasksApi.submit).mockResolvedValue({
      success: true,
      status: 'PENDING',
    } as any)
  })

  it('completes the full interactive camera workflow with retake and submit', async () => {
    renderLiveModal({ config: { camera_only: true } })

    // 1. Open Camera
    const openBtn = screen.getByTestId('open-camera-button')
    fireEvent.click(openBtn)

    // 2. Video preview is active, capture button appears
    const captureBtn = await screen.findByTestId('capture-button')
    expect(captureBtn).toBeInTheDocument()

    // Mock video element width/height for snapshot
    const video = screen.getByTestId('camera-preview')
    Object.defineProperty(video, 'videoWidth', { value: 1280, configurable: true })
    Object.defineProperty(video, 'videoHeight', { value: 960, configurable: true })

    // 3. Take Photo
    fireEvent.click(captureBtn)

    // 4. Preview appears with Retake and Use Photo buttons
    const retakeBtn = await screen.findByTestId('retake-button')
    const usePhotoBtn = await screen.findByTestId('use-photo-button')
    expect(retakeBtn).toBeInTheDocument()
    expect(usePhotoBtn).toBeInTheDocument()
    expect(screen.getByTestId('capture-preview')).toBeInTheDocument()

    // 5. Retake Photo -> Re-opens camera stream
    fireEvent.click(retakeBtn)
    const captureBtn2 = await screen.findByTestId('capture-button')
    expect(captureBtn2).toBeInTheDocument()

    // 6. Take Photo again
    const video2 = screen.getByTestId('camera-preview')
    Object.defineProperty(video2, 'videoWidth', { value: 1280, configurable: true })
    Object.defineProperty(video2, 'videoHeight', { value: 960, configurable: true })
    fireEvent.click(captureBtn2)

    // 7. Use Photo -> Submits to backend
    const usePhotoBtn2 = await screen.findByTestId('use-photo-button')
    fireEvent.click(usePhotoBtn2)

    // 8. Verifies uploadTaskProof & tasksApi.submit were called
    await waitFor(() => {
      expect(compressModule.uploadTaskProof).toHaveBeenCalledTimes(1)
      expect(tasksApi.submit).toHaveBeenCalledWith(9, {
        payload: expect.objectContaining({
          file_url: 'https://storage.supabase.co/task-proofs/fake.jpg',
          file_name: 'capture-123.jpg',
        }),
      })
    })

    // 9. Submitted UI displays PENDING verification status
    await screen.findByText(/Foto Berhasil Dikirim! 📸/i)
    expect(screen.getByText(/masuk antrean verifikasi admin/i)).toBeInTheDocument()
    expect(screen.getByText(/Lanjut Tugas Berikutnya/i)).toBeInTheDocument()
  })
})

describe('Failure Handling: Clear error displayed without crash', () => {
  it('displays unsupported error when navigator.mediaDevices is absent', () => {
    vi.stubGlobal('navigator', {})
    renderLiveModal({ config: { camera_only: true } })

    const errorBox = screen.getByTestId('camera-error')
    expect(errorBox).toBeInTheDocument()
    expect(errorBox).toHaveTextContent(/tidak mendukung akses kamera/i)
  })

  it('displays permission denied error when user denies camera access', async () => {
    const error = new Error('Permission denied')
    error.name = 'NotAllowedError'
    const getUserMedia = vi.fn().mockRejectedValue(error)
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })

    renderLiveModal({ config: { camera_only: true } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    const errorBox = await screen.findByTestId('camera-error')
    expect(errorBox).toBeInTheDocument()
    expect(errorBox).toHaveTextContent(/Izinkan akses kamera pada browser\/perangkat/i)
  })

  it('displays camera not found error when no camera device exists', async () => {
    const error = new Error('No camera found')
    error.name = 'NotFoundError'
    const getUserMedia = vi.fn().mockRejectedValue(error)
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })

    renderLiveModal({ config: { camera_only: true } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    const errorBox = await screen.findByTestId('camera-error')
    expect(errorBox).toBeInTheDocument()
    expect(errorBox).toHaveTextContent(/Kamera tidak ditemukan di perangkat ini/i)
  })

  it('displays in-use error when camera is locked by another app', async () => {
    const error = new Error('In use')
    error.name = 'NotReadableError'
    const getUserMedia = vi.fn().mockRejectedValue(error)
    vi.stubGlobal('navigator', { mediaDevices: { getUserMedia } })

    renderLiveModal({ config: { camera_only: true } })
    fireEvent.click(screen.getByTestId('open-camera-button'))

    const errorBox = await screen.findByTestId('camera-error')
    expect(errorBox).toBeInTheDocument()
    expect(errorBox).toHaveTextContent(/Kamera sedang digunakan aplikasi lain/i)
  })
})
