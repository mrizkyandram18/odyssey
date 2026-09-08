import React from 'react'
import {
  CheckCircle2,
  Coins,
  Users,
  Calendar,
  ArrowRight,
  Sparkles,
  Check,
  PlusCircle,
  Sliders,
} from 'lucide-react'
import { useAdminSubmissions } from '../../hooks/useAdminSubmissions'
import { useAdminClaims } from '../../hooks/useAdminClaims'
import { useAdminMembers } from '../../hooks/useAdminMembers'
import { useAdminConfig } from '../../hooks/useAdminConfig'

export type AdminTab = 'overview' | 'submissions' | 'claims' | 'tasks' | 'members' | 'rewards' | 'settings'

export interface AdminOverviewProps {
  onNavigateTab: (tab: AdminTab) => void
  submissionsController?: ReturnType<typeof useAdminSubmissions>
  claimsController?: ReturnType<typeof useAdminClaims>
}

export const AdminOverview: React.FC<AdminOverviewProps> = ({
  onNavigateTab,
  submissionsController,
  claimsController,
}) => {
  const defaultSubmissions = useAdminSubmissions({ enabled: !submissionsController })
  const defaultClaims = useAdminClaims({ enabled: !claimsController })
  const { submissions, pendingTotal, isFetching: isFetchingSubs } = submissionsController || defaultSubmissions
  const { claims, pendingTotal: pendingClaimsTotal, isFetching: isFetchingClaims } = claimsController || defaultClaims
  const { members, isFetching: isFetchingMembers } = useAdminMembers()
  const { config, isFetching: isFetchingConfig } = useAdminConfig()

  const isLoading = isFetchingSubs || isFetchingClaims || isFetchingMembers || isFetchingConfig

  // 1. Pending Submissions (Verifikasi)
  const pendingSubCount = pendingTotal ?? submissions.filter((s) => s.status === 'PENDING').length
  const latestPendingSub = submissions.find((s) => s.status === 'PENDING')

  // 2. Pending Claims (Pencairan)
  const pendingClaims = claims.filter((c) => c.status === 'PENDING')
  const pendingClaimsCount = pendingClaimsTotal ?? pendingClaims.length
  const totalCoinsRequested = pendingClaims.reduce(
    (sum, c) => sum + (c.coins_redeemed ?? (c as any).coins_requested ?? 0),
    0
  )
  const conversionRate = config?.conversion_rate || 100
  const estimatedRupiah = totalCoinsRequested * conversionRate

  // 3. Members Attention (Anggota)
  const inactiveDaysLimit = config?.auto_block_inactivity_days ?? 7
  const activeMembers = members.filter((m) => m.is_active)
  const inactiveMembers = members.filter(
    (m) => !m.is_active || (m.inactive_days != null && m.inactive_days >= inactiveDaysLimit)
  )
  const cappedMembers = members.filter((m) => {
    const cap = (m as any).effective_earning_cap ?? m.monthly_earning_cap ?? 0
    const earned = (m as any).earned_coins_this_month ?? m.earned_this_period ?? 0
    return cap > 0 && earned >= cap
  })
  const membersNeedingAttention = inactiveMembers.length + cappedMembers.length

  // 4. Redemption Period
  const isOpen = config?.is_open ?? false
  const payoutDay = config?.payout_day ?? 24

  // Check if any action is needed
  const hasActions = pendingSubCount > 0 || pendingClaimsCount > 0 || membersNeedingAttention > 0

  return (
    <div className="w-full flex flex-col gap-6 animate-in fade-in duration-200">
      {/* Overview Intro Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 sm:p-5 rounded-2xl bg-gradient-to-r from-accent-magic/10 via-surface to-accent-magic/5 border border-accent-magic/20">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-accent-magic animate-pulse" />
            <h2 className="text-sm sm:text-base font-bold text-text-primary tracking-tight">
              Ringkasan Operasional Harian
            </h2>
          </div>
          <p className="text-xs text-text-secondary">
            Pantau status verifikasi, persetujuan klaim koin, dan kondisi anggota keluarga.
          </p>
        </div>
        {isLoading && (
          <span className="text-[11px] text-text-secondary bg-surface px-2.5 py-1 rounded-full border border-border-subtle self-start sm:self-auto flex items-center gap-1.5 shrink-0">
            <span className="w-1.5 h-1.5 rounded-full bg-accent-magic animate-ping" />
            Sinkronisasi data...
          </span>
        )}
      </div>

      {/* ACTION QUEUE SECTION */}
      <section className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-accent-magic" />
            <h3 className="text-xs sm:text-sm font-bold text-text-primary tracking-tight">
              Perlu Tindakan Admin
            </h3>
            {hasActions && (
              <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-accent-magic/15 text-accent-magic border border-accent-magic/20">
                {(pendingSubCount > 0 ? 1 : 0) + (pendingClaimsCount > 0 ? 1 : 0) + (membersNeedingAttention > 0 ? 1 : 0)} Antrean
              </span>
            )}
          </div>
        </div>

        {!hasActions ? (
          <div className="p-5 rounded-2xl bg-surface border border-border-subtle flex items-center gap-3.5 shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-status-success/10 text-status-success flex items-center justify-center shrink-0">
              <Check className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-xs sm:text-sm font-bold text-text-primary">
                Semua Antrean Bersih & Terkendali
              </h4>
              <p className="text-xs text-text-secondary mt-0.5">
                Tidak ada bukti tugas yang menunggu verifikasi maupun klaim pencairan yang tertunda.
              </p>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {/* Action Item: Submissions Queue */}
            {pendingSubCount > 0 && (
              <div className="p-4 rounded-2xl bg-surface border border-accent-magic/30 shadow-xs flex flex-col justify-between gap-3 hover:border-accent-magic transition-colors">
                <div className="flex items-start gap-3">
                  <div className="w-9 h-9 rounded-xl bg-accent-magic/15 text-accent-magic flex items-center justify-center shrink-0 mt-0.5">
                    <CheckCircle2 className="w-4 h-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="text-xs sm:text-sm font-bold text-text-primary truncate">
                        {pendingSubCount} Bukti Tugas Menunggu
                      </h4>
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-extrabold bg-accent-magic text-white shrink-0">
                        Penting
                      </span>
                    </div>
                    <p className="text-xs text-text-secondary mt-1 line-clamp-2">
                      {latestPendingSub
                        ? `Terbaru: "${latestPendingSub.task_title}" oleh ${latestPendingSub.user_name || 'Anggota'}`
                        : 'Bukti pengerjaan tugas siap diperiksa untuk pemberian koin reward.'}
                    </p>
                  </div>
                </div>
                <div className="flex items-center justify-end pt-2 border-t border-border-subtle/60">
                  <button
                    type="button"
                    onClick={() => onNavigateTab('submissions')}
                    className="px-3.5 py-1.5 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 active:scale-95 transition-all flex items-center gap-1.5 cursor-pointer shadow-xs"
                  >
                    <span>Periksa Tugas</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            )}

            {/* Action Item: Claims Queue */}
            {pendingClaimsCount > 0 && (
              <div className="p-4 rounded-2xl bg-surface border border-accent-gold/30 shadow-xs flex flex-col justify-between gap-3 hover:border-accent-gold transition-colors">
                <div className="flex items-start gap-3">
                  <div className="w-9 h-9 rounded-xl bg-accent-gold/15 text-accent-gold flex items-center justify-center shrink-0 mt-0.5">
                    <Coins className="w-4 h-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="text-xs sm:text-sm font-bold text-text-primary truncate">
                        {pendingClaimsCount} Permintaan Pencairan
                      </h4>
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-extrabold bg-accent-gold text-white shrink-0">
                        Klaim
                      </span>
                    </div>
                    <p className="text-xs text-text-secondary mt-1">
                      Total <strong>{totalCoinsRequested.toLocaleString('id-ID')} Koin</strong> (estimasi Rp {estimatedRupiah.toLocaleString('id-ID')}) menunggu ditransfer.
                    </p>
                  </div>
                </div>
                <div className="flex items-center justify-end pt-2 border-t border-border-subtle/60">
                  <button
                    type="button"
                    onClick={() => onNavigateTab('claims')}
                    className="px-3.5 py-1.5 rounded-xl bg-accent-gold text-white text-xs font-bold hover:brightness-110 active:scale-95 transition-all flex items-center gap-1.5 cursor-pointer shadow-xs"
                  >
                    <span>Proses Pencairan</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            )}

            {/* Action Item: Members Attention */}
            {membersNeedingAttention > 0 && (
              <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-xs flex flex-col justify-between gap-3 hover:border-text-secondary/40 transition-colors">
                <div className="flex items-start gap-3">
                  <div className="w-9 h-9 rounded-xl bg-status-warning/15 text-status-warning flex items-center justify-center shrink-0 mt-0.5">
                    <Users className="w-4 h-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <h4 className="text-xs sm:text-sm font-bold text-text-primary">
                      Perhatian Anggota ({membersNeedingAttention})
                    </h4>
                    <p className="text-xs text-text-secondary mt-1">
                      {inactiveMembers.length > 0 && `${inactiveMembers.length} anggota inaktif / diblokir.`}
                      {inactiveMembers.length > 0 && cappedMembers.length > 0 && ' '}
                      {cappedMembers.length > 0 && `${cappedMembers.length} anggota telah mencapai batas koin bulanan.`}
                    </p>
                  </div>
                </div>
                <div className="flex items-center justify-end pt-2 border-t border-border-subtle/60">
                  <button
                    type="button"
                    onClick={() => onNavigateTab('members')}
                    className="px-3.5 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary hover:bg-surface text-xs font-bold transition-all flex items-center gap-1.5 cursor-pointer"
                  >
                    <span>Lihat Anggota</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </section>

      {/* 4 ESSENTIAL METRIC CARDS */}
      <section className="space-y-3">
        <h3 className="text-xs sm:text-sm font-bold text-text-primary tracking-tight">
          Metrik Operasional Utama
        </h3>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          {/* Card 1: Antrean Verifikasi Bukti Tugas */}
          <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-xs flex flex-col justify-between gap-3">
            <div>
              <div className="flex items-center justify-between text-text-secondary">
                <span className="text-[11px] font-bold">Antrean Verifikasi Bukti Tugas</span>
                <CheckCircle2 className="w-4 h-4 text-accent-magic" />
              </div>
              <div className="mt-2">
                <p className="text-xl sm:text-2xl font-black text-text-primary tracking-tight">
                  {pendingSubCount}
                  <span className="text-xs font-medium text-text-secondary ml-1.5">menunggu</span>
                </p>
                <p className="text-[11px] text-text-secondary mt-1 line-clamp-1">
                  {pendingSubCount === 0
                    ? 'Tidak Ada Antrean Verifikasi'
                    : latestPendingSub
                      ? `Terbaru: ${latestPendingSub.task_title}`
                      : `${pendingSubCount} tugas menunggu review`}
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => onNavigateTab('submissions')}
              className="text-xs font-bold text-accent-magic hover:underline flex items-center gap-1 cursor-pointer pt-2 border-t border-border-subtle/50"
            >
              <span>Buka Antrean</span>
              <ArrowRight className="w-3 h-3" />
            </button>
          </div>

          {/* Card 2: Permintaan Pencairan */}
          <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-xs flex flex-col justify-between gap-3">
            <div>
              <div className="flex items-center justify-between text-text-secondary">
                <span className="text-[11px] font-bold">Permintaan Pencairan</span>
                <Coins className="w-4 h-4 text-accent-gold" />
              </div>
              <div className="mt-2">
                <p className="text-xl sm:text-2xl font-black text-text-primary tracking-tight">
                  {pendingClaimsCount}
                  <span className="text-xs font-medium text-text-secondary ml-1.5">klaim</span>
                </p>
                <p className="text-[11px] text-text-secondary mt-1">
                  {pendingClaimsCount === 0
                    ? 'Belum ada klaim tertunda'
                    : `${totalCoinsRequested.toLocaleString('id-ID')} koin (Rp ${estimatedRupiah.toLocaleString('id-ID')})`}
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => onNavigateTab('claims')}
              className="text-xs font-bold text-accent-gold hover:underline flex items-center gap-1 cursor-pointer pt-2 border-t border-border-subtle/50"
            >
              <span>Lihat Klaim</span>
              <ArrowRight className="w-3 h-3" />
            </button>
          </div>

          {/* Card 3: Status Anggota */}
          <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-xs flex flex-col justify-between gap-3">
            <div>
              <div className="flex items-center justify-between text-text-secondary">
                <span className="text-[11px] font-bold">Status Anggota</span>
                <Users className="w-4 h-4 text-accent-nature" />
              </div>
              <div className="mt-2">
                <p className="text-xl sm:text-2xl font-black text-text-primary tracking-tight">
                  {activeMembers.length}
                  <span className="text-xs font-medium text-text-secondary ml-1.5">
                    / {members.length} aktif
                  </span>
                </p>
                <p className="text-[11px] text-text-secondary mt-1">
                  {membersNeedingAttention > 0
                    ? `${membersNeedingAttention} anggota perlu perhatian`
                    : 'Semua anggota aktif berkegiatan'}
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => onNavigateTab('members')}
              className="text-xs font-bold text-accent-nature hover:underline flex items-center gap-1 cursor-pointer pt-2 border-t border-border-subtle/50"
            >
              <span>Kelola Anggota</span>
              <ArrowRight className="w-3 h-3" />
            </button>
          </div>

          {/* Card 4: Jadwal Pencairan */}
          <div className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-xs flex flex-col justify-between gap-3">
            <div>
              <div className="flex items-center justify-between text-text-secondary">
                <span className="text-[11px] font-bold">Jadwal Pencairan</span>
                <Calendar className="w-4 h-4 text-accent-magic" />
              </div>
              <div className="mt-2">
                <div className="flex items-baseline gap-2">
                  <p className="text-lg sm:text-xl font-black text-text-primary tracking-tight">
                    Tgl {config ? `${config.redemption_start_day} s/d ${config.redemption_end_day}` : '21 s/d 26'}
                  </p>
                  <span
                    className={`text-[10px] font-bold px-1.5 py-0.5 rounded-full ${
                      isOpen
                        ? 'bg-status-success/15 text-status-success'
                        : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                    }`}
                  >
                    {isOpen ? 'Buka' : 'Tutup'}
                  </span>
                </div>
                <p className="text-[11px] text-text-secondary mt-1">
                  Hari transfer rutin: tanggal <strong>{payoutDay}</strong>
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => onNavigateTab('settings')}
              className="text-xs font-bold text-accent-magic hover:underline flex items-center gap-1 cursor-pointer pt-2 border-t border-border-subtle/50"
            >
              <span>Atur Periode</span>
              <ArrowRight className="w-3 h-3" />
            </button>
          </div>
        </div>
      </section>

      {/* QUICK SHORTCUTS ROW */}
      <section className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-accent-magic shrink-0" />
          <span className="text-xs font-bold text-text-primary">Akses Cepat Pengelolaan:</span>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => onNavigateTab('tasks')}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface transition-colors cursor-pointer flex items-center gap-1.5"
          >
            <PlusCircle className="w-3.5 h-3.5 text-accent-magic" />
            <span>Jadwal Tugas</span>
          </button>
          <button
            type="button"
            onClick={() => onNavigateTab('rewards')}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface transition-colors cursor-pointer flex items-center gap-1.5"
          >
            <Sparkles className="w-3.5 h-3.5 text-accent-gold" />
            <span>Katalog Hadiah</span>
          </button>
          <button
            type="button"
            onClick={() => onNavigateTab('settings')}
            className="px-3 py-1.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary hover:bg-surface transition-colors cursor-pointer flex items-center gap-1.5"
          >
            <Sliders className="w-3.5 h-3.5 text-text-secondary" />
            <span>Pengaturan Ekonomi</span>
          </button>
        </div>
      </section>
    </div>
  )
}
