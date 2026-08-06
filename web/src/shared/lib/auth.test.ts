import { describe, it, expect, vi, beforeEach } from 'vitest'
import { HttpAuthClient } from './auth'
import type { ApiClient } from './api'
import type { LoginResponse, Explorer } from '../types'

const localStorageMock = (() => {
	let store: Record<string, string> = {}
	return {
		getItem: (key: string): string | null => store[key] || null,
		setItem: (key: string, value: string): void => {
			store[key] = value
		},
		removeItem: (key: string): void => {
			delete store[key]
		},
		clear: (): void => {
			store = {}
		},
	}
})()

beforeEach(() => {
	localStorageMock.clear()
	vi.stubGlobal('localStorage', localStorageMock)
})

function createMockClient() {
	const request = vi.fn()
	const client = { request } as unknown as ApiClient
	return { client, request }
}

describe('HttpAuthClient.current', () => {
	it('returns the profile on success', async () => {
		const profile: Explorer = {
			uid: 'user-1',
			crew_id: 'crew-1',
			explorer_name: 'Alice',
			role: 'SEEKER',
			level: 1,
			xp: 0,
			created_at: '2026-08-03T00:00:00Z',
			updated_at: '2026-08-03T00:00:00Z',
		}
		const { client, request } = createMockClient()
		request.mockResolvedValue(profile)

		const auth = new HttpAuthClient(client)
		const result = await auth.current()

		expect(result).toEqual(profile)
		expect(request).toHaveBeenCalledWith('/api/me', { method: 'GET' })
	})

	it('returns null on HTTP error', async () => {
		const { client, request } = createMockClient()
		request.mockRejectedValue(new Error('unauthorized'))

		const auth = new HttpAuthClient(client)
		const result = await auth.current()

		expect(result).toBeNull()
	})

	it('returns null on network error', async () => {
		const { client, request } = createMockClient()
		request.mockRejectedValue(new Error('network error'))

		const auth = new HttpAuthClient(client)
		const result = await auth.current()

		expect(result).toBeNull()
	})
})

describe('HttpAuthClient.login', () => {
	it('returns ok session on success', async () => {
		const { client, request } = createMockClient()
		const response: LoginResponse = {
			status: 'success',
			session: 'token-abc',
			uid: 'user-1',
			crew_id: 'crew-1',
			kind: 'user',
			role: 'SEEKER',
			expires: 9999999999,
		}
		request.mockResolvedValue(response)

		const auth = new HttpAuthClient(client)
		const result = await auth.login('user-1', 'secret', { device_id: 'web-pwa', login_method: 'BOTH' })

		expect(result.ok).toBe(true)
		if (result.ok) {
			expect(result.session.uid).toBe('user-1')
			expect(result.session.token).toBe('token-abc')
			expect(result.session.crew_id).toBe('crew-1')
			expect(result.session.kind).toBe('user')
			expect(result.session.role).toBe('SEEKER')
			expect(result.session.expires).toBe(9999999999)
		}
	})

	it('returns password_required variant', async () => {
		const { client, request } = createMockClient()
		const response: LoginResponse = {
			status: 'password_required',
			uid: 'user-1',
			message: 'Password is required.',
		}
		request.mockResolvedValue(response)

		const auth = new HttpAuthClient(client)
		const result = await auth.login('user-1', '', { device_id: 'web-pwa', login_method: 'BOTH' })

		expect(result.ok).toBe(false)
		if (!result.ok) {
			expect(result.variant).toBe('password_required')
		}
	})

	it('returns setup_needed variant with setup token', async () => {
		const { client, request } = createMockClient()
		const response: LoginResponse = {
			status: 'setup_needed',
			setup_token: 'setup-token',
			uid: 'user-1',
			kind: 'setup',
			expires: 9999999999,
		}
		request.mockResolvedValue(response)

		const auth = new HttpAuthClient(client)
		const result = await auth.login('user-1', 'secret', { device_id: 'web-pwa', login_method: 'BOTH' })

		expect(result.ok).toBe(false)
		if (!result.ok && result.variant === 'setup_needed') {
			expect(result.setupToken).toBe('setup-token')
		}
	})

	it('throws on unknown response status', async () => {
		const { client, request } = createMockClient()
		const response: LoginResponse = {
			status: 'success',
			message: 'something went wrong',
		}
		request.mockResolvedValue(response)

		const auth = new HttpAuthClient(client)
		await expect(
			auth.login('user-1', 'secret', { device_id: 'web-pwa', login_method: 'BOTH' }),
		).rejects.toThrow('something went wrong')
	})
})

describe('HttpAuthClient.logout', () => {
	it('clears session without error', () => {
		const { client } = createMockClient()
		const auth = new HttpAuthClient(client)
		expect(() => auth.logout()).not.toThrow()
	})
})
