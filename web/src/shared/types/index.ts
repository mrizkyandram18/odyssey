export type Role = 'ADMIN' | 'MEMBER' | 'SEEKER' | 'BUILDER' | 'GUIDE'

export type LoginMethod = 'BOTH' | 'PASSWORD' | 'GATEKEEPER'

export interface DevicePayload {
  device_id: string
  login_method: LoginMethod
  device_label?: string
  device_name?: string
  device_model?: string
  platform?: string
  browser?: string
}

export interface Session {
  uid: string
  family_id: string
  kind: 'user' | 'setup'
  role?: Role
  expires: number
  token: string
}

export interface LoginRequest {
  uid: string
  credential: string
  device: DevicePayload
}

export interface LoginResponse {
  status: 'success' | 'setup_needed' | 'password_required'
  session?: string
  setup_token?: string
  uid?: string
  family_id?: string
  kind?: 'user' | 'setup'
  role?: Role
  expires?: number
  message?: string
}

export interface Family {
  id: string
  name?: string
  banner_url?: string
  theme?: string
  created_at: string
  updated_at: string
}

export interface Explorer {
  uid: string
  family_id: string
  explorer_name: string
  role: Role
  level: number
  xp: number
  coins: number
  streak_days?: number
  last_active_date?: string
  avatar_style: string
  avatar_seed: string
  avatar_frame?: string
  equipped_explorer_effect?: string
  created_at: string
  updated_at: string
}

export type MeResponse = Explorer

export interface ApiError {
  error: string
}

export type TaskType =
  | 'VIDEO'
  | 'QUIZ'
  | 'PHOTO_UPLOAD'
  | 'DOCUMENT_UPLOAD'
  | 'TEXT_RESPONSE'
  | 'MINI_GAME'
  | 'VIDEO_QUIZ'
  | 'PHOTO_PROOF'
  | 'GENERAL'
  | 'YOUTUBE_VIDEO'

export type EvaluationType = 'AUTO' | 'ADMIN_REVIEW'

export type TaskStatus = 'LOCKED' | 'UNLOCKED' | 'PENDING' | 'APPROVED' | 'REJECTED'

export interface QuizQuestion {
  id: string | number
  question: string
  options: string[]
}

export interface VideoConfig {
  video_url?: string
  youtube_url?: string
  minimum_duration_seconds?: number
  minimum_watch_seconds?: number
}

export interface QuizConfig {
  questions?: QuizQuestion[]
}

export interface PhotoUploadConfig {
  instruction?: string
  max_files?: number
  accepted_mime_types?: string[]
}

export interface DocUploadConfig {
  instruction?: string
  attachment_url?: string
  attachment_name?: string
  accepted_extensions?: string[]
  max_file_size_mb?: number
}

export interface TextResponseConfig {
  instruction?: string
  prompt?: string
  minimum_characters?: number
  maximum_characters?: number
}

export interface MiniGameConfig {
  game?: 'MEMORY' | 'MEMORY_CHALLENGE' | string
  difficulty?: 'EASY' | 'MEDIUM' | 'HARD'
  target_score?: number
  time_limit_seconds?: number
}

export interface TaskConfig extends VideoConfig, QuizConfig, PhotoUploadConfig, DocUploadConfig, TextResponseConfig, MiniGameConfig {
  [key: string]: any
}

export interface MemberView {
  uid: string
  family_id: string
  explorer_name: string
  username: string
  role: Role
  is_active: boolean
  level: number
  xp: number
  coins: number
  created_at: string
}

export interface CreateMemberInput {
  username: string
  password: string
  explorer_name: string
  role?: Role
}

export interface UpdateMemberInput {
  explorer_name?: string
  role?: Role
  is_active?: boolean
  password?: string
  reset_device?: boolean
}

export interface TaskView {
  id: number
  title: string
  description?: string
  task_type: TaskType
  evaluation_type?: EvaluationType
  step_order: number
  reward_coins: number
  reward_xp: number
  config: TaskConfig
  target_scope?: 'ALL' | 'FAMILY' | 'USER'
  target_user_uid?: string
  is_locked: boolean
  status: TaskStatus
  admin_notes?: string
  coins_earned: number
  xp_earned: number
  submitted_at?: string
}

export interface SubmitTaskResponse {
  success: boolean
  submission_id?: number
  status?: string
  coins_earned?: number
  xp_earned?: number
  new_balance?: number
  new_xp?: number
  streak?: number
  error?: string
}

export interface RedemptionConfig {
  redemption_start_day: number
  redemption_end_day: number
  payout_day: number
  earning_period_days: number
  is_open: boolean
  is_payout_day: boolean
  current_day: number
  conversion_rate: number
  payout_target_rupiah: number
  payout_target_coins: number
  max_payout_coins: number
  timezone: string
}

export interface RewardCatalogItem {
  id: number
  title: string
  description?: string
  category: 'PULSA' | 'EWALLET' | 'CASH' | 'SPECIAL'
  cost_coins: number
  icon_name?: string
  is_available: boolean
}

export interface ClaimView {
  id: number
  user_uid: string
  user_name?: string
  reward_id?: number
  reward_title?: string
  coins_redeemed: number
  target_type: string
  target_value: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  admin_notes?: string
  created_at: string
  processed_at?: string
}

export interface PendingSubmissionView {
  id: number
  task_id: number
  task_title: string
  task_type: TaskType
  user_uid: string
  user_name: string
  submission_type: string
  status: string
  payload: {
    answers?: Record<string, string>
    file_url?: string
    file_name?: string
    file_size?: number
    captured_at?: string
    note?: string
    text?: string
    score?: number
    game?: string
    moves?: number
    [key: string]: any
  }
  created_at: string
  reward_coins: number
  reward_xp: number
}
