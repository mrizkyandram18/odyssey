// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { TaskRevisionBanner } from './TaskRevisionBanner'

describe('TaskRevisionBanner component', () => {
  beforeEach(() => {
    cleanup()
    document.body.innerHTML = ''
  })

  afterEach(() => {
    cleanup()
    document.body.innerHTML = ''
  })

  it('renders "Perlu Diperbaiki" header and instruction', () => {
    render(<TaskRevisionBanner adminNotes="Perbaiki pencahayaan foto" />)
    expect(screen.getByText('Perlu Diperbaiki')).toBeInTheDocument()
    expect(screen.getByText('Catatan dari admin:')).toBeInTheDocument()
    expect(screen.getByText(/Perbaiki pencahayaan foto/)).toBeInTheDocument()
    expect(screen.getByText('Silakan perbaiki bukti lalu kirim ulang agar dapat disetujui admin.')).toBeInTheDocument()
  })

  it('renders gracefully when adminNotes is null', () => {
    render(<TaskRevisionBanner adminNotes={null} />)
    expect(screen.getByText('Perlu Diperbaiki')).toBeInTheDocument()
    expect(screen.queryByText('Catatan dari admin:')).toBeNull()
    expect(screen.getByText('Silakan perbaiki bukti lalu kirim ulang agar dapat disetujui admin.')).toBeInTheDocument()
  })

  it('renders gracefully when adminNotes is whitespace only', () => {
    render(<TaskRevisionBanner adminNotes={`   
      `} />)
    expect(screen.getByText('Perlu Diperbaiki')).toBeInTheDocument()
    expect(screen.queryByText('Catatan dari admin:')).toBeNull()
    expect(screen.getByText('Silakan perbaiki bukti lalu kirim ulang agar dapat disetujui admin.')).toBeInTheDocument()
  })

  it('renders multiline admin note accurately', () => {
    const multiline = 'Baris 1: Foto kurang jelas.\nBaris 2: Wajah harus terlihat menghadap kamera.\nBaris 3: Jangan blur.'
    render(<TaskRevisionBanner adminNotes={multiline} />)
    expect(screen.getByText(/Baris 1: Foto kurang jelas/)).toBeInTheDocument()
    expect(screen.getByText(/Baris 2: Wajah harus terlihat menghadap kamera/)).toBeInTheDocument()
    expect(screen.getByText(/Baris 3: Jangan blur/)).toBeInTheDocument()
  })

  it('renders special characters, emojis, and quotes safely without breaking', () => {
    const special = 'Catatan: "Foto & CV <wajib> @2026 #100%!" — Tolong diulang ya 👍✨'
    render(<TaskRevisionBanner adminNotes={special} />)
    expect(screen.getByText(new RegExp(special.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))).toBeInTheDocument()
  })

  it('renders very long admin note safely with break-words container', () => {
    const longNote = 'A'.repeat(500) + ' ' + 'B'.repeat(500)
    render(<TaskRevisionBanner adminNotes={longNote} />)
    expect(screen.getByText(new RegExp('A{500}'))).toBeInTheDocument()
  })
})
