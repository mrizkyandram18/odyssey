import React, { useState } from 'react'
import { Navigate } from 'react-router-dom'
import { ShieldCheck, CheckCircle2, Coins, Calendar, Users, Sliders } from 'lucide-react'
import { useSession } from '../../shared/hooks/useSession'
import { useAdminConfig } from './hooks/useAdminConfig'
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

  const periodRange = config
    ? `${config.redemption_start_day}–${config.redemption_end_day}`
    : '21–26'
  const isOpen = config?.is_open ?? false

  return (
    <div className="w-full flex flex-col gap-4">
      {/* Unified Admin Header */}
      <header className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 pb-1 border-b border-border-subtle/60">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center">
            <ShieldCheck className="w-4 h-4" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold text-text-primary tracking-tight">
                Panel Operasional Admin
              </h1>
              <span className="text-[10px] font-extrabold px-2 py-0.5 rounded-full bg-accent-magic/15 text-accent-magic border border-accent-magic/20 tracking-wider">
                ADMIN
              </span>
            </div>
            <p className="text-[11px] text-text-secondary">
              Pusat kendali operasional, verifikasi, jadwal tugas, dan anggota
            </p>
          </div>
        </div>

        {/* Live Period Status Tag */}
        <div className="flex items-center gap-2 self-start sm:self-auto">
          <div className="flex items-center gap-1.5 px-3 py-1 rounded-xl bg-surface border border-border-subtle text-xs">
            <span className="text-[11px] text-text-secondary font-medium">Periode Penukaran:</span>
            <span className="font-bold text-text-primary font-mono">{periodRange}</span>
            <span
              className={`inline-flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 rounded-full ${
                isOpen
                  ? 'bg-status-success/15 text-status-success'
                  : 'bg-surface-elevated text-text-secondary border border-border-subtle'
              }`}
            >
              <span className={`w-1.5 h-1.5 rounded-full ${isOpen ? 'bg-status-success' : 'bg-text-secondary'}`} />
              {isOpen ? 'Buka' : 'Tutup'}
            </span>
          </div>
        </div>
      </header>

      {/* Single Primary Tab Navigation */}
      <nav
        aria-label="Admin Navigation"
        className="grid grid-cols-2 sm:grid-cols-5 gap-1.5 p-1 bg-surface rounded-2xl border border-border-subtle shadow-xs"
      >
        <button
          type="button"
          data-testid="admin-tab-submissions"
          onClick={() => setActiveTab('submissions')}
          aria-current={activeTab === 'submissions' ? 'page' : undefined}
          className={`py-2 px-3 rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
            activeTab === 'submissions'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          <CheckCircle2 className="w-3.5 h-3.5" />
          <span>Verifikasi</span>
        </button>

        <button
          type="button"
          data-testid="admin-tab-claims"
          onClick={() => setActiveTab('claims')}
          aria-current={activeTab === 'claims' ? 'page' : undefined}
          className={`py-2 px-3 rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
            activeTab === 'claims'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          <Coins className="w-3.5 h-3.5" />
          <span>Pencairan</span>
        </button>

        <button
          type="button"
          data-testid="admin-tab-tasks"
          onClick={() => setActiveTab('tasks')}
          aria-current={activeTab === 'tasks' ? 'page' : undefined}
          className={`py-2 px-3 rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
            activeTab === 'tasks'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          <Calendar className="w-3.5 h-3.5" />
          <span>Tugas</span>
        </button>

        <button
          type="button"
          data-testid="admin-tab-members"
          onClick={() => setActiveTab('members')}
          aria-current={activeTab === 'members' ? 'page' : undefined}
          className={`py-2 px-3 rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
            activeTab === 'members'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          <Users className="w-3.5 h-3.5" />
          <span>Anggota</span>
        </button>

        <button
          type="button"
          data-testid="admin-tab-settings"
          onClick={() => setActiveTab('settings')}
          aria-current={activeTab === 'settings' ? 'page' : undefined}
          className={`col-span-2 sm:col-span-1 py-2 px-3 rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
            activeTab === 'settings'
              ? 'bg-accent-magic text-white shadow-sm'
              : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
          }`}
        >
          <Sliders className="w-3.5 h-3.5" />
          <span>Periode</span>
        </button>
      </nav>

      {/* Tab Content */}
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
