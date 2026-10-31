// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup, act } from '@testing-library/react'
import { PWAUpdatePrompt } from './PWAUpdatePrompt'

describe('PWAUpdatePrompt (Phase 2A - F-03 / F-POST-02 / F-POST-04)', () => {
  const originalLocation = window.location
  let deleteCacheMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()

    // Mock caches API
    deleteCacheMock = vi.fn().mockResolvedValue(true)
    Object.assign(window, {
      caches: {
        keys: vi.fn().mockResolvedValue(['cache-v1', 'cache-v2']),
        delete: deleteCacheMock,
      },
    })

    // Mock window.location.reload
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        ...originalLocation,
        reload: vi.fn(),
      },
    })
  })

  afterEach(() => {
    cleanup()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    })
    vi.restoreAllMocks()
  })

  it('does not display prompt when no update is detected', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({ build_time: '' }),
    } as any)

    render(<PWAUpdatePrompt />)

    expect(screen.queryByTestId('pwa-update-prompt')).toBeNull()
    expect(window.location.reload).not.toHaveBeenCalled()
  })

  it('displays user-initiated prompt when update is detected without auto-reloading', async () => {
    let callCount = 0
    vi.spyOn(globalThis, 'fetch').mockImplementation(async () => {
      callCount++
      return {
        ok: true,
        json: async () => ({
          build_time: callCount === 1 ? '2026-10-31T01:00:00Z' : '2026-10-31T02:00:00Z',
        }),
      } as any
    })

    render(<PWAUpdatePrompt />)

    await waitFor(() => expect(callCount).toBeGreaterThanOrEqual(1))

    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      value: 'visible',
    })
    await act(async () => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await waitFor(() => {
      expect(screen.getByTestId('pwa-update-prompt')).toBeInTheDocument()
    })

    expect(screen.getByText('Pembaruan Sistem Tersedia')).toBeInTheDocument()
    expect(screen.getByTestId('pwa-update-button')).toBeInTheDocument()
    expect(screen.getByTestId('pwa-later-button')).toBeInTheDocument()

    // Crucial: Must NOT auto-reload after 800ms or 4000ms!
    await new Promise((r) => setTimeout(r, 1000))
    expect(window.location.reload).not.toHaveBeenCalled()
    expect(deleteCacheMock).not.toHaveBeenCalled()
  })

  it('clicking Nanti dismisses the prompt without reloading', async () => {
    let callCount = 0
    vi.spyOn(globalThis, 'fetch').mockImplementation(async () => {
      callCount++
      return {
        ok: true,
        json: async () => ({
          build_time: callCount === 1 ? 'build-1' : 'build-2',
        }),
      } as any
    })

    render(<PWAUpdatePrompt />)

    await waitFor(() => expect(callCount).toBeGreaterThanOrEqual(1))

    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      value: 'visible',
    })
    await act(async () => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await waitFor(() => {
      expect(screen.getByTestId('pwa-update-prompt')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByTestId('pwa-later-button'))

    expect(screen.queryByTestId('pwa-update-prompt')).toBeNull()
    expect(window.location.reload).not.toHaveBeenCalled()
  })

  it('clicking Perbarui reloads directly when no waiting service worker exists', async () => {
    let callCount = 0
    vi.spyOn(globalThis, 'fetch').mockImplementation(async () => {
      callCount++
      return {
        ok: true,
        json: async () => ({
          build_time: callCount === 1 ? 'build-1' : 'build-2',
        }),
      } as any
    })

    render(<PWAUpdatePrompt />)

    await waitFor(() => expect(callCount).toBeGreaterThanOrEqual(1))

    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      value: 'visible',
    })
    await act(async () => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await waitFor(() => {
      expect(screen.getByTestId('pwa-update-prompt')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByTestId('pwa-update-button'))

    await waitFor(() => {
      expect(window.location.reload).toHaveBeenCalledTimes(1)
    })
    expect(deleteCacheMock).not.toHaveBeenCalled()
  })

  it('clicking Perbarui sends SKIP_WAITING to waiting worker and reloads ONLY on controllerchange', async () => {
    let callCount = 0
    vi.spyOn(globalThis, 'fetch').mockImplementation(async () => {
      callCount++
      return {
        ok: true,
        json: async () => ({
          build_time: callCount === 1 ? 'build-1' : 'build-2',
        }),
      } as any
    })

    const postMessageMock = vi.fn()
    const waitingWorker = {
      postMessage: postMessageMock,
    }

    const swListeners: Record<string, ((e?: any) => void)[]> = {}
    const mockSWContainer = {
      ready: Promise.resolve({
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      }),
      getRegistration: vi.fn().mockResolvedValue({
        waiting: waitingWorker,
      }),
      addEventListener: vi.fn((event: string, handler: (e?: any) => void) => {
        swListeners[event] = swListeners[event] || []
        swListeners[event].push(handler)
      }),
      removeEventListener: vi.fn((event: string, handler: (e?: any) => void) => {
        if (swListeners[event]) {
          swListeners[event] = swListeners[event].filter((fn) => fn !== handler)
        }
      }),
      controller: {},
    }

    Object.defineProperty(navigator, 'serviceWorker', {
      configurable: true,
      value: mockSWContainer,
    })

    render(<PWAUpdatePrompt />)

    await waitFor(() => expect(callCount).toBeGreaterThanOrEqual(1))

    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      value: 'visible',
    })
    await act(async () => {
      document.dispatchEvent(new Event('visibilitychange'))
    })

    await waitFor(() => {
      expect(screen.getByTestId('pwa-update-prompt')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByTestId('pwa-update-button'))

    await waitFor(() => {
      expect(postMessageMock).toHaveBeenCalledWith({ type: 'SKIP_WAITING' })
    })

    // Before controllerchange fires, reload MUST NOT have happened!
    expect(window.location.reload).not.toHaveBeenCalled()

    // Trigger controllerchange
    act(() => {
      swListeners['controllerchange']?.forEach((fn) => fn())
    })

    expect(window.location.reload).toHaveBeenCalledTimes(1)
  })

  it('does NOT reload if controllerchange fails to fire within timeout, displays retry message (Mandatory Correction)', async () => {
    vi.useFakeTimers()
    try {
      let callCount = 0
      vi.spyOn(globalThis, 'fetch').mockImplementation(async () => {
        callCount++
        return {
          ok: true,
          json: async () => ({
            build_time: callCount === 1 ? 'build-1' : 'build-2',
          }),
        } as any
      })

      const postMessageMock = vi.fn()
      const waitingWorker = {
        postMessage: postMessageMock,
      }

      const swListeners: Record<string, ((e?: any) => void)[]> = {}
      const mockSWContainer = {
        ready: Promise.resolve({
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
        }),
        getRegistration: vi.fn().mockResolvedValue({
          waiting: waitingWorker,
        }),
        addEventListener: vi.fn((event: string, handler: (e?: any) => void) => {
          swListeners[event] = swListeners[event] || []
          swListeners[event].push(handler)
        }),
        removeEventListener: vi.fn((event: string, handler: (e?: any) => void) => {
          if (swListeners[event]) {
            swListeners[event] = swListeners[event].filter((fn) => fn !== handler)
          }
        }),
        controller: {},
      }

      Object.defineProperty(navigator, 'serviceWorker', {
        configurable: true,
        value: mockSWContainer,
      })

      render(<PWAUpdatePrompt />)

      // Allow fetch promise to resolve
      await act(async () => {
        await vi.advanceTimersByTimeAsync(10)
      })

      Object.defineProperty(document, 'visibilityState', {
        configurable: true,
        value: 'visible',
      })
      await act(async () => {
        document.dispatchEvent(new Event('visibilitychange'))
        await vi.advanceTimersByTimeAsync(10)
      })

      expect(screen.getByTestId('pwa-update-prompt')).toBeInTheDocument()

      // Click update
      await act(async () => {
        fireEvent.click(screen.getByTestId('pwa-update-button'))
      })

      expect(postMessageMock).toHaveBeenCalledWith({ type: 'SKIP_WAITING' })

      // Advance timers past 10 seconds without controllerchange firing
      await act(async () => {
        await vi.advanceTimersByTimeAsync(10005)
      })

      // Crucial: Must NEVER reload when controllerchange does not fire!
      expect(window.location.reload).not.toHaveBeenCalled()

      // UI must show failure/retry state
      expect(screen.getByTestId('pwa-error-message')).toHaveTextContent(
        /Pembaruan belum dapat diterapkan. Silakan coba lagi./i
      )
      expect(screen.getByTestId('pwa-update-button')).toHaveTextContent('Coba Lagi')
      expect(screen.getByTestId('pwa-update-button')).not.toBeDisabled()
    } finally {
      vi.useRealTimers()
    }
  })

  it('cleans up event listeners on unmount (F-POST-04)', async () => {
    const regRemoveEventListener = vi.fn()
    const mockReg = {
      addEventListener: vi.fn(),
      removeEventListener: regRemoveEventListener,
    }

    Object.defineProperty(navigator, 'serviceWorker', {
      configurable: true,
      value: {
        ready: Promise.resolve(mockReg),
        getRegistration: vi.fn().mockResolvedValue(null),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      },
    })

    const { unmount } = render(<PWAUpdatePrompt />)
    await act(async () => {
      await Promise.resolve()
    })

    unmount()

    expect(mockReg.removeEventListener).toHaveBeenCalledWith('updatefound', expect.any(Function))
  })
})
