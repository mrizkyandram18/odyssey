// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent, cleanup } from '@testing-library/react'
import { CosmeticsCatalogSection } from './CosmeticsCatalogSection'
import { adminCosmeticsApi } from '../../../../shared/lib/api'

vi.mock('../../../../shared/lib/api', () => ({
  adminCosmeticsApi: {
    getCosmetics: vi.fn(),
    createCosmetic: vi.fn(),
    updateCosmetic: vi.fn(),
  },
}))

describe('CosmeticsCatalogSection', () => {
  afterEach(() => {
    cleanup()
  })
  const mockCosmetics = [
    {
      id: 'frame-gold',
      name: 'Bingkai Emas',
      slot: 'frame' as const,
      asset: 'gold',
      tier: 1,
      is_active: true,
    },
    {
      id: 'effect-sparkle',
      name: 'Efek Kilau',
      slot: 'effect' as const,
      asset: 'sparkle',
      tier: 1,
      is_active: false,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(adminCosmeticsApi.getCosmetics).mockResolvedValue({ items: mockCosmetics })
    vi.mocked(adminCosmeticsApi.updateCosmetic).mockResolvedValue({ status: 'ok' })
    vi.mocked(adminCosmeticsApi.createCosmetic).mockResolvedValue({
      id: 'frame-neon',
      name: 'Bingkai Neon',
      slot: 'frame',
      asset: 'neon',
      tier: 1,
      is_active: true,
    })
  })

  it('renders cosmetic items with active and inactive statuses', async () => {
    render(<CosmeticsCatalogSection />)

    await waitFor(() => {
      expect(screen.getByText('Bingkai Emas')).toBeInTheDocument()
      expect(screen.getByText('Efek Kilau')).toBeInTheDocument()
      expect(screen.getByText('● Aktif di Gacha Box')).toBeInTheDocument()
      expect(screen.getByText('○ Dinonaktifkan')).toBeInTheDocument()
    })
  })

  it('toggles cosmetic active status when toggle button clicked', async () => {
    render(<CosmeticsCatalogSection />)

    await waitFor(() => {
      expect(screen.getByText('Bingkai Emas')).toBeInTheDocument()
    })

    const toggleButtons = screen.getAllByRole('button', { name: /nonaktifkan kosmetik/i })
    fireEvent.click(toggleButtons[0])

    await waitFor(() => {
      expect(adminCosmeticsApi.updateCosmetic).toHaveBeenCalledWith('frame-gold', {
        is_active: false,
      })
    })
  })

  it('opens create modal, submits new cosmetic, and triggers reload', async () => {
    render(<CosmeticsCatalogSection />)

    await waitFor(() => {
      expect(screen.getByText('Katalog Kosmetik Hadiah')).toBeInTheDocument()
    })

    const addBtn = screen.getByRole('button', { name: /Tambah Kosmetik/i })
    fireEvent.click(addBtn)

    expect(screen.getByText('Tambah Kosmetik Baru')).toBeInTheDocument()

    const idInput = screen.getByPlaceholderText(/contoh: frame-emerald/i)
    const nameInput = screen.getByPlaceholderText(/contoh: Bingkai Zamrud/i)
    const assetInput = screen.getByPlaceholderText(/contoh: emerald \/ sparkle/i)

    fireEvent.change(idInput, { target: { value: 'frame-neon' } })
    fireEvent.change(nameInput, { target: { value: 'Bingkai Neon' } })
    fireEvent.change(assetInput, { target: { value: 'neon' } })

    const submitBtn = screen.getByRole('button', { name: /Simpan Kosmetik/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(adminCosmeticsApi.createCosmetic).toHaveBeenCalledWith({
        id: 'frame-neon',
        name: 'Bingkai Neon',
        slot: 'frame',
        asset: 'neon',
        tier: 1,
        is_active: true,
      })
    })
  })
})
