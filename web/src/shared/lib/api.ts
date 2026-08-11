import type {
  ApiError,
  ChestView,
  OpenResult,
  InventoryItem,
  RelicDefinition,
  QuestView,
  QuestWithChallenges,
  CompleteChallengeResult,
  AchievementView,
  LoreView,
} from '../types'
import { getSession, isSessionExpired } from './session'

const API_BASE = import.meta.env.DEV ? '' : ''

export class ApiClient {
  private baseURL: string
  private csrfToken: string | null = null

  constructor(baseURL: string = API_BASE) {
    this.baseURL = baseURL
  }

  async get<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: 'GET' })
  }

  async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'POST',
      body: JSON.stringify(body),
    })
  }

  async patch<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'PATCH',
      body: JSON.stringify(body),
    })
  }

  async delete<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'DELETE',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  /** Fetch CSRF token once per session for state-changing creative routes. */
  private async ensureCsrfToken(): Promise<string | null> {
    if (this.csrfToken && this.csrfToken.length >= 32) {
      return this.csrfToken
    }
    try {
      const resp = await fetch(this.baseURL + '/api/csrf', {
        method: 'GET',
        credentials: 'include',
      })
      if (!resp.ok) return null
      const data = (await resp.json()) as { csrf_token?: string }
      if (data.csrf_token && data.csrf_token.length >= 32) {
        this.csrfToken = data.csrf_token
        return this.csrfToken
      }
    } catch {
      // Best-effort: server may still accept cookie-based CSRF if set.
    }
    return this.csrfToken
  }

  async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const session = getSession()
    const isAuthenticated = session !== null && !isSessionExpired(session)
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (isAuthenticated && session) {
      headers.Authorization = `Bearer ${session.token}`
      headers['X-User-Session'] = session.token
    }

    const method = (options.method || 'GET').toUpperCase()
    // Creative write routes require X-CSRF-Token (see pkg/server CSRF middleware).
    if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
      if (path.startsWith('/api/creative') || path.startsWith('/api/board') || path.startsWith('/api/push')) {
        const csrf = await this.ensureCsrfToken()
        if (csrf) {
          headers['X-CSRF-Token'] = csrf
        }
      }
    }

    const existingHeaders = (options.headers as Record<string, string> | undefined) ?? {}
    const mergedHeaders = { ...headers, ...existingHeaders }

    const resp = await fetch(this.baseURL + path, {
      ...options,
      credentials: 'include',
      headers: mergedHeaders,
    })

    if (!resp.ok) {
      let err: ApiError
      try {
        err = await resp.json()
      } catch {
        err = { error: resp.statusText }
      }
      throw new Error(err.error || 'Request failed')
    }

    if (resp.status === 204) {
      return null as T
    }

    return resp.json()
  }
}

export const apiClient = new ApiClient()

export const chestsApi = {
  list: () => apiClient.get<ChestView[]>('/api/chests'),
  get: (id: number) => apiClient.get<ChestView>(`/api/chests/${id}`),
  open: (id: number) => apiClient.post<OpenResult>(`/api/chests/${id}/open`, {}),
}

export const relicsApi = {
  list: () => apiClient.get<InventoryItem[]>('/api/relics'),
  get: (slug: string) => apiClient.get<RelicDefinition>(`/api/relics/${slug}`),
  inventory: () => apiClient.get<InventoryItem[]>('/api/relics/inventory'),
}

export const questsApi = {
  list: () => apiClient.get<QuestView[]>('/api/quests'),
  get: (id: number) => apiClient.get<QuestWithChallenges>(`/api/quests/${id}`),
  start: (id: number) => apiClient.post<{ started: boolean }>(`/api/quests/${id}/start`, {}),
  completeChallenge: (questId: number, challengeId: number) =>
    apiClient.post<CompleteChallengeResult>(`/api/quests/${questId}/challenges/${challengeId}/complete`, {}),
}

export const achievementsApi = {
  list: () => apiClient.get<AchievementView[]>('/api/achievements'),
}

export const loreApi = {
  list: () => apiClient.get<LoreView[]>('/api/lore'),
}

export type ReactionType = 'HEART' | 'CLAP' | 'STAR'
export type TargetType = 'JOURNAL' | 'QUEST' | 'TEXT_BOARD'

export interface BoardPost {
  id: number
  crew_id: string
  realm: string
  author_uid: string
  kind: string
  payload: string
  created_at: string
}

export const boardApi = {
  list: (): Promise<{ posts: BoardPost[] }> =>
    apiClient.get<{ posts: BoardPost[] }>('/api/board'),
  post: (content: string): Promise<BoardPost> =>
    apiClient.post<BoardPost>('/api/board', { content }),
}

export interface ReactionRow {
  id: number
  crew_id: string
  target_type: TargetType
  target_id: number
  actor_uid: string
  reaction_type: ReactionType
  created_at: string
}

export interface ReactionsResponse {
  reactions: ReactionRow[]
}

/** Derived state computed client-side from the raw API response.
 *  The backend returns full rows including actor_uid. We derive counts
 *  and myReaction by matching actor_uid == session UID.
 *  We NEVER send actor_uid to the backend — only receive it here for display.
 */
export interface ReactionState {
  counts: Record<ReactionType, number>
  myReaction: ReactionType | null
}

export function deriveReactionState(rows: ReactionRow[], myUID: string): ReactionState {
  const counts: Record<ReactionType, number> = { HEART: 0, CLAP: 0, STAR: 0 }
  let myReaction: ReactionType | null = null
  for (const r of rows) {
    if (r.reaction_type in counts) {
      counts[r.reaction_type as ReactionType]++
    }
    if (r.actor_uid === myUID) {
      myReaction = r.reaction_type as ReactionType
    }
  }
  return { counts, myReaction }
}

export const reactionsApi = {
  list: (targetType: TargetType, targetId: number): Promise<ReactionsResponse> =>
    apiClient.get<ReactionsResponse>(`/api/reactions?target_type=${targetType}&target_id=${targetId}`),
  upsert: (targetType: TargetType, targetId: number, reactionType: ReactionType): Promise<ReactionRow> =>
    apiClient.post<ReactionRow>('/api/reactions', {
      target_type: targetType,
      target_id: targetId,
      reaction_type: reactionType,
    }),
}

export interface PushSubscribePayload {
  endpoint: string
  keys: {
    p256dh: string
    auth: string
  }
}

export const pushApi = {
  subscribe: (payload: PushSubscribePayload) => apiClient.post<{ status: string }>('/api/push/subscribe', payload),
  unsubscribe: (endpoint?: string) => apiClient.delete<{ status: string }>('/api/push/subscribe', endpoint ? { endpoint } : undefined),
}
