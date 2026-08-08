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

    const existingHeaders = (options.headers as Record<string, string> | undefined) ?? {}
    const mergedHeaders = { ...headers, ...existingHeaders }

    const resp = await fetch(this.baseURL + path, {
      ...options,
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
