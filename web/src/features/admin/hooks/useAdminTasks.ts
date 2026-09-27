import { useState, useCallback, useEffect } from 'react'
import { adminTasksApi } from '../../../shared/lib/api'
import type { TaskView, TaskType } from '../../../shared/types'
import { buildTaskConfig } from './taskConfigBuilder'

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
  video_mode: 'youtube' | 'recording'
  video_max_duration: number
  video_camera_facing: 'user' | 'environment'
  video_instruction: string
  questions: Array<{ id: string; question: string; options: string[]; correct_answer: string }>
  photo_min_count: number
  photo_camera_only: boolean
  photo_camera_facing: 'user' | 'environment'
  photo_camera_instruction: string
  doc_allowed_extensions: string
  doc_max_size_mb: number
  text_prompt: string
  text_min_chars: number
  text_max_chars: number
  game_type: string
  game_target_score: number
  game_max_moves: number
  /** Raw JSON for decision/finance scenario. When non-empty, parsed as config.scenario (MINI_GAME). */
  game_scenario_json: string
}

const getInitialNewTask = (date: string): NewTaskFormState => ({
  title: '',
  description: '',
  task_type: 'VIDEO',
  video_answer_mode: 'none',
  video_mode: 'youtube',
  video_max_duration: 60,
  video_camera_facing: 'user',
  video_instruction: '',
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
  photo_camera_only: false,
  photo_camera_facing: 'environment',
  photo_camera_instruction: '',
  doc_allowed_extensions: 'pdf,docx,xlsx,txt',
  doc_max_size_mb: 10,
  text_prompt: '',
  text_min_chars: 10,
  text_max_chars: 500,
  game_type: 'MEMORY_MATCH',
  game_target_score: 500,
  game_max_moves: 20,
  game_scenario_json: '',
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
    video_mode: 'youtube' | 'recording'
    video_max_duration: number
    video_camera_facing: 'user' | 'environment'
    video_instruction: string
    questions: any[]
    photo_min_count: number
    photo_camera_only: boolean
    photo_camera_facing: 'user' | 'environment'
    photo_camera_instruction: string
    doc_allowed_extensions: string
    doc_max_size_mb: number
    text_prompt: string
    text_min_chars: number
    text_max_chars: number
    game_type: string
    game_target_score: number
    game_max_moves: number
    game_scenario_json: string
    config: Record<string, any>
  }>({
    title: '',
    description: '',
    task_type: 'VIDEO',
    reward_coins: 50,
    reward_xp: 100,
    video_url: '',
    video_answer_mode: 'none',
    video_mode: 'youtube',
    video_max_duration: 60,
    video_camera_facing: 'user',
    video_instruction: '',
    questions: [{ id: '1', question: '', options: ['', ''], correct_answer: '' }],
    photo_min_count: 1,
    photo_camera_only: false,
    photo_camera_facing: 'environment',
    photo_camera_instruction: '',
    doc_allowed_extensions: 'pdf,docx,xlsx,txt',
    doc_max_size_mb: 10,
    text_prompt: '',
    text_min_chars: 10,
    text_max_chars: 500,
    game_type: 'MEMORY_MATCH',
    game_target_score: 500,
    game_max_moves: 20,
    game_scenario_json: '',
    config: {},
  })
  const [isSavingTask, setIsSavingTask] = useState(false)
  const [isReordering, setIsReordering] = useState(false)

  const fetchTasks = useCallback(async (date: string) => {
    const effectiveDate = date || getLocalTodayString()
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
  }, [selectedDate, fetchTasks])

  const openCreateModal = () => {
    // Auto-position: tasks state always holds exactly the selectedDate list
    // (replaced on every fetchTasks(selectedDate): date change, retry, and
    // after each create/duplicate/delete/save), so max+1 here is per-date.
    // TaskView carries no active_date field, hence no extra filter possible.
    // Uniqueness is still enforced server-side (duplicate 400 backstop).
    const nextStep =
      tasks.reduce((m, t) => Math.max(m, t.step_order || 0), 0) + 1
    setNewTask({ ...getInitialNewTask(selectedDate), step_order: nextStep })
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

    // Single canonical task_type → config mapping (taskConfigBuilder).
    // Create parity: YouTube URL required, min_photos without max_files,
    // standalone QUIZ drops video_url, no stored-config preservation.
    const config = buildTaskConfig(newTask, {
      requireVideoUrl: true,
      includePhotoMaxFiles: false,
      includeQuizYoutubeUrl: false,
      messages: {
        videoUrlRequired: 'URL Video YouTube wajib diisi',
        videoUrlInvalid: 'URL video harus http(s)',
        quizVideoEmpty: (i) => `Pertanyaan ke-${i + 1} belum diisi`,
        quizVideoOptions: (i) => `Opsi ke-${i + 1} belum lengkap`,
        quizVideoKey: (i) => `Kunci #${i + 1} wajib dipilih`,
        quizEmpty: (i) => `Pertanyaan ke-${i + 1} belum diisi`,
        quizOptions: (i) => `Semua opsi pilihan jawaban pada pertanyaan ke-${i + 1} wajib diisi`,
        quizKey: (i) => `Kunci jawaban pada pertanyaan ke-${i + 1} wajib dipilih`,
        textPromptEmpty: 'Instruksi/pertanyaan esai tidak boleh kosong',
      },
    })
    if (!config) return

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
      const msg = err?.message || 'Terjadi kesalahan'
      // Concurrency backstop (message-only, no new backend): another admin may
      // have taken the suggested position. Refresh the suggestion in place so
      // the admin keeps title/content and can simply save again.
      if (msg.includes('step_order sudah digunakan')) {
        try {
          const fresh = await adminTasksApi.getTasks(selectedDate)
          const nextStep = (fresh || []).reduce(
            (m: number, t: TaskView) => Math.max(m, t.step_order || 0),
            0
          ) + 1
          setNewTask((prev) => ({ ...prev, step_order: nextStep }))
          await fetchTasks(selectedDate)
        } catch {
          // keep existing state; fall through to the message below
        }
        alert('Posisi tersebut baru saja terisi. Posisi otomatis sudah diperbarui — silakan simpan lagi.')
      } else {
        alert(`Gagal membuat tugas: ${msg}`)
      }
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

  const handleReorder = async (orderedIds: (number | string)[]) => {
    if (isReordering || isFetching) return
    if (orderedIds.length !== tasks.length) {
      setError('Gagal reorder: jumlah tugas tidak sesuai')
      return
    }
    const previousTasks = [...tasks]
    const idToTask = new Map<string, TaskView>(tasks.map((t) => [String(t.id), t]))
    const reordered: TaskView[] = []
    for (let i = 0; i < orderedIds.length; i++) {
      const t = idToTask.get(String(orderedIds[i]))
      if (!t) {
        setError('Gagal reorder: task tidak ditemukan')
        return
      }
      reordered.push({ ...t, step_order: i + 1 })
    }
    setTasks(reordered)
    setIsReordering(true)
    try {
      const res = await adminTasksApi.reorderTasks(orderedIds)
      if (Array.isArray(res) && res.length > 0) {
        // backend returns ordered tasks; ensure deterministic sort
        const sorted = [...res].sort((a: TaskView, b: TaskView) => a.step_order - b.step_order || a.id - b.id)
        setTasks(sorted)
      } else {
        await fetchTasks(selectedDate)
      }
    } catch (err: any) {
      setTasks(previousTasks)
      const msg = err?.message || 'Gagal menyimpan urutan'
      setError(msg)
      alert(`Gagal menyimpan urutan: ${msg}`)
    } finally {
      setIsReordering(false)
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
    const rec = (cfg.recording || {}) as Record<string, any>
    const isRecording = Boolean(rec.enabled)
    setEditTaskForm({
      title: task.title,
      description: task.description || '',
      task_type: task.task_type as TaskView['task_type'],
      reward_coins: task.reward_coins,
      reward_xp: task.reward_xp || 100,
      video_url: cfg.video_url || cfg.youtube_url || '',
      video_answer_mode: vMode,
      video_mode: isRecording ? 'recording' : 'youtube',
      video_max_duration: Number(rec.max_duration_seconds) || 60,
      video_camera_facing: rec.camera_facing === 'environment' ? 'environment' : 'user',
      video_instruction: (rec.instruction as string) || '',
      questions: Array.isArray(cfg.questions) && cfg.questions.length > 0 ? cfg.questions as any[] : [{ id: '1', question: '', options: ['', ''], correct_answer: '' }],
      photo_min_count: (cfg.min_photos as number) || (cfg.max_files as number) || 1,
      photo_camera_only: Boolean(cfg.camera_only),
      photo_camera_facing: cfg.camera_facing === 'user' ? 'user' : 'environment',
      photo_camera_instruction: (cfg.camera_instruction as string) || '',
      doc_allowed_extensions: Array.isArray(cfg.allowed_extensions) ? (cfg.allowed_extensions as unknown as string[]).join(',') : (cfg.accepted_extensions as unknown as string) || 'pdf,docx,xlsx,txt',
      doc_max_size_mb: (cfg.max_file_size_mb as number) || 10,
      text_prompt: (cfg.prompt as string) || (cfg.instruction as string) || '',
      text_min_chars: (cfg.minimum_characters as number) || 10,
      text_max_chars: (cfg.maximum_characters as number) || 500,
      game_type: (cfg.game as string) || 'MEMORY_MATCH',
      game_target_score: (cfg.target_score as number) || 500,
      game_max_moves: (cfg.max_moves as number) || 20,
      game_scenario_json: cfg.scenario ? JSON.stringify({ scenario: cfg.scenario }, null, 2) : '',
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
    // Single canonical task_type → config mapping (taskConfigBuilder).
    // Edit parity: YouTube URL optional (http(s) checked when set), stored
    // custom keys preserved on recording/GENERAL, max_files written,
    // standalone QUIZ carries youtube_url over.
    const config = buildTaskConfig(editTaskForm, {
      requireVideoUrl: false,
      preserveConfig: editTaskForm.config,
      includePhotoMaxFiles: true,
      includeQuizYoutubeUrl: true,
      messages: {
        videoUrlRequired: 'URL Video YouTube wajib diisi',
        videoUrlInvalid: 'URL video harus http(s)',
        quizVideoEmpty: (i) => `Pertanyaan #${i + 1} belum diisi`,
        quizVideoOptions: (i) => `Opsi #${i + 1} belum lengkap`,
        quizVideoKey: (i) => `Kunci #${i + 1} belum dipilih`,
        quizEmpty: (i) => `Pertanyaan #${i + 1} belum diisi`,
        quizOptions: (i) => `Opsi pertanyaan #${i + 1} belum lengkap`,
        quizKey: (i) => `Kunci jawaban #${i + 1} belum dipilih`,
        textPromptEmpty: 'Prompt esai tidak boleh kosong',
      },
    })
    if (!config) return

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
    isReordering,
    handleReorder,
  }
}
