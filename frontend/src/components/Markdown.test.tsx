import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Markdown } from './Markdown'

describe('Markdown', () => {
  it('CA-5: muestra los bloques de código con botón de copiar', async () => {
    // userEvent.setup() instala su propio clipboard: el mock va después.
    const user = userEvent.setup()
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })

    render(<Markdown content={'Reiniciar:\n\n```bash\nsudo systemctl restart nginx\n```'} />)

    await user.click(screen.getByRole('button', { name: 'Copiar' }))
    expect(writeText).toHaveBeenCalledWith('sudo systemctl restart nginx')
    expect(await screen.findByRole('button', { name: 'Copiado' })).toBeInTheDocument()
  })

  it('CA-7: no interpreta HTML crudo (<script>, <img onerror>)', () => {
    const { container } = render(
      <Markdown content={'Hola <script>alert(1)</script> <img src=x onerror="alert(2)">'} />,
    )
    expect(container.querySelector('script')).toBeNull()
    expect(container.querySelector('img')).toBeNull()
  })

  it('CA-7: neutraliza links javascript:', () => {
    render(<Markdown content={'[clic](javascript:alert(1)) y [ok](https://example.com)'} />)
    const bad = screen.getByText('clic').closest('a')
    expect(bad?.getAttribute('href') ?? '').not.toMatch(/javascript:/i)
    const good = screen.getByRole('link', { name: 'ok' })
    expect(good).toHaveAttribute('href', 'https://example.com')
    expect(good).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('renderiza tablas de GitHub (remark-gfm)', () => {
    render(<Markdown content={'| Host | IP |\n|---|---|\n| web-01 | 192.0.2.10 |'} />)
    expect(screen.getByRole('table')).toBeInTheDocument()
    expect(screen.getByText('192.0.2.10')).toBeInTheDocument()
  })
})
