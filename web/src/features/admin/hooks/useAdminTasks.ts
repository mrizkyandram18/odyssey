import { useState, useCallback, useEffect } from 'react'
import { adminTasksApi } from '../../../shared/lib/api'
import type { TaskView, TaskType } from '../../../shared/types'

export interface NewTaskFormState {
  title: string
  description: string
  task_type: TaskType
  reward_coins: number
  reward_xp: number
  target_scope: 'ALL' | 'USER'
  target_user_uid: string
  active_date: string
  step_order: number
  // Config fields
  video_url: string
  video_answer_mode: 'none' | 'quiz' | 'essay'
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
}

const getInitialNewTask = (date: string): NewTaskFormState => ({
  title: '',
  description: '',
  task_type: 'VIDEO',
  video_answer_mode: 'none',
  reward_coins: 50,
  reward_xp: 100,
  target_scope: 'ALL',
  target_user_uid: '',
  active_date: date,
  step_order: 1,
  video_url: '',
  questions: [
    {
      id: '1',
      question: '',
      options: ['', ''],
      correct_answer: '',
    },
  ],
  photo_min_count: 1,
  doc_allowed_extensions: 'pdf,docx,xlsx,txt',
  doc_max_size_mb: 10,
  text_prompt: '',
  text_min_chars: 10,
  text_max_chars: 500,
  game_type: 'MEMORY_MATCH',
  game_target_score: 500,
  game_max_moves: 20,
})

const getLocalTodayString = (): string =>
  new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Jakarta' })

export function useAdminTasks() {
  const [selectedDate, setSelectedDate] = useState(() => getLocalTodayString())
  const [tasks, setTasks] = useState<TaskView[]>([])
  const [isFetching, setIsFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [processingId, setProcessingId] = useState<number | null>(null)

  // Create modal state
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [newTask, setNewTask] = useState<NewTaskFormState>(() => getInitialNewTask(selectedDate))
  const [isCreatingTask, setIsCreatingTask] = useState(false)

  // Edit modal state
  const [editingTask, setEditingTask] = useState<TaskView | null>(null)
  const [editTaskForm, setEditTaskForm] = useState<{
    title: string
    description: string
    task_type: TaskView['task_type']
    reward_coins: number
    reward_xp: number
    video_url: string
    video_answer_mode: 'none' | 'quiz' | 'essay'
    questions: any[]
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
  }>({
    title: '',
    description: '',
    task_type: 'VIDEO',
    reward_coins: 50,
    reward_xp: 100,
    video_url: '',
    video_answer_mode: 'none',
    questions: [{ id: '1', question: '', options: ['', ''], correct_answer: '' }],
    photo_min_count: 1,
    doc_allowed_extensions: 'pdf,docx,xlsx,txt',
    doc_max_size_mb: 10,
    text_prompt: '',
    text_min_chars: 10,
    text_max_chars: 500,
    game_type: 'MEMORY_MATCH',
    game_target_score: 500,
    game_max_moves: 20,
    config: {},
  })
  const [isSavingTask, setIsSavingTask] = useState(false)

  const fetchTasks = useCallback(async (date: string) => {
    const effectiveDate = date || selectedDate || getLocalTodayString()
    setIsFetching(true)
    setError(null)
    try {
      const res = await adminTasksApi.getTasks(effectiveDate)
      setTasks(res || [])
    } catch (err: any) {
      setError(err?.message || 'Gagal memuat daftar tugas')
    } finally {
      setIsFetching(false)
    }
  }, [])

  useEffect(() => {
    fetchTasks(selectedDate)
  }, [selectedDate])

  const openCreateModal = () => {
    setNewTask(getInitialNewTask(selectedDate))
    setIsCreateModalOpen(true)
  }

  const closeCreateModal = () => {
    setIsCreateModalOpen(false)
  }

  const handleCreateTask = async () => {
    if (!newTask.title.trim()) {
      alert('Judul tugas tidak boleh kosong')
      return
    }

    let config: Record<string, any> = {}
    if (newTask.task_type === 'VIDEO') {
      if (!newTask.video_url.trim()) {
        alert('URL Video YouTube wajib diisi')
        return
      }
      config = { video_url: newTask.video_url.trim(), youtube_url: newTask.video_url.trim() }
      if (newTask.video_answer_mode === 'quiz') {
        for (let i = 0; i < newTask.questions.length; i++) {
          const q = newTask.questions[i]
          if (!q.question.trim()) { alert(`Pertanyaan ke-${i + 1} belum diisi`); return }
          if (q.options.some((opt) => !opt.trim())) { alert(`Opsi ke-${i + 1} belum lengkap`); return }
          if (!q.correct_answer.trim()) { alert(`Kunci #${i + 1} wajib dipilih`); return }
        }
        config.questions = newTask.questions
      } else if (newTask.video_answer_mode === 'essay') {
        if (!newTask.text_prompt.trim()) { alert('Prompt essay tidak boleh kosong'); return }
        config.prompt = newTask.text_prompt.trim()
        config.minimum_characters = newTask.text_min_chars
        config.maximum_characters = newTask.text_max_chars
      }
    } else if (newTask.task_type === 'QUIZ') {
      for (let i = 0; i < newTask.questions.length; i++) {
        const q = newTask.questions[i]
        if (!q.question.trim()) {
          alert(`Pertanyaan ke-${i + 1} belum diisi`)
          return
        }
        if (q.options.some((opt) => !opt.trim())) {
          alert(`Semua opsi pilihan jawaban pada pertanyaan ke-${i + 1} wajib diisi`)
          return
        }
        if (!q.correct_answer.trim()) {
          alert(`Kunci jawaban pada pertanyaan ke-${i + 1} wajib dipilih`)
          return
        }
      }
      config = { questions: newTask.questions }
    } else if (newTask.task_type === 'PHOTO_UPLOAD') {
      config = { min_photos: newTask.photo_min_count }
    } else if (newTask.task_type === 'DOCUMENT_UPLOAD') {
      const exts = newTask.doc_allowed_extensions
        .split(',')
        .map((s) => s.trim().toLowerCase().replace(/^\./, ''))
        .filter(Boolean)
      config = {
        allowed_extensions: exts,
        max_file_size_mb: newTask.doc_max_size_mb,
      }
    } else if (newTask.task_type === 'TEXT_RESPONSE') {
      if (!newTask.text_prompt.trim()) {
        alert('Instruksi/pertanyaan esai tidak boleh kosong')
        return
      }
      config = {
        prompt: newTask.text_prompt.trim(),
        minimum_characters: newTask.text_min_chars,
        maximum_characters: newTask.text_max_chars,
      }
    } else if (newTask.task_type === 'MINI_GAME') {
      config = {
        game: newTask.game_type,
        target_score: newTask.game_target_score,
        max_moves: newTask.game_max_moves,
      }
    }

    setIsCreatingTask(true)
    try {
      await adminTasksApi.createTask({
        title: newTask.title.trim(),
        description: newTask.description.trim(),
        task_type: newTask.task_type,
        reward_coins: newTask.reward_coins,
        reward_xp: newTask.reward_xp || 100,
        target_scope: newTask.target_scope,
        target_user_uid: newTask.target_scope === 'USER' ? newTask.target_user_uid : undefined,
        active_date: newTask.active_date || selectedDate,
        step_order: Number(newTask.step_order) || 1,
        config,
      })
      closeCreateModal()
      await fetchTasks(selectedDate)
    } catch (err: any) {
      alert(`Gagal membuat tugas: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setIsCreatingTask(false)
    }
  }

  const handleDuplicateTask = async (id: number) => {
    setProcessingId(id)
    try {
      await adminTasksApi.duplicateTask(id)
      await fetchTasks(selectedDate)
    } catch (err: any) {
      alert(`Gagal menduplikasi tugas: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  const handleDeleteTask = async (id: number) => {
    if (!confirm('Apakah Anda yakin ingin menghapus tugas ini?')) return
    setProcessingId(id)
    try {
      await adminTasksApi.deleteTask(id)
      await fetchTasks(selectedDate)
    } catch (err: any) {
      alert(`Gagal menghapus tugas: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setProcessingId(null)
    }
  }

  const openEditModal = (task: TaskView) => {
    setEditingTask(task)
    const cfg = task.config || {}
    // Detect VIDEO answer mode from existing config
    let vMode: 'none' | 'quiz' | 'essay' = 'none'
    if (task.task_type === 'VIDEO') {
      if (Array.isArray(cfg.questions) && cfg.questions.length > 0) vMode = 'quiz'
      else if (cfg.prompt) vMode = 'essay'
    }
    setEditTaskForm({
      title: task.title,
      description: task.description || '',
      task_type: task.task_type as TaskView['task_type'],
      reward_coins: task.reward_coins,
      reward_xp: task.reward_xp || 100,
      video_url: cfg.video_url || cfg.youtube_url || '',
      video_answer_mode: vMode,
      questions: Array.isArray(cfg.questions) && cfg.questions.length > 0 ? cfg.questions as any[] : [{ id: '1', question: '', options: ['', ''], correct_answer: '' }],
      photo_min_count: (cfg.min_photos as number) || (cfg.max_files as number) || 1,
      doc_allowed_extensions: Array.isArray(cfg.allowed_extensions) ? (cfg.allowed_extensions as unknown as string[]).join(',') : (cfg.accepted_extensions as unknown as string) || 'pdf,docx,xlsx,txt',
      doc_max_size_mb: (cfg.max_file_size_mb as number) || 10,
      text_prompt: (cfg.prompt as string) || (cfg.instruction as string) || '',
      text_min_chars: (cfg.minimum_characters as number) || 10,
      text_max_chars: (cfg.maximum_characters as number) || 500,
      game_type: (cfg.game as string) || 'MEMORY_MATCH',
      game_target_score: (cfg.target_score as number) || 500,
      game_max_moves: (cfg.max_moves as number) || 20,
      config: cfg,
    })
  }

  const closeEditModal = () => {
    setEditingTask(null)
  }

  const handleSaveEditTask = async () => {
    if (!editingTask) return
    if (!editTaskForm.title.trim()) {
      alert('Judul tugas tidak boleh kosong')
      return
    }
    // Build config based on current task_type
    let config: Record<string, any> = {}
    const t = editTaskForm.task_type
    if (t === 'VIDEO') {
      if (editTaskForm.video_url.trim() && !editTaskForm.video_url.trim().startsWith('http')) {
        alert('URL video harus http(s)')
        return
      }
      config = { video_url: editTaskForm.video_url.trim() }
      if (editTaskForm.video_url.trim()) config.youtube_url = editTaskForm.video_url.trim()
      if (editTaskForm.video_answer_mode === 'quiz') {
        for (let i = 0; i < editTaskForm.questions.length; i++) {
          const q = editTaskForm.questions[i]
          if (!q.question.trim()) { alert(`Pertanyaan #${i + 1} belum diisi`); return }
          if (q.options.some((o: string) => !o.trim())) { alert(`Opsi #${i + 1} belum lengkap`); return }
          if (!q.correct_answer.trim()) { alert(`Kunci #${i + 1} belum dipilih`); return }
        }
        config.questions = editTaskForm.questions
      } else if (editTaskForm.video_answer_mode === 'essay') {
        if (!editTaskForm.text_prompt.trim()) { alert('Prompt essay tidak boleh kosong'); return }
        config.prompt = editTaskForm.text_prompt.trim()
        config.minimum_characters = editTaskForm.text_min_chars
        config.maximum_characters = editTaskForm.text_max_chars
      }
    } else if (t === 'QUIZ') {
      for (let i = 0; i < editTaskForm.questions.length; i++) {
        const q = editTaskForm.questions[i]
        if (!q.question.trim()) { alert(`Pertanyaan #${i + 1} belum diisi`); return }
        if (q.options.some((o: string) => !o.trim())) { alert(`Opsi pertanyaan #${i + 1} belum lengkap`); return }
        if (!q.correct_answer.trim()) { alert(`Kunci jawaban #${i + 1} belum dipilih`); return }
      }
      config = { questions: editTaskForm.questions }
      if (editTaskForm.video_url.trim()) config.youtube_url = editTaskForm.video_url.trim()
      } else if (t === 'PHOTO_UPLOAD') {
      config = { max_files: editTaskForm.photo_min_count, min_photos: editTaskForm.photo_min_count }
    } else if (t === 'DOCUMENT_UPLOAD') {
      const exts = editTaskForm.doc_allowed_extensions.split(',').map((s: string) => s.trim().toLowerCase().replace(/^\./, '')).filter(Boolean)
      config = { allowed_extensions: exts, max_file_size_mb: editTaskForm.doc_max_size_mb }
    } else if (t === 'TEXT_RESPONSE') {
      if (!editTaskForm.text_prompt.trim()) { alert('Prompt esai tidak boleh kosong'); return }
      config = { prompt: editTaskForm.text_prompt.trim(), minimum_characters: editTaskForm.text_min_chars, maximum_characters: editTaskForm.text_max_chars }
    } else if (t === 'MINI_GAME') {
      config = { game: editTaskForm.game_type, target_score: editTaskForm.game_target_score, max_moves: editTaskForm.game_max_moves }
    } else {
      config = editTaskForm.config || {}
    }

    setIsSavingTask(true)
    try {
      const patch: any = {
        title: editTaskForm.title.trim(),
        description: editTaskForm.description.trim(),
        task_type: editTaskForm.task_type,
        reward_coins: editTaskForm.reward_coins,
        reward_xp: editTaskForm.reward_xp || 100,
        config,
      }
      await adminTasksApi.updateTask(editingTask.id, patch)
      closeEditModal()
      await fetchTasks(selectedDate)
    } catch (err: any) {
      alert(`Gagal menyimpan perubahan tugas: ${err?.message || 'Terjadi kesalahan'}`)
    } finally {
      setIsSavingTask(false)
    }
  }

  return {
    tasks,
    selectedDate,
    setSelectedDate,
    isFetching,
    error,
    processingId,
    fetchTasks,
    isCreateModalOpen,
    newTask,
    setNewTask,
    isCreatingTask,
    openCreateModal,
    closeCreateModal,
    handleCreateTask,
    handleDuplicateTask,
    handleDeleteTask,
    editingTask,
    editTaskForm,
    setEditTaskForm,
    isSavingTask,
    openEditModal,
    closeEditModal,
    handleSaveEditTask,
  }
}
