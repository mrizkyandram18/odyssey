import React from 'react'
import { Calendar, Plus, FileText, Edit3, Copy, Trash2, RefreshCw } from 'lucide-react'
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
  } = useAdminTasks()

  return (
    <div className="space-y-4">
      {error && (
        <div className="p-4 rounded-2xl bg-status-error/10 border border-status-error/20 text-status-error text-xs flex items-center justify-between">
          <span>{error}</span>
          <button
            type="button"
            onClick={() => fetchTasks(selectedDate)}
            className="font-bold underline ml-2"
          >
            Coba lagi
          </button>
        </div>
      )}

      {/* Date selector & Create button */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm">
        <div className="flex items-center gap-2">
          <Calendar className="w-4 h-4 text-accent-magic" />
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            aria-label="Pilih tanggal tugas"
            className="p-2 rounded-xl bg-surface border border-border-subtle text-xs font-bold text-text-primary focus:outline-none focus:border-accent-magic"
          />
          {isFetching && <RefreshCw className="w-3.5 h-3.5 animate-spin text-text-secondary" />}
        </div>
        <button
          onClick={openCreateModal}
          className="px-3.5 py-2 rounded-xl bg-accent-magic text-white font-heading font-bold text-xs flex items-center gap-1.5 shadow-sm shadow-accent-magic/30 hover:brightness-110 active:scale-95 transition-all self-start sm:self-auto"
        >
          <Plus className="w-4 h-4" />
          <span>Tambah Tugas</span>
        </button>
      </div>

      {tasks.length === 0 && !isFetching ? (
        <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
          <div className="w-10 h-10 mx-auto rounded-xl bg-accent-cyan/10 text-accent-cyan flex items-center justify-center">
            <FileText className="w-5 h-5" />
          </div>
          <p className="font-bold text-text-primary text-sm">Belum Ada Tugas pada Tanggal Ini</p>
          <p className="text-xs text-text-secondary max-w-xs mx-auto">
            Klik &quot;Tambah Tugas&quot; untuk membuat urutan tugas.
          </p>
        </div>
      ) : (
        tasks.map((task) => (
          <div
            key={task.id}
            className="p-3 rounded-2xl bg-surface border border-border-subtle shadow-sm flex items-center justify-between gap-3"
          >
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-accent-magic/15 text-accent-magic flex items-center justify-center font-bold text-sm shrink-0">
                #{task.step_order}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h4 className="font-heading font-bold text-text-primary text-sm">
                    {task.title}
                  </h4>
                  {task.target_scope === 'USER' && (
                    <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-accent-magic/20 text-accent-magic">
                      Personal:{' '}
                      {members.find((m) => m.uid === task.target_user_uid)?.explorer_name ||
                        task.target_user_uid}
                    </span>
                  )}
                </div>
                <p className="text-xs text-text-secondary line-clamp-1">
                  {task.description || 'Tanpa deskripsi'}
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <TaskTypeBadge type={task.task_type} />
                  <span className="text-[10px] font-bold text-accent-gold">
                    +{task.reward_coins} 🪙
                  </span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-1 shrink-0">
              <button
                type="button"
                onClick={() => openEditModal(task)}
                aria-label={`Edit tugas ${task.title}`}
                title="Edit Tugas"
                className="p-2 rounded-xl bg-surface border border-border-subtle text-text-secondary hover:text-accent-magic hover:bg-accent-magic/10 active:scale-95 transition-all"
              >
                <Edit3 className="w-4 h-4" />
              </button>

              <button
                type="button"
                onClick={() => handleDuplicateTask(task.id)}
                aria-label={`Duplikasi tugas ${task.title}`}
                title="Duplikasi Tugas"
                className="p-2 rounded-xl bg-surface border border-border-subtle text-accent-magic hover:bg-accent-magic/10 active:scale-95 transition-all"
              >
                <Copy className="w-4 h-4" />
              </button>

              <button
                type="button"
                onClick={() => handleDeleteTask(task.id)}
                aria-label={`Hapus tugas ${task.title}`}
                title="Hapus Tugas — tindakan tidak dapat dibatalkan"
                className="p-2 rounded-xl bg-surface border border-border-subtle text-status-error/70 hover:text-status-error hover:bg-status-error/10 hover:border-status-error/20 active:scale-95 transition-all"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </div>
        ))
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
