import React, { useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Camera, RefreshCw, X, CheckCircle2, AlertCircle, Sparkles, Clock, ArrowRight } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { compressImage, uploadTaskProof } from '../../shared/lib/compress'
import { isEarningCapError, EARNING_CAP_MESSAGE } from '../../shared/lib/earning'

interface CameraCaptureModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
  onNextTask?: () => void
}

export const CameraCaptureModal: React.FC<CameraCaptureModalProps> = ({ task, onClose, onSuccess, onNextTask }) => {
  const isAlreadyDone = task.status === 'APPROVED' || task.status === 'PENDING'
  const cameraInputRef = useRef<HTMLInputElement>(null)
  const [capturedFile, setCapturedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [compressing, setCompressing] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(isAlreadyDone)

  const handleCapture = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const originalFile = e.target.files[0]
      setCompressing(true)
      setErrorMessage(null)
      try {
        // Compress client-side to <= 1280px, quality 0.7, with timestamp watermark
        const result = await compressImage(originalFile, {
          maxWidth: 1280,
          maxHeight: 1280,
          quality: 0.7,
        })
        setCapturedFile(result.file)
        setPreviewUrl(result.dataUrl)
      } catch (err: any) {
        setErrorMessage('Gagal memproses foto kamera: ' + err.message)
      } finally {
        setCompressing(false)
      }
    }
  }

  const handleSubmit = async () => {
    if (!capturedFile) return
    setUploading(true)
    setErrorMessage(null)
    try {
      // 1. Upload compressed image to Supabase storage bucket
      const uploadRes = await uploadTaskProof(capturedFile)

      // 2. Submit task manual verification
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
        setErrorMessage(err.message || 'Gagal mengunggah foto. Silakan periksa koneksi internet.')
      }
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.97, opacity: 0, y: 12 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.97, opacity: 0, y: 12 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[92vh] sm:max-h-[90vh]"
      >
        {/* Header — compact unified */}
        <div className="flex items-center justify-between px-5 py-3.5 border-b border-border-subtle bg-surface shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            <span className="w-7 h-7 rounded-full bg-accent-magic/12 text-accent-magic flex items-center justify-center font-bold text-xs shrink-0">
              #{task.step_order}
            </span>
            <h3 className="font-bold text-text-primary text-[14px] line-clamp-1">
              {task.title}
            </h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup"
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors shrink-0"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content Body */}
        <div className="flex-1 overflow-y-auto p-5 space-y-5">
          <AnimatePresence mode="wait">
            {!submitted ? (
              <motion.div
                key="camera-form"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-5"
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">
                    📸 Ambil Foto Bukti
                  </span>
                  <div className="text-xs text-text-secondary font-bold">
                    +{task.reward_coins} 🪙 | +{task.reward_xp} XP
                  </div>
                </div>

                {task.description && (
                  <p className="text-sm text-text-secondary bg-surface p-4 rounded-2xl border border-border-subtle leading-relaxed">
                    {task.description}
                  </p>
                )}

                {/* Hidden native camera input */}
                <input
                  ref={cameraInputRef}
                  type="file"
                  accept="image/*"
                  capture="environment"
                  className="hidden"
                  onChange={handleCapture}
                />

                {/* Camera Viewfinder / Preview */}
                {!previewUrl ? (
                  <div
                    onClick={() => cameraInputRef.current?.click()}
                    className="aspect-square w-full max-w-[280px] mx-auto border-2 border-dashed border-accent-magic/50 bg-accent-magic/5 rounded-2xl flex flex-col items-center justify-center gap-3 cursor-pointer hover:bg-accent-magic/10 active:scale-95 transition-all p-6 text-center shadow-inner"
                  >
                    <div className="w-16 h-16 rounded-full bg-accent-magic text-white flex items-center justify-center shadow-lg shadow-accent-magic/30">
                      <Camera className="w-8 h-8" />
                    </div>
                    <div>
                      <p className="font-heading font-bold text-text-primary text-base">
                        Buka Kamera HP
                      </p>
                      <p className="text-xs text-text-secondary mt-1">
                        Ketuk untuk mengambil foto bukti langsung
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-3">
                    <div className="relative aspect-square w-full max-w-[280px] mx-auto rounded-2xl overflow-hidden border-2 border-accent-magic shadow-md bg-black">
                      <img
                        src={previewUrl}
                        alt="Preview Foto"
                        className="w-full h-full object-cover"
                      />
                      <div className="absolute top-2 right-2 px-2.5 py-1 rounded-full bg-black/60 backdrop-blur-md text-[10px] text-white font-medium">
                        {(capturedFile ? capturedFile.size / 1024 : 0).toFixed(0)} KB (Optimal)
                      </div>
                    </div>
                    <div className="text-center">
                      <button
                        type="button"
                        onClick={() => cameraInputRef.current?.click()}
                        className="inline-flex items-center gap-1.5 text-xs text-accent-magic font-bold hover:underline"
                      >
                        <RefreshCw className="w-3.5 h-3.5" />
                        <span>Ambil Ulang Foto</span>
                      </button>
                    </div>
                  </div>
                )}

                {compressing && (
                  <div className="p-3 text-center text-xs text-accent-magic animate-pulse">
                    Mengompres foto & menambahkan watermark...
                  </div>
                )}

                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2">
                    <AlertCircle className="w-5 h-5 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                <button
                  type="button"
                  disabled={!capturedFile || uploading || compressing}
                  onClick={handleSubmit}
                  className="w-full py-4 rounded-2xl bg-accent-magic text-white font-heading font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all flex items-center justify-center gap-2"
                >
                  {uploading ? (
                    <span>Mengirim Bukti Foto...</span>
                  ) : (
                    <>
                      <Sparkles className="w-4 h-4" />
                      <span>Kirim Foto Bukti</span>
                    </>
                  )}
                </button>
              </motion.div>
            ) : (
              <motion.div
                key="submitted-state"
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="py-6 text-center space-y-5"
              >
                <div className={`w-20 h-20 mx-auto rounded-full ${task.status === 'APPROVED' ? 'bg-status-success/20 text-status-success' : 'bg-accent-gold/20 text-accent-gold'} flex items-center justify-center`}>
                  {task.status === 'APPROVED' ? <CheckCircle2 className="w-12 h-12" /> : <Clock className="w-12 h-12" />}
                </div>
                <div>
                  <h4 className="font-heading font-bold text-2xl text-text-primary">
                    {task.status === 'APPROVED' ? 'Foto Selesai (Disetujui)! 📸' : 'Foto Berhasil Dikirim! 📸'}
                  </h4>
                  <p className="text-sm text-text-secondary mt-1">
                    {task.status === 'APPROVED'
                      ? `Tugas telah disetujui dan kamu mendapatkan +${task.coins_earned || task.reward_coins} Koin & +${task.xp_earned || task.reward_xp} EXP.`
                      : 'Bukti foto telah masuk ke antrean verifikasi admin. Koin & EXP akan otomatis masuk saat disetujui.'}
                  </p>
                </div>

                <div className="p-4 rounded-2xl bg-surface border border-border-subtle text-left flex items-start gap-3">
                  <CheckCircle2 className="w-5 h-5 text-status-success shrink-0 mt-0.5" />
                  <p className="text-xs text-text-secondary leading-relaxed">
                    <strong className="text-text-primary">Tugas berikutnya sudah terbuka!</strong> Kamu tidak perlu menunggu verifikasi admin untuk lanjut ke tugas berikutnya.
                  </p>
                </div>

                <button
                  onClick={() => {
                    if (onNextTask) {
                      onNextTask()
                    } else {
                      onSuccess()
                      onClose()
                    }
                  }}
                  className="w-full py-4 rounded-2xl bg-accent-magic text-white font-heading font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 active:scale-[0.98] transition-all flex items-center justify-center gap-2"
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
