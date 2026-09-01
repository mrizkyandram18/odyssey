import React, { useState } from 'react'
import { Navigate } from 'react-router-dom'
import { ShieldCheck, Users } from 'lucide-react'
import { useSession } from '../../shared/hooks/useSession'
import { useAdminConfig } from './hooks/useAdminConfig'
import { AdminMetricsBar } from './components/AdminMetricsBar'
import { SubmissionsQueue } from './components/SubmissionsTab/SubmissionsQueue'
import { ClaimsQueue } from './components/ClaimsTab/ClaimsQueue'
import { TaskScheduleList } from './components/TasksTab/TaskScheduleList'
import { MemberList } from './components/MembersTab/MemberList'
import { EconomySettingsForm } from './components/SettingsTab/EconomySettingsForm'

type AdminTab = 'submissions' | 'claims' | 'tasks' | 'members' | 'settings'

export const AdminPage: React.FC = () => {
  const { profile, loading } = useSession()
  const [activeTab, setActiveTab] = useState<AdminTab>('submissions')
  const { config } = useAdminConfig()

  if (loading) {
    return (
      <div className="w-full py-12 flex items-center justify-center">
        <div className="w-8 h-8 rounded-full border-2 border-accent-magic border-t-transparent animate-spin" />
      </div>
    )
  }

  const role = profile?.role
  const isAdmin = role === 'ADMIN' || role === 'GUIDE' || role === 'BUILDER'

  if (!isAdmin) {
    return <Navigate to="/home" replace />
  }

  return (
    <div className="w-full flex flex-col gap-4">
      {/* Header */}
      <header className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-lg font-bold text-text-primary flex items-center gap-2">
            <ShieldCheck className="w-5 h-5 text-accent-magic" />
            <span>Panel Operasional Admin</span>
            <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/20">
              ADMIN
            </span>
          </h1>
        </div>
      </header>

      {/* Metrics Bar */}
      <AdminMetricsBar
        activeTab={activeTab}
        onSelectTab={setActiveTab}
        config={config}
      />

      {/* Tab Navigation */}
      <div className="flex flex-wrap gap-1 p-1 bg-surface rounded-xl border border-border-subtle">
        <button
          type="button"
          data-testid="admin-tab-submissions"
          onClick={() => setActiveTab('submissions')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'submissions'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Verifikasi
        </button>

        <button
          type="button"
          data-testid="admin-tab-claims"
          onClick={() => setActiveTab('claims')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'claims'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Pencairan
        </button>

        <button
          type="button"
          data-testid="admin-tab-tasks"
          onClick={() => setActiveTab('tasks')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'tasks'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Tugas
        </button>

        <button
          type="button"
          data-testid="admin-tab-members"
          onClick={() => setActiveTab('members')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'members'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Users className="w-3.5 h-3.5" />
          <span>Anggota</span>
        </button>

        <button
          type="button"
          data-testid="admin-tab-settings"
          onClick={() => setActiveTab('settings')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'settings'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Periode
        </button>
      </div>

      {/* Tab Content (Lazy Render per Tab) */}
      <main className="w-full">
        {activeTab === 'submissions' && <SubmissionsQueue />}
        {activeTab === 'claims' && <ClaimsQueue />}
        {activeTab === 'tasks' && <TaskScheduleList />}
        {activeTab === 'members' && <MemberList />}
        {activeTab === 'settings' && <EconomySettingsForm />}
      </main>
    </div>
  )
}
export default AdminPage
