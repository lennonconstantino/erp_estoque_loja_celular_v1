import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export function Field({
  label,
  children,
  disabled,
}: {
  label: string
  children: ReactNode
  disabled?: boolean
}) {
  return (
    <div>
      <label className={cn('block text-sm font-medium mb-1', disabled ? 'text-gray-400' : 'text-gray-700')}>
        {label}
      </label>
      {children}
    </div>
  )
}

/** Classe padrão para <input>/<select>/<textarea>. */
export function inputClasses(disabled = false) {
  return cn(
    'w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900',
    disabled ? 'border-gray-200 bg-gray-50 text-gray-400 cursor-not-allowed' : 'border-gray-300',
  )
}
