// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, cleanup, waitFor } from '@testing-library/react'
import { useAdminTasks } from './useAdminTasks'
import { adminTasksApi } from '../../../shared/lib/api'

vi.mock('../../../shared/lib/api', () => ({
  adminTasksApi: {
    getTasks: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    duplicateTask: vi.fn(),
    deleteTask: vi.fn(),
    reorderTasks: vi.fn(),
  },
}))

const task = (id: number, step_order: number) =>
  ({
    id,
    title: `Tugas ${id}`,
    task_type: 'PHOTO_UPLOAD',
    step_order,
    reward_coins: 50,
    reward_xp: 100,
    config: {},
  } as any)

beforeEach(() => {
  vi.clearAllMocks()
  window.alert = vi.fn()
})

afterEach(() => {
  cleanup()
})

describe('Gate 2-B — auto-position suggestion', () => {
  it('suggests max(step_order)+1 for the listed date, 1 when empty', async () => {
    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([task(1, 1), task(2, 3)])
    const { result } = renderHook(() => useAdminTasks())

    await waitFor(() => expect(result.current.tasks).toHaveLength(2))

    act(() => {
      result.current.openCreateModal()
    })
    expect(result.current.newTask.step_order).toBe(4)
  })

  it('suggests position 1 when no tasks exist', async () => {
    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([])
    const { result } = renderHook(() => useAdminTasks())

    await waitFor(() => expect(result.current.isFetching).toBe(false))

    act(() => {
      result.current.openCreateModal()
    })
    expect(result.current.newTask.step_order).toBe(1)
  })

  it('refreshes the suggestion in place on duplicate-step race (message-only)', async () => {
    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([task(1, 1)])
    vi.mocked(adminTasksApi.createTask).mockRejectedValue(
      new Error('step_order sudah digunakan untuk tanggal tersebut')
    )
    // Fresh list after the race: another admin took position 2.
    vi.mocked(adminTasksApi.getTasks).mockResolvedValueOnce([task(1, 1)])
    const { result } = renderHook(() => useAdminTasks())

    await waitFor(() => expect(result.current.tasks).toHaveLength(1))

    act(() => {
      result.current.openCreateModal()
    })
    expect(result.current.newTask.step_order).toBe(2)

    act(() => {
      result.current.setNewTask((prev) => ({ ...prev, title: 'Tugas Baru', task_type: 'PHOTO_UPLOAD' }))
    })
    vi.mocked(adminTasksApi.getTasks).mockResolvedValue([task(1, 1), task(2, 2)])

    await act(async () => {
      await result.current.handleCreateTask()
    })

    // Payload kept the suggested position; race recovered without losing content.
    expect(adminTasksApi.createTask).toHaveBeenCalledWith(expect.objectContaining({ step_order: 2 }))
    expect(result.current.newTask.step_order).toBe(3)
    expect(result.current.newTask.title).toBe('Tugas Baru')
    expect(window.alert).toHaveBeenCalledWith(expect.stringMatching(/baru saja terisi/))
  })
})
