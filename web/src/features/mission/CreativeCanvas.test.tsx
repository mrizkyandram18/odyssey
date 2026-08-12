// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { CreativeCanvas } from '../mission/CreativeCanvas'

vi.mock('react-sketch-canvas', () => ({
  ReactSketchCanvas: ({ ref, style, className }: any) => {
    React.useImperativeHandle(ref, () => ({
      exportSvg: vi.fn().mockResolvedValue('<svg xmlns="http://www.w3.org/2000/svg"><path d="M0 0"/></svg>'),
      clearCanvas: vi.fn(),
      undo: vi.fn(),
      eraseMode: vi.fn(),
    }))
    return <div data-testid="mock-canvas" className={className} style={style} />
  },
}))

describe('CreativeCanvas', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('renders without crashing', () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Submit Drawing')).toBeInTheDocument()
  })

  it('hides stamp picker when enableStamps is false', () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.queryByTitle('Star')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Heart')).not.toBeInTheDocument()
  })

  it('shows stamp picker when enableStamps is true', () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} enableStamps />)
    expect(screen.getByTitle('Star')).toBeInTheDocument()
    expect(screen.getByTitle('Heart')).toBeInTheDocument()
  })

  it('shows background colors when enableStamps is true', () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} enableStamps />)
    expect(screen.getByLabelText('Background white')).toBeInTheDocument()
    expect(screen.getByLabelText('Background blue')).toBeInTheDocument()
  })

  it('does not show background colors when enableStamps is false', () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.queryByLabelText('Background white')).not.toBeInTheDocument()
  })

  it('selects a stamp when clicked in picker', async () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} enableStamps />)
    const starButton = screen.getByTitle('Star')
    starButton.click()
    await waitFor(() => {
      expect(screen.getByText('Click on the canvas to place the stamp')).toBeInTheDocument()
    })
  })

  it('shows cancel button when stamp is active', async () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} enableStamps />)
    screen.getByTitle('Star').click()
    await waitFor(() => {
      expect(screen.getByText('Cancel stamp placement')).toBeInTheDocument()
    })
  })

  it('cancels stamp placement when cancel is clicked', async () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} enableStamps />)
    screen.getByTitle('Star').click()
    await waitFor(() => {
      expect(screen.getByText('Cancel stamp placement')).toBeInTheDocument()
    })
    screen.getByText('Cancel stamp placement').click()
    await waitFor(() => {
      expect(screen.queryByText('Click on the canvas to place the stamp')).not.toBeInTheDocument()
    })
  })

  it('does not show stamp controls when no stamp is selected', () => {
    render(<CreativeCanvas onSubmit={vi.fn()} onCancel={vi.fn()} enableStamps />)
    expect(screen.queryByText('Bring Fwd')).not.toBeInTheDocument()
    expect(screen.queryByText('Send Bwd')).not.toBeInTheDocument()
    expect(screen.queryByText('Remove')).not.toBeInTheDocument()
  })
})
