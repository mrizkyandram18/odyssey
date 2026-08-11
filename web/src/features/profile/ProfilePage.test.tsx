// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ProfilePage } from './ProfilePage'
import { useSession } from '../../shared/hooks/useSession'

const mockApiClientGet = vi.fn((url: string) => {
  if (url === '/api/cosmetics') {
    return Promise.resolve({
      coins: 0,
      items: [
        { id: 'explorer_effect_sparkle', name: 'Sparkle Effect', description: 'desc', price: 0, kind: 'explorer_effect', value: 'sparkle', unlocked: true },
      ],
      avatar_frame: 'none',
      explorer_effect: 'none',
    })
  }
  if (url === '/api/relics/inventory') {
    return Promise.resolve([])
  }
  if (url === '/api/quests') {
    return Promise.resolve([])
  }
  if (url === '/api/crews') {
    return Promise.resolve({ banner_url: '', theme: 'default' })
  }
  if (url === '/api/rewards') {
    return Promise.resolve([])
  }
  return Promise.resolve([])
})
const mockApiClientPost = vi.fn(() => Promise.resolve({}))
const mockApiClientPatch = vi.fn(() => Promise.resolve({}))

vi.mock('../../shared/hooks/useSession', () => ({
  useSession: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  apiClient: {
    get: (...args: any[]) => mockApiClientGet(...args),
    post: (...args: any[]) => mockApiClientPost(...args),
    patch: (...args: any[]) => mockApiClientPatch(...args),
  },
  crewsApi: {
    get: (...args: any[]) => mockApiClientGet(...args),
    patch: (...args: any[]) => mockApiClientPatch(...args),
  },
  questsApi: {
    list: () => Promise.resolve([]),
  },
}))

describe('ProfilePage crew customization', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mockApiClientGet.mockImplementation((url: string) => {
      if (url === '/api/relics/inventory') return Promise.resolve([])
      if (url === '/api/quests') return Promise.resolve([])
      if (url === '/api/crews') return Promise.resolve({ banner_url: '', theme: 'default' })
      if (url === '/api/rewards') return Promise.resolve([])
      return Promise.resolve([])
    })
    mockApiClientPost.mockResolvedValue({})
    mockApiClientPatch.mockResolvedValue({})

    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1', crew_id: 'crew-1', role: 'SEEKER' } as any,
      profile: {
        uid: 'u1',
        crew_id: 'crew-1',
        explorer_name: 'Tester',
        role: 'SEEKER',
        level: 1,
        xp: 0,
        coins: 0,
        avatar_style: 'adventurer',
        avatar_seed: 'seed',
        created_at: '',
        updated_at: '',
      } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })
  })

  afterEach(() => {
    cleanup()
  })

  it('renders crew customization section', async () => {
    mockApiClientGet.mockResolvedValueOnce([])

    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('crew-customization')).toBeInTheDocument()
    })
    expect(screen.getByTestId('banner-url-input')).toBeInTheDocument()
    expect(screen.getByTestId('theme-select')).toBeInTheDocument()
    expect(screen.getByTestId('save-crew-settings')).toBeInTheDocument()
  })

  it('saves crew settings on button click', async () => {
    mockApiClientPatch.mockResolvedValueOnce({
      id: 'crew-1',
      name: 'Test Crew',
      banner_url: 'https://example.com/banner.png',
      theme: 'forest',
      created_at: '',
      updated_at: '',
    })

    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('crew-customization')).toBeInTheDocument()
    })

    const bannerInput = screen.getByTestId('banner-url-input')
    const themeSelect = screen.getByTestId('theme-select')
    const saveBtn = screen.getByTestId('save-crew-settings')

    fireEvent.change(bannerInput, { target: { value: 'https://example.com/banner.png' } })
    fireEvent.change(themeSelect, { target: { value: 'forest' } })

    saveBtn.click()

    await waitFor(() => {
      expect(mockApiClientPatch).toHaveBeenCalledWith({
        banner_url: 'https://example.com/banner.png',
        theme: 'forest',
      })
    })

    await waitFor(() => {
      expect(screen.getByTestId('crew-save-msg')).toHaveTextContent('Crew customization saved!')
    })
  })
})

describe('ProfilePage explorer effects', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mockApiClientGet.mockImplementation((url: string) => {
      if (url === '/api/cosmetics') {
        return Promise.resolve({
          coins: 0,
          items: [
            { id: 'explorer_effect_sparkle', name: 'Sparkle Effect', description: 'desc', price: 0, kind: 'explorer_effect', value: 'sparkle', unlocked: true },
          ],
          avatar_frame: 'none',
          explorer_effect: 'none',
        })
      }
      if (url === '/api/relics/inventory') {
        return Promise.resolve([])
      }
      if (url === '/api/quests') {
        return Promise.resolve([])
      }
      if (url === '/api/crews') {
        return Promise.resolve({ banner_url: '', theme: 'default' })
      }
      if (url === '/api/rewards') {
        return Promise.resolve([])
      }
      return Promise.resolve([])
    })
    mockApiClientPost.mockResolvedValue({})
    mockApiClientPatch.mockResolvedValue({})

    vi.mocked(useSession).mockReturnValue({
      session: { uid: 'u1', crew_id: 'crew-1', role: 'SEEKER' } as any,
      profile: {
        uid: 'u1',
        crew_id: 'crew-1',
        explorer_name: 'Tester',
        role: 'SEEKER',
        level: 1,
        xp: 0,
        coins: 0,
        avatar_style: 'adventurer',
        avatar_seed: 'seed',
        equipped_explorer_effect: 'none',
        created_at: '',
        updated_at: '',
      } as any,
      loading: false,
      error: null,
      login: vi.fn(),
      logout: vi.fn(),
      refreshProfile: vi.fn(),
    })
  })

  afterEach(() => {
    cleanup()
  })

  it('renders effect items in the shop', async () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('cosmetic-item-explorer_effect_sparkle')).toBeInTheDocument()
    })
    expect(screen.getByText('Sparkle Effect')).toBeInTheDocument()
  })

  it('equips effect and updates state', async () => {
    mockApiClientPost.mockResolvedValueOnce({ status: 'equipped' })

    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByTestId('equip-explorer_effect_sparkle')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByTestId('equip-explorer_effect_sparkle'))

    await waitFor(() => {
      expect(screen.getByTestId('cosmetic-shop-msg')).toHaveTextContent('Effect equipped.')
    })
  })
})
