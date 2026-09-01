// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { ForceChangePasswordModal } from './ForceChangePasswordModal'
import { useSessionContext } from '../../app/SessionProvider'
import { profileApi } from '../../shared/lib/api'

vi.mock('../../app/SessionProvider', () => ({
  useSessionContext: vi.fn(),
}))

vi.mock('../../shared/lib/api', () => ({
  profileApi: {
    changePasswordFull: vi.fn(),
  },
}))

describe('ForceChangePasswordModal', () => {
  const refreshProfileMock = vi.fn()

  beforeEach(() => {
    vi.resetAllMocks()
    refreshProfileMock.mockReset()
  })

  afterEach(() => {
    cleanup()
  })

  it('does not render if must_change_password is false', () => {
    vi.mocked(useSessionContext).mockReturnValue({
      profile: {
        uid: 'u1',
        role: 'MEMBER',
        must_change_password: false,
      } as any,
      refreshProfile: refreshProfileMock,
    } as any)

    const { container } = render(<ForceChangePasswordModal />)
    expect(container.firstChild).toBeNull()
  })

  it('renders modal if must_change_password is true and role is MEMBER', () => {
    vi.mocked(useSessionContext).mockReturnValue({
      profile: {
        uid: 'u1',
        role: 'MEMBER',
        must_change_password: true,
      } as any,
      refreshProfile: refreshProfileMock,
    } as any)

    render(<ForceChangePasswordModal />)
    expect(screen.getByText('Ubah Kata Sandi')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Minimal 6 karakter')).toBeInTheDocument()
  })

  it('validates password length and confirmation match', async () => {
    vi.mocked(useSessionContext).mockReturnValue({
      profile: {
        uid: 'u1',
        role: 'MEMBER',
        must_change_password: true,
      } as any,
      refreshProfile: refreshProfileMock,
    } as any)

    render(<ForceChangePasswordModal />)

    const newPassInput = screen.getByPlaceholderText('Minimal 6 karakter')
    const confirmInput = screen.getByPlaceholderText('Ulangi kata sandi baru')
    const submitBtn = screen.getByText('Simpan Kata Sandi')

    // Test < 6 chars
    fireEvent.change(newPassInput, { target: { value: '123' } })
    fireEvent.change(confirmInput, { target: { value: '123' } })
    fireEvent.click(submitBtn)

    expect(await screen.findByText('Kata sandi minimal 6 karakter')).toBeInTheDocument()
    expect(profileApi.changePasswordFull).not.toHaveBeenCalled()

    // Test mismatch
    fireEvent.change(newPassInput, { target: { value: 'newsecret123' } })
    fireEvent.change(confirmInput, { target: { value: 'different123' } })
    fireEvent.click(submitBtn)

    expect(await screen.findByText('Konfirmasi kata sandi tidak cocok')).toBeInTheDocument()
    expect(profileApi.changePasswordFull).not.toHaveBeenCalled()
  })

  it('submits successfully and calls refreshProfile', async () => {
    vi.mocked(profileApi.changePasswordFull).mockResolvedValueOnce({
      status: 'success',
      message: 'Kata sandi berhasil diubah',
    })

    vi.mocked(useSessionContext).mockReturnValue({
      profile: {
        uid: 'u1',
        role: 'MEMBER',
        must_change_password: true,
      } as any,
      refreshProfile: refreshProfileMock,
    } as any)

    render(<ForceChangePasswordModal />)

    const newPassInput = screen.getByPlaceholderText('Minimal 6 karakter')
    const confirmInput = screen.getByPlaceholderText('Ulangi kata sandi baru')
    const submitBtn = screen.getByText('Simpan Kata Sandi')

    fireEvent.change(newPassInput, { target: { value: 'myNewPass123' } })
    fireEvent.change(confirmInput, { target: { value: 'myNewPass123' } })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(profileApi.changePasswordFull).toHaveBeenCalledWith({
        new_password: 'myNewPass123',
        confirm_password: 'myNewPass123',
      })
      expect(refreshProfileMock).toHaveBeenCalled()
    })
  })
})
