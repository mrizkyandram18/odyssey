import React, { useState, useRef, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Video, RefreshCw, X, CheckCircle2, AlertCircle, Clock, ArrowRight, Square, Circle } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { uploadTaskProof } from '../../shared/lib/compress'
import { isEarningCapError, EARNING_CAP_MESSAGE } from '../../shared/lib/earning'

interface VideoRecordModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

function pickSupportedMimeType(): { mimeType?: string; extension: string } {
  if (typeof MediaRecorder === 'undefined' || typeof (MediaRecorder as any).isTypeSupported !== 'function') {
    return { mimeType: undefined, extension: 'webm' }
  }
  const candidates: Array<{ mimeType: string; extension: string }> = [
    { mimeType: 'video/webm;codecs=vp9', extension: 'webm' },
    { mimeType: 'video/webm;codecs=vp8', extension: 'webm' },
    { mimeType: 'video/webm', extension: 'webm' },
    { mimeType: 'video/mp4', extension: 'mp4' },
  ]
  for (const c of candidates) {
    try {
      if ((MediaRecorder as any).isTypeSupported(c.mimeType)) return c
    } catch { /* try next */ }
  }
  return { mimeType: undefined, extension: 'webm' }
}

function safePlay(video: HTMLVideoElement | null) {
  if (!video) return
  try {
    const p = video.play() as unknown as Promise<void> | undefined
    if (p && typeof p.catch === 'function') {
      p.catch((err) => console.warn('Video play error:', err))
    }
  } catch (err) {
    console.warn('Video play error:', err)
  }
}

function formatTimer(totalSeconds: number): string {
  const m = Math.floor(totalSeconds / 60)
  const s = totalSeconds % 60
  return `${m}:${String(s).padStart(2, '0')}`
}

export const VideoRecordModal: React.FC<VideoRecordModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const isAlreadyDone = task.status === 'APPROVED' || task.status === 'PENDING'
  const videoRef = useRef<HTMLVideoElement>(null)
  const previewRef = useRef<HTMLVideoElement>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const recorderRef = useRef<MediaRecorder | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const timerRef = useRef<number | null>(null)

  const recordingCfg = task.config?.recording || {}
  const maxDuration = Math.max(1, Math.min(600, Number(recordingCfg.max_duration_seconds) || 60))
  const rawFacing = recordingCfg.camera_facing || 'user'
  const cameraFacing: 'user' | 'environment' = rawFacing === 'environment' ? 'environment' : 'user'
  const instruction = (recordingCfg.instruction as string | undefined)?.trim() || ''

  const [isSupported] = useState(() => typeof navigator !== 'undefined' && !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia))
  const [recorderSupported] = useState(() => typeof MediaRecorder !== 'undefined')
  const [isCameraOpen, setIsCameraOpen] = useState(false)
  const [isRequesting, setIsRequesting] = useState(false)
  const [isRecording, setIsRecording] = useState(false)
  const [elapsed, setElapsed] = useState(0)
  const [recordedFile, setRecordedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [actualMime, setActualMime] = useState<string>('')
  const [uploading, setUploading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(isAlreadyDone)

  const stopStream = useCallback(() => {
    if (timerRef.current !== null) {
      window.clearInterval(timerRef.current)
      timerRef.current = null
    }
    try { recorderRef.current?.stream?.getTracks()?.forEach((t) => t.stop()) } catch { /* ignore */ }
    recorderRef.current = null
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((track) => track.stop())
      streamRef.current = null
    }
    if (videoRef.current) videoRef.current.srcObject = null
  }, [])

  useEffect(() => {
    return () => {
      stopStream()
      if (previewUrl && typeof URL.revokeObjectURL === 'function') URL.revokeObjectURL(previewUrl)
    }
  }, [stopStream, previewUrl])

  useEffect(() => {
    if (isCameraOpen && streamRef.current && videoRef.current && !previewUrl) {
      const video = videoRef.current
      if (video.srcObject !== streamRef.current) video.srcObject = streamRef.current
      safePlay(video)
    }
  }, [isCameraOpen, previewUrl])

  const handleOpenCamera = async () => {
    if (!isSupported) {
      setErrorMessage('Perangkat/browser Anda tidak mendukung akses kamera. Silakan gunakan browser terbaru di HP Anda.')
      return
    }
    if (!recorderSupported) {
      setErrorMessage('Browser Anda tidak mendukung perekaman video (MediaRecorder). Silakan gunakan Chrome atau Safari terbaru.')
      return
    }
    setIsRequesting(true)
    setErrorMessage(null)
    try {
      let stream: MediaStream
      try {
        stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: cameraFacing }, audio: true })
      } catch {
        stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      }
      streamRef.current = stream
      setIsCameraOpen(true)
    } catch (err: any) {
      const name = err?.name || ''
      if (name === 'NotAllowedError' || name === 'PermissionDeniedError') {
        setErrorMessage('Izin kamera/mikrofon belum aktif. Ketuk ikon gembok 🔒 di sebelah alamat web lalu pilih Izinkan akses kamera & mikrofon.')
      } else if (name === 'NotFoundError' || name === 'OverconstrainedError') {
        setErrorMessage('Kamera tidak ditemukan di perangkat ini.')
      } else if (name === 'NotReadableError') {
        setErrorMessage('Kamera sedang digunakan aplikasi lain. Tutup aplikasi kamera lain lalu coba lagi.')
      } else {
        setErrorMessage('Gagal membuka kamera: ' + (err?.message || 'Pastikan izin kamera aktif'))
      }
    } finally {
      setIsRequesting(false)
    }
  }

  const stopRecording = useCallback(() => {
    const rec = recorderRef.current
    if (rec && rec.state !== 'inactive') {
      rec.stop()
    }
    if (timerRef.current !== null) {
      window.clearInterval(timerRef.current)
      timerRef.current = null
    }
    setIsRecording(false)
  }, [])

  const handleStartRecording = () => {
    if (!streamRef.current) return
    setErrorMessage(null)
    chunksRef.current = []
    const picked = pickSupportedMimeType()
    try {
      const rec = picked.mimeType
        ? new MediaRecorder(streamRef.current, { mimeType: picked.mimeType, videoBitsPerSecond: 800_000 })
        : new MediaRecorder(streamRef.current)
      recorderRef.current = rec
      setActualMime(rec.mimeType || picked.mimeType || '')
      rec.ondataavailable = (e: BlobEvent) => {
        if (e.data && e.data.size > 0) chunksRef.current.push(e.data)
      }
      rec.onerror = () => setErrorMessage('Perekaman gagal. Coba lagi.')
      rec.onstop = () => {
        const type = rec.mimeType || picked.mimeType || 'video/webm'
        const ext = type.includes('mp4') ? 'mp4' : 'webm'
        const blob = new Blob(chunksRef.current, { type })
        const file = new File([blob], `video-${Date.now()}.${ext}`, { type })
        setRecordedFile(file)
        setPreviewUrl((old) => {
          if (old && typeof URL.revokeObjectURL === 'function') URL.revokeObjectURL(old)
          return URL.createObjectURL(blob)
        })
        stopStream()
        setIsCameraOpen(false)
      }
      rec.start(500)
      setIsRecording(true)
      setElapsed(0)
      timerRef.current = window.setInterval(() => {
        setElapsed((prev) => {
          const next = prev + 1
          if (next >= maxDuration) {
            window.setTimeout(() => stopRecording(), 0)
          }
          return Math.min(next, maxDuration)
        })
      }, 1000)
    } catch (err: any) {
      setErrorMessage('Tidak dapat memulai perekaman: ' + (err?.message || 'browser tidak mendukung format video'))
    }
  }

  const handleRetake = () => {
    if (previewUrl && typeof URL.revokeObjectURL === 'function') URL.revokeObjectURL(previewUrl)
    setPreviewUrl(null)
    setRecordedFile(null)
    setElapsed(0)
    setErrorMessage(null)
    handleOpenCamera()
  }

  const handleSubmit = async () => {
    if (!recordedFile) return
    setUploading(true)
    setErrorMessage(null)
    try {
      const uploadRes = await uploadTaskProof(recordedFile)
      const res = await tasksApi.submit(task.id, {
        payload: {
          file_url: uploadRes.file_url,
          file_name: uploadRes.file_name,
          file_size: uploadRes.file_size,
          // Client-reported only — admin verifies actual duration by watching.
          duration_seconds: elapsed,
          mime_type: actualMime || recordedFile.type,
          camera_facing: cameraFacing,
          captured_at: new Date().toISOString(),
        },
      })
      if (res.success) {
        setSubmitted(true)
      } else {
        setErrorMessage(res.error || 'Gagal menyimpan submission video')
      }
    } catch (err: any) {
      if (isEarningCapError(err)) {
        setErrorMessage(EARNING_CAP_MESSAGE)
        try { onSuccess() } catch { /* ignore */ }
      } else {
        setErrorMessage(err.message || 'Gagal mengunggah video. Periksa koneksi internet.')
      }
    } finally {
      setUploading(false)
    }
  }

  const handleClose = () => {
    stopStream()
    onClose()
  }

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
              <motion.div key="video-record" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-5">
                <div className="flex items-center justify-between">
                  <span className="text-xs bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">🎥 Rekam Video</span>
                  <div className="text-xs text-text-secondary font-bold">+{task.reward_coins} 🪙 | +{task.reward_xp} XP</div>
                </div>

                {task.description && <p className="text-sm text-text-secondary bg-surface p-4 rounded-2xl border border-border-subtle leading-relaxed">{task.description}</p>}

                <div className="rounded-xl bg-amber-50 border border-amber-200 p-3 text-xs leading-relaxed text-amber-800">
                  <p className="font-bold">Petunjuk (maksimal {maxDuration} detik):</p>
                  {instruction ? (
                    <p className="mt-1 whitespace-pre-wrap">{instruction}</p>
                  ) : (
                    <ol className="list-decimal list-inside space-y-1 mt-1">
                      <li>Perkenalkan siapa kamu.</li>
                      <li>Ceritakan satu hal yang kamu kuasai.</li>
                      <li>Ceritakan satu hal yang sedang kamu pelajari.</li>
                      <li>Ceritakan pekerjaan/aktivitas yang ingin kamu coba.</li>
                    </ol>
                  )}
                  <p className="mt-2 text-[11px] text-amber-700">Video harus direkam langsung dari kamera — bukan upload file.</p>
                </div>

                {!isSupported || !recorderSupported ? (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2" data-testid="camera-error">
                    <AlertCircle className="w-5 h-5 shrink-0" />
                    <span>Perangkat/browser Anda tidak mendukung perekaman video langsung. Gunakan Chrome/Safari terbaru di HP Anda.</span>
                  </div>
                ) : !previewUrl ? (
                  <>
                    {!isCameraOpen ? (
                      <button
                        type="button"
                        data-testid="open-camera-button"
                        onClick={handleOpenCamera}
                        disabled={isRequesting}
                        className="w-full py-4 rounded-2xl bg-accent-magic text-white font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 disabled:opacity-50 flex items-center justify-center gap-2"
                      >
                        <Video className="w-5 h-5" />
                        {isRequesting ? 'Membuka Kamera...' : 'Buka Kamera'}
                      </button>
                    ) : (
                      <div className="space-y-3">
                        <div className="relative aspect-[4/3] w-full max-w-[360px] mx-auto rounded-2xl overflow-hidden border-2 border-accent-magic bg-black">
                          <video ref={videoRef} autoPlay playsInline muted className="w-full h-full object-cover" data-testid="camera-preview" />
                          {isRecording && (
                            <div className="absolute top-2 left-2 flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-black/60 text-xs text-white font-bold" data-testid="recording-indicator">
                              <Circle className="w-2.5 h-2.5 fill-red-500 text-red-500 animate-pulse" />
                              <span data-testid="recording-timer">{formatTimer(elapsed)} / {formatTimer(maxDuration)}</span>
                            </div>
                          )}
                        </div>
                        <div className="flex gap-2">
                          {!isRecording ? (
                            <>
                              <button type="button" onClick={() => { stopStream(); setIsCameraOpen(false) }} className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary font-bold hover:bg-surface">
                                Batal
                              </button>
                              <button
                                type="button"
                                data-testid="start-recording-button"
                                onClick={handleStartRecording}
                                className="flex-1 py-3 rounded-xl bg-red-500 text-white font-bold shadow hover:brightness-110 flex items-center justify-center gap-2"
                              >
                                <Circle className="w-4 h-4 fill-white" /> Mulai Rekam
                              </button>
                            </>
                          ) : (
                            <button
                              type="button"
                              data-testid="stop-recording-button"
                              onClick={stopRecording}
                              className="w-full py-3 rounded-xl bg-surface-elevated border-2 border-red-500 text-red-500 font-bold flex items-center justify-center gap-2"
                            >
                              <Square className="w-4 h-4 fill-red-500" /> Berhenti ({formatTimer(maxDuration - elapsed)} tersisa)
                            </button>
                          )}
                        </div>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="space-y-3">
                    <div className="relative aspect-[4/3] w-full max-w-[360px] mx-auto rounded-2xl overflow-hidden border-2 border-accent-magic bg-black">
                      <video ref={previewRef} src={previewUrl} controls playsInline className="w-full h-full object-contain" data-testid="record-preview" />
                    </div>
                    <p className="text-center text-[11px] text-text-secondary">
                      Durasi: {formatTimer(elapsed)} • Ukuran: {recordedFile ? (recordedFile.size / 1024 / 1024).toFixed(1) : 0} MB
                    </p>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        data-testid="retake-button"
                        onClick={handleRetake}
                        className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-secondary font-bold flex items-center justify-center gap-1.5"
                      >
                        <RefreshCw className="w-4 h-4" /> Rekam Ulang
                      </button>
                      <button
                        type="button"
                        data-testid="use-video-button"
                        onClick={handleSubmit}
                        disabled={uploading}
                        className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold shadow hover:brightness-110 disabled:opacity-50 flex items-center justify-center gap-2"
                      >
                        {uploading ? 'Mengirim...' : (
                          <>
                            <CheckCircle2 className="w-4 h-4" /> Gunakan Video
                          </>
                        )}
                      </button>
                    </div>
                  </div>
                )}

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
                  <h4 className="font-bold text-2xl text-text-primary">{task.status === 'APPROVED' ? 'Video Selesai (Disetujui)! 🎥' : 'Video Berhasil Dikirim! 🎥'}</h4>
                  <p className="text-sm text-text-secondary mt-1">{task.status === 'APPROVED' ? `Tugas disetujui +${task.coins_earned || task.reward_coins} Koin` : 'Video masuk antrean verifikasi admin.'}</p>
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
