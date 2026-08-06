import type { Session, DevicePayload, LoginResponse, Explorer } from '../types'
import type { ApiClient } from './api'
import { clearSession } from './session'

export type LoginOutcome =
    | { ok: true; session: Session }
    | { ok: false; variant: 'setup_needed'; setupToken: string }
    | { ok: false; variant: 'password_required' }

export interface AuthClient {
    login(uid: string, credential: string, device: DevicePayload): Promise<LoginOutcome>
    current(): Promise<Explorer | null>
    logout(): void
}

export class HttpAuthClient {
    private client: ApiClient

    constructor(client: ApiClient) {
        this.client = client
    }

    async login(uid: string, credential: string, device: DevicePayload): Promise<LoginOutcome> {
        const response = await this.client.request<LoginResponse>('/api/login', {
            method: 'POST',
            body: JSON.stringify({ uid, credential, device }),
        })

        if (response.status === 'success' && response.session && response.uid) {
            return {
                ok: true,
                session: {
                    uid: response.uid,
                    crew_id: response.crew_id || '',
                    kind: response.kind || 'user',
                    role: response.role,
                    expires: response.expires || 0,
                    token: response.session,
                },
            }
        }

        if (response.status === 'setup_needed' && response.setup_token) {
            return { ok: false, variant: 'setup_needed', setupToken: response.setup_token }
        }

        if (response.status === 'password_required') {
            return { ok: false, variant: 'password_required' }
        }

        throw new Error(response.message || 'Login failed')
    }

    async current(): Promise<Explorer | null> {
        try {
            return await this.client.request<Explorer>('/api/me', {
                method: 'GET',
            })
        } catch {
            return null
        }
    }

    logout(): void {
        clearSession()
    }
}
