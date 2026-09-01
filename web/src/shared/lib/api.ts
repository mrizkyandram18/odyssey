import type {
  ApiError,
  Family,
  Explorer,
  TaskView,
  SubmitTaskResponse,
  RewardCatalogItem,
  ClaimView,
  PendingSubmissionView,
  RedemptionConfig,
  MemberView,
  CreateMemberInput,
  UpdateMemberInput,
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
      // Best-effort
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
    if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
      if (path.startsWith('/api/push')) {
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
      throw new Error(err.error || 'request failed')
    }

    if (resp.status === 204) {
      return null as T
    }

    // Guard against empty body (e.g. Supabase PATCH without Prefer: return=representation)
    const text = await resp.text()
    if (!text || text.trim() === '' || text.trim() === '[]' || text.trim() === 'null') {
      return null as T
    }
    return JSON.parse(text) as T
  }
}

export const apiClient = new ApiClient()

export const crewsApi = {
  get: () => apiClient.get<Family>('/api/families'),
  patch: (body: { banner_url?: string; theme?: string }) =>
    apiClient.patch<Family>('/api/families', body),
  members: () => apiClient.get<Explorer[]>('/api/families/members').then(d => d || []),
}

export interface PushSubscribePayload {
  endpoint: string
  keys: {
    p256dh: string
    auth: string
  }
}

export const pushApi = {
  subscribe: (payload: PushSubscribePayload) => apiClient.post<{ status: string }>('/api/push', payload),
  unsubscribe: (endpoint?: string) => apiClient.delete<{ status: string }>('/api/push' + (endpoint ? '?endpoint=' + encodeURIComponent(endpoint) : ''), endpoint ? { endpoint } : undefined),
  delete: () => apiClient.delete<{ success: boolean }>('/api/push'),
}

export const profileApi = {
  getProfile: () => apiClient.get<Explorer>('/api/me'),
  updateAvatar: (data: { avatar_style: string; avatar_seed: string }) => apiClient.patch<{ status: string }>('/api/me/avatar', data),
  changePassword: (newPassword: string, currentPassword?: string, confirmPassword?: string) =>
    apiClient.post<{ status: string; message: string }>('/api/me/change-password', {
      new_password: newPassword,
      ...(currentPassword !== undefined ? { current_password: currentPassword } : {}),
      ...(confirmPassword !== undefined ? { confirm_password: confirmPassword } : {}),
    }),
  changePasswordFull: (data: { current_password?: string; new_password: string; confirm_password?: string }) =>
    apiClient.post<{ status: string; message: string }>('/api/me/change-password', data),
}

export const tasksApi = {
  getToday: () => apiClient.get<{ tasks: TaskView[] }>('/api/tasks/today'),
  getTask: (taskId: number) => apiClient.get<TaskView>(`/api/tasks/${taskId}`),
  submit: (taskId: number, data: { submission_type?: string; answers?: Record<string, any>; payload?: Record<string, any> }) =>
    apiClient.post<SubmitTaskResponse>(`/api/tasks/${taskId}/submit`, data),
}

export const shopApi = {
  getConfig: () => apiClient.get<RedemptionConfig>('/api/shop/config'),
  getCatalog: () => apiClient.get<RewardCatalogItem[]>('/api/shop/items'),
  redeem: (data: { reward_id?: number; coins: number; target_type: string; target_value: string }) =>
    apiClient.post<{ success: boolean; claim_id: number; new_balance: number }>('/api/shop/redeem', data),
  getMyClaims: () => apiClient.get<ClaimView[]>('/api/shop/claims'),
}

export const adminMembersApi = {
  getMembers: () => apiClient.get<MemberView[]>('/api/admin/members'),
  createMember: (data: CreateMemberInput) => apiClient.post<MemberView>('/api/admin/members', data),
  updateMember: (uid: string, patch: UpdateMemberInput) => apiClient.patch<MemberView>(`/api/admin/members/${uid}`, patch),
}

export const adminTasksApi = {
  getConfig: () => apiClient.get<RedemptionConfig>('/api/admin/config'),
  updateConfig: (data: { start_day?: number; end_day?: number; payout_day?: number; earning_period_days?: number; conversion_rate?: number; payout_target_rupiah?: number; payout_target_coins?: number; max_payout_coins?: number; timezone?: string }) =>
    apiClient.post<RedemptionConfig>('/api/admin/config', data),
  getTasks: (date?: string) => apiClient.get<TaskView[]>(`/api/admin/tasks${date ? '?date=' + date : ''}`),
  createTask: (data: any) => apiClient.post<TaskView>('/api/admin/tasks', data),
  updateTask: (id: number, patch: any) => apiClient.patch<any>(`/api/admin/tasks/${id}`, patch),
  duplicateTask: (id: number) => apiClient.post<TaskView>(`/api/admin/tasks/${id}/duplicate`, {}),
  deleteTask: (id: number) => apiClient.delete<{ status: string }>(`/api/admin/tasks/${id}`),
  getSubmissions: (status?: string) =>
    apiClient.get<PendingSubmissionView[]>(`/api/admin/submissions${status ? '?status=' + status : ''}`),
  getPendingSubmissions: (status?: string) =>
    apiClient.get<PendingSubmissionView[]>(`/api/admin/submissions${status ? '?status=' + status : ''}`),
  verifySubmission: (id: number, status: 'APPROVED' | 'REJECTED', notes?: string, penaltyCoins?: number) =>
    apiClient.post<{ success: boolean; status: string; coins_earned?: number; coins_deducted?: number }>(
      `/api/admin/submissions/${id}/verify`,
      { status, notes, penalty_coins: penaltyCoins }
    ),
  editSubmission: (id: number, payload: Record<string, any>, notes?: string) =>
    apiClient.patch<{ success: boolean; submission_id: number; status: string; payload: Record<string, any> }>(
      `/api/admin/submissions/${id}`,
      { payload, notes }
    ),
  getClaims: (status?: string) => apiClient.get<ClaimView[]>(`/api/admin/claims${status ? '?status=' + status : ''}`),
  processClaim: (id: number, status: 'APPROVED' | 'REJECTED', notes?: string) =>
    apiClient.post<{ success: boolean; status: string }>(`/api/admin/claims/${id}/process`, { status, notes }),
}
