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

export function useAdminTasks() {
  const [selectedDate, setSelectedDate] = useState(() => new Date().toISOString().split('T')[0])
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
  const [editTaskForm, setEditTaskForm] = useState({
    title: '',
    description: '',
    reward_coins: 50,
    reward_xp: 100,
    video_url: '',
  })
  const [isSavingTask, setIsSavingTask] = useState(false)

  const fetchTasks = useCallback(async (date = selectedDate) => {
    setIsFetching(true)
    setError(null)
    try {
      const res = await adminTasksApi.getTasks(date)
      setTasks(res || [])
    } catch (err: any) {
      setError(err?.message || 'Gagal memuat daftar tugas')
    } finally {
      setIsFetching(false)
    }
  }, [selectedDate])

  useEffect(() => {
    fetchTasks(selectedDate)
  }, [selectedDate, fetchTasks])

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
      config = { video_url: newTask.video_url.trim() }
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
    setEditTaskForm({
      title: task.title,
      description: task.description || '',
      reward_coins: task.reward_coins,
      reward_xp: task.reward_xp || 100,
      video_url: task.config?.video_url || '',
    })
  }

  const closeEditModal = () => {
    setEditingTask(null)
  }

  const handleSaveEditTask = async () => {
    if (!editingTask) return
    setIsSavingTask(true)
    try {
      const patch: any = {
        title: editTaskForm.title,
        description: editTaskForm.description,
        reward_coins: editTaskForm.reward_coins,
        reward_xp: editTaskForm.reward_xp,
      }
      if (editingTask.task_type === 'VIDEO' && editTaskForm.video_url) {
        patch.config = {
          ...editingTask.config,
          video_url: editTaskForm.video_url.trim(),
        }
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
