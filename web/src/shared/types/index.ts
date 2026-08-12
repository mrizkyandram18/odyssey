export type Role = 'SEEKER' | 'BUILDER' | 'GUIDE'

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

export interface Loginrequest {
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
  avatar_style: string
  avatar_seed: string
  /** Slice 2.2 equipped frame: none | gold */
  avatar_frame?: string
  /** Slice 4.5 equipped explorer effect: none | sparkle | float | trail */
  equipped_explorer_effect?: string
  created_at: string
  updated_at: string
}

export interface CosmeticCatalogItem {
  id: string
  name: string
  description: string
  price: number
  kind: string
  value: string
  unlocked: boolean
}

export interface CosmeticsResponse {
  coins: number
  items: CosmeticCatalogItem[]
  avatar_frame: string
  explorer_effect: string
}

export interface CosmeticPurchaseResult {
  status: 'purchased' | 'already_owned' | string
  cosmetic_id: string
  price: number
  coins: number
  avatar_frame: string
  explorer_effect: string
  already_owned: boolean
}

export type MeResponse = Explorer

export type MissionStatus = 'PENDING' | 'ACTIVE' | 'DONE'
export type MissionType = 'SOLO' | 'RELAY' | 'CREATIVE' | 'GROUP'

export interface Mission {
  id: number
  family_id: string
  template_slug: string
  title: string
  status: MissionStatus
  Mission_type?: MissionType | string
  active_challenge_assigned_to?: string
  started_at?: string | null
  completed_at?: string | null
  created_at: string
}

export type ChallengeStatus = 'PENDING' | 'DONE'
export type ChallengeType = 'OBSERVATION' | 'RESEARCH' | 'PUZZLE' | 'MOVEMENT' | 'DRAW' | 'WRITE'

export interface Exercise {
  id: number
  mission_id: number
  slug: string
  description: string
  status: ChallengeStatus
  assigned_to?: string | null
  completed_by?: string | null
  completed_at?: string | null
  created_at: string
  type?: string
  question?: string
  options?: string[]
  explanation?: string
}

export type RealmStatus = 'LOCKED' | 'ACTIVE' | 'COMPLETE'

export interface JourneyProgress {
  family_id: string
  journey: string
  status: RealmStatus
  story_branch?: string | null
  progress: number
  last_unlocked_at?: string | null
  updated_at: string
}

export type SubmissionKind = 'STORY' | 'COMIC' | 'PHOTO' | 'VIDEO' | 'DRAWING'
export type SubmissionStatus = 'PENDING' | 'APPROVED' | 'REJECTED'

export interface CreativeItem {
  id: number
  family_id: string
  journey: string
  author_uid: string
  kind: SubmissionKind
  payload: string
  created_at: string
}

export interface CreativeSubmission {
  id: number
  mission_id: number
  exercise_id: number
  family_id: string
  author_uid: string
  kind: SubmissionKind
  content: string
  status: SubmissionStatus
  reviewed_by?: string | null
  reviewed_at?: string | null
  rejection_reason?: string | null
  created_at: string
  updated_at: string
}

export interface DailyMission {
  id: number
  uid: string
  date: string
  mission_slug: string
  completed: boolean
  created_at: string
}

export type AchievementKind = 'PERSONAL' | 'GROUP'

export type AchievementTrigger =
  | 'Mission_COMPLETED'
  | 'REALM_COMPLETED'
  | 'RELIC_COLLECTED'
  | 'DAILY_STREAK'
  | 'CREATIVE_SUBMISSION'
  | 'LEVEL_REACHED'

export interface Achievement {
  id: number
  uid: string
  family_id?: string | null
  code: string
  title?: string
  description?: string
  kind: AchievementKind
  trigger?: AchievementTrigger
  awarded_at: string
  created_at: string
}

export interface AchievementView {
  id: number
  code: string
  slug?: string
  title: string
  description?: string
  icon?: string
  kind: AchievementKind
  trigger: AchievementTrigger
  threshold: number
  progress: number
  unlocked: boolean
  awarded_at?: string | null
}

export interface ChapterDefinition {
  id: number
  slug: string
  journey: string
  title: string
  description: string
  order: number
  created_at: string
  updated_at: string
}

export type ChapterStatus = 'LOCKED' | 'ACTIVE' | 'COMPLETE'

export interface ChapterSummary {
  definition: ChapterDefinition
  status: ChapterStatus
  completed_at?: string | null
}

export interface courseProgressView {
  current_course?: ChapterSummary | null
  next_course?: ChapterSummary | null
  completed_chapters: ChapterSummary[]
  unlocked_chapters: ChapterSummary[]
  all_chapters: ChapterSummary[]
}

export interface LoreDefinition {
  id: number
  slug: string
  journey: string
  course: string
  title: string
  content: string
  order: number
  season_slug?: string | null
}

export interface LoreView {
  slug: string
  journey: string
  course: string
  title: string
  content: string
  order: number
  unlocked: boolean
  unlocked_at?: string | null
}

export interface LoreSummary {
  locked_count: number
  unlocked_count: number
  latest?: LoreView | null
}

export interface SeasonDefinition {
  id: number
  slug: string
  name: string
  description: string
  start_at: string
  end_at: string
  journey: string
  created_at: string
  updated_at: string
}

export type SeasonState = 'UPCOMING' | 'ACTIVE' | 'EXPIRED' | 'INACTIVE'

export interface SeasonSummary {
  definition: SeasonDefinition
  state: SeasonState
}

export interface Collection {
  id: number
  uid: string
  code: string
  name: string
  description: string
  journey: string
  rarity: string
  image: string
  concept: string
  awarded_at: string
  created_at: string
}

export interface Gift {
  id: number
  uid: string
  source: string
  gift_slug: string
  rarity: string
  icon: string
  description: string
  drop_table: string
  opened: boolean
  opened_at?: string | null
  created_at: string
}

export interface PlayerRelic {
  collection_id: number
  collection_slug: string
  name: string
  description: string
  journey: string
  rarity: string
  image: string
  concept: string
  owned_count: number
  is_new: boolean
  discovered_at: string
  created_at: string
}

export interface ChestView {
  id: number
  uid: string
  source: string
  gift_slug: string
  name: string
  rarity: string
  icon: string
  description: string
  opened: boolean
  opened_at?: string | null
  created_at: string
}

export interface RewardItem {
  collection_slug: string
  name: string
  rarity: string
  is_new: boolean
}

export interface OpenResult {
  chest: ChestView
  rewards: RewardItem[]
  new_count: number
  duplicate_count: number
}

export interface RelicDefinition {
  id: number
  slug: string
  name: string
  description: string
  journey: string
  rarity: string
  image: string
  concept: string
}

export interface InventoryItem {
  collection_id: number
  collection_slug: string
  name: string
  description: string
  journey: string
  rarity: string
  image: string
  concept: string
  owned_count: number
  is_new: boolean
  discovered_at: string
  created_at: string
}

export interface GiftRelicrequest {
  recipient_uid: string
  collection_slug: string
}

export interface GiftRelicResult {
  collection_slug: string
  relic_name: string
  recipient_uid: string
  recipient_name: string
  sender_remaining_count: number
}

export interface MissionView {
  id: number
  family_id: string
  template_slug: string
  title: string
  status: MissionStatus
  Mission_type?: MissionType | string
  started_at?: string | null
  completed_at?: string | null
  created_at: string
  challenge_count: number
  completed_count: number
  active_challenge_assigned_to?: string
  season_slug?: string
}

export interface BranchOption {
  slug: string
  title: string
  description: string
}

export interface SelectBranchResult {
  success: boolean
  story_branch: string
  journey: string
}

export interface CrewMember {
  uid: string
  explorer_name: string
  role?: string
  level?: number
}

export interface MissionWithChallenges {
  id: number
  family_id: string
  template_slug: string
  title: string
  status: MissionStatus
  Mission_type?: MissionType | string
  active_challenge_assigned_to?: string
  started_at?: string | null
  completed_at?: string | null
  created_at: string
  learn_text?: string
  result_text?: string
  exercises: Exercise[]
  members?: CrewMember[]
  branch_options?: BranchOption[]
}

export interface CompleteChallengeResult {
  Mission?: MissionWithChallenges | null
  Mission_completed: boolean
  next_action?: string
  xp: number
  new_level: number
  level_up: boolean
}

export interface dailyMissionView {
  today: string
  completed: boolean
  available: boolean
  streak_days: number
  crew_streak: number
  remaining_turns: number
  mission_slug?: string
}

export interface HomeResponse {
  player: Explorer
  missions: MissionView[]
  daily_mission: dailyMissionView
  journey_progress: JourneyProgress[]
  relic_count: number
  active_missions: MissionView[]
  completed_missions_today: MissionView[]
  pending_creative_review: number
  last_submission?: CreativeSubmission | null
  sections: HomeSections
  available_gifts: ChestView[]
  latest_relic?: InventoryItem | null
  collection_progress: { collected: number; total: number }
  course_progress?: courseProgressView | null
  concept_summary?: LoreSummary | null
  achievements?: AchievementView[] | null
  current_season?: SeasonSummary | null
  season_progress?: {
    season_slug: string
    season_name: string
    missions_completed: number
    journey_progress: number
    journey_status: string
  }
}

export interface HomeSections {
  player: PlayerSection
  missions: MissionsSection
  daily_mission: dailyMissionSection
  journey: RealmSection
  world: WorldSection
  creative: CreativeSection
  gifts: ChestsSection
  collections: RelicsSection
  concept: LoreSection
  achievements: AchievementsSection
}

export interface PlayerSection {
  uid: string
  family_id: string
  explorer_name: string
  role: Role
  level: number
  xp: number
  coins: number
  avatar_style: string
  avatar_seed: string
  created_at: string
  updated_at: string
  xp_to_next: number
}

export interface RewardLedgerEntry {
  id: string
  user_id: string
  source: string
  amount: number
  reward_type: string
  metadata?: string | null
  created_at: string
}

export interface MissionsSection {
  all: MissionView[]
  active: MissionView[]
  done: MissionView[]
  done_today: MissionView[]
}

export interface dailyMissionSection {
  today: string
  completed: boolean
  available: boolean
  streak_days: number
  crew_streak: number
  remaining_turns: number
  mission_slug?: string
}

export interface RealmSection {
  progress: JourneyProgress[]
}

export interface WorldSection {
  current_course?: ChapterSummary | null
  next_course?: ChapterSummary | null
  completed_chapters: ChapterSummary[]
  unlocked_chapters: ChapterSummary[]
  all_chapters: ChapterSummary[]
}

export interface CreativeSection {
  pending_review_count: number
  last_submission?: CreativeSubmission | null
}

export interface ChestsSection {
  available: ChestView[]
}

export interface RelicsSection {
  latest?: InventoryItem | null
  collection_progress: { collected: number; total: number }
}

export interface LoreSection {
  summary?: LoreSummary | null
  unlocked: LoreView[]
}

export interface AchievementsSection {
  all: AchievementView[]
  recent: AchievementView[]
  count: number
}

export interface learningConceptView {
  slug: string
  journey: string
  title: string
  content: string
  set_name: string
  is_hidden: boolean
  discovered: boolean
  discovered_at?: string | null
}

export interface DiscoverResult {
  fragment: learningConceptView
  discovered: boolean
  xp_granted: number
}

export interface ReplayResult {
  journey: string
  is_replay: boolean
  bonus_dialogue: string
  unlocked_fragments: learningConceptView[]
}

export interface ApiError {
  error: string
}
