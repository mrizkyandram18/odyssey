import React, { useState, useEffect, useCallback } from 'react'
import { Navigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  ShieldCheck,
  Calendar,
  Plus,
  CheckCircle2,
  XCircle,
  X,
  ExternalLink,
  Copy,
  Trash2,
  RefreshCw,
  Coins,
  AlertCircle,
  FileText,
  Gamepad2,
  PenLine,
  Camera,
  Play,
  HelpCircle,
  Settings,
  Sliders,
  Check,
  Building2,
  Wallet,
  Smartphone,
  Users,
  UserPlus,
  Edit3,
} from 'lucide-react'
import { useSession } from '../../shared/hooks/useSession'
import { adminTasksApi, adminMembersApi } from '../../shared/lib/api'
import type { TaskView, PendingSubmissionView, ClaimView, TaskType, RedemptionConfig, MemberView, Role } from '../../shared/types'

export const AdminPage: React.FC = () => {
  const { session, profile, loading: sessionLoading } = useSession()

  const [activeTab, setActiveTab] = useState<'submissions' | 'claims' | 'tasks' | 'members' | 'settings'>('submissions')
  const [selectedDate, setSelectedDate] = useState<string>(
    new Date().toISOString().split('T')[0]
  )

  // Data states
  const [tasks, setTasks] = useState<TaskView[]>([])
  const [submissions, setSubmissions] = useState<PendingSubmissionView[]>([])
  const [claims, setClaims] = useState<ClaimView[]>([])
  const [members, setMembers] = useState<MemberView[]>([])
  const [config, setConfig] = useState<RedemptionConfig | null>(null)
  const [isFetching, setIsFetching] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Config settings form states — fully config-driven economy
  const [startDayInput, setStartDayInput] = useState<number>(24)
  const [endDayInput, setEndDayInput] = useState<number>(26)
  const [payoutDayInput, setPayoutDayInput] = useState<number>(24)
  const [earningPeriodInput, setEarningPeriodInput] = useState<number>(30)
  const [conversionRateInput, setConversionRateInput] = useState<number>(100)
  const [targetRupiahInput, setTargetRupiahInput] = useState<number>(320000)
  const [maxPayoutInput, setMaxPayoutInput] = useState<number>(3200)
  const [timezoneInput, setTimezoneInput] = useState<string>('Asia/Jakarta')
  const [isSavingConfig, setIsSavingConfig] = useState(false)
  const [configSuccessMsg, setConfigSuccessMsg] = useState<string | null>(null)
  const [configErrorMsg, setConfigErrorMsg] = useState<string | null>(null)

  // Modals & UI states
  const [isCreateTaskModalOpen, setIsCreateTaskModalOpen] = useState(false)
  const [isCreateMemberModalOpen, setIsCreateMemberModalOpen] = useState(false)
  const [isEditMemberModalOpen, setIsEditMemberModalOpen] = useState(false)
  const [selectedMember, setSelectedMember] = useState<MemberView | null>(null)
  const [previewImage, setPreviewImage] = useState<string | null>(null)
  const [actionNotes, setActionNotes] = useState<Record<number, string>>({})
  const [processingId, setProcessingId] = useState<number | null>(null)

  // Member form states
  const [newMember, setNewMember] = useState({
    username: '',
    password: '',
    explorer_name: '',
    role: 'MEMBER' as Role,
  })

  const [editMemberForm, setEditMemberForm] = useState({
    explorer_name: '',
    role: 'MEMBER' as Role,
    is_active: true,
    password: '',
    reset_device: false,
  })

  // New task form state
  const [newTask, setNewTask] = useState({
    title: '',
    description: '',
    task_type: 'VIDEO' as TaskType,
    step_order: 1,
    active_date: selectedDate,
    reward_coins: 50,
    reward_xp: 100,
    target_scope: 'ALL' as 'ALL' | 'USER',
    target_user_uid: '',
    // Video
    youtube_url: '',
    min_watch_seconds: 60,
    // Quiz
    question_text: '',
    opt_a: '',
    opt_b: '',
    opt_c: '',
    opt_d: '',
    correct_ans: 'A',
    // Photo
    photo_instruction: 'Ambil foto bukti setelah menyelesaikan aktivitas.',
    max_photos: 1,
    // Doc
    attachment_url: '',
    attachment_name: 'Template Tugas',
    doc_instruction: 'Download template, kerjakan, lalu upload kembali file yang sudah selesai.',
    doc_extensions: '.xlsx,.docx,.pdf',
    // Text
    text_prompt: 'Ceritakan apa yang kamu pelajari hari ini:',
    min_chars: 20,
    max_chars: 1000,
    // Game
    game_type: 'MEMORY',
    game_difficulty: 'MEDIUM',
    target_score: 80,
  })

  const fetchData = useCallback(async () => {
    try {
      setIsFetching(true)
      setError(null)
      const [tList, subList, claimList, cfg, mList] = await Promise.all([
        adminTasksApi.getTasks(selectedDate),
        adminTasksApi.getPendingSubmissions(),
        adminTasksApi.getClaims('PENDING'),
        adminTasksApi.getConfig(),
        adminMembersApi.getMembers().catch(() => []),
      ])
      setTasks(tList || [])
      setSubmissions(subList || [])
      setClaims(claimList || [])
      setMembers(mList || [])
      if (cfg) {
        setConfig(cfg)
        setStartDayInput(cfg.redemption_start_day)
        setEndDayInput(cfg.redemption_end_day)
        if (cfg.payout_day) setPayoutDayInput(cfg.payout_day)
        if (cfg.earning_period_days) setEarningPeriodInput(cfg.earning_period_days)
        if (cfg.conversion_rate) setConversionRateInput(cfg.conversion_rate)
        if (cfg.payout_target_rupiah) setTargetRupiahInput(cfg.payout_target_rupiah)
        if (cfg.max_payout_coins) setMaxPayoutInput(cfg.max_payout_coins)
        if (cfg.timezone) setTimezoneInput(cfg.timezone)
      }
    } catch (err: any) {
      setError(err.message || 'Gagal memuat data admin dashboard')
    } finally {
      setIsFetching(false)
    }
  }, [selectedDate])

  useEffect(() => {
    if (session?.role === 'GUIDE' || profile?.role === 'GUIDE') {
      fetchData()
    }
  }, [fetchData, session, profile])

  if (sessionLoading) {
    return (
      <div className="flex h-64 w-full items-center justify-center">
        <div className="animate-pulse flex flex-col items-center gap-2">
          <ShieldCheck className="w-10 h-10 text-accent-magic" />
          <p className="text-xs text-text-secondary">Memuat Admin Operations Panel...</p>
        </div>
      </div>
    )
  }

  // Security guard
  if (session?.role !== 'GUIDE' && profile?.role !== 'GUIDE') {
    return <Navigate to="/" replace />
  }

  // --- Handlers ---
  const handleVerifySubmission = async (subId: number, status: 'APPROVED' | 'REJECTED') => {
    setProcessingId(subId)
    try {
      const notes = actionNotes[subId] || ''
      await adminTasksApi.verifySubmission(subId, status, notes)
      setSubmissions((prev) => prev.filter((s) => s.id !== subId))
      fetchData()
    } catch (err: any) {
      alert('Gagal memproses verifikasi: ' + err.message)
    } finally {
      setProcessingId(null)
    }
  }

  const handleProcessClaim = async (claimId: number, status: 'APPROVED' | 'REJECTED') => {
    setProcessingId(claimId)
    try {
      const notes = actionNotes[claimId] || ''
      await adminTasksApi.processClaim(claimId, status, notes)
      setClaims((prev) => prev.filter((c) => c.id !== claimId))
      fetchData()
    } catch (err: any) {
      alert('Gagal memproses klaim: ' + err.message)
    } finally {
      setProcessingId(null)
    }
  }

  const handleDeleteTask = async (taskId: number) => {
    if (!confirm('Apakah kamu yakin ingin menghapus tugas ini?')) return
    try {
      await adminTasksApi.deleteTask(taskId)
      setTasks((prev) => prev.filter((t) => t.id !== taskId))
    } catch (err: any) {
      alert('Gagal menghapus tugas: ' + err.message)
    }
  }

  const handleDuplicateTask = async (taskId: number) => {
    try {
      await adminTasksApi.duplicateTask(taskId)
      fetchData()
    } catch (err: any) {
      alert('Gagal menduplikasi tugas: ' + err.message)
    }
  }

  const handleCreateMember = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await adminMembersApi.createMember(newMember)
      setIsCreateMemberModalOpen(false)
      setNewMember({
        username: '',
        password: '',
        explorer_name: '',
        role: 'SEEKER',
      })
      fetchData()
    } catch (err: any) {
      alert('Gagal membuat anggota: ' + err.message)
    }
  }

  const handleUpdateMember = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedMember) return
    try {
      const patch: any = {
        explorer_name: editMemberForm.explorer_name,
        role: editMemberForm.role,
        is_active: editMemberForm.is_active,
      }
      if (editMemberForm.password.trim()) {
        patch.password = editMemberForm.password.trim()
      }
      if (editMemberForm.reset_device) {
        patch.reset_device = true
      }
      await adminMembersApi.updateMember(selectedMember.uid, patch)
      setIsEditMemberModalOpen(false)
      setSelectedMember(null)
      fetchData()
    } catch (err: any) {
      alert('Gagal memperbarui anggota: ' + err.message)
    }
  }

  const handleSaveConfig = async (e: React.FormEvent) => {
    e.preventDefault()
    setConfigSuccessMsg(null)
    setConfigErrorMsg(null)

    if (startDayInput < 1 || startDayInput > 31 || endDayInput < 1 || endDayInput > 31) {
      setConfigErrorMsg('Tanggal mulai dan berakhir harus antara 1 sampai 31')
      return
    }
    if (startDayInput > endDayInput) {
      setConfigErrorMsg('Tanggal mulai tidak boleh lebih besar dari tanggal berakhir')
      return
    }
    if (payoutDayInput < 1 || payoutDayInput > 31 || earningPeriodInput < 1 || earningPeriodInput > 365) {
      setConfigErrorMsg('Tanggal gajian atau durasi periode tidak valid')
      return
    }
    if (conversionRateInput <= 0 || maxPayoutInput <= 0 || targetRupiahInput < 0) {
      setConfigErrorMsg('Nilai konversi atau batas payout tidak valid')
      return
    }
    if (targetRupiahInput > 0 && conversionRateInput > 0) {
      const derived = Math.floor(targetRupiahInput / conversionRateInput)
      if (derived > maxPayoutInput) {
        setConfigErrorMsg(`Target ${derived} koin melebihi batas maksimum ${maxPayoutInput} koin`)
        return
      }
    }

    setIsSavingConfig(true)
    try {
      const updated = await adminTasksApi.updateConfig({
        start_day: Number(startDayInput),
        end_day: Number(endDayInput),
        payout_day: Number(payoutDayInput),
        earning_period_days: Number(earningPeriodInput),
        conversion_rate: Number(conversionRateInput),
        payout_target_rupiah: Number(targetRupiahInput),
        max_payout_coins: Number(maxPayoutInput),
        timezone: timezoneInput,
      })
      setConfig(updated)
      setStartDayInput(updated.redemption_start_day)
      setEndDayInput(updated.redemption_end_day)
      setPayoutDayInput(updated.payout_day)
      setEarningPeriodInput(updated.earning_period_days)
      setConversionRateInput(updated.conversion_rate)
      setTargetRupiahInput(updated.payout_target_rupiah)
      setMaxPayoutInput(updated.max_payout_coins)
      setTimezoneInput(updated.timezone)
      setConfigSuccessMsg(
        `Konfigurasi tersimpan: Target ${updated.payout_target_coins} koin = Rp ${updated.payout_target_rupiah.toLocaleString('id-ID')} • Maks ${updated.max_payout_coins} koin • Gajian tgl ${updated.payout_day} • Periode ${updated.earning_period_days} hari`
      )
    } catch (err: any) {
      setConfigErrorMsg(err.message || 'Gagal menyimpan pengaturan.')
    } finally {
      setIsSavingConfig(false)
    }
  }

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const taskConfig: Record<string, any> = {}

      switch (newTask.task_type) {
        case 'VIDEO':
        case 'YOUTUBE_VIDEO':
          taskConfig.youtube_url = newTask.youtube_url
          taskConfig.video_url = newTask.youtube_url
          taskConfig.minimum_duration_seconds = Number(newTask.min_watch_seconds)
          break

        case 'QUIZ':
        case 'VIDEO_QUIZ':
          if (newTask.youtube_url) {
            taskConfig.youtube_url = newTask.youtube_url
          }
          if (newTask.question_text) {
            taskConfig.questions = [
              {
                id: 1,
                question: newTask.question_text,
                options: [
                  `A. ${newTask.opt_a}`,
                  `B. ${newTask.opt_b}`,
                  `C. ${newTask.opt_c}`,
                  `D. ${newTask.opt_d}`,
                ].filter((o) => !o.endsWith('. ')),
                correct_answer: newTask.correct_ans,
              },
            ]
          }
          break

        case 'PHOTO_UPLOAD':
        case 'PHOTO_PROOF':
          taskConfig.instruction = newTask.photo_instruction
          taskConfig.max_files = Number(newTask.max_photos)
          break

        case 'DOCUMENT_UPLOAD':
          taskConfig.instruction = newTask.doc_instruction
          taskConfig.attachment_url = newTask.attachment_url
          taskConfig.attachment_name = newTask.attachment_name
          taskConfig.accepted_extensions = newTask.doc_extensions.split(',').map((s) => s.trim())
          break

        case 'TEXT_RESPONSE':
          taskConfig.prompt = newTask.text_prompt
          taskConfig.minimum_characters = Number(newTask.min_chars)
          taskConfig.maximum_characters = Number(newTask.max_chars)
          break

        case 'MINI_GAME':
          taskConfig.game = newTask.game_type
          taskConfig.difficulty = newTask.game_difficulty
          taskConfig.target_score = Number(newTask.target_score)
          break
      }

      await adminTasksApi.createTask({
        title: newTask.title,
        description: newTask.description,
        task_type: newTask.task_type,
        step_order: Number(newTask.step_order),
        active_date: newTask.active_date || selectedDate,
        reward_coins: Number(newTask.reward_coins),
        reward_xp: Number(newTask.reward_xp),
        target_scope: newTask.target_scope,
        target_user_uid: newTask.target_scope === 'USER' ? newTask.target_user_uid : undefined,
        config: taskConfig,
        is_active: true,
      })

      setIsCreateTaskModalOpen(false)
      setNewTask({
        title: '',
        description: '',
        task_type: 'VIDEO',
        step_order: tasks.length + 1,
        active_date: selectedDate,
        reward_coins: 50,
        reward_xp: 100,
        target_scope: 'ALL',
        target_user_uid: '',
        youtube_url: '',
        min_watch_seconds: 60,
        question_text: '',
        opt_a: '',
        opt_b: '',
        opt_c: '',
        opt_d: '',
        correct_ans: 'A',
        photo_instruction: 'Ambil foto bukti setelah menyelesaikan aktivitas.',
        max_photos: 1,
        attachment_url: '',
        attachment_name: 'Template Tugas',
        doc_instruction: 'Download template, kerjakan, lalu upload kembali file yang sudah selesai.',
        doc_extensions: '.xlsx,.docx,.pdf',
        text_prompt: 'Ceritakan apa yang kamu pelajari hari ini:',
        min_chars: 20,
        max_chars: 1000,
        game_type: 'MEMORY',
        game_difficulty: 'MEDIUM',
      target_score: 80,
    })
    fetchData()
  } catch (err: any) {
    alert('Gagal membuat tugas: ' + err.message)
  }
}

  const getTaskTypeBadge = (tType: string) => {
    switch (tType) {
      case 'VIDEO':
      case 'YOUTUBE_VIDEO':
        return (
          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-accent-magic/10 text-accent-magic">
            <Play className="w-3 h-3" />
            <span>Video</span>
          </span>
        )
      case 'QUIZ':
      case 'VIDEO_QUIZ':
        return (
          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-accent-magic/10 text-accent-magic">
            <HelpCircle className="w-3 h-3" />
            <span>Kuis</span>
          </span>
        )
      case 'PHOTO_UPLOAD':
      case 'PHOTO_PROOF':
        return (
          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-accent-cyan/15 text-accent-cyan">
            <Camera className="w-3 h-3" />
            <span>Foto Bukti</span>
          </span>
        )
      case 'DOCUMENT_UPLOAD':
        return (
          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-accent-gold/15 text-accent-gold">
            <FileText className="w-3 h-3" />
            <span>Dokumen</span>
          </span>
        )
      case 'TEXT_RESPONSE':
        return (
          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-accent-magic/15 text-accent-magic">
            <PenLine className="w-3 h-3" />
            <span>Respon Teks</span>
          </span>
        )
      case 'MINI_GAME':
        return (
          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-status-success/15 text-status-success">
            <Gamepad2 className="w-3 h-3" />
            <span>Mini Game</span>
          </span>
        )
      default:
        return (
          <span className="text-[10px] font-bold px-2 py-0.5 rounded-md bg-surface text-text-secondary">
            {tType}
          </span>
        )
    }
  }

  return (
    <div className="w-full flex flex-col gap-4">
      <header className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-lg font-bold text-text-primary flex items-center gap-2">
            <ShieldCheck className="w-5 h-5 text-accent-magic" />
            <span>Panel Operasional Admin</span>
            <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-accent-magic/10 text-accent-magic border border-accent-magic/20">GUIDE</span>
          </h1>
        </div>
        <button
          type="button"
          onClick={fetchData}
          disabled={isFetching}
          aria-label="Segarkan data admin"
          aria-busy={isFetching}
          className="self-start sm:self-auto px-3 py-2 rounded-xl bg-surface border border-border-subtle text-text-secondary hover:text-text-primary transition-colors flex items-center gap-2 text-xs font-bold shrink-0"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isFetching ? 'animate-spin' : ''}`} />
          <span>Segarkan</span>
        </button>
      </header>

      {/* Needs Attention — operator's 5-second clarity */}
      {(submissions.length > 0 || claims.length > 0) && (
        <div className="p-4 rounded-2xl bg-accent-magic/5 border border-accent-magic/20">
          <h2 className="text-xs font-bold tracking-wide uppercase text-text-secondary">Butuh Perhatian</h2>
          <div className="mt-2 flex flex-wrap gap-2">
            {submissions.length > 0 && (
              <button onClick={() => setActiveTab('submissions')} className="inline-flex items-center gap-2 px-3 py-2 rounded-xl bg-surface border border-accent-magic/30 text-sm font-bold text-text-primary hover:bg-surface-elevated transition-colors">
                <span className="w-6 h-6 rounded-full bg-accent-magic text-white flex items-center justify-center text-xs font-bold">{submissions.length}</span>
                <span>Verifikasi Tugas</span>
              </button>
            )}
            {claims.length > 0 && (
              <button onClick={() => setActiveTab('claims')} className="inline-flex items-center gap-2 px-3 py-2 rounded-xl bg-surface border border-status-success/30 text-sm font-bold text-text-primary hover:bg-surface-elevated transition-colors">
                <span className="w-6 h-6 rounded-full bg-status-success text-white flex items-center justify-center text-xs font-bold">{claims.length}</span>
                <span>Pencairan Koin</span>
              </button>
            )}
            {submissions.length === 0 && claims.length === 0 && (
              <span className="text-xs text-status-success font-bold flex items-center gap-1"><CheckCircle2 className="w-4 h-4" /> Semua antrean bersih</span>
            )}
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
        <button
          onClick={() => setActiveTab('submissions')}
          aria-current={activeTab === 'submissions' ? 'true' : undefined}
          className={`text-left p-3 rounded-2xl border transition-colors ${
            activeTab === 'submissions'
              ? 'bg-accent-magic/10 border-accent-magic/30'
              : 'bg-surface border-border-subtle hover:bg-surface-elevated'
          }`}
        >
          <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
            <CheckCircle2 className="w-3.5 h-3.5 text-accent-magic" />
            Verifikasi
          </span>
          <p className="mt-1 flex items-baseline gap-1">
            <span className="text-xl font-bold text-text-primary">{submissions.length}</span>
            <span className="text-[11px] text-text-secondary">antrean</span>
          </p>
        </button>

        <button
          onClick={() => setActiveTab('claims')}
          aria-current={activeTab === 'claims' ? 'true' : undefined}
          className={`text-left p-3 rounded-2xl border transition-colors ${
            activeTab === 'claims'
              ? 'bg-status-success/10 border-status-success/30'
              : 'bg-surface border-border-subtle hover:bg-surface-elevated'
          }`}
        >
          <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
            <Coins className="w-3.5 h-3.5 text-status-success" />
            Pencairan
          </span>
          <p className="mt-1 flex items-baseline gap-1">
            <span className="text-xl font-bold text-text-primary">{claims.length}</span>
            <span className="text-[11px] text-text-secondary">klaim</span>
          </p>
        </button>

        <button
          onClick={() => setActiveTab('tasks')}
          aria-current={activeTab === 'tasks' ? 'true' : undefined}
          className={`text-left p-3 rounded-2xl border transition-colors ${
            activeTab === 'tasks'
              ? 'bg-accent-cyan/10 border-accent-cyan/30'
              : 'bg-surface border-border-subtle hover:bg-surface-elevated'
          }`}
        >
          <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
            <Calendar className="w-3.5 h-3.5 text-accent-cyan" />
            Tugas
          </span>
          <p className="mt-1 flex items-baseline gap-1">
            <span className="text-xl font-bold text-text-primary">{tasks.length}</span>
            <span className="text-[11px] text-text-secondary">item</span>
          </p>
        </button>

        <button
          onClick={() => setActiveTab('members')}
          aria-current={activeTab === 'members' ? 'true' : undefined}
          className={`text-left p-3 rounded-2xl border transition-colors ${
            activeTab === 'members'
              ? 'bg-accent-magic/10 border-accent-magic/30'
              : 'bg-surface border-border-subtle hover:bg-surface-elevated'
          }`}
        >
          <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
            <Users className="w-3.5 h-3.5 text-accent-magic" />
            Anggota
          </span>
          <p className="mt-1 flex items-baseline gap-1">
            <span className="text-xl font-bold text-text-primary">{members.length}</span>
            <span className="text-[11px] text-text-secondary">orang</span>
          </p>
        </button>

        <button
          onClick={() => setActiveTab('settings')}
          aria-current={activeTab === 'settings' ? 'true' : undefined}
          className={`text-left p-3 rounded-2xl border transition-colors col-span-2 lg:col-span-1 ${
            activeTab === 'settings'
              ? 'bg-accent-gold/10 border-accent-gold/30'
              : 'bg-surface border-border-subtle hover:bg-surface-elevated'
          }`}
        >
          <span className="text-[11px] font-bold text-text-secondary flex items-center gap-1">
            <Sliders className="w-3.5 h-3.5 text-accent-gold" />
            Periode
          </span>
          <p className="mt-1 flex items-baseline gap-1">
            <span className="text-lg font-bold text-text-primary">{config ? `${config.redemption_start_day}–${config.redemption_end_day}` : '24–26'}</span>
            <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded-full ${config?.is_open ? 'bg-status-success/15 text-status-success' : 'bg-surface-elevated text-text-secondary border border-border-subtle'}`}>{config?.is_open ? 'Buka' : 'Tutup'}</span>
          </p>
        </button>
      </div>

      <div className="flex flex-wrap gap-1 p-1 bg-surface rounded-xl border border-border-subtle">
        <button
          data-testid="admin-tab-submissions"
          onClick={() => setActiveTab('submissions')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'submissions' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Verifikasi
          {submissions.length > 0 && <span className="px-1.5 py-0.5 rounded-full bg-white/20 text-white font-mono text-[10px]">{submissions.length}</span>}
        </button>

        <button
          data-testid="admin-tab-claims"
          onClick={() => setActiveTab('claims')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'claims' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Pencairan
          {claims.length > 0 && <span className="px-1.5 py-0.5 rounded-full bg-white/20 text-white font-mono text-[10px]">{claims.length}</span>}
        </button>

        <button
          data-testid="admin-tab-tasks"
          onClick={() => setActiveTab('tasks')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'tasks' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          Tugas
        </button>

        <button
          data-testid="admin-tab-members"
          onClick={() => setActiveTab('members')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'members' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Users className="w-3.5 h-3.5" />
          Anggota
          {members.length > 0 && <span className="px-1.5 py-0.5 rounded-full bg-white/20 text-white font-mono text-[10px]">{members.length}</span>}
        </button>

        <button
          data-testid="admin-tab-settings"
          onClick={() => setActiveTab('settings')}
          className={`flex-1 min-w-[100px] py-2.5 px-3 rounded-lg font-bold text-xs flex items-center justify-center gap-1.5 transition-colors ${
            activeTab === 'settings' ? 'bg-accent-magic text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'
          }`}
        >
          <Settings className="w-3.5 h-3.5" />
          Pengaturan
        </button>
      </div>

      {error && (
        <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* --- TAB 1: VERIFIKASI SUBMISSION --- */}
      {activeTab === 'submissions' && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
            <h3 className="font-bold text-text-primary text-sm">Antrean Verifikasi Bukti Tugas ({submissions.length})</h3>
            <span className="text-xs text-text-secondary">Approve memberi koin & EXP otomatis</span>
          </div>

          {submissions.length === 0 ? (
            <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
              <div className="w-10 h-10 mx-auto rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center"><CheckCircle2 className="w-5 h-5" /></div>
              <p className="font-bold text-text-primary text-sm">Tidak Ada Antrean Verifikasi</p>
              <p className="text-xs text-text-secondary max-w-xs mx-auto">Semua tugas dari anggota sudah selesai diperiksa.</p>
            </div>
          ) : (
            submissions.map((sub) => (
              <div
                key={sub.id}
                className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm space-y-3"
              >
                <div className="flex items-start justify-between gap-2">
                  <div>
                    {getTaskTypeBadge(sub.task_type)}
                    <h4 className="font-heading font-bold text-text-primary text-base mt-1">
                      {sub.task_title}
                    </h4>
                    <p className="text-xs text-text-secondary mt-0.5">
                      Oleh: <strong className="text-text-primary">{sub.user_name}</strong> •{' '}
                      {new Date(sub.created_at).toLocaleTimeString('id-ID', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </p>
                  </div>
                  <div className="text-right shrink-0">
                    <span className="text-xs font-bold text-accent-gold">
                      +{sub.reward_coins} 🪙
                    </span>
                    <p className="text-[10px] text-text-secondary">+{sub.reward_xp} XP</p>
                  </div>
                </div>

                {/* Proof Media Preview based on Task Type */}
                {/* 1. PHOTO or IMAGE */}
                {sub.payload?.file_url && (sub.task_type === 'PHOTO_UPLOAD' || sub.task_type === 'PHOTO_PROOF' || sub.payload.file_url.match(/\.(jpg|jpeg|png|webp)$/i)) && (
                  <div className="p-3 rounded-xl bg-surface border border-border-subtle space-y-2">
                    <div
                      onClick={() => setPreviewImage(sub.payload.file_url!)}
                      className="relative aspect-video max-h-56 rounded-xl overflow-hidden cursor-pointer bg-black group"
                    >
                      <img
                        src={sub.payload.file_url}
                        alt="Bukti Foto"
                        className="w-full h-full object-contain group-hover:scale-105 transition-transform"
                      />
                      <div className="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white text-xs font-bold transition-opacity">
                        Ketuk untuk Memperbesar 🔍
                      </div>
                    </div>
                    {sub.payload.file_size && (
                      <p className="text-[10px] text-text-secondary text-right">
                        Ukuran: {(sub.payload.file_size / 1024).toFixed(0)} KB
                      </p>
                    )}
                  </div>
                )}

                {/* 2. DOCUMENT */}
                {sub.payload?.file_url && !(sub.task_type === 'PHOTO_UPLOAD' || sub.task_type === 'PHOTO_PROOF' || sub.payload.file_url.match(/\.(jpg|jpeg|png|webp)$/i)) && (
                  <div className="p-3 rounded-xl bg-surface border border-border-subtle flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <FileText className="w-5 h-5 text-accent-magic" />
                      <div>
                        <p className="text-xs font-bold text-text-primary line-clamp-1">
                          {sub.payload.file_name || 'Dokumen Bukti'}
                        </p>
                        {sub.payload.file_size && (
                          <p className="text-[10px] text-text-secondary">
                            {(sub.payload.file_size / 1024).toFixed(1)} KB
                          </p>
                        )}
                      </div>
                    </div>
                    <a
                      href={sub.payload.file_url}
                      target="_blank"
                      rel="noreferrer"
                      className="px-3 py-1.5 rounded-lg bg-surface-elevated border border-border-subtle text-xs font-bold text-accent-magic hover:underline flex items-center gap-1"
                    >
                      <span>Buka / Unduh</span>
                      <ExternalLink className="w-3 h-3" />
                    </a>
                  </div>
                )}

                {/* 3. TEXT RESPONSE */}
                {sub.payload?.text && (
                  <div className="p-3 rounded-xl bg-surface border border-border-subtle space-y-1">
                    <p className="text-[10px] text-text-secondary uppercase font-bold flex items-center gap-1">
                      <PenLine className="w-3 h-3" />
                      <span>Respon Tertulis:</span>
                    </p>
                    <p className="text-xs text-text-primary whitespace-pre-wrap leading-relaxed">
                      {sub.payload.text}
                    </p>
                  </div>
                )}

                {/* 4. MINI GAME STATS */}
                {sub.payload?.score !== undefined && (
                  <div className="p-3 rounded-xl bg-surface border border-border-subtle flex items-center justify-between text-xs">
                    <span className="font-bold text-text-secondary flex items-center gap-1">
                      <Gamepad2 className="w-4 h-4 text-accent-magic" />
                      <span>Skor Game:</span>
                    </span>
                    <span className="font-bold text-status-success">
                      {sub.payload.score} Poin {sub.payload.moves ? `(${sub.payload.moves} langkah)` : ''}
                    </span>
                  </div>
                )}

                {sub.payload?.note && (
                  <p className="text-xs text-text-secondary italic">
                    Catatan: &quot;{sub.payload.note}&quot;
                  </p>
                )}

                {/* Admin note — associated for rejection context */}
                <label htmlFor={`note-${sub.id}`} className="text-[11px] font-bold text-text-secondary">Catatan untuk anggota (opsional, wajib jika menolak)</label>
                <input
                  id={`note-${sub.id}`}
                  type="text"
                  placeholder="Contoh: Foto kurang jelas, ulangi dari sudut lain"
                  value={actionNotes[sub.id] || ''}
                  onChange={(e) =>
                    setActionNotes((prev) => ({ ...prev, [sub.id]: e.target.value }))
                  }
                  className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                />

                {/* Action Buttons — Approve primary, Reject secondary */}
                <div className="flex gap-2 pt-1">
                  <button
                    type="button"
                    aria-label={`Tolak verifikasi ${sub.task_title}`}
                    disabled={processingId === sub.id}
                    onClick={() => handleVerifySubmission(sub.id, 'REJECTED')}
                    className="flex-1 py-2.5 rounded-xl bg-surface border border-status-error/20 text-status-error hover:bg-status-error/10 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <XCircle className="w-4 h-4" />
                    <span>{processingId === sub.id ? 'Memproses...' : 'Tolak'}</span>
                  </button>

                  <button
                    type="button"
                    aria-label={`Setujui ${sub.task_title}`}
                    disabled={processingId === sub.id}
                    aria-busy={processingId === sub.id}
                    onClick={() => handleVerifySubmission(sub.id, 'APPROVED')}
                    className="flex-1 py-2.5 rounded-xl bg-status-success hover:brightness-110 text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Setujui (+{sub.reward_coins}🪙)</span>
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* --- TAB 2: PENCAIRAN KOIN & KLAIM --- */}
      {activeTab === 'claims' && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
            <h3 className="font-bold text-text-primary text-sm">Permintaan Pencairan Koin ({claims.length})</h3>
            <span className="text-xs text-text-secondary">Transfer / Top-up saldo anggota</span>
          </div>

          {claims.length === 0 ? (
            <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
              <div className="w-10 h-10 mx-auto rounded-xl bg-status-success/10 text-status-success flex items-center justify-center"><CheckCircle2 className="w-5 h-5" /></div>
              <p className="font-bold text-text-primary text-sm">Tidak Ada Klaim Pending</p>
              <p className="text-xs text-text-secondary max-w-xs mx-auto">Semua pengajuan penukaran koin sudah selesai diproses.</p>
            </div>
          ) : (
            claims.map((claim) => (
              <div
                key={claim.id}
                className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm space-y-3"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-1.5">
                      {claim.target_type === 'EWALLET' ? (
                        <Wallet className="w-4 h-4 text-accent-magic" />
                      ) : claim.target_type === 'BANK' ? (
                        <Building2 className="w-4 h-4 text-status-success" />
                      ) : (
                        <Smartphone className="w-4 h-4 text-accent-cyan" />
                      )}
                      <h4 className="font-heading font-bold text-text-primary text-base">
                        Pencairan {claim.target_type}
                      </h4>
                    </div>
                    <p className="text-xs text-text-secondary mt-0.5">
                      Pemohon: <strong className="text-text-primary">{claim.user_name}</strong> •{' '}
                      {new Date(claim.created_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </p>
                  </div>
                  <span className="px-3 py-1 rounded-full bg-accent-gold/15 text-accent-gold font-bold text-xs flex items-center gap-1">
                    <Coins className="w-3.5 h-3.5" />
                    <span>{claim.coins_redeemed.toLocaleString('id-ID')} Koin</span>
                  </span>
                </div>

                {/* Destination info with Copy Button */}
                <div className="p-3 rounded-xl bg-surface border border-border-subtle flex items-center justify-between">
                  <div>
                    <p className="text-[10px] text-text-secondary uppercase font-bold">
                      Tujuan Penukaran / No. Rekening / No. HP
                    </p>
                    <p className="text-sm font-mono font-bold text-text-primary mt-0.5">
                      {claim.target_value}
                    </p>
                  </div>
                  <button
                    onClick={() => {
                      navigator.clipboard.writeText(claim.target_value)
                      alert('Nomor tujuan berhasil disalin!')
                    }}
                    className="p-2 rounded-lg bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary active:scale-95 transition-all"
                    title="Salin Nomor"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                </div>

                <label htmlFor={`claim-note-${claim.id}`} className="text-[11px] font-bold text-text-secondary">Catatan / No. Ref (opsional)</label>
                <input
                  id={`claim-note-${claim.id}`}
                  type="text"
                  placeholder="Contoh: TRF BCA 123456 — 28 Agu 2026"
                  value={actionNotes[claim.id] || ''}
                  onChange={(e) =>
                    setActionNotes((prev) => ({ ...prev, [claim.id]: e.target.value }))
                  }
                  className="w-full p-2.5 rounded-xl bg-surface border border-border-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                />

                <div className="flex gap-2 pt-1">
                  <button
                    type="button"
                    aria-label={`Tolak pencairan ${claim.target_value}`}
                    disabled={processingId === claim.id}
                    onClick={() => handleProcessClaim(claim.id, 'REJECTED')}
                    className="flex-1 py-2.5 rounded-xl bg-surface border border-status-error/20 text-status-error hover:bg-status-error/10 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <XCircle className="w-4 h-4" />
                    <span>{processingId === claim.id ? 'Memproses...' : 'Tolak & Refund'}</span>
                  </button>

                  <button
                    type="button"
                    aria-label={`Selesaikan pencairan ${claim.target_value}`}
                    disabled={processingId === claim.id}
                    aria-busy={processingId === claim.id}
                    onClick={() => handleProcessClaim(claim.id, 'APPROVED')}
                    className="flex-1 py-2.5 rounded-xl bg-status-success hover:brightness-110 text-white font-bold text-xs flex items-center justify-center gap-1.5 shadow-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Sudah Ditransfer</span>
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* --- TAB 3: JADWAL TUGAS HARIAN --- */}
      {activeTab === 'tasks' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <Calendar className="w-4 h-4 text-accent-magic" />
              <input
                type="date"
                value={selectedDate}
                onChange={(e) => setSelectedDate(e.target.value)}
                aria-label="Pilih tanggal tugas"
                className="p-2 rounded-xl bg-surface border border-border-subtle text-xs font-bold text-text-primary focus:outline-none focus:border-accent-magic"
              />
            </div>
            <button
              onClick={() => {
                setNewTask((prev) => ({ ...prev, step_order: tasks.length + 1 }))
                setIsCreateTaskModalOpen(true)
              }}
              className="px-3.5 py-2 rounded-xl bg-accent-magic text-white font-heading font-bold text-xs flex items-center gap-1.5 shadow-sm shadow-accent-magic/30 hover:brightness-110 active:scale-95 transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Tambah Tugas</span>
            </button>
          </div>

          {tasks.length === 0 ? (
            <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
              <div className="w-10 h-10 mx-auto rounded-xl bg-accent-cyan/10 text-accent-cyan flex items-center justify-center"><FileText className="w-5 h-5" /></div>
              <p className="font-bold text-text-primary text-sm">Belum Ada Tugas pada Tanggal Ini</p>
              <p className="text-xs text-text-secondary max-w-xs mx-auto">Klik &quot;Tambah Tugas&quot; untuk membuat urutan tugas.</p>
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
                          Personal: {members.find(m => m.uid === task.target_user_uid)?.explorer_name || task.target_user_uid}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-text-secondary line-clamp-1">
                      {task.description || 'Tanpa deskripsi'}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      {getTaskTypeBadge(task.task_type)}
                      <span className="text-[10px] font-bold text-accent-gold">
                        +{task.reward_coins} 🪙
                      </span>
                      <span className="text-[10px] font-bold text-accent-magic">
                        +{task.reward_xp} XP
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-1 shrink-0">
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
        </div>
      )}

      {/* --- TAB: MANAJEMEN ANGGOTA --- */}
      {activeTab === 'members' && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Users className="w-4 h-4 text-accent-magic" />
                <span>Daftar Anggota ({members.length})</span>
              </h3>
              <p className="text-xs text-text-secondary mt-0.5">
                Kelola profil anggota, role, status aktif, dan buat akun baru.
              </p>
            </div>
            <button
              onClick={() => setIsCreateMemberModalOpen(true)}
              className="px-3.5 py-2 rounded-xl bg-accent-magic text-white font-heading font-bold text-xs flex items-center gap-1.5 shadow-sm shadow-accent-magic/30 hover:brightness-110 active:scale-95 transition-all shrink-0"
            >
              <UserPlus className="w-4 h-4" />
              <span>+ Tambah Anggota</span>
            </button>
          </div>

          {members.length === 0 ? (
            <div className="py-12 text-center bg-surface rounded-2xl border border-border-subtle space-y-2 p-6">
              <div className="w-10 h-10 mx-auto rounded-xl bg-accent-magic/10 text-accent-magic flex items-center justify-center">
                <Users className="w-5 h-5" />
              </div>
              <p className="font-bold text-text-primary text-sm">Belum Ada Anggota</p>
              <p className="text-xs text-text-secondary max-w-xs mx-auto">
                Klik &quot;+ Tambah Anggota&quot; untuk mendaftarkan anggota baru.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {members.map((member) => (
                <div
                  key={member.uid}
                  className="p-4 rounded-2xl bg-surface border border-border-subtle shadow-sm flex items-start justify-between gap-3"
                >
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <h4 className="font-heading font-bold text-text-primary text-sm">
                        {member.explorer_name}
                      </h4>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                        member.role === 'GUIDE'
                          ? 'bg-accent-magic/15 text-accent-magic'
                          : 'bg-surface-elevated text-text-secondary border border-border-subtle'
                      }`}>
                        {member.role}
                      </span>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                        member.is_active
                          ? 'bg-status-success/15 text-status-success'
                          : 'bg-status-error/15 text-status-error'
                      }`}>
                        {member.is_active ? 'Aktif' : 'Nonaktif'}
                      </span>
                    </div>

                    <p className="text-xs text-text-secondary font-mono">
                      @{member.username} • <span className="text-[10px] opacity-70">UID: {member.uid}</span>
                    </p>

                    <div className="flex items-center gap-3 text-xs pt-1">
                      <span className="font-bold text-accent-gold flex items-center gap-1">
                        <Coins className="w-3.5 h-3.5" /> {member.coins} Koin
                      </span>
                      <span className="font-bold text-accent-magic">
                        Level {member.level} ({member.xp} XP)
                      </span>
                    </div>

                    <p className="text-[10px] text-text-secondary pt-0.5">
                      Bergabung: {new Date(member.created_at).toLocaleDateString('id-ID', { year: 'numeric', month: 'short', day: 'numeric' })}
                    </p>
                  </div>

                  <button
                    onClick={() => {
                      setSelectedMember(member)
                      setEditMemberForm({
                        explorer_name: member.explorer_name,
                        role: member.role,
                        is_active: member.is_active,
                        password: '',
                        reset_device: false,
                      })
                      setIsEditMemberModalOpen(true)
                    }}
                    className="p-2 rounded-xl bg-surface border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-surface-elevated transition-all shrink-0"
                    title="Edit Anggota"
                  >
                    <Edit3 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* --- TAB 4: PENGATURAN PERIODE PENUKARAN --- */}
      {activeTab === 'settings' && (
        <div className="space-y-4">
          <div className="p-5 rounded-2xl bg-surface border border-border-subtle shadow-sm space-y-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="font-heading font-bold text-text-primary text-base flex items-center gap-2">
                  <Sliders className="w-5 h-5 text-accent-magic" />
                  <span>Pengaturan Periode Penukaran Koin</span>
                </h3>
                <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                  Tentukan tanggal kalender setiap bulan di mana anggota dapat mengajukan pencairan koin menjadi uang tunai.
                </p>
              </div>

              {config && (
                <span
                  className={`text-xs font-bold px-3 py-1 rounded-full shrink-0 ${
                    config.is_open
                      ? 'bg-status-success/20 text-status-success'
                      : 'bg-surface text-text-secondary border border-border-subtle'
                  }`}
                >
                  {config.is_open ? '● Sedang Dibuka' : '○ Sedang Ditutup'}
                </span>
              )}
            </div>

            <form onSubmit={handleSaveConfig} className="space-y-4 pt-2 border-t border-border-subtle">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Nilai 1 Koin (Rupiah):
                  </label>
                  <input
                    type="number"
                    min={1}
                    required
                    value={conversionRateInput}
                    onChange={(e) => setConversionRateInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <p className="text-[11px] text-text-secondary">
                    Contoh: <strong>100</strong> (1 Koin = Rp 100)
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Target Penghasilan Normal (Rupiah):
                  </label>
                  <input
                    type="number"
                    min={0}
                    required
                    value={targetRupiahInput}
                    onChange={(e) => setTargetRupiahInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <p className="text-[11px] text-text-secondary">
                    Hasil kalkulasi koin: <strong>{conversionRateInput > 0 ? Math.floor(targetRupiahInput / conversionRateInput) : 0} Koin</strong>
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Batas Maksimum Pencairan (Koin):
                  </label>
                  <input
                    type="number"
                    min={1}
                    required
                    value={maxPayoutInput}
                    onChange={(e) => setMaxPayoutInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <p className="text-[11px] text-text-secondary">
                    Batas keras (cap) pencairan per periode
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Tanggal Gajian / Payday (1–31):
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={31}
                    required
                    value={payoutDayInput}
                    onChange={(e) => setPayoutDayInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <p className="text-[11px] text-text-secondary">
                    Hari kalender perayaan gajian
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Durasi Periode Earning (Hari):
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={365}
                    required
                    value={earningPeriodInput}
                    onChange={(e) => setEarningPeriodInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                  <p className="text-[11px] text-text-secondary">
                    Contoh: <strong>30</strong> (30 hari kerja)
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Timezone:
                  </label>
                  <input
                    type="text"
                    required
                    value={timezoneInput}
                    onChange={(e) => setTimezoneInput(e.target.value)}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                  />
                  <p className="text-[11px] text-text-secondary">
                    Contoh: <strong>Asia/Jakarta</strong>
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Tanggal Mulai Penukaran (1–31):
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={31}
                    required
                    value={startDayInput}
                    onChange={(e) => setStartDayInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-text-secondary">
                    Tanggal Akhir Penukaran (1–31):
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={31}
                    required
                    value={endDayInput}
                    onChange={(e) => setEndDayInput(Number(e.target.value))}
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                  />
                </div>
              </div>

              <div className="p-3 rounded-xl bg-surface border border-border-subtle text-xs">
                <span className="text-text-secondary font-medium">Pratinjau:</span>
                <span className="font-bold text-text-primary ml-1">
                  Target Rp {targetRupiahInput.toLocaleString('id-ID')} ({conversionRateInput > 0 ? Math.floor(targetRupiahInput / conversionRateInput) : 0} koin) • Maks {maxPayoutInput} koin • Gajian tgl {payoutDayInput} • Penukaran tgl {startDayInput}–{endDayInput} • {earningPeriodInput} hari ({timezoneInput})
                </span>
              </div>

              {configErrorMsg && (
                <div className="p-3.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{configErrorMsg}</span>
                </div>
              )}

              {configSuccessMsg && (
                <div className="p-3.5 rounded-xl bg-status-success/15 border border-status-success/30 text-status-success text-xs flex items-center gap-2">
                  <Check className="w-4 h-4 shrink-0" />
                  <span>{configSuccessMsg}</span>
                </div>
              )}

              <button
                type="submit"
                disabled={isSavingConfig}
                className="w-full py-3.5 rounded-2xl bg-accent-magic hover:brightness-110 active:scale-[0.98] text-white font-heading font-bold text-sm shadow-md shadow-accent-magic/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
              >
                {isSavingConfig ? (
                  <span>Menyimpan Pengaturan...</span>
                ) : (
                  <>
                    <Check className="w-4 h-4" />
                    <span>Simpan Pengaturan Periode</span>
                  </>
                )}
              </button>
            </form>
          </div>
        </div>
      )}

      {/* --- MODAL: CONFIGURABLE TASK BUILDER --- */}
      {isCreateTaskModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <motion.div
            initial={{ scale: 0.96, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]"
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
              <div>
                <h3 className="font-bold text-text-primary text-sm">Buat Tugas Baru</h3>
                <p className="text-xs text-text-secondary">Isi informasi dasar, pilih tipe, lalu atur konfigurasi.</p>
              </div>
              <button
                type="button"
                onClick={() => setIsCreateTaskModalOpen(false)}
                aria-label="Tutup"
                className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleCreateTask} className="p-5 space-y-5 overflow-y-auto">
              {/* 1 — Informasi Dasar */}
              <div className="space-y-3">
                <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary">1 — Informasi Dasar</h4>
                <div className="space-y-1">
                  <label htmlFor="task-title" className="text-xs font-bold text-text-secondary">Judul Tugas <span className="text-status-error">*</span></label>
                  <input
                    id="task-title"
                    type="text"
                    required
                    value={newTask.title}
                    onChange={(e) => setNewTask({ ...newTask, title: e.target.value })}
                    placeholder="Contoh: Merapikan Meja Belajar & Kamar"
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                  />
                </div>

                <div className="space-y-1">
                  <label htmlFor="task-desc" className="text-xs font-bold text-text-secondary">Deskripsi / Petunjuk</label>
                  <textarea
                    id="task-desc"
                    rows={2}
                    value={newTask.description}
                    onChange={(e) => setNewTask({ ...newTask, description: e.target.value })}
                    placeholder="Jelaskan apa yang harus dilakukan oleh anggota..."
                    className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic resize-none"
                  />
                </div>
              </div>

              {/* 2 — Target & Tipe */}
              <div className="space-y-3">
                <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary">2 — Target & Tipe</h4>
                
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <label htmlFor="task-target-scope" className="text-xs font-bold text-text-secondary">Target Penerima</label>
                    <select
                      id="task-target-scope"
                      value={newTask.target_scope}
                      onChange={(e) => setNewTask({ ...newTask, target_scope: e.target.value as 'ALL' | 'USER' })}
                      className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                    >
                      <option value="ALL">🌐 Semua Anggota Program</option>
                      <option value="USER">👤 User Tertentu (Personal Task)</option>
                    </select>
                  </div>

                  {newTask.target_scope === 'USER' ? (
                    <div className="space-y-1">
                      <label htmlFor="task-target-user" className="text-xs font-bold text-text-secondary">Pilih User Target</label>
                      <select
                        id="task-target-user"
                        required
                        value={newTask.target_user_uid}
                        onChange={(e) => setNewTask({ ...newTask, target_user_uid: e.target.value })}
                        className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                      >
                        <option value="">-- Pilih Anggota --</option>
                        {members.map((m) => (
                          <option key={m.uid} value={m.uid}>
                            {m.explorer_name} (@{m.username})
                          </option>
                        ))}
                      </select>
                    </div>
                  ) : (
                    <div className="space-y-1">
                      <label htmlFor="task-type" className="text-xs font-bold text-text-secondary">Tipe Tugas</label>
                      <select
                        id="task-type"
                        value={newTask.task_type}
                        onChange={(e) => setNewTask({ ...newTask, task_type: e.target.value as TaskType })}
                        className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                      >
                        <option value="VIDEO">🎥 Video YouTube</option>
                        <option value="QUIZ">🧠 Kuis Pilihan Ganda</option>
                        <option value="PHOTO_UPLOAD">📸 Upload Foto Bukti</option>
                        <option value="DOCUMENT_UPLOAD">📄 Upload Dokumen (Excel/Word/PDF)</option>
                        <option value="TEXT_RESPONSE">✍️ Respon Teks / Esai</option>
                        <option value="MINI_GAME">🎮 Mini Game Memori</option>
                      </select>
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {newTask.target_scope === 'USER' && (
                    <div className="space-y-1">
                      <label htmlFor="task-type" className="text-xs font-bold text-text-secondary">Tipe Tugas</label>
                      <select
                        id="task-type"
                        value={newTask.task_type}
                        onChange={(e) => setNewTask({ ...newTask, task_type: e.target.value as TaskType })}
                        className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-bold"
                      >
                        <option value="VIDEO">🎥 Video YouTube</option>
                        <option value="QUIZ">🧠 Kuis Pilihan Ganda</option>
                        <option value="PHOTO_UPLOAD">📸 Upload Foto Bukti</option>
                        <option value="DOCUMENT_UPLOAD">📄 Upload Dokumen (Excel/Word/PDF)</option>
                        <option value="TEXT_RESPONSE">✍️ Respon Teks / Esai</option>
                        <option value="MINI_GAME">🎮 Mini Game Memori</option>
                      </select>
                    </div>
                  )}

                  <div className="space-y-1">
                    <label htmlFor="task-step" className="text-xs font-bold text-text-secondary">Urutan Step (1, 2..)</label>
                    <input
                      id="task-step"
                      type="number"
                      min={1}
                      value={newTask.step_order}
                      onChange={(e) => setNewTask({ ...newTask, step_order: Number(e.target.value) })}
                      className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                    />
                  </div>
                </div>
              </div>

              {/* 3 — Hadiah */}
              <div className="space-y-3">
                <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary">3 — Hadiah</h4>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <label htmlFor="task-coins" className="text-xs font-bold text-text-secondary">Koin (🪙)</label>
                    <input
                      id="task-coins"
                      type="number"
                      min={1}
                      value={newTask.reward_coins}
                      onChange={(e) => setNewTask({ ...newTask, reward_coins: Number(e.target.value) })}
                      className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                    />
                  </div>
                  <div className="space-y-1">
                    <label htmlFor="task-xp" className="text-xs font-bold text-text-secondary">EXP</label>
                    <input
                      id="task-xp"
                      type="number"
                      min={1}
                      value={newTask.reward_xp}
                      onChange={(e) => setNewTask({ ...newTask, reward_xp: Number(e.target.value) })}
                      className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic focus:ring-1 focus:ring-accent-magic"
                    />
                  </div>
                </div>
              </div>

              {/* 4 — Konfigurasi spesifik tipe (progressive disclosure) */}
              {/* VIDEO */}
              {(newTask.task_type === 'VIDEO' || newTask.task_type === 'YOUTUBE_VIDEO') && (
                <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                    <Play className="w-3.5 h-3.5 text-accent-magic" /> 4 — Video YouTube
                  </h4>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Link YouTube:</label>
                    <input
                      type="url"
                      required
                      value={newTask.youtube_url}
                      onChange={(e) => setNewTask({ ...newTask, youtube_url: e.target.value })}
                      placeholder="https://www.youtube.com/watch?v=..."
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                </div>
              )}

              {/* KUIS */}
              {(newTask.task_type === 'QUIZ' || newTask.task_type === 'VIDEO_QUIZ') && (
                <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                    <HelpCircle className="w-3.5 h-3.5 text-accent-magic" /> 4 — Soal Kuis
                  </h4>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Link YouTube (Opsional):</label>
                    <input
                      type="url"
                      value={newTask.youtube_url}
                      onChange={(e) => setNewTask({ ...newTask, youtube_url: e.target.value })}
                      placeholder="https://www.youtube.com/watch?v=..."
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Pertanyaan Kuis:</label>
                    <input
                      type="text"
                      required
                      value={newTask.question_text}
                      onChange={(e) => setNewTask({ ...newTask, question_text: e.target.value })}
                      placeholder="Contoh: Apa manfaat menabung sejak dini?"
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <input
                      type="text"
                      required
                      value={newTask.opt_a}
                      onChange={(e) => setNewTask({ ...newTask, opt_a: e.target.value })}
                      placeholder="Pilihan A"
                      className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary"
                    />
                    <input
                      type="text"
                      required
                      value={newTask.opt_b}
                      onChange={(e) => setNewTask({ ...newTask, opt_b: e.target.value })}
                      placeholder="Pilihan B"
                      className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary"
                    />
                    <input
                      type="text"
                      value={newTask.opt_c}
                      onChange={(e) => setNewTask({ ...newTask, opt_c: e.target.value })}
                      placeholder="Pilihan C (opsional)"
                      className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary"
                    />
                    <input
                      type="text"
                      value={newTask.opt_d}
                      onChange={(e) => setNewTask({ ...newTask, opt_d: e.target.value })}
                      placeholder="Pilihan D (opsional)"
                      className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Kunci Jawaban Benar (Server-Side):</label>
                    <select
                      value={newTask.correct_ans}
                      onChange={(e) => setNewTask({ ...newTask, correct_ans: e.target.value })}
                      className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary font-bold"
                    >
                      <option value="A">Pilihan A</option>
                      <option value="B">Pilihan B</option>
                      <option value="C">Pilihan C</option>
                      <option value="D">Pilihan D</option>
                    </select>
                  </div>
                </div>
              )}

              {/* 3. PHOTO UPLOAD CONFIG */}
              {(newTask.task_type === 'PHOTO_UPLOAD' || newTask.task_type === 'PHOTO_PROOF') && (
                <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                    <Camera className="w-3.5 h-3.5 text-accent-cyan" /> 4 — Foto Bukti
                  </h4>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Instruksi Pengambilan Foto:</label>
                    <input
                      type="text"
                      value={newTask.photo_instruction}
                      onChange={(e) => setNewTask({ ...newTask, photo_instruction: e.target.value })}
                      placeholder="Contoh: Foto meja belajar yang sudah rapi dan bersih."
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                </div>
              )}

              {/* 4. DOCUMENT UPLOAD CONFIG */}
              {newTask.task_type === 'DOCUMENT_UPLOAD' && (
                <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                    <FileText className="w-3.5 h-3.5 text-accent-gold" /> 4 — Dokumen & Template
                  </h4>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Link File Template / Attachment (Opsional):</label>
                    <input
                      type="url"
                      value={newTask.attachment_url}
                      onChange={(e) => setNewTask({ ...newTask, attachment_url: e.target.value })}
                      placeholder="https://cdn.example.com/template-laporan.xlsx"
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Nama Dokumen Template:</label>
                    <input
                      type="text"
                      value={newTask.attachment_name}
                      onChange={(e) => setNewTask({ ...newTask, attachment_name: e.target.value })}
                      placeholder="Contoh: Lembar_Laporan_Keuangan.xlsx"
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Format File yang Diterima:</label>
                    <input
                      type="text"
                      value={newTask.doc_extensions}
                      onChange={(e) => setNewTask({ ...newTask, doc_extensions: e.target.value })}
                      placeholder=".xlsx, .docx, .pdf"
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                </div>
              )}

              {/* 5. TEXT RESPONSE CONFIG */}
              {newTask.task_type === 'TEXT_RESPONSE' && (
                <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                    <PenLine className="w-3.5 h-3.5 text-accent-magic" /> 4 — Respon Teks
                  </h4>
                  <div className="space-y-1">
                    <label className="text-[11px] text-text-secondary">Pertanyaan / Topik Respon:</label>
                    <input
                      type="text"
                      required
                      value={newTask.text_prompt}
                      onChange={(e) => setNewTask({ ...newTask, text_prompt: e.target.value })}
                      placeholder="Contoh: Apa pelajaran terpenting hari ini?"
                      className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    <div className="space-y-1">
                      <label className="text-[10px] text-text-secondary">Min Karakter:</label>
                      <input
                        type="number"
                        min={1}
                        value={newTask.min_chars}
                        onChange={(e) => setNewTask({ ...newTask, min_chars: Number(e.target.value) })}
                        className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                      />
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] text-text-secondary">Maks Karakter:</label>
                      <input
                        type="number"
                        min={50}
                        value={newTask.max_chars}
                        onChange={(e) => setNewTask({ ...newTask, max_chars: Number(e.target.value) })}
                        className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* 6. MINI GAME CONFIG */}
              {newTask.task_type === 'MINI_GAME' && (
                <div className="p-4 rounded-2xl bg-surface border border-border-subtle space-y-3">
                  <h4 className="text-xs font-bold tracking-wide uppercase text-text-secondary flex items-center gap-1.5">
                    <Gamepad2 className="w-3.5 h-3.5 text-status-success" /> 4 — Mini Game
                  </h4>
                  <div className="grid grid-cols-2 gap-2">
                    <div className="space-y-1">
                      <label className="text-[11px] text-text-secondary">Tingkat Kesulitan:</label>
                      <select
                        value={newTask.game_difficulty}
                        onChange={(e) => setNewTask({ ...newTask, game_difficulty: e.target.value })}
                        className="w-full p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary font-bold"
                      >
                        <option value="EASY">Mudah (8 Kartu)</option>
                        <option value="MEDIUM">Sedang (12 Kartu)</option>
                        <option value="HARD">Tantangan (16 Kartu)</option>
                      </select>
                    </div>
                    <div className="space-y-1">
                      <label className="text-[11px] text-text-secondary">Target Skor Minimum:</label>
                      <input
                        type="number"
                        min={50}
                        max={100}
                        value={newTask.target_score}
                        onChange={(e) => setNewTask({ ...newTask, target_score: Number(e.target.value) })}
                        className="p-2 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary w-full"
                      />
                    </div>
                  </div>
                </div>
              )}

              <div className="sticky bottom-0 -mx-5 -mb-5 mt-2 px-5 py-4 border-t border-border-subtle bg-surface flex gap-3">
                <button
                  type="button"
                  onClick={() => setIsCreateTaskModalOpen(false)}
                  className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm hover:bg-surface transition-colors"
                >
                  Batalkan
                </button>
                <button
                  type="submit"
                  className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-sm hover:brightness-110 transition-all"
                >
                  Simpan & Terbitkan
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}

      {/* Preview foto — consistent rounded-2xl + X icon */}
      {previewImage && (
        <div
          onClick={() => setPreviewImage(null)}
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm cursor-pointer"
          role="dialog"
          aria-label="Pratinjau foto"
        >
          <div className="relative max-w-xl max-h-[85vh] rounded-2xl overflow-hidden shadow-xl border border-white/20 bg-black">
            <img src={previewImage} alt="Pratinjau foto bukti" className="w-full h-full object-contain max-h-[85vh]" />
            <button
              type="button"
              onClick={() => setPreviewImage(null)}
              aria-label="Tutup pratinjau"
              className="absolute top-3 right-3 w-8 h-8 rounded-full bg-black/60 backdrop-blur text-white flex items-center justify-center hover:bg-black/80 transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
      {/* --- MODAL: CREATE MEMBER --- */}
      {isCreateMemberModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <motion.div
            initial={{ scale: 0.96, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="w-full max-w-md bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden"
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
              <div>
                <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                  <UserPlus className="w-4 h-4 text-accent-magic" />
                  <span>Tambah Anggota Baru</span>
                </h3>
                <p className="text-xs text-text-secondary">Akun baru dapat langsung login di HP anggota.</p>
              </div>
              <button
                type="button"
                onClick={() => setIsCreateMemberModalOpen(false)}
                className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleCreateMember} className="p-5 space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Nama Lengkap / Panggilan <span className="text-status-error">*</span></label>
                <input
                  type="text"
                  required
                  placeholder="Contoh: Andi Wijaya"
                  value={newMember.explorer_name}
                  onChange={(e) => setNewMember({ ...newMember, explorer_name: e.target.value })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Username (untuk Login) <span className="text-status-error">*</span></label>
                <input
                  type="text"
                  required
                  placeholder="Contoh: andiwijaya"
                  value={newMember.username}
                  onChange={(e) => setNewMember({ ...newMember, username: e.target.value })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Password Awal <span className="text-status-error">*</span></label>
                <input
                  type="password"
                  required
                  minLength={6}
                  placeholder="Minimal 6 karakter"
                  value={newMember.password}
                  onChange={(e) => setNewMember({ ...newMember, password: e.target.value })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Role</label>
                <select
                  value={newMember.role}
                  onChange={(e) => setNewMember({ ...newMember, role: e.target.value as Role })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                >
                  <option value="MEMBER">MEMBER (Anggota biasa)</option>
                  <option value="ADMIN">ADMIN (Administrator)</option>
                </select>
              </div>

              <div className="pt-2 flex gap-3">
                <button
                  type="button"
                  onClick={() => setIsCreateMemberModalOpen(false)}
                  className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-md hover:brightness-110"
                >
                  Buat Akun
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}

      {/* --- MODAL: EDIT MEMBER --- */}
      {isEditMemberModalOpen && selectedMember && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <motion.div
            initial={{ scale: 0.96, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="w-full max-w-md bg-surface border border-border-subtle rounded-2xl shadow-xl overflow-hidden"
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-border-subtle bg-surface">
              <div>
                <h3 className="font-bold text-text-primary text-sm flex items-center gap-2">
                  <Edit3 className="w-4 h-4 text-accent-magic" />
                  <span>Edit Anggota @{selectedMember.username}</span>
                </h3>
              </div>
              <button
                type="button"
                onClick={() => {
                  setIsEditMemberModalOpen(false)
                  setSelectedMember(null)
                }}
                className="w-8 h-8 rounded-full bg-surface-elevated border border-border-subtle text-text-secondary hover:text-text-primary flex items-center justify-center transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleUpdateMember} className="p-5 space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Nama Lengkap</label>
                <input
                  type="text"
                  required
                  value={editMemberForm.explorer_name}
                  onChange={(e) => setEditMemberForm({ ...editMemberForm, explorer_name: e.target.value })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Role</label>
                <select
                  value={editMemberForm.role}
                  onChange={(e) => setEditMemberForm({ ...editMemberForm, role: e.target.value as Role })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                >
                  <option value="MEMBER">MEMBER (Anggota biasa)</option>
                  <option value="ADMIN">ADMIN (Administrator)</option>
                </select>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Status Akun</label>
                <select
                  value={editMemberForm.is_active ? 'active' : 'inactive'}
                  onChange={(e) => setEditMemberForm({ ...editMemberForm, is_active: e.target.value === 'active' })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm font-bold text-text-primary focus:outline-none focus:border-accent-magic"
                >
                  <option value="active">● Aktif (Bisa Login & Mengerjakan Tugas)</option>
                  <option value="inactive">○ Nonaktif (Akses Login Diblokir)</option>
                </select>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">Reset Password (Opsional)</label>
                <input
                  type="password"
                  placeholder="Kosongkan jika tidak diubah"
                  value={editMemberForm.password}
                  onChange={(e) => setEditMemberForm({ ...editMemberForm, password: e.target.value })}
                  className="w-full p-3 rounded-xl bg-surface border border-border-subtle text-sm text-text-primary focus:outline-none focus:border-accent-magic font-mono"
                />
              </div>

              <div className="p-3 rounded-xl bg-accent-gold/10 border border-accent-gold/20 flex items-center justify-between">
                <div>
                  <p className="text-xs font-bold text-text-primary">Reset Binding Perangkat</p>
                  <p className="text-[11px] text-text-secondary">Izinkan akun login di HP/perangkat baru</p>
                </div>
                <input
                  type="checkbox"
                  checked={editMemberForm.reset_device}
                  onChange={(e) => setEditMemberForm({ ...editMemberForm, reset_device: e.target.checked })}
                  className="w-4 h-4 rounded border-border-subtle text-accent-magic focus:ring-accent-magic"
                />
              </div>

              <div className="pt-2 flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setIsEditMemberModalOpen(false)
                    setSelectedMember(null)
                  }}
                  className="flex-1 py-3 rounded-xl bg-surface-elevated border border-border-subtle text-text-primary font-bold text-sm"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  className="flex-1 py-3 rounded-xl bg-accent-magic text-white font-bold text-sm shadow-md hover:brightness-110"
                >
                  Simpan Perubahan
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </div>
  )
}
