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
  crew_id: string
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
  crew_id?: string
  kind?: 'user' | 'setup'
  role?: Role
  expires?: number
  message?: string
}

export interface Crew {
  id: string
  name?: string
  created_at: string
  updated_at: string
}

export interface Explorer {
  uid: string
  crew_id: string
  explorer_name: string
  role: Role
  level: number
  xp: number
  created_at: string
  updated_at: string
}

export type MeResponse = Explorer

export type QuestStatus = 'PENDING' | 'ACTIVE' | 'DONE'

export interface Quest {
  id: number
  crew_id: string
  template_slug: string
  title: string
  status: QuestStatus
  started_at?: string | null
  completed_at?: string | null
  created_at: string
}

export type ChallengeStatus = 'PENDING' | 'DONE'
export type ChallengeType = 'OBSERVATION' | 'RESEARCH' | 'PUZZLE' | 'MOVEMENT' | 'DRAW' | 'WRITE'

export interface Challenge {
  id: number
  quest_id: number
  slug: string
  description: string
  status: ChallengeStatus
  completed_by?: string | null
  completed_at?: string | null
  created_at: string
}

export type RealmStatus = 'LOCKED' | 'ACTIVE' | 'COMPLETE'

export interface RealmProgress {
  crew_id: string
  realm: string
  status: RealmStatus
  story_branch?: string | null
  progress: number
  last_unlocked_at?: string | null
  updated_at: string
}

export type SubmissionKind = 'STORY' | 'COMIC' | 'PHOTO' | 'VIDEO'
export type SubmissionStatus = 'PENDING' | 'APPROVED' | 'REJECTED'

export interface CreativeItem {
  id: number
  crew_id: string
  realm: string
  author_uid: string
  kind: SubmissionKind
  payload: string
  created_at: string
}

export interface CreativeSubmission {
  id: number
  quest_id: number
  challenge_id: number
  crew_id: string
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

export interface DailyTurn {
  id: number
  uid: string
  date: string
  quest_slug: string
  completed: boolean
  created_at: string
}

export type AchievementKind = 'PERSONAL' | 'GROUP'

export type AchievementTrigger =
  | 'QUEST_COMPLETED'
  | 'REALM_COMPLETED'
  | 'RELIC_COLLECTED'
  | 'DAILY_STREAK'
  | 'CREATIVE_SUBMISSION'
  | 'LEVEL_REACHED'

export interface Achievement {
  id: number
  uid: string
  crew_id?: string | null
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
  title: string
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
  realm: string
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

export interface ChapterProgressView {
  current_chapter?: ChapterSummary | null
  next_chapter?: ChapterSummary | null
  completed_chapters: ChapterSummary[]
  unlocked_chapters: ChapterSummary[]
  all_chapters: ChapterSummary[]
}

export interface LoreDefinition {
  id: number
  slug: string
  realm: string
  chapter: string
  title: string
  content: string
  order: number
  season_slug?: string | null
}

export interface LoreView {
  slug: string
  realm: string
  chapter: string
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
  realm: string
  created_at: string
  updated_at: string
}

export type SeasonState = 'UPCOMING' | 'ACTIVE' | 'EXPIRED' | 'INACTIVE'

export interface SeasonSummary {
  definition: SeasonDefinition
  state: SeasonState
}

export interface Relic {
  id: number
  uid: string
  code: string
  name: string
  description: string
  realm: string
  rarity: string
  image: string
  lore: string
  awarded_at: string
  created_at: string
}

export interface Chest {
  id: number
  uid: string
  source: string
  chest_slug: string
  rarity: string
  icon: string
  description: string
  drop_table: string
  opened: boolean
  opened_at?: string | null
  created_at: string
}

export interface PlayerRelic {
  relic_id: number
  relic_slug: string
  name: string
  description: string
  realm: string
  rarity: string
  image: string
  lore: string
  owned_count: number
  is_new: boolean
  discovered_at: string
  created_at: string
}

export interface ChestView {
  id: number
  uid: string
  source: string
  chest_slug: string
  name: string
  rarity: string
  icon: string
  description: string
  opened: boolean
  opened_at?: string | null
  created_at: string
}

export interface RewardItem {
  relic_slug: string
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
  realm: string
  rarity: string
  image: string
  lore: string
}

export interface InventoryItem {
  relic_id: number
  relic_slug: string
  name: string
  description: string
  realm: string
  rarity: string
  image: string
  lore: string
  owned_count: number
  is_new: boolean
  discovered_at: string
  created_at: string
}

export interface QuestView {
  id: number
  crew_id: string
  template_slug: string
  title: string
  status: QuestStatus
  started_at?: string | null
  completed_at?: string | null
  created_at: string
  challenge_count: number
  completed_count: number
}

export interface QuestWithChallenges {
  id: number
  crew_id: string
  template_slug: string
  title: string
  status: QuestStatus
  started_at?: string | null
  completed_at?: string | null
  created_at: string
  challenges: Challenge[]
}

export interface DailyTurnView {
  today: string
  completed: boolean
  available: boolean
  streak_days: number
  remaining_turns: number
  quest_slug?: string
}

export interface HomeResponse {
  player: Explorer
  quests: QuestView[]
  daily_turn: DailyTurnView
  realm_progress: RealmProgress[]
  relic_count: number
  active_quests: QuestView[]
  completed_quests_today: QuestView[]
  pending_creative_review: number
  last_submission?: CreativeSubmission | null
  sections: HomeSections
  available_chests: ChestView[]
  latest_relic?: InventoryItem | null
  collection_progress: { collected: number; total: number }
  chapter_progress?: ChapterProgressView | null
  lore_summary?: LoreSummary | null
  achievements?: AchievementView[] | null
}

export interface HomeSections {
  player: PlayerSection
  quests: QuestsSection
  daily_turn: DailyTurnSection
  realm: RealmSection
  world: WorldSection
  creative: CreativeSection
  chests: ChestsSection
  relics: RelicsSection
  lore: LoreSection
  achievements: AchievementsSection
}

export interface PlayerSection {
  uid: string
  crew_id: string
  explorer_name: string
  role: Role
  level: number
  xp: number
  created_at: string
  updated_at: string
  xp_to_next: number
}

export interface QuestsSection {
  all: QuestView[]
  active: QuestView[]
  done: QuestView[]
  done_today: QuestView[]
}

export interface DailyTurnSection {
  today: string
  completed: boolean
  available: boolean
  streak_days: number
  remaining_turns: number
  quest_slug?: string
}

export interface RealmSection {
  progress: RealmProgress[]
}

export interface WorldSection {
  current_chapter?: ChapterSummary | null
  next_chapter?: ChapterSummary | null
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

export interface ApiError {
  error: string
}
