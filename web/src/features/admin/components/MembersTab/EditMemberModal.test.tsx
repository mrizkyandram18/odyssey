// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { EditMemberModal } from './EditMemberModal'
import { adminMembersApi } from '../../../../shared/lib/api'

vi.mock('../../../../shared/lib/api', () => ({
  adminMembersApi: {
    getMembers: vi.fn(),
    createMember: vi.fn(),
    updateMember: vi.fn(),
    resetPassword: vi.fn(),
  },
  apiClient: { get: vi.fn(), post: vi.fn(), patch: vi.fn(), delete: vi.fn() },
}))

const member = {
  uid: 'usr_target',
  family_id: 'fam_1',
  explorer_name: 'Selvica Hyani',
  username: 'selvica',
  role: 'MEMBER' as const,
  is_active: true,
  level: 1,
  xp: 0,
  coins: 0,
  created_at: new Date().toISOString(),
}

function renderModal(overrides: Partial<React.ComponentProps<typeof EditMemberModal>> = {}) {
  const form = {
    explorer_name: 'Selvica Hyani',
    role: 'MEMBER' as const,
    is_active: true,
    reset_device: false,
  }
  const setForm = vi.fn()
  const onClose = vi.fn()
  const onSave = vi.fn()
  const props = {
    member,
    form,
    setForm,
    isSaving: false,
    onClose,
    onSave,
    ...overrides,
  } as React.ComponentProps<typeof EditMemberModal>
  const utils = render(<EditMemberModal {...props} />)
  return { ...utils, setForm, onClose, onSave }
}

describe('EditMemberModal - Reset Password', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // mock clipboard
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    })
  })

  afterEach(() => cleanup())

  it('Reset Password button exists', () => {
    renderModal()
    expect(screen.getByTestId('reset-password-button')).toBeInTheDocument()
    expect(screen.getByTestId('reset-password-button')).toHaveTextContent('Reset Password')
    expect(screen.getByText('Atur ulang password akun member jika lupa password.')).toBeInTheDocument()
  })

  it('Reset Password section is separate from Reset Device Binding', () => {
    renderModal()
    expect(screen.getByTestId('reset-password-button')).toBeInTheDocument()
    expect(screen.getByText('Reset Binding Perangkat')).toBeInTheDocument()
    // Ensure they are in different containers
    const resetPassBtn = screen.getByTestId('reset-password-button')
    const deviceCheckbox = document.getElementById('reset-device-checkbox')
    expect(resetPassBtn).toBeInTheDocument()
    expect(deviceCheckbox).toBeInTheDocument()
    expect(resetPassBtn.closest('div')?.textContent).not.toContain('Izinkan akun login di HP')
  })

  it('Confirmation dialog opens on Reset Password click', async () => {
    renderModal()
    fireEvent.click(screen.getByTestId('reset-password-button'))
    expect(await screen.findByText('Reset Password?')).toBeInTheDocument()
    expect(screen.getByText(/Password akun/)).toBeInTheDocument()
    expect(screen.getByText(/Selvica Hyani/)).toBeInTheDocument()
    expect(screen.getByTestId('reset-cancel-button')).toBeInTheDocument()
    expect(screen.getByTestId('reset-confirm-button')).toBeInTheDocument()
  })

  it('Cancel does not reset', async () => {
    renderModal()
    fireEvent.click(screen.getByTestId('reset-password-button'))
    await screen.findByText('Reset Password?')
    fireEvent.click(screen.getByTestId('reset-cancel-button'))
    await waitFor(() => {
      expect(screen.queryByText('Reset Password?')).not.toBeInTheDocument()
    })
    expect(adminMembersApi.resetPassword).not.toHaveBeenCalled()
  })

  it('Successful reset displays temporary password', async () => {
    vi.mocked(adminMembersApi.resetPassword).mockResolvedValueOnce({ temporary_password: 'X7mK9QaP2nB4!x' })
    renderModal()
    fireEvent.click(screen.getByTestId('reset-password-button'))
    await screen.findByText('Reset Password?')
    fireEvent.click(screen.getByTestId('reset-confirm-button'))
    await waitFor(() => {
      expect(screen.getByText('Password Berhasil Di-reset')).toBeInTheDocument()
    })
    expect(screen.getByTestId('temporary-password-display')).toHaveTextContent('X7mK9QaP2nB4!x')
    expect(adminMembersApi.resetPassword).toHaveBeenCalledWith('usr_target')
  })

  it('Copy button works and shows Tersalin', async () => {
    vi.mocked(adminMembersApi.resetPassword).mockResolvedValueOnce({ temporary_password: 'CopyPass123!@#' })
    renderModal()
    fireEvent.click(screen.getByTestId('reset-password-button'))
    await screen.findByText('Reset Password?')
    fireEvent.click(screen.getByTestId('reset-confirm-button'))
    await screen.findByText('Password Berhasil Di-reset')
    const copyBtn = screen.getByTestId('copy-password-button')
    fireEvent.click(copyBtn)
    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('CopyPass123!@#')
    })
    expect((await screen.findAllByText('Tersalin')).length).toBeGreaterThan(0)
  })

  it('Success state does not expose password after close', async () => {
    vi.mocked(adminMembersApi.resetPassword).mockResolvedValueOnce({ temporary_password: 'SecretTemp999!' })
    renderModal()
    fireEvent.click(screen.getByTestId('reset-password-button'))
    await screen.findByText('Reset Password?')
    fireEvent.click(screen.getByTestId('reset-confirm-button'))
    await screen.findByText('Password Berhasil Di-reset')
    expect(screen.getByTestId('temporary-password-display')).toBeInTheDocument()
    fireEvent.click(screen.getByTestId('close-success-button'))
    await waitFor(() => {
      expect(screen.queryByText('Password Berhasil Di-reset')).not.toBeInTheDocument()
    })
    expect(screen.queryByTestId('temporary-password-display')).not.toBeInTheDocument()
    expect(screen.queryByText('SecretTemp999!')).not.toBeInTheDocument()
  })

  it('Existing Edit Member workflow still works', () => {
    const onSave = vi.fn()
    const setForm = vi.fn()
    render(<EditMemberModal member={member} form={{ explorer_name: 'Selvica', role: 'MEMBER', is_active: true, reset_device: false }} setForm={setForm} isSaving={false} onClose={vi.fn()} onSave={onSave} />)
    const input = screen.getByDisplayValue('Selvica')
    fireEvent.change(input, { target: { value: 'Selvica Updated' } })
    expect(setForm).toHaveBeenCalled()
    // Save button
    fireEvent.click(screen.getByText('Simpan Perubahan'))
    expect(onSave).toHaveBeenCalled()
  })

  it('Device binding reset remains separate and functional', () => {
    const setForm = vi.fn()
    render(<EditMemberModal member={member} form={{ explorer_name: 'Selvica Hyani', role: 'MEMBER', is_active: true, reset_device: false }} setForm={setForm} isSaving={false} onClose={vi.fn()} onSave={vi.fn()} />)
    const checkbox = document.getElementById('reset-device-checkbox') as HTMLInputElement
    expect(checkbox).toBeInTheDocument()
    expect(checkbox.checked).toBe(false)
    fireEvent.click(checkbox)
    expect(setForm).toHaveBeenCalledWith(expect.objectContaining({ reset_device: true }))
    // Reset password button should still exist separately
    expect(screen.getByTestId('reset-password-button')).toBeInTheDocument()
  })

  it('does not use browser alert/confirm', async () => {
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
    const confirmSpy = vi.spyOn(window, 'confirm').mockImplementation(() => true)
    vi.mocked(adminMembersApi.resetPassword).mockResolvedValueOnce({ temporary_password: 'NoAlertPass123!' })
    renderModal()
    fireEvent.click(screen.getByTestId('reset-password-button'))
    await screen.findByText('Reset Password?')
    fireEvent.click(screen.getByTestId('reset-confirm-button'))
    await screen.findByText('Password Berhasil Di-reset')
    expect(alertSpy).not.toHaveBeenCalled()
    expect(confirmSpy).not.toHaveBeenCalled()
    alertSpy.mockRestore()
    confirmSpy.mockRestore()
  })
})
