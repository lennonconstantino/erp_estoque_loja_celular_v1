import type { ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'success' | 'ghost'
export type ButtonSize = 'sm' | 'md' | 'icon'

const variants: Record<ButtonVariant, string> = {
  primary: 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm border border-primary/20',
  secondary: 'border border-border bg-card text-foreground hover:bg-accent hover:text-accent-foreground',
  danger: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
  success: 'bg-green-600 text-white hover:bg-green-700',
  ghost: 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
}

const sizes: Record<ButtonSize, string> = {
  sm: 'px-3 py-1.5 text-xs',
  md: 'px-4 py-2 text-sm',
  icon: 'p-2',
}

/** Classes do botão padrão — útil para aplicar em <Link> do react-router. */
export function buttonClasses(variant: ButtonVariant = 'primary', size: ButtonSize = 'md') {
  return cn(
    'inline-flex items-center justify-center gap-2 rounded-full font-medium transition-all active:scale-[0.98] cursor-pointer',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
    'disabled:opacity-50 disabled:cursor-not-allowed',
    variants[variant],
    sizes[size],
  )
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
}

export function Button({ variant = 'primary', size = 'md', className, ...props }: ButtonProps) {
  return <button className={cn(buttonClasses(variant, size), className)} {...props} />
}
