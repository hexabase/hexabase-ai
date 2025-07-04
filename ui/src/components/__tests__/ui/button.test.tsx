import React from 'react'
import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { Button } from '@/components/ui/button'

describe('Button', () => {
  it('renders with text', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument()
  })

  it('handles click events', async () => {
    const handleClick = jest.fn()
    const user = userEvent.setup()
    
    render(<Button onClick={handleClick}>Click me</Button>)
    
    await user.click(screen.getByRole('button', { name: 'Click me' }))
    
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('can be disabled', () => {
    render(<Button disabled>Disabled</Button>)
    
    expect(screen.getByRole('button', { name: 'Disabled' })).toBeDisabled()
  })

  it('applies variant prop', () => {
    const { rerender } = render(<Button variant="destructive">Delete</Button>)
    
    const button = screen.getByRole('button', { name: 'Delete' })
    expect(button).toBeInTheDocument()
    
    // Test that component accepts different variants
    rerender(<Button variant="outline">Outline</Button>)
    expect(screen.getByRole('button', { name: 'Outline' })).toBeInTheDocument()
  })

  it('applies size prop', () => {
    const { rerender } = render(<Button size="sm">Small</Button>)
    
    const button = screen.getByRole('button', { name: 'Small' })
    expect(button).toBeInTheDocument()
    
    // Test that component accepts different sizes
    rerender(<Button size="lg">Large</Button>)
    expect(screen.getByRole('button', { name: 'Large' })).toBeInTheDocument()
  })
})