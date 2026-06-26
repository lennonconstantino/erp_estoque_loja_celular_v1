import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export type BadgeTone = 'success' | 'neutral' | 'warning' | 'danger'

const tones: Record<BadgeTone, string> = {
  success: 'bg-green-50 text-green-700',
  neutral: 'bg-gray-100 text-gray-500',
  warning: 'bg-yellow-100 text-yellow-800',
  danger: 'bg-red-100 text-red-700',
}

export function StatusBadge({
  tone = 'neutral',
  children,
  className,
}: {
  tone?: BadgeTone
  children: ReactNode
  className?: string
}) {
  return (
    <span className={cn('inline-block px-2 py-0.5 rounded text-xs font-medium', tones[tone], className)}>
      {children}
    </span>
  )
}
