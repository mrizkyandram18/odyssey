import React, { useState } from 'react'
import { Calendar, Plus, FileText, Edit3, Copy, Trash2, RefreshCw, GripVertical } from 'lucide-react'
import { useAdminTasks } from '../../hooks/useAdminTasks'
import { TaskTypeBadge } from '../shared/TaskTypeBadge'
import { CreateTaskModal } from './CreateTaskModal'
import { EditTaskModal } from './EditTaskModal'
import type { MemberView } from '../../../../shared/types'

interface TaskScheduleListProps {
  members?: MemberView[]
}

export const TaskScheduleList: React.FC<TaskScheduleListProps> = ({ members = [] }) => {
  const {
    tasks,
    selectedDate,
    setSelectedDate,
    isFetching,
    error,
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
  } = useAdminTasks()

  const [draggedId, setDraggedId] = useState<number | null>(null)
  const [dragOverId, setDragOverId] = useState<number | null>(null)

  const handleDragStart = (e: React.DragEvent, id: number) => {
    if (isReordering) {
      e.preventDefault()
      return
    }
    setDraggedId(id)
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(id))
  }

  const handleDragOver = (e: React.DragEvent, id: number) => {
    e.preventDefault()
    if (draggedId === null || draggedId === id) return
    setDragOverId(id)
  }

  const handleDragLeave = () => {
    setDragOverId(null)
  }

  const handleDrop = (e: React.DragEvent, targetId: number) => {
    e.preventDefault()
    if (draggedId === null || draggedId === targetId) {
      setDraggedId(null)
      setDragOverId(null)
      return
    }
    const fromIndex = tasks.findIndex((t) => t.id === draggedId)
    const toIndex = tasks.findIndex((t) => t.id === targetId)
    if (fromIndex < 0 || toIndex < 0) {
      setDraggedId(null)
      setDragOverId(null)
      return
    }
    const newOrder = [...tasks]
    const [moved] = newOrder.splice(fromIndex, 1)
    newOrder.splice(toIndex, 0, moved)
    const orderedIds = newOrder.map((t) => t.id)
    setDraggedId(null)
    setDragOverId(null)
    handleReorder(orderedIds)
  }

  const handleDragEnd = () => {
    setDraggedId(null)
    setDragOverId(null)
  }

  return (
    <div className="space-y-3.5">
      {error && (
        <div className="p-3.5 rounded-2xl bg-status-error/10 border border-status-error/20 text-status-error text-xs flex items-center justify-between">
          <span>{error}</span>
          <button
            type="button"
            onClick={() => fetchTasks(selectedDate)}
            className="font-bold underline ml-2 cursor-pointer"
          >
            Coba lagi
          </button>
        </div>
      )}

      {/* Date selector & Create button Toolbar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-2xl bg-surface border border-border-subtle shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle">
            <Calendar className="w-4 h-4 text-accent-magic shrink-0" />
            <input
              type="date"
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              aria-label="Pilih tanggal tugas"
              className="bg-transparent text-xs font-bold text-text-primary focus:outline-none cursor-pointer"
            />
          </div>
          {isFetching && <RefreshCw className="w-3.5 h-3.5 animate-spin text-text-secondary" />}
          <span className="text-xs text-text-secondary hidden md:inline">
            {tasks.length} tugas terjadwal
          </span>
        </div>

        <button
          type="button"
          onClick={openCreateModal}
          className="px-4 py-2 rounded-xl bg-accent-magic text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-xs hover:brightness-110 active:scale-95 transition-all self-stretch sm:self-auto cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          <span>Tambah Tugas</span>
        </button>
      </div>

      {/* Tasks List */}
      {tasks.length === 0 && !isFetching ? (
        <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
          <div className="w-10 h-10 mx-auto rounded-xl bg-accent-cyan/10 text-accent-cyan flex items-center justify-center">
            <FileText className="w-5 h-5" />
          </div>
          <p className="font-bold text-text-primary text-sm">Belum Ada Tugas pada Tanggal Ini</p>
          <p className="text-xs text-text-secondary max-w-xs mx-auto">
            Klik &quot;Tambah Tugas&quot; untuk membuat urutan tugas harian.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {tasks.length > 1 && !isReordering && (
            <p className="text-[11px] text-text-secondary flex items-center gap-1">
              <GripVertical className="w-3 h-3" /> Seret untuk mengatur urutan
            </p>
          )}
          {isReordering && (
            <p className="text-[11px] font-medium text-accent-magic flex items-center gap-1">
              <RefreshCw className="w-3 h-3 animate-spin" /> Menyimpan urutan...
            </p>
          )}
          {tasks.map((task) => (
            <div
              key={task.id}
              draggable={!isReordering}
              onDragStart={(e) => handleDragStart(e, task.id)}
              onDragOver={(e) => handleDragOver(e, task.id)}
              onDragLeave={handleDragLeave}
              onDrop={(e) => handleDrop(e, task.id)}
              onDragEnd={handleDragEnd}
              className={`p-3 sm:p-3.5 rounded-2xl bg-surface border shadow-xs flex items-center justify-between gap-3 transition-all hover:bg-surface-elevated/40 ${draggedId === task.id ? 'opacity-50 border-dashed border-accent-magic' : 'border-border-subtle'} ${dragOverId === task.id ? 'ring-2 ring-accent-magic ring-offset-1' : ''} ${isReordering ? 'opacity-60 cursor-wait' : 'cursor-grab active:cursor-grabbing'}`}
            >
              <div className="flex items-center gap-3 min-w-0">
                <span className="text-text-secondary cursor-grab active:cursor-grabbing p-1" aria-hidden="true">
                  <GripVertical className="w-4 h-4" />
                </span>
                <div className="w-8 h-8 rounded-xl bg-accent-magic/15 text-accent-magic flex items-center justify-center font-bold text-xs shrink-0 font-mono">
                  #{task.step_order}
                </div>

                <div className="min-w-0 space-y-0.5">
                  <div className="flex flex-wrap items-center gap-1.5">
                    <TaskTypeBadge type={task.task_type} />
                    <h4 className="font-heading font-bold text-text-primary text-xs sm:text-sm truncate">
                      {task.title}
                    </h4>
                    {task.target_scope === 'USER' && (
                      <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-accent-magic/15 text-accent-magic border border-accent-magic/20">
                        Personal:{' '}
                        {members.find((m) => m.uid === task.target_user_uid)?.explorer_name ||
                          task.target_user_uid}
                      </span>
                    )}
                  </div>
                  {task.description && (
                    <p className="text-[11px] text-text-secondary truncate max-w-md">
                      {task.description}
                    </p>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2 shrink-0">
                <span className="text-xs font-bold text-accent-gold font-mono px-2 py-0.5 rounded-md bg-accent-gold/10 border border-accent-gold/20">
                  +{task.reward_coins} 🪙
                </span>

                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => openEditModal(task)}
                    aria-label={`Edit tugas ${task.title}`}
                    title="Edit Tugas"
                    className="p-1.5 rounded-lg bg-surface border border-border-subtle text-text-secondary hover:text-accent-magic hover:bg-accent-magic/10 active:scale-95 transition-all cursor-pointer"
                  >
                    <Edit3 className="w-3.5 h-3.5" />
                  </button>

                  <button
                    type="button"
                    onClick={() => handleDuplicateTask(task.id)}
                    aria-label={`Duplikasi tugas ${task.title}`}
                    title="Duplikasi Tugas"
                    className="p-1.5 rounded-lg bg-surface border border-border-subtle text-accent-magic hover:bg-accent-magic/10 active:scale-95 transition-all cursor-pointer"
                  >
                    <Copy className="w-3.5 h-3.5" />
                  </button>

                  <button
                    type="button"
                    onClick={() => handleDeleteTask(task.id)}
                    aria-label={`Hapus tugas ${task.title}`}
                    title="Hapus Tugas"
                    className="p-1.5 rounded-lg bg-surface border border-border-subtle text-status-error/70 hover:text-status-error hover:bg-status-error/10 hover:border-status-error/20 active:scale-95 transition-all cursor-pointer"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modals */}
      <CreateTaskModal
        isOpen={isCreateModalOpen}
        newTask={newTask}
        setNewTask={setNewTask}
        members={members}
        isCreating={isCreatingTask}
        onClose={closeCreateModal}
        onSubmit={handleCreateTask}
        existingTasks={tasks}
      />

      <EditTaskModal
        task={editingTask}
        form={editTaskForm}
        setForm={setEditTaskForm}
        isSaving={isSavingTask}
        onClose={closeEditModal}
        onSave={handleSaveEditTask}
      />
    </div>
  )
}
