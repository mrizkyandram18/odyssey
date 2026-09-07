// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React, { useState } from 'react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { CreateTaskModal } from './CreateTaskModal'
import { EditTaskModal } from './EditTaskModal'
import type { NewTaskFormState } from '../../hooks/useAdminTasks'
import type { TaskView } from '../../../../shared/types'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const initialCreateState: NewTaskFormState = {
  title: 'Foto Bukti Diskusi',
  description: 'Ambil foto setelah berdiskusi.',
  task_type: 'PHOTO_UPLOAD',
  reward_coins: 50,
  reward_xp: 100,
  target_scope: 'ALL',
  target_user_uid: '',
  active_date: '2026-09-08',
  step_order: 1,
  video_url: '',
  video_answer_mode: 'none',
  questions: [],
  photo_min_count: 1,
  photo_camera_only: false,
  photo_camera_facing: 'environment',
  photo_camera_instruction: '',
  doc_allowed_extensions: 'pdf,docx',
  doc_max_size_mb: 10,
  text_prompt: '',
  text_min_chars: 10,
  text_max_chars: 500,
  game_type: 'MEMORY_MATCH',
  game_target_score: 500,
  game_max_moves: 20,
}

function CreateModalWrapper({ onSubmit }: { onSubmit: (state: NewTaskFormState) => void }) {
  const [taskState, setTaskState] = useState<NewTaskFormState>(initialCreateState)
  return (
    <CreateTaskModal
      isOpen={true}
      newTask={taskState}
      setNewTask={setTaskState}
      members={[]}
      isCreating={false}
      onClose={vi.fn()}
      onSubmit={() => onSubmit(taskState)}
    />
  )
}

function EditModalWrapper({
  initialTask,
  onSave,
}: {
  initialTask: TaskView
  onSave: (formState: any) => void
}) {
  const cfg = initialTask.config || {}
  const [form, setForm] = useState({
    title: initialTask.title,
    description: initialTask.description || '',
    task_type: initialTask.task_type,
    reward_coins: initialTask.reward_coins,
    reward_xp: initialTask.reward_xp || 100,
    video_url: '',
    video_answer_mode: 'none',
    questions: [],
    photo_min_count: (cfg.min_photos as number) || (cfg.max_files as number) || 1,
    photo_camera_only: Boolean(cfg.camera_only),
    photo_camera_facing: cfg.camera_facing === 'user' ? 'user' : 'environment',
    photo_camera_instruction: (cfg.camera_instruction as string) || '',
    doc_allowed_extensions: 'pdf,docx',
    doc_max_size_mb: 10,
    text_prompt: '',
    text_min_chars: 10,
    text_max_chars: 500,
    game_type: 'MEMORY_MATCH',
    game_target_score: 500,
    game_max_moves: 20,
    config: cfg,
  })

  return (
    <EditTaskModal
      task={initialTask}
      form={form}
      setForm={setForm}
      onClose={vi.fn()}
      onSave={() => onSave(form)}
      isSaving={false}
    />
  )
}

describe('Admin PHOTO_UPLOAD Configuration (Tests 1, 2, 3)', () => {
  it('Test 1 — Admin camera ON: enables camera_only and sets facing/instruction', () => {
    let submittedState: NewTaskFormState | null = null
    render(<CreateModalWrapper onSubmit={(state) => { submittedState = state }} />)

    // Initially "Upload Foto / Gallery" is checked
    const cameraRadio = screen.getByTestId('photo-mode-camera')
    expect(cameraRadio).not.toBeChecked()

    // Switch to "Wajib Ambil dari Kamera"
    fireEvent.click(cameraRadio)
    expect(cameraRadio).toBeChecked()

    // Facing select and instruction input become available
    const facingSelect = screen.getByTestId('photo-camera-facing-select')
    fireEvent.change(facingSelect, { target: { value: 'user' } })

    const instructionInput = screen.getByTestId('photo-camera-instruction-input')
    fireEvent.change(instructionInput, { target: { value: 'Ambil swafoto bersama teman.' } })

    // Submit form via Simpan & Terbitkan button
    fireEvent.submit(screen.getByRole('button', { name: /Simpan & Terbitkan/i }))

    expect(submittedState).not.toBeNull()
    expect(submittedState!.photo_camera_only).toBe(true)
    expect(submittedState!.photo_camera_facing).toBe('user')
    expect(submittedState!.photo_camera_instruction).toBe('Ambil swafoto bersama teman.')
  })

  it('Test 2 — Admin camera OFF: sets camera_only to false', () => {
    let submittedState: NewTaskFormState | null = null
    render(<CreateModalWrapper onSubmit={(state) => { submittedState = state }} />)

    const uploadRadio = screen.getByTestId('photo-mode-upload')
    expect(uploadRadio).toBeChecked()

    fireEvent.submit(screen.getByRole('button', { name: /Simpan & Terbitkan/i }))

    expect(submittedState).not.toBeNull()
    expect(submittedState!.photo_camera_only).toBe(false)
  })

  it('Test 3 — Edit existing PHOTO task: toggles OFF -> ON and ON -> OFF', () => {
    // 3a. OFF -> ON
    let savedFormState: any = null
    const existingTaskOff: TaskView = {
      id: 42,
      title: 'Upload Bukti Screenshot Canva',
      task_type: 'PHOTO_UPLOAD',
      config: { camera_only: false },
      reward_coins: 50,
      reward_xp: 100,
    } as any

    const { unmount } = render(
      <EditModalWrapper
        initialTask={existingTaskOff}
        onSave={(state) => { savedFormState = state }}
      />
    )

    const cameraRadio = screen.getByTestId('edit-photo-mode-camera')
    expect(cameraRadio).not.toBeChecked()

    fireEvent.click(cameraRadio)
    expect(cameraRadio).toBeChecked()

    fireEvent.submit(screen.getByRole('button', { name: /Simpan Perubahan/i }))
    expect(savedFormState.photo_camera_only).toBe(true)

    unmount()

    // 3b. ON -> OFF
    savedFormState = null
    const existingTaskOn: TaskView = {
      id: 43,
      title: 'Bukti Diskusi',
      task_type: 'PHOTO_UPLOAD',
      config: { camera_only: true },
      reward_coins: 50,
      reward_xp: 100,
    } as any

    render(
      <EditModalWrapper
        initialTask={existingTaskOn}
        onSave={(state) => { savedFormState = state }}
      />
    )

    const uploadRadio = screen.getByTestId('edit-photo-mode-upload')
    const cameraRadioOn = screen.getByTestId('edit-photo-mode-camera')
    expect(cameraRadioOn).toBeChecked()
    expect(uploadRadio).not.toBeChecked()

    fireEvent.click(uploadRadio)
    expect(uploadRadio).toBeChecked()

    fireEvent.submit(screen.getByRole('button', { name: /Simpan Perubahan/i }))
    expect(savedFormState.photo_camera_only).toBe(false)
  })
})
