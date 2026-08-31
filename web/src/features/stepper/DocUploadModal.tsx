import React, { useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { FileUp, FileText, X, CheckCircle2, AlertCircle, Sparkles, Clock, ArrowRight, Download } from 'lucide-react'
import type { TaskView } from '../../shared/types'
import { tasksApi } from '../../shared/lib/api'
import { uploadTaskProof } from '../../shared/lib/compress'

interface DocUploadModalProps {
  task: TaskView
  onClose: () => void
  onSuccess: () => void
}

export const DocUploadModal: React.FC<DocUploadModalProps> = ({ task, onClose, onSuccess }) => {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [note, setNote] = useState('')
  const [uploading, setUploading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(false)

  const attachmentUrl = task.config?.attachment_url || ''
  const attachmentName = task.config?.attachment_name || 'Dokumen Template'
  const acceptedExtensions = task.config?.accepted_extensions?.join(',') || '.pdf,.xlsx,.xls,.docx,.doc,.csv,.txt'
  const maxSizeBytes = (task.config?.max_file_size_mb || 10) * 1024 * 1024

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0]
      if (file.size > maxSizeBytes) {
        setErrorMessage(`Ukuran file maksimal ${(maxSizeBytes / (1024 * 1024)).toFixed(0)} MB`)
        return
      }
      setSelectedFile(file)
      setErrorMessage(null)
    }
  }

  const handleSubmit = async () => {
    if (!selectedFile) return
    setUploading(true)
    setErrorMessage(null)
    try {
      // 1. Upload document file to backend storage
      const uploadRes = await uploadTaskProof(selectedFile)

      // 2. Submit task manual verification
      const res = await tasksApi.submit(task.id, {
        payload: {
          file_url: uploadRes.file_url,
          file_name: uploadRes.file_name,
          file_size: uploadRes.file_size,
          note: note.trim(),
          submitted_at: new Date().toISOString(),
        },
      })

      if (res.success) {
        setSubmitted(true)
      } else {
        setErrorMessage(res.error || 'Gagal menyimpan submission dokumen')
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Gagal mengunggah dokumen. Silakan periksa koneksi.')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fadeIn">
      <motion.div
        initial={{ scale: 0.95, opacity: 0, y: 20 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0.95, opacity: 0, y: 20 }}
        className="w-full max-w-lg bg-surface-elevated border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border-subtle bg-surface">
          <div className="flex items-center gap-2">
            <span className="w-8 h-8 rounded-full bg-accent-magic/20 text-accent-magic flex items-center justify-center font-bold text-sm">
              #{task.step_order}
            </span>
            <h3 className="font-heading font-bold text-text-primary text-base md:text-lg line-clamp-1">
              {task.title}
            </h3>
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup"
            className="w-8 h-8 rounded-full bg-surface-elevated text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          <AnimatePresence mode="wait">
            {!submitted ? (
              <motion.div
                key="upload-form"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-5"
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">
                    📄 Upload Dokumen Tugas
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

                {/* Optional Template / Attachment Download Box */}
                {attachmentUrl && (
                  <div className="p-4 rounded-2xl bg-surface border border-border-subtle flex items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-xl bg-accent-magic/15 text-accent-magic flex items-center justify-center shrink-0">
                        <FileText className="w-5 h-5" />
                      </div>
                      <div>
                        <p className="font-heading font-bold text-text-primary text-xs line-clamp-1">
                          {attachmentName}
                        </p>
                        <p className="text-[10px] text-text-secondary">
                          Download template untuk dikerjakan
                        </p>
                      </div>
                    </div>
                    <a
                      href={attachmentUrl}
                      target="_blank"
                      rel="noreferrer"
                      download
                      className="px-3.5 py-2 rounded-xl bg-accent-magic/10 hover:bg-accent-magic/20 text-accent-magic font-bold text-xs flex items-center gap-1.5 transition-colors shrink-0"
                    >
                      <Download className="w-3.5 h-3.5" />
                      <span>Download</span>
                    </a>
                  </div>
                )}

                {/* File Dropzone */}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept={acceptedExtensions}
                  className="hidden"
                  onChange={handleFileChange}
                />

                <div
                  onClick={() => fileInputRef.current?.click()}
                  className={`p-6 border-2 border-dashed rounded-2xl cursor-pointer text-center transition-all ${
                    selectedFile
                      ? 'border-accent-magic bg-accent-magic/10'
                      : 'border-border-subtle hover:border-accent-magic/40 bg-surface'
                  }`}
                >
                  {selectedFile ? (
                    <div className="flex flex-col items-center gap-2 text-text-primary">
                      <FileText className="w-10 h-10 text-accent-magic animate-bounce" />
                      <p className="font-bold text-sm line-clamp-1">{selectedFile.name}</p>
                      <p className="text-xs text-text-secondary">
                        {(selectedFile.size / 1024).toFixed(1)} KB — Ketuk untuk ganti file
                      </p>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center gap-2 text-text-secondary">
                      <FileUp className="w-10 h-10 text-accent-magic" />
                      <p className="font-bold text-text-primary text-sm">Pilih Dokumen Selesai</p>
                      <p className="text-xs">Mendukung {acceptedExtensions} (Maks {(maxSizeBytes / (1024 * 1024)).toFixed(0)} MB)</p>
                    </div>
                  )}
                </div>

                {/* Optional Note */}
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">Catatan Tambahan (Opsional):</label>
                  <textarea
                    rows={2}
                    value={note}
                    onChange={(e) => setNote(e.target.value)}
                    placeholder="Tuliskan keterangan singkat jika ada..."
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic resize-none"
                  />
                </div>

                {errorMessage && (
                  <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-sm flex items-center gap-2">
                    <AlertCircle className="w-5 h-5 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                <button
                  type="button"
                  disabled={!selectedFile || uploading}
                  onClick={handleSubmit}
                  className="w-full py-4 rounded-2xl bg-accent-magic text-white font-heading font-bold shadow-lg shadow-accent-magic/30 hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all flex items-center justify-center gap-2"
                >
                  {uploading ? (
                    <span>Mengunggah Dokumen...</span>
                  ) : (
                    <>
                      <Sparkles className="w-4 h-4" />
                      <span>Kirim Dokumen Bukti</span>
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
                <div className="w-20 h-20 mx-auto rounded-full bg-accent-gold/20 text-accent-gold flex items-center justify-center">
                  <Clock className="w-12 h-12" />
                </div>
                <div>
                  <h4 className="font-heading font-bold text-2xl text-text-primary">
                    Dokumen Berhasil Terkirim! 📄
                  </h4>
                  <p className="text-sm text-text-secondary mt-1">
                    Dokumenmu sedang dalam antrean verifikasi admin. Koin & EXP akan otomatis masuk setelah disetujui.
                  </p>
                </div>

                <div className="p-4 rounded-2xl bg-surface border border-border-subtle text-left flex items-start gap-3">
                  <CheckCircle2 className="w-5 h-5 text-status-success shrink-0 mt-0.5" />
                  <p className="text-xs text-text-secondary leading-relaxed">
                    <strong className="text-text-primary">Tugas berikutnya sudah terbuka!</strong> Kamu bisa langsung melanjutkan tugas step berikutnya tanpa perlu menunggu verifikasi admin.
                  </p>
                </div>

                <button
                  onClick={() => {
                    onSuccess()
                    onClose()
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
