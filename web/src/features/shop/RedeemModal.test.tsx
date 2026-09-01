// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { RedeemModal } from './RedeemModal'
import { shopApi } from '../../shared/lib/api'

vi.mock('../../shared/lib/api', () => ({
  shopApi: {
    redeem: vi.fn(),
  },
}))

describe('RedeemModal Component', () => {
  const defaultProps = {
    userCoins: 1000,
    conversionRate: 100,
    onClose: vi.fn(),
    onSuccess: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders E-Wallet and Bank options with mandatory warning notice', () => {
    render(<RedeemModal {...defaultProps} />)

    expect(screen.getByText('E-Wallet')).toBeInTheDocument()
    expect(screen.getByText('Transfer Bank')).toBeInTheDocument()
    expect(
      screen.getByText(/Pastikan data tujuan pencairan sudah benar/)
    ).toBeInTheDocument()
    expect(screen.getByTestId('proceed-confirm-btn')).toBeInTheDocument()
  })

  it('validates minimum digits and format for E-Wallet phone number', async () => {
    render(<RedeemModal {...defaultProps} />)

    const input = screen.getByTestId('account-number-input')
    const proceedBtn = screen.getByTestId('proceed-confirm-btn')

    // Empty number submit
    fireEvent.change(input, { target: { value: '' } })
    fireEvent.click(proceedBtn)
    expect(
      await screen.findByText('Nomor HP / akun e-wallet tujuan wajib diisi')
    ).toBeInTheDocument()

    // Too short number (<9 digits for ewallet)
    fireEvent.change(input, { target: { value: '0812' } })
    fireEvent.click(proceedBtn)
    expect(
      await screen.findByText(/minimal 9 digit angka/)
    ).toBeInTheDocument()
  })

  it('allows switching to Bank and validates bank account number', async () => {
    render(<RedeemModal {...defaultProps} />)

    const bankBtn = screen.getByTestId('method-bank')
    fireEvent.click(bankBtn)

    expect(screen.getByText('Pilihan Bank:')).toBeInTheDocument()
    expect(screen.getByText('Nomor Rekening:')).toBeInTheDocument()

    const input = screen.getByTestId('account-number-input')
    const proceedBtn = screen.getByTestId('proceed-confirm-btn')

    fireEvent.change(input, { target: { value: '123' } })
    fireEvent.click(proceedBtn)
    expect(
      await screen.findByText('Nomor rekening minimal 5 digit angka')
    ).toBeInTheDocument()
  })

  it('transitions to confirmation step, displays masked number, nominal, warning, and submits on confirm', async () => {
    vi.mocked(shopApi.redeem).mockResolvedValue({
      success: true,
      claim_id: 88,
      new_balance: 500,
    } as any)

    render(<RedeemModal {...defaultProps} />)

    // Fill valid E-Wallet details
    const input = screen.getByTestId('account-number-input')
    fireEvent.change(input, { target: { value: '081234567890' } })

    const nameInput = screen.getByTestId('account-name-input')
    fireEvent.change(nameInput, { target: { value: 'Budi Santoso' } })

    const proceedBtn = screen.getByTestId('proceed-confirm-btn')
    fireEvent.click(proceedBtn)

    // Step 2: Confirmation review screen
    expect(await screen.findByText('Konfirmasi Pencairan Dana')).toBeInTheDocument()
    expect(screen.getByText('Ringkasan Pencairan')).toBeInTheDocument()
    expect(screen.getByText('0812****7890')).toBeInTheDocument()
    expect(screen.getByText('Budi Santoso')).toBeInTheDocument()
    expect(screen.getByText('1.000 Koin')).toBeInTheDocument()
    expect(screen.getByText('Rp 100.000')).toBeInTheDocument()
    expect(
      screen.getByText('PENTING — BACA SEBELUM KONFIRMASI')
    ).toBeInTheDocument()

    // Submit final confirmation
    const finalConfirmBtn = screen.getByTestId('final-confirm-btn')
    fireEvent.click(finalConfirmBtn)

    await waitFor(() => {
      expect(shopApi.redeem).toHaveBeenCalledWith({
        coins: 1000,
        target_type: 'EWALLET',
        target_value: 'GoPay - 081234567890 (a.n Budi Santoso)',
      })
    })

    // Step 3: Success screen
    expect(await screen.findByText('Pengajuan Berhasil Dikirim! 🎉')).toBeInTheDocument()
  })

  it('allows going back to modify data from confirmation step', async () => {
    render(<RedeemModal {...defaultProps} />)

    const input = screen.getByTestId('account-number-input')
    fireEvent.change(input, { target: { value: '081234567890' } })

    fireEvent.click(screen.getByTestId('proceed-confirm-btn'))
    expect(await screen.findByText('Konfirmasi Pencairan Dana')).toBeInTheDocument()

    const ubahBtn = screen.getByText('Ubah Data')
    fireEvent.click(ubahBtn)

    expect(await screen.findByText('Pencairan Koin ke Cash')).toBeInTheDocument()
    expect(screen.getByDisplayValue('081234567890')).toBeInTheDocument()
  })
})
