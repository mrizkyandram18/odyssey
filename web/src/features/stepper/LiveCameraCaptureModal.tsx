import React, { useState, useRef, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Camera, RefreshCw, X, CheckCircle2, AlertCircle, Clock, ArrowRight, SwitchCamera } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { compressImage, uploadTaskProof } from '../../shared/lib/compress'
import { isEarningCapError, EARNING_CAP_MESSAGE } from '../../shared/lib/earning'
import { TaskRevisionBanner } from './TaskRevisionBanner'

interface LiveCameraCaptureModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

export const LiveCameraCaptureModal: React.FC<LiveCameraCaptureModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const isRevision = task.status === 'REJECTED'
  const isAlreadyDone = task.status === 'APPROVED' || task.status === 'PENDING'
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const streamRef = useRef<MediaStream | null>(null)

  const [isCameraOpen, setIsCameraOpen] = useState(false)
  const [isRequesting, setIsRequesting] = useState(false)
  const [capturedFile, setCapturedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [compressing, setCompressing] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(isAlreadyDone)
  const [isSupported, setIsSupported] = useState(true)

  // Per-task camera params (optional). Default to rear camera ('environment') — most photo-proof tasks
  // are taken of other people/objects. Front camera ('user') is only used when explicitly configured.
  const rawFacing = task.config?.camera_facing || task.config?.photo_camera_facing
  const cameraFacing: 'user' | 'environment' = rawFacing === 'user' ? 'user' : 'environment'
  const [activeFacing, setActiveFacing] = useState<'user' | 'environment'>(cameraFacing)
  const [isSwitchingCamera, setIsSwitchingCamera] = useState(false)

  useEffect(() => {
    setActiveFacing(cameraFacing)
  }, [cameraFacing])

  const customInstruction =
    (task.config?.camera_instruction as string | undefined)?.trim() ||
    (task.config?.instruction as string | undefined)?.trim() ||
    ''
  const isCvTask = task.title === 'Foto Langsung Kesiapan Profil CV (Rapi & Profesional)'
  const isDiscussionTask = task.title === 'Bukti Diskusi'

  useEffect(() => {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      setIsSupported(false)
    }
  }, [])

  const stopStream = useCallback(() => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((track) => track.stop())
      streamRef.current = null
    }
    if (videoRef.current) {
      videoRef.current.srcObject = null
    }
  }, [])

  useEffect(() => {
    return () => {
      stopStream()
      if (previewUrl && typeof URL.revokeObjectURL === 'function') URL.revokeObjectURL(previewUrl)
    }
  }, [stopStream, previewUrl])

  useEffect(() => {
    if (isCameraOpen && streamRef.current && videoRef.current) {
      const video = videoRef.current
      if (video.srcObject !== streamRef.current) {
        video.srcObject = streamRef.current
      }
      video.onloadedmetadata = () => {
        video.play().catch((err) => console.warn('Video play error:', err))
      }
      video.play().catch((err) => console.warn('Video play error:', err))
    }
  }, [isCameraOpen])

  const handleOpenCamera = async () => {
    if (!isSupported) {
      setErrorMessage('Perangkat/browser Anda tidak mendukung akses kamera. Silakan gunakan browser terbaru di HP Anda.')
      return
    }
    setIsRequesting(true)
    setErrorMessage(null)
    try {
      // Use { ideal: activeFacing } so the browser selects the preferred camera without
      // throwing OverconstrainedError. The 'ideal' hint is honored on Android Chrome and
      // iOS Safari — front ('user') and rear ('environment') are both correctly resolved.
      // Do NOT fall back to { video: true } here: that silently opens the wrong camera.
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: { ideal: activeFacing } },
      })
      streamRef.current = stream
      setIsCameraOpen(true)
    } catch (err: any) {
      const name = err?.name || ''
      if (name === 'NotAllowedError' || name === 'PermissionDeniedError') {
        setErrorMessage('Izin kamera belum aktif. Ketuk ikon gembok 🔒 di sebelah alamat web (atas browser) lalu pilih Izinkan akses kamera pada browser/perangkat Anda.')
      } else if (name === 'NotFoundError') {
        setErrorMessage('Kamera tidak ditemukan di perangkat ini.')
      } else if (name === 'OverconstrainedError') {
        // Device does not have the requested camera (e.g. no front camera).
        // Report clearly — do NOT silently open the wrong camera.
        const requested = activeFacing === 'user' ? 'depan' : 'belakang'
        setErrorMessage(`Kamera ${requested} tidak tersedia di perangkat ini. Pastikan perangkat memiliki kamera ${requested}.`)
      } else if (name === 'NotReadableError') {
        setErrorMessage('Kamera sedang digunakan aplikasi lain. Tutup aplikasi kamera lain lalu coba lagi.')
      } else {
        setErrorMessage('Gagal membuka kamera: ' + (err?.message || 'Pastikan izin kamera aktif'))
      }
    } finally {
      setIsRequesting(false)
    }
  }

  const handleSwitchCamera = async () => {
    if (isSwitchingCamera || isRequesting) return
    const targetFacing = activeFacing === 'user' ? 'environment' : 'user'
    setIsSwitchingCamera(true)
    setErrorMessage(null)
    try {
      if (streamRef.current) {
        streamRef.current.getTracks().forEach((track) => track.stop())
        streamRef.current = null
      }
      if (videoRef.current) {
        videoRef.current.srcObject = null
      }
      const newStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: { ideal: targetFacing } },
      })
      streamRef.current = newStream
      setActiveFacing(targetFacing)
      if (videoRef.current) {
        videoRef.current.srcObject = newStream
        videoRef.current.onloadedmetadata = () => {
          videoRef.current?.play().catch((err) => console.warn('Video play error:', err))
        }
        videoRef.current.play().catch((err) => console.warn('Video play error:', err))
      }
    } catch (err: any) {
      const name = err?.name || ''
      const requested = targetFacing === 'user' ? 'depan' : 'belakang'
      if (name === 'OverconstrainedError' || name === 'NotFoundError') {
        setErrorMessage(`Kamera ${requested} tidak tersedia di perangkat ini.`)
      } else if (name === 'NotAllowedError' || name === 'PermissionDeniedError') {
        setErrorMessage('Izin kamera belum aktif. Berikan izin akses kamera pada browser.')
      } else {
        setErrorMessage(`Gagal mengganti ke kamera ${requested}: ` + (err?.message || 'Terjadi kesalahan'))
      }
    } finally {
      setIsSwitchingCamera(false)
    }
  }

  const handleCancelCamera = () => {
    stopStream()
    setIsCameraOpen(false)
    setErrorMessage(null)
  }

  const handleCapture = async () => {
    if (!videoRef.current || !canvasRef.current) return
    const video = videoRef.current
    const canvas = canvasRef.current
    const width = video.videoWidth
    const height = video.videoHeight
    if (!width || !height) {
      setErrorMessage('Gagal menangkap gambar. Coba lagi.')
      return
    }
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      setErrorMessage('Gagal memproses gambar.')
      return
    }
    if (ctx.save) ctx.save()
    if (activeFacing === 'user') {
      if (ctx.translate) ctx.translate(width, 0)
      if (ctx.scale) ctx.scale(-1, 1)
    }
    ctx.drawImage(video, 0, 0, width, height)
    if (ctx.restore) ctx.restore()
    canvas.toBlob(
      async (blob) => {
        if (!blob) {
          setErrorMessage('Gagal membuat file foto.')
          return
        }
        const file = new File([blob], `capture-${Date.now()}.jpg`, { type: 'image/jpeg' })
        setCompressing(true)
        setErrorMessage(null)
        try {
          const result = await compressImage(file, { maxWidth: 1280, maxHeight: 1280, quality: 0.7 })
          setCapturedFile(result.file)
          setPreviewUrl(result.dataUrl)
          stopStream()
          setIsCameraOpen(false)
        } catch (e: any) {
          setErrorMessage('Gagal memproses foto: ' + e.message)
        } finally {
          setCompressing(false)
        }
      },
      'image/jpeg',
      0.85,
    )
  }

  const handleRetake = () => {
    if (previewUrl && typeof URL.revokeObjectURL === 'function') URL.revokeObjectURL(previewUrl)
    setPreviewUrl(null)
    setCapturedFile(null)
    setErrorMessage(null)
    handleOpenCamera()
  }

  const handleSubmit = async () => {
    if (!capturedFile) return
    setUploading(true)
    setErrorMessage(null)
    try {
      const uploadRes = await uploadTaskProof(capturedFile)
      const res = await tasksApi.submit(task.id, {
        payload: {
          file_url: uploadRes.file_url,
          file_name: uploadRes.file_name,
          file_size: uploadRes.file_size,
          captured_at: new Date().toISOString(),
        },
      })
      if (res.success) {
        setSubmitted(true)
      } else {
        setErrorMessage(res.error || 'Gagal menyimpan submission foto')
      }
    } catch (err: any) {
      if (isEarningCapError(err)) {
        setErrorMessage(EARNING_CAP_MESSAGE)
        try { onSuccess() } catch { /* ignore */ }
      } else {
        setErrorMessage(err.message || 'Gagal mengunggah foto. Periksa koneksi internet.')
      }
    } finally {
      setUploading(false)
    }
  }

  const handleClose = () => {
    stopStream()
    onClose()
  }

  // Derive today string for instruction
  const todayStr = new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.97, opacity: 0, y: 12 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.97, opacity: 0, y: 12 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[92vh] sm:max-h-[90vh]"
      >
        <div className="flex items-center justify-between px-5 py-3.5 border-b border-border-subtle bg-surface shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            <span className="w-7 h-7 rounded-full bg-accent-magic/12 text-accent-magic flex items-center justify-center font-bold text-xs shrink-0">#{task.step_order}</span>
            <h3 className="font-bold text-text-primary text-[14px] line-clamp-1">{task.title}</h3>
          </div>
          <button onClick={handleClose} aria-label="Tutup" className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center shrink-0">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-5 space-y-5">
          <AnimatePresence mode="wait">
            {!submitted ? (
              <motion.div key="live-camera" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-5">
                <div className="flex items-center justify-between">
                  <span className="text-xs bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">📸 Foto Langsung</span>
                  <div className="text-xs text-text-secondary font-bold">+{task.reward_coins} 🪙 | +{task.reward_xp} XP</div>
                </div>

                {isRevision && <TaskRevisionBanner adminNotes={task.admin_notes} />}

                {task.description && <p className="text-sm text-text-secondary bg-surface p-4 rounded-2xl border border-border-subtle leading-relaxed">{task.description}</p>}

                <div className="rounded-xl bg-amber-50 border border-amber-200 p-3 text-xs leading-relaxed text-amber-800">
                  <p className="font-bold">Petunjuk:</p>
                  {customInstruction ? (
                    <p className="mt-1 whitespace-pre-wrap">{customInstruction}</p>
                  ) : isCvTask ? (
                    <ol className="list-decimal list-inside space-y-1 mt-1">
                      <li>Gunakan kemeja/pakaian berkerah rapi seperti standar foto CV.</li>
                      <li>Pegang kertas kecil di depan dada bertuliskan: <span className="font-bold">[Nama Lengkap] - Siap Kerja {todayStr}</span></li>
                      <li>Wajah tegak menghadap kamera, tersenyum ramah, latar dinding polos.</li>
                      <li>Ambil foto langsung menggunakan kamera HP lalu kirim.</li>
                    </ol>
                  ) : isDiscussionTask ? (
                    <ol className="list-decimal list-inside space-y-1 mt-1">
                      <li>Selesaikan sesi diskusi bersama teman/rekanmu.</li>
                      <li>Buka kamera dan ambil foto bersama teman diskusimu langsung di tempat.</li>
                      <li>Pastikan wajah atau bukti aktivitas terlihat jelas lalu kirim.</li>
                    </ol>
                  ) : (
                    <ol className="list-decimal list-inside space-y-1 mt-1">
                      <li>Ambil foto bukti aktivitas langsung menggunakan kamera perangkat.</li>
                      <li>Pastikan objek foto terlihat jelas dan memiliki pencahayaan yang cukup.</li>
                      <li>Kirim foto langsung setelah diambil untuk verifikasi admin.</li>
                    </ol>
                  )}
                  <p className="mt-2 text-[11px] text-amber-700">Foto harus diambil langsung dari kamera — tidak ada pilihan galeri.</p>
                </div>

                {!isSupported && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2" data-testid="camera-error">
                    <AlertCircle className="w-5 h-5 shrink-0" />
                    <span>Perangkat/browser Anda tidak mendukung akses kamera.</span>
                  </div>
                )}

                {/* Hidden canvas for capture */}
                <canvas ref={canvasRef} className="hidden" />

                {!previewUrl ? (
                  <>
                    {!isCameraOpen ? (
                      <button
                        type="button"
                        data-testid="open-camera-button"
                        onClick={handleOpenCamera}
                        disabled={isRequesting || !isSupported}
                        className="w-full py-4 rounded-2xl bg-accent-magic text-white font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 disabled:opacity-50 flex items-center justify-center gap-2 cursor-pointer"
                      >
                        <Camera className="w-5 h-5" />
                        {isRequesting ? 'Membuka Kamera...' : isRevision ? 'Ambil Foto Ulang' : 'Buka Kamera'}
                      </button>
                    ) : (
                      <div className="space-y-3">
                        <div className="relative aspect-[4/3] w-full max-w-[360px] mx-auto rounded-2xl overflow-hidden border-2 border-accent-magic bg-black">
                          <video
                            ref={(el) => {
                              videoRef.current = el
                              if (el && streamRef.current && el.srcObject !== streamRef.current) {
                                el.srcObject = streamRef.current
                                el.onloadedmetadata = () => {
                                  el.play().catch((err) => console.warn('Video play error:', err))
                                }
                                el.play().catch((err) => console.warn('Video play error:', err))
                              }
                            }}
                            autoPlay
                            playsInline
                            muted
                            className={`w-full h-full object-cover transition-transform ${activeFacing === 'user' ? 'scale-x-[-1]' : ''}`}
                            data-testid="camera-preview"
                          />
                          {/* Floating Switch Camera Button on Viewfinder */}
                          <button
                            type="button"
                            data-testid="switch-camera-overlay-button"
                            onClick={handleSwitchCamera}
                            disabled={isSwitchingCamera || isRequesting}
                            className="absolute top-2.5 right-2.5 z-10 px-3 py-1.5 rounded-full bg-black/65 hover:bg-black/85 active:scale-95 text-white backdrop-blur-md flex items-center gap-1.5 text-xs font-bold transition-all shadow-md cursor-pointer border border-white/20"
                            title="Putar Kamera (Depan / Belakang)"
                            aria-label="Putar Kamera"
                          >
                            <SwitchCamera className={`w-3.5 h-3.5 text-accent-magic ${isSwitchingCamera ? 'animate-spin' : ''}`} />
                            <span>{activeFacing === 'user' ? 'Kamera Depan' : 'Kamera Belakang'}</span>
                          </button>
                        </div>
                        <div className="flex gap-2">
                          <button
                            type="button"
                            onClick={handleCancelCamera}
                            className="py-3 px-3.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary font-bold hover:bg-surface cursor-pointer"
                          >
                            Batal
                          </button>
                          <button
                            type="button"
                            data-testid="switch-camera-button"
                            onClick={handleSwitchCamera}
                            disabled={isSwitchingCamera || isRequesting}
                            className="py-3 px-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold hover:bg-surface flex items-center justify-center gap-1.5 cursor-pointer shadow-xs active:scale-95 transition-all"
                            title="Putar Kamera Depan / Belakang"
                            aria-label="Putar Kamera"
                          >
                            <SwitchCamera className={`w-4 h-4 text-accent-magic ${isSwitchingCamera ? 'animate-spin' : ''}`} />
                            <span className="text-xs whitespace-nowrap">Putar Kamera</span>
                          </button>
                          <button
                            type="button"
                            data-testid="capture-button"
                            onClick={handleCapture}
                            disabled={compressing || isSwitchingCamera}
                            className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold shadow hover:brightness-110 disabled:opacity-50 flex items-center justify-center gap-2 cursor-pointer active:scale-95 transition-all"
                          >
                            <Camera className="w-4 h-4" /> Ambil Foto
                          </button>
                        </div>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="space-y-3">
                    <div className="relative aspect-[4/3] w-full max-w-[360px] mx-auto rounded-2xl overflow-hidden border-2 border-accent-magic bg-black">
                      <img src={previewUrl} alt="Preview Foto" className="w-full h-full object-cover" data-testid="capture-preview" />
                      <div className="absolute top-2 right-2 px-2.5 py-1 rounded-full bg-black/60 text-[10px] text-white">{capturedFile ? (capturedFile.size / 1024).toFixed(0) : 0} KB</div>
                    </div>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        data-testid="retake-button"
                        onClick={handleRetake}
                        className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary font-bold flex items-center justify-center gap-1.5"
                      >
                        <RefreshCw className="w-4 h-4" /> Ambil Ulang
                      </button>
                      <button
                        type="button"
                        data-testid="use-photo-button"
                        onClick={handleSubmit}
                        disabled={uploading || compressing}
                        className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold shadow hover:brightness-110 disabled:opacity-50 flex items-center justify-center gap-2"
                      >
                        {uploading ? 'Mengirim...' : (
                          <>
                            <CheckCircle2 className="w-4 h-4" /> Gunakan Foto
                          </>
                        )}
                      </button>
                    </div>
                  </div>
                )}

                {compressing && <div className="p-3 text-center text-xs text-accent-magic animate-pulse">Mengompres foto...</div>}
                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2" data-testid="camera-error">
                    <AlertCircle className="w-5 h-5 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}
              </motion.div>
            ) : (
              <motion.div key="submitted" initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="py-6 text-center space-y-5">
                <div className={`w-20 h-20 mx-auto rounded-full ${task.status === 'APPROVED' ? 'bg-status-success/20 text-status-success' : 'bg-accent-gold/20 text-accent-gold'} flex items-center justify-center`}>
                  {task.status === 'APPROVED' ? <CheckCircle2 className="w-12 h-12" /> : <Clock className="w-12 h-12" />}
                </div>
                <div>
                  <h4 className="font-bold text-2xl text-text-primary">{task.status === 'APPROVED' ? 'Foto Selesai (Disetujui)! 📸' : 'Foto Berhasil Dikirim! 📸'}</h4>
                  <p className="text-sm text-text-secondary mt-1">{task.status === 'APPROVED' ? `Tugas disetujui +${task.coins_earned || task.reward_coins} Koin` : 'Bukti foto masuk antrean verifikasi admin.'}</p>
                </div>
                <button
                  onClick={() => {
                    if (onNextTask) onNextTask()
                    else { onSuccess(); onClose() }
                  }}
                  className="w-full py-4 rounded-2xl bg-accent-magic text-white font-bold shadow flex items-center justify-center gap-2"
                >
                  <span>Lanjut Tugas Berikutnya</span>
                  <ArrowRight className="w-5 h-5" />
                </button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  )
}
