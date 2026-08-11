// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { PushNotificationToggle } from './PushNotificationToggle'
import * as pushHook from '../../hooks/usePushSubscription'

vi.mock('../../hooks/usePushSubscription')

describe('PushNotificationToggle', () => {
  const mockSubscribe = vi.fn()
  const mockUnsubscribe = vi.fn()
  const mockCheckSubscription = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders permission_required state with Enable button', () => {
    vi.spyOn(pushHook, 'usePushSubscription').mockReturnValue({
      status: 'permission_required',
      error: null,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      checkSubscription: mockCheckSubscription,
    })

    render(<PushNotificationToggle />)

    expect(screen.getByText(/Permission required/i)).toBeInTheDocument()
    const enableBtn = screen.getByTestId('enable-push-btn')
    expect(enableBtn).toBeInTheDocument()

    fireEvent.click(enableBtn)
    expect(mockSubscribe).toHaveBeenCalled()
  })

  it('renders enabled state with Disable button', () => {
    vi.spyOn(pushHook, 'usePushSubscription').mockReturnValue({
      status: 'enabled',
      error: null,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      checkSubscription: mockCheckSubscription,
    })

    render(<PushNotificationToggle />)

    expect(screen.getByTestId('push-status-enabled')).toBeInTheDocument()
    const disableBtn = screen.getByTestId('disable-push-btn')
    expect(disableBtn).toBeInTheDocument()

    fireEvent.click(disableBtn)
    expect(mockUnsubscribe).toHaveBeenCalled()
  })

  it('renders blocked state when permission is denied', () => {
    vi.spyOn(pushHook, 'usePushSubscription').mockReturnValue({
      status: 'blocked',
      error: null,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      checkSubscription: mockCheckSubscription,
    })

    render(<PushNotificationToggle />)

    expect(screen.getByTestId('push-status-blocked')).toBeInTheDocument()
  })

  it('renders unsupported state when Web Push is unavailable', () => {
    vi.spyOn(pushHook, 'usePushSubscription').mockReturnValue({
      status: 'unsupported',
      error: null,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      checkSubscription: mockCheckSubscription,
    })

    render(<PushNotificationToggle />)

    expect(screen.getByTestId('push-status-unsupported')).toBeInTheDocument()
  })
})
