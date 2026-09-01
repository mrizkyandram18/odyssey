import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Edit3, Play, HelpCircle, Camera, FileText, PenLine, Gamepad2 } from 'lucide-react'
import type { TaskView, TaskType } from '../../../../shared/types'

interface EditTaskModalProps {
  task: TaskView | null
  form: {
    title: string
    description: string
    task_type: TaskType
    reward_coins: number
    reward_xp: number
    video_url: string
    questions: Array<{ id: string; question: string; options: string[]; correct_answer: string }>
    photo_min_count: number
    doc_allowed_extensions: string
    doc_max_size_mb: number
    text_prompt: string
    text_min_chars: number
    text_max_chars: number
    game_type: string
    game_target_score: number
    game_max_moves: number
    config: Record<string, any>
  }
  setForm: React.Dispatch<
    React.SetStateAction<{
      title: string
      description: string
      task_type: TaskType
      reward_coins: number
      reward_xp: number
      video_url: string
      questions: Array<{ id: string; question: string; options: string[]; correct_answer: string }>
      photo_min_count: number
      doc_allowed_extensions: string
      doc_max_size_mb: number
      text_prompt: string
      text_min_chars: number
      text_max_chars: number
      game_type: string
      game_target_score: number
      game_max_moves: number
      config: Record<string, any>
    }>
  >
  isSaving: boolean
  onClose: () => void
  onSave: () => void
}

const TYPE_OPTIONS: { value: TaskType; label: string }[] = [
  { value: 'PHOTO_UPLOAD', label: '📸 Foto' },
  { value: 'MINI_GAME', label: '🎮 Mini Games' },
  { value: 'TEXT_RESPONSE', label: '✍️ Esai' },
  { value: 'QUIZ', label: '🧠 Kuis' },
  { value: 'VIDEO', label: '🎥 Video' },
  { value: 'DOCUMENT_UPLOAD', label: '📄 Dokumen' },
  { value: 'GENERAL', label: '⚙️ Umum' },
]

const isLegacyType = (t: string) => ['VIDEO_QUIZ', 'PHOTO_PROOF', 'YOUTUBE_VIDEO'].includes(t)

export const EditTaskModal: React.FC<EditTaskModalProps> = ({
  task,
  form,
  setForm,
  isSaving,
  onClose,
  onSave,
}) => {
  if (!task) return null

  const handleTypeChange = (newType: TaskType) => {
    if (newType === form.task_type) return
    // Check if config is incompatible — always confirm when switching
    const confirmMsg = 'Mengganti jenis tugas akan mengganti konfigurasi tugas yang tidak kompatibel. Konfigurasi lama tidak akan digunakan untuk jenis baru. Lanjutkan?'
    if (!window.confirm(confirmMsg)) return
    setForm({ ...form, task_type: newType })
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave()
  }

  const showLegacyBadge = isLegacyType(task.task_type)

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <motion.div
          initial={{ scale: 0.96, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.96, opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
            <div>
              <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Edit3 className="w-4 h-4 text-accent-magic" />
                <span>Edit Tugas</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5">
                #{task.step_order} — {task.task_type} {showLegacyBadge && <span className="text-[11px] bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">Legacy</span>}
              </p>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle flex items-center justify-center text-text-secondary hover:text-text-primary transition-colors cursor-pointer"
              title="Tutup"
              aria-label="Tutup"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-5 space-y-4">
            {/* Judul */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Judul Tugas <span className="text-status-error">*</span>
              </label>
              <input
                type="text"
                required
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic"
              />
            </div>

            {/* Deskripsi */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">
                Deskripsi
              </label>
              <textarea
                rows={3}
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm text-text-primary focus:outline-none focus:border-accent-magic resize-none"
              />
            </div>

            {/* Jenis Tugas */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Jenis Tugas *</label>
              <select
                value={form.task_type}
                onChange={(e) => handleTypeChange(e.target.value as TaskType)}
                className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
              >
                {TYPE_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
                {showLegacyBadge && <option value={task.task_type}>{task.task_type} (Legacy)</option>}
              </select>
              <p className="text-[11px] text-text-secondary">Mengubah jenis akan mengganti konfigurasi. Perubahan hanya untuk submission berikutnya.</p>
            </div>

            {/* Dynamic Config */}
            <div className="space-y-3 pt-2 border-t border-border-subtle">
              {form.task_type === 'VIDEO' && (
                <div className="p-3.5 rounded-2xl bg-surface-elevated border border-border-subtle space-y-2.5">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
                    <Play className="w-3.5 h-3.5 text-accent-magic" /> Video
                  </h4>
                  <input
                    type="url"
                    value={form.video_url}
                    onChange={(e) => setForm({ ...form, video_url: e.target.value })}
                    placeholder="https://www.youtube.com/watch?v=..."
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary"
                  />
                </div>
              )}
              {form.task_type === 'QUIZ' && (
                <div className="p-3.5 rounded-2xl bg-surface-elevated border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
                    <HelpCircle className="w-3.5 h-3.5 text-accent-magic" /> Kuis
                  </h4>
                  <input
                    type="url"
                    value={form.video_url}
                    onChange={(e) => setForm({ ...form, video_url: e.target.value })}
                    placeholder="YouTube URL opsional"
                    className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary"
                  />
                  {form.questions.map((q, qi) => (
                    <div key={q.id || qi} className="p-3 bg-surface rounded-xl border border-border-subtle space-y-2">
                      <input
                        type="text"
                        value={q.question}
                        onChange={(e) => {
                          const next = [...form.questions]
                          next[qi].question = e.target.value
                          setForm({ ...form, questions: next })
                        }}
                        placeholder={`Pertanyaan #${qi + 1}`}
                        className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs"
                      />
                      <div className="grid grid-cols-2 gap-2">
                        {q.options.map((opt, oi) => (
                          <input
                            key={oi}
                            type="text"
                            value={opt}
                            onChange={(e) => {
                              const next = [...form.questions]
                              next[qi].options[oi] = e.target.value
                              setForm({ ...form, questions: next })
                            }}
                            placeholder={`Opsi ${String.fromCharCode(65 + oi)}`}
                            className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs"
                          />
                        ))}
                      </div>
                      <select
                        value={q.correct_answer}
                        onChange={(e) => {
                          const next = [...form.questions]
                          next[qi].correct_answer = e.target.value
                          setForm({ ...form, questions: next })
                        }}
                        className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold"
                      >
                        <option value="">-- Kunci Jawaban --</option>
                        {q.options.map((opt, oi) => (
                          <option key={oi} value={opt || String.fromCharCode(65 + oi)}>
                            {String.fromCharCode(65 + oi)}: {opt || '(kosong)'}
                          </option>
                        ))}
                      </select>
                    </div>
                  ))}
                  <button type="button" onClick={() => setForm({ ...form, questions: [...form.questions, { id: String(Date.now()), question: '', options: ['', ''], correct_answer: '' }] })} className="text-xs font-bold text-accent-magic hover:underline">+ Tambah Pertanyaan</button>
                </div>
              )}
              {form.task_type === 'PHOTO_UPLOAD' && (
                <div className="p-3.5 rounded-2xl bg-surface-elevated border border-border-subtle space-y-2.5">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5"><Camera className="w-3.5 h-3.5 text-accent-cyan" /> Foto</h4>
                  <label className="text-[11px] text-text-secondary">Maksimal file (1-10)</label>
                  <input type="number" min={1} max={10} value={form.photo_min_count} onChange={(e) => setForm({ ...form, photo_min_count: Number(e.target.value) || 1 })} className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs font-mono" />
                </div>
              )}
              {form.task_type === 'DOCUMENT_UPLOAD' && (
                <div className="p-3.5 rounded-2xl bg-surface-elevated border border-border-subtle space-y-2.5">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5"><FileText className="w-3.5 h-3.5 text-accent-gold" /> Dokumen</h4>
                  <input type="text" value={form.doc_allowed_extensions} onChange={(e) => setForm({ ...form, doc_allowed_extensions: e.target.value })} placeholder="pdf,docx,xlsx" className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs font-mono" />
                  <input type="number" min={1} max={50} value={form.doc_max_size_mb} onChange={(e) => setForm({ ...form, doc_max_size_mb: Number(e.target.value) || 10 })} className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs font-mono" placeholder="Max MB" />
                </div>
              )}
              {form.task_type === 'TEXT_RESPONSE' && (
                <div className="p-3.5 rounded-2xl bg-surface-elevated border border-border-subtle space-y-2.5">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5"><PenLine className="w-3.5 h-3.5 text-accent-magic" /> Esai</h4>
                  <input type="text" value={form.text_prompt} onChange={(e) => setForm({ ...form, text_prompt: e.target.value })} placeholder="Prompt pertanyaan" className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs" />
                  <div className="grid grid-cols-2 gap-2">
                    <input type="number" value={form.text_min_chars} onChange={(e) => setForm({ ...form, text_min_chars: Number(e.target.value) || 10 })} className="p-2 rounded-xl bg-surface border border-border-subtle text-xs font-mono" placeholder="Min" />
                    <input type="number" value={form.text_max_chars} onChange={(e) => setForm({ ...form, text_max_chars: Number(e.target.value) || 500 })} className="p-2 rounded-xl bg-surface border border-border-subtle text-xs font-mono" placeholder="Max" />
                  </div>
                </div>
              )}
              {form.task_type === 'MINI_GAME' && (
                <div className="p-3.5 rounded-2xl bg-surface-elevated border border-border-subtle space-y-2.5">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5"><Gamepad2 className="w-3.5 h-3.5 text-status-success" /> Mini Game</h4>
                  <input type="text" value={form.game_type} onChange={(e) => setForm({ ...form, game_type: e.target.value })} placeholder="MEMORY" className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs" />
                  <div className="grid grid-cols-2 gap-2">
                    <input type="number" value={form.game_target_score} onChange={(e) => setForm({ ...form, game_target_score: Number(e.target.value) || 0 })} className="p-2 rounded-xl bg-surface border border-border-subtle text-xs font-mono" placeholder="Target score" />
                    <input type="number" value={form.game_max_moves} onChange={(e) => setForm({ ...form, game_max_moves: Number(e.target.value) || 0 })} className="p-2 rounded-xl bg-surface border border-border-subtle text-xs font-mono" placeholder="Max moves" />
                  </div>
                </div>
              )}
              {form.task_type === 'GENERAL' && <p className="text-xs text-text-secondary">Tidak ada konfigurasi khusus untuk tipe Umum.</p>}
            </div>

            {/* Bobot */}
            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Bobot Tugas</label>
              <input type="number" min={1} value={form.reward_coins} onChange={(e) => setForm({ ...form, reward_coins: Number(e.target.value) })} className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs sm:text-sm font-mono font-bold" />
              <p className="text-[11px] text-text-secondary leading-relaxed">Bobot bukan koin final. Koin final = Target Bulanan × Bobot ÷ Total Bobot.</p>
              <p className="text-[11px] text-text-secondary">Estimasi: Target 1000→~34, 2000→~67, 3200→~107 (untuk bobot 90, tergantung total bobot periode).</p>
            </div>

            <div className="space-y-1">
              <label className="text-xs font-bold text-text-secondary">Reward XP</label>
              <input type="number" min={0} value={form.reward_xp} onChange={(e) => setForm({ ...form, reward_xp: Number(e.target.value) })} className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs font-mono" />
            </div>

            <p className="text-[11px] text-amber-700 bg-amber-50 border border-amber-200 rounded-xl p-2.5">Perubahan konfigurasi hanya berlaku untuk submission berikutnya dan tidak mengubah reward/submission yang sudah ada.</p>

            {/* Footer */}
            <div className="sticky bottom-0 -mx-5 -mb-5 mt-4 px-5 py-3.5 border-t border-border-subtle bg-surface flex gap-3">
              <button type="button" onClick={onClose} className="flex-1 py-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-xs">Batal</button>
              <button type="submit" disabled={isSaving} className="flex-1 py-2.5 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 disabled:opacity-60 shadow-xs">
                {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
              </button>
            </div>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  )
}
