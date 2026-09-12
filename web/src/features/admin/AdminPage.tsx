import React, { useState } from 'react'
import { Navigate, useSearchParams } from 'react-router-dom'
import {
  ShieldCheck,
  LayoutDashboard,
  CheckCircle2,
  Coins,
  Calendar,
  Users,
  Sliders,
  Sparkles,
} from 'lucide-react'
import { useSession } from '../../shared/hooks/useSession'
import { useAdminConfig } from './hooks/useAdminConfig'
import { useAdminSubmissions } from './hooks/useAdminSubmissions'
import { useAdminClaims } from './hooks/useAdminClaims'
import { useAdminMembers } from './hooks/useAdminMembers'
import { AdminOverview } from './components/DashboardTab/AdminOverview'
import { SubmissionsQueue } from './components/SubmissionsTab/SubmissionsQueue'
import { ClaimsQueue } from './components/ClaimsTab/ClaimsQueue'
import { TaskScheduleList } from './components/TasksTab/TaskScheduleList'
import { MemberList } from './components/MembersTab/MemberList'
import { EconomySettingsForm } from './components/SettingsTab/EconomySettingsForm'
import { CosmeticsCatalogSection } from './components/SettingsTab/CosmeticsCatalogSection'

export type AdminTab = 'overview' | 'submissions' | 'claims' | 'tasks' | 'members' | 'rewards' | 'settings'

const ADMIN_TABS: AdminTab[] = ['overview', 'submissions', 'claims', 'tasks', 'members', 'rewards', 'settings']

function isAdminTab(value: string | null): value is AdminTab {
  return value !== null && (ADMIN_TABS as string[]).includes(value)
}

export interface AdminPageProps {
  initialTab?: AdminTab
}

interface TabButtonProps {
  tab: AdminTab
  testId: string
  label: string
  icon: React.ReactNode
  isActive: boolean
  onSelect: (tab: AdminTab) => void
  badge?: React.ReactNode
}

const tabButtonClass = (isActive: boolean) =>
  `shrink-0 flex-1 py-2 px-3 rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer whitespace-nowrap ${
    isActive
      ? 'bg-accent-magic text-white shadow-xs'
      : 'text-text-secondary hover:text-text-primary hover:bg-surface-elevated'
  }`

const TabButton: React.FC<TabButtonProps> = ({ tab, testId, label, icon, isActive, onSelect, badge }) => (
  <button
    type="button"
    data-testid={testId}
    onClick={() => onSelect(tab)}
    aria-current={isActive ? 'page' : undefined}
    className={tabButtonClass(isActive)}
  >
    {icon}
    <span>{label}</span>
    {badge}
  </button>
)

export const AdminPage: React.FC<AdminPageProps> = ({ initialTab = 'overview' }) => {
  const { profile, loading } = useSession()
  const [searchParams, setSearchParams] = useSearchParams()
  const tabFromUrl = searchParams.get('tab')
  const [activeTab, setActiveTab] = useState<AdminTab>(
    isAdminTab(tabFromUrl) ? tabFromUrl : initialTab
  )
  const { config } = useAdminConfig()
  const submissionsController = useAdminSubmissions()
  const claimsController = useAdminClaims()
  // Lifted here (instead of inside AdminOverview) so the existing member data
  // can also feed TaskScheduleList's assignee labels without an extra fetch.
  const membersController = useAdminMembers()

  const navigateTab = (tab: AdminTab) => {
    setActiveTab(tab)
    // Deep-link: /admin stays Ringkasan; any other tab is encoded so refresh
    // and shared links reopen the same tab. No nested routes.
    setSearchParams(tab === 'overview' ? {} : { tab }, { replace: true })
  }

  const pendingSubCount = submissionsController.pendingTotal
  const pendingClaimsCount = claimsController.pendingTotal ?? claimsController.claims.filter((c) => c.status === 'PENDING').length

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
            <span className="text-[11px] text-text-secondary font-medium">Jadwal Pencairan:</span>
            <span className="font-bold text-text-primary font-mono">Tgl {periodRange}</span>
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

      {/* Flat Primary Tab Navigation: 7 destinations, no grouping.
          Ringkasan = landing/triase; badge hanya di 2 antrean time-sensitive. */}
      <nav
        aria-label="Admin Navigation"
        className="flex items-center gap-1 p-1 bg-surface rounded-2xl border border-border-subtle shadow-xs overflow-x-auto no-scrollbar"
      >
        <TabButton
          tab="overview"
          testId="admin-tab-overview"
          label="Ringkasan"
          icon={<LayoutDashboard className="w-3.5 h-3.5" />}
          isActive={activeTab === 'overview'}
          onSelect={navigateTab}
        />
        <TabButton
          tab="submissions"
          testId="admin-tab-submissions"
          label="Verifikasi"
          icon={<CheckCircle2 className="w-3.5 h-3.5" />}
          isActive={activeTab === 'submissions'}
          onSelect={navigateTab}
          badge={
            pendingSubCount != null && pendingSubCount > 0 ? (
              <span
                className={`px-1.5 py-0.2 rounded-full text-[10px] font-extrabold ${
                  activeTab === 'submissions'
                    ? 'bg-white/30 text-white'
                    : 'bg-accent-magic/20 text-accent-magic'
                }`}
              >
                {pendingSubCount}
              </span>
            ) : undefined
          }
        />
        <TabButton
          tab="claims"
          testId="admin-tab-claims"
          label="Pencairan"
          icon={<Coins className="w-3.5 h-3.5" />}
          isActive={activeTab === 'claims'}
          onSelect={navigateTab}
          badge={
            pendingClaimsCount > 0 ? (
              <span
                className={`px-1.5 py-0.2 rounded-full text-[10px] font-extrabold ${
                  activeTab === 'claims'
                    ? 'bg-white/30 text-white'
                    : 'bg-accent-gold/25 text-amber-700 dark:text-amber-300'
                }`}
              >
                {pendingClaimsCount}
              </span>
            ) : undefined
          }
        />
        <TabButton
          tab="tasks"
          testId="admin-tab-tasks"
          label="Tugas"
          icon={<Calendar className="w-3.5 h-3.5" />}
          isActive={activeTab === 'tasks'}
          onSelect={navigateTab}
        />
        <TabButton
          tab="members"
          testId="admin-tab-members"
          label="Anggota"
          icon={<Users className="w-3.5 h-3.5" />}
          isActive={activeTab === 'members'}
          onSelect={navigateTab}
        />
        <TabButton
          tab="rewards"
          testId="admin-tab-rewards"
          label="Hadiah"
          icon={<Sparkles className="w-3.5 h-3.5" />}
          isActive={activeTab === 'rewards'}
          onSelect={navigateTab}
        />
        <TabButton
          tab="settings"
          testId="admin-tab-settings"
          label="Pengaturan"
          icon={<Sliders className="w-3.5 h-3.5" />}
          isActive={activeTab === 'settings'}
          onSelect={navigateTab}
        />
      </nav>

      {/* Tab Content */}
      <main className="w-full">
        {activeTab === 'overview' && (
          <AdminOverview
            onNavigateTab={navigateTab}
            submissionsController={submissionsController}
            claimsController={claimsController}
            membersController={membersController}
          />
        )}
        {activeTab === 'submissions' && <SubmissionsQueue controller={submissionsController} />}
        {activeTab === 'claims' && <ClaimsQueue controller={claimsController} />}
        {activeTab === 'tasks' && <TaskScheduleList members={membersController.members} />}
        {activeTab === 'members' && <MemberList controller={membersController} />}
        {activeTab === 'rewards' && <CosmeticsCatalogSection />}
        {activeTab === 'settings' && <EconomySettingsForm />}
      </main>
    </div>
  )
}
export default AdminPage
