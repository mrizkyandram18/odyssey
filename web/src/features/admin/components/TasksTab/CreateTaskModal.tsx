import React from 'react'
import { motion } from 'framer-motion'
import {
  X,
  Play,
  HelpCircle,
  Camera,
  FileText,
  PenLine,
  Gamepad2,
} from 'lucide-react'
import type { TaskType, MemberView } from '../../../../shared/types'
import type { NewTaskFormState } from '../../hooks/useAdminTasks'

interface CreateTaskModalProps {
  isOpen: boolean
  newTask: NewTaskFormState
  setNewTask: React.Dispatch<React.SetStateAction<NewTaskFormState>>
  members: MemberView[]
  isCreating: boolean
  onClose: () => void
  onSubmit: () => void
}

export const CreateTaskModal: React.FC<CreateTaskModalProps> = ({
  isOpen,
  newTask,
  setNewTask,
  members,
  isCreating,
  onClose,
  onSubmit,
}) => {
  if (!isOpen) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <motion.div
        initial={{ scale: 0.96, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.96, opacity: 0 }}
        className="w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
          <div>
            <h3 className="font-bold text-text-primary text-sm">Buat Tugas Baru</h3>
            <p className="text-xs text-text-secondary">
              Isi informasi dasar, pilih tipe, lalu atur konfigurasi.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-5 overflow-y-auto">
          {/* 1 — Informasi Dasar */}
          <div className="space-y-3">
            <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary">
              1 — Informasi Dasar
            </h4>
            <div className="space-y-1">
              <label htmlFor="task-title" className="text-xs font-bold text-text-secondary">
                Judul Tugas <span className="text-status-error">*</span>
              </label>
              <input
                id="task-title"
                type="text"
                required
                value={newTask.title}
                onChange={(e) => setNewTask({ ...newTask, title: e.target.value })}
                placeholder="Contoh: Merapikan Meja Belajar & Kamar"
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
              />
            </div>

            <div className="space-y-1">
              <label htmlFor="task-desc" className="text-xs font-bold text-text-secondary">
                Deskripsi / Petunjuk
              </label>
              <textarea
                id="task-desc"
                rows={2}
                value={newTask.description}
                onChange={(e) => setNewTask({ ...newTask, description: e.target.value })}
                placeholder="Jelaskan apa yang harus dilakukan oleh anggota..."
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic resize-none"
              />
            </div>
          </div>

          {/* 2 — Target & Tipe */}
          <div className="space-y-3">
            <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary">
              2 — Target & Tipe
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1">
                <label htmlFor="task-target-scope" className="text-xs font-bold text-text-secondary">
                  Target Penerima
                </label>
                <select
                  id="task-target-scope"
                  value={newTask.target_scope}
                  onChange={(e) =>
                    setNewTask({ ...newTask, target_scope: e.target.value as 'ALL' | 'USER' })
                  }
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                >
                  <option value="ALL">🌐 Semua Anggota Program</option>
                  <option value="USER">👤 User Tertentu (Personal Task)</option>
                </select>
              </div>

              {newTask.target_scope === 'USER' ? (
                <div className="space-y-1">
                  <label htmlFor="task-target-user" className="text-xs font-bold text-text-secondary">
                    Pilih User Target
                  </label>
                  <select
                    id="task-target-user"
                    required
                    value={newTask.target_user_uid}
                    onChange={(e) => setNewTask({ ...newTask, target_user_uid: e.target.value })}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                  >
                    <option value="">-- Pilih Anggota --</option>
                    {members.map((m) => (
                      <option key={m.uid} value={m.uid}>
                        {m.explorer_name} (@{m.username})
                      </option>
                    ))}
                  </select>
                </div>
              ) : (
                <div className="space-y-1">
                  <label htmlFor="task-type" className="text-xs font-bold text-text-secondary">
                    Tipe Tugas
                  </label>
                  <select
                    id="task-type"
                    value={newTask.task_type}
                    onChange={(e) =>
                      setNewTask({ ...newTask, task_type: e.target.value as TaskType })
                    }
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                  >
                    <option value="VIDEO">🎥 Video YouTube</option>
                    <option value="QUIZ">🧠 Kuis Pilihan Ganda</option>
                    <option value="PHOTO_UPLOAD">📸 Upload Foto Bukti</option>
                    <option value="DOCUMENT_UPLOAD">📄 Upload Dokumen (Excel/Word/PDF)</option>
                    <option value="TEXT_RESPONSE">✍️ Respon Teks / Esai</option>
                    <option value="MINI_GAME">🎮 Mini Game Memori</option>
                  </select>
                </div>
              )}
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {newTask.target_scope === 'USER' && (
                <div className="space-y-1">
                  <label htmlFor="task-type-user" className="text-xs font-bold text-text-secondary">
                    Tipe Tugas
                  </label>
                  <select
                    id="task-type-user"
                    value={newTask.task_type}
                    onChange={(e) =>
                      setNewTask({ ...newTask, task_type: e.target.value as TaskType })
                    }
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                  >
                    <option value="VIDEO">🎥 Video YouTube</option>
                    <option value="QUIZ">🧠 Kuis Pilihan Ganda</option>
                    <option value="PHOTO_UPLOAD">📸 Upload Foto Bukti</option>
                    <option value="DOCUMENT_UPLOAD">📄 Upload Dokumen (Excel/Word/PDF)</option>
                    <option value="TEXT_RESPONSE">✍️ Respon Teks / Esai</option>
                    <option value="MINI_GAME">🎮 Mini Game Memori</option>
                  </select>
                </div>
              )}

              <div className="space-y-1">
                <label htmlFor="task-step" className="text-xs font-bold text-text-secondary">
                  Urutan Step (1, 2..)
                </label>
                <input
                  id="task-step"
                  type="number"
                  min={1}
                  value={newTask.step_order}
                  onChange={(e) => setNewTask({ ...newTask, step_order: Number(e.target.value) })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                />
              </div>
            </div>
          </div>

          {/* 3 — Hadiah */}
          <div className="space-y-3">
            <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary">
              3 — Hadiah
            </h4>
            <div className="space-y-1">
              <label htmlFor="task-coins" className="text-xs font-bold text-text-secondary">
                Koin (🪙)
              </label>
              <input
                id="task-coins"
                type="number"
                min={1}
                value={newTask.reward_coins}
                onChange={(e) => setNewTask({ ...newTask, reward_coins: Number(e.target.value) })}
                className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
              />
            </div>
          </div>

          {/* 4 — Konfigurasi spesifik tipe */}
          {/* VIDEO */}
          {newTask.task_type === 'VIDEO' && (
            <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
              <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                <Play className="w-3.5 h-3.5 text-accent-magic" /> 4 — Video YouTube
              </h4>
              <div className="space-y-1">
                <label className="text-[11px] text-text-secondary">Link YouTube:</label>
                <input
                  type="url"
                  required
                  value={newTask.video_url}
                  onChange={(e) => setNewTask({ ...newTask, video_url: e.target.value })}
                  placeholder="https://www.youtube.com/watch?v=..."
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
            </div>
          )}

          {/* QUIZ */}
          {newTask.task_type === 'QUIZ' && (
            <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
              <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                <HelpCircle className="w-3.5 h-3.5 text-accent-magic" /> 4 — Soal Kuis
              </h4>
              <div className="space-y-1">
                <label className="text-[11px] text-text-secondary">Link YouTube (Opsional):</label>
                <input
                  type="url"
                  value={newTask.video_url}
                  onChange={(e) => setNewTask({ ...newTask, video_url: e.target.value })}
                  placeholder="https://www.youtube.com/watch?v=..."
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>

              {newTask.questions.map((q, qIndex) => (
                <div key={q.id || qIndex} className="space-y-2 p-3 bg-surface-elevated rounded-xl border border-border-subtle">
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Pertanyaan #{qIndex + 1}:</label>
                    <input
                      type="text"
                      required
                      value={q.question}
                      onChange={(e) => {
                        const nextQs = [...newTask.questions]
                        nextQs[qIndex].question = e.target.value
                        setNewTask({ ...newTask, questions: nextQs })
                      }}
                      placeholder="Contoh: Apa manfaat menabung sejak dini?"
                      className="w-full p-2 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    {q.options.map((opt, optIndex) => (
                      <input
                        key={optIndex}
                        type="text"
                        required
                        value={opt}
                        onChange={(e) => {
                          const nextQs = [...newTask.questions]
                          nextQs[qIndex].options[optIndex] = e.target.value
                          setNewTask({ ...newTask, questions: nextQs })
                        }}
                        placeholder={`Pilihan ${String.fromCharCode(65 + optIndex)}`}
                        className="p-2 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary"
                      />
                    ))}
                  </div>

                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Kunci Jawaban Benar:</label>
                    <select
                      value={q.correct_answer}
                      onChange={(e) => {
                        const nextQs = [...newTask.questions]
                        nextQs[qIndex].correct_answer = e.target.value
                        setNewTask({ ...newTask, questions: nextQs })
                      }}
                      className="w-full p-2 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary font-bold"
                    >
                      <option value="">-- Pilih Kunci Jawaban --</option>
                      {q.options.map((opt, optIndex) => (
                        <option key={optIndex} value={opt || String.fromCharCode(65 + optIndex)}>
                          Pilihan {String.fromCharCode(65 + optIndex)}: {opt || '(Belum diisi)'}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* PHOTO UPLOAD */}
          {newTask.task_type === 'PHOTO_UPLOAD' && (
            <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
              <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                <Camera className="w-3.5 h-3.5 text-accent-cyan" /> 4 — Foto Bukti
              </h4>
              <div className="space-y-1">
                <label className="text-[11px] text-text-secondary">Jumlah Minimal Foto:</label>
                <input
                  type="number"
                  min={1}
                  max={5}
                  value={newTask.photo_min_count}
                  onChange={(e) =>
                    setNewTask({ ...newTask, photo_min_count: Number(e.target.value) || 1 })
                  }
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
            </div>
          )}

          {/* DOCUMENT UPLOAD */}
          {newTask.task_type === 'DOCUMENT_UPLOAD' && (
            <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
              <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                <FileText className="w-3.5 h-3.5 text-accent-gold" /> 4 — Dokumen & Template
              </h4>
              <div className="space-y-1">
                <label className="text-[11px] text-text-secondary">Ekstensi Diizinkan (pisahkan koma):</label>
                <input
                  type="text"
                  value={newTask.doc_allowed_extensions}
                  onChange={(e) =>
                    setNewTask({ ...newTask, doc_allowed_extensions: e.target.value })
                  }
                  placeholder="pdf, docx, xlsx, txt"
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
              <div className="space-y-1">
                <label className="text-[11px] text-text-secondary">Maksimum Ukuran File (MB):</label>
                <input
                  type="number"
                  min={1}
                  max={50}
                  value={newTask.doc_max_size_mb}
                  onChange={(e) =>
                    setNewTask({ ...newTask, doc_max_size_mb: Number(e.target.value) || 10 })
                  }
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
            </div>
          )}

          {/* TEXT RESPONSE */}
          {newTask.task_type === 'TEXT_RESPONSE' && (
            <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
              <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                <PenLine className="w-3.5 h-3.5 text-accent-magic" /> 4 — Respon Teks
              </h4>
              <div className="space-y-1">
                <label className="text-[11px] text-text-secondary">Instruksi / Prompt Esai:</label>
                <input
                  type="text"
                  required
                  value={newTask.text_prompt}
                  onChange={(e) => setNewTask({ ...newTask, text_prompt: e.target.value })}
                  placeholder="Contoh: Apa pelajaran terpenting hari ini?"
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div className="space-y-1">
                  <label className="text-[10px] text-text-secondary">Min Karakter:</label>
                  <input
                    type="number"
                    min={1}
                    value={newTask.text_min_chars}
                    onChange={(e) =>
                      setNewTask({ ...newTask, text_min_chars: Number(e.target.value) || 10 })
                    }
                    className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-[10px] text-text-secondary">Maks Karakter:</label>
                  <input
                    type="number"
                    min={50}
                    value={newTask.text_max_chars}
                    onChange={(e) =>
                      setNewTask({ ...newTask, text_max_chars: Number(e.target.value) || 500 })
                    }
                    className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                  />
                </div>
              </div>
            </div>
          )}

          {/* MINI GAME */}
          {newTask.task_type === 'MINI_GAME' && (
            <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
              <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                <Gamepad2 className="w-3.5 h-3.5 text-status-success" /> 4 — Mini Game
              </h4>
              <div className="grid grid-cols-2 gap-2">
                <div className="space-y-1">
                  <label className="text-[11px] text-text-secondary">Target Skor Minimum:</label>
                  <input
                    type="number"
                    min={100}
                    value={newTask.game_target_score}
                    onChange={(e) =>
                      setNewTask({ ...newTask, game_target_score: Number(e.target.value) || 500 })
                    }
                    className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-[11px] text-text-secondary">Maks Langkah (Moves):</label>
                  <input
                    type="number"
                    min={5}
                    value={newTask.game_max_moves}
                    onChange={(e) =>
                      setNewTask({ ...newTask, game_max_moves: Number(e.target.value) || 20 })
                    }
                    className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                  />
                </div>
              </div>
            </div>
          )}

          <div className="sticky bottom-0 -mx-5 -mb-5 mt-2 px-5 py-4 border-t border-border-subtle bg-surface flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm hover:bg-surface transition-colors"
            >
              Batalkan
            </button>
            <button
              type="submit"
              disabled={isCreating}
              className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-sm hover:brightness-110 disabled:opacity-50 transition-all"
            >
              {isCreating ? 'Menyimpan...' : 'Simpan & Terbitkan'}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  )
}
