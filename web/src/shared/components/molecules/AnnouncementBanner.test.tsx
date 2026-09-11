// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { AnnouncementBanner } from './AnnouncementBanner'

function stubStorage() {
  const store = new Map<string, string>()
  const storage = {
    getItem: (k: string) => (store.has(k) ? store.get(k)! : null),
    setItem: (k: string, v: string) => {
      store.set(k, v)
    },
    removeItem: (k: string) => {
      store.delete(k)
    },
    clear: () => store.clear(),
  }
  vi.stubGlobal('localStorage', storage)
}

beforeEach(() => {
  stubStorage()
})

afterEach(() => {
  cleanup()
})

describe('AnnouncementBanner', () => {
  it('renders title and body from backend config', () => {
    render(
      <AnnouncementBanner
        announcement={{
          enabled: true,
          title: 'Judul dari Admin',
          body: 'Isi dari Admin.',
          audience: 'ALL',
          priority: 'normal',
          visible: true,
        }}
      />
    )
    expect(screen.getByTestId('announcement-banner')).toBeTruthy()
    expect(screen.getByTestId('announcement-title').textContent).toBe('Judul dari Admin')
    expect(screen.getByTestId('announcement-body').textContent).toBe('Isi dari Admin.')
  })

  it('renders nothing when not visible', () => {
    const { container } = render(
      <AnnouncementBanner
        announcement={{
          enabled: false,
          title: 'Judul',
          body: 'Isi',
          audience: 'ALL',
          priority: 'normal',
          visible: false,
        }}
      />
    )
    expect(container.innerHTML).toBe('')
  })

  it('renders nothing for null config', () => {
    const { container } = render(<AnnouncementBanner announcement={null} />)
    expect(container.innerHTML).toBe('')
  })

  it('supports dismiss behavior via localStorage', () => {
    const cfg = {
      enabled: true,
      title: 'Judul Dismiss',
      body: 'Isi Dismiss.',
      audience: 'ALL',
      priority: 'normal',
      visible: true,
    }
    const { unmount } = render(<AnnouncementBanner announcement={cfg} />)
    fireEvent.click(screen.getByTestId('announcement-dismiss'))
    expect(screen.queryByTestId('announcement-banner')).toBeNull()
    unmount()
    // Re-mount with same content stays dismissed via storage.
    const second = render(<AnnouncementBanner announcement={cfg} />)
    expect(second.container.innerHTML).toBe('')
    second.unmount()
  })
})
