import { cloneElement, isValidElement, useId, type ReactElement, type ReactNode } from 'react'
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
  const autoId = useId()
  // Associa o rótulo ao controle: reaproveita um `id` já existente no filho
  // ou injeta um gerado, garantindo `htmlFor`/`id` casados (WCAG 1.3.1/3.3.2).
  let controlId = autoId
  let control = children
  if (isValidElement(children)) {
    const child = children as ReactElement<{ id?: string }>
    controlId = child.props.id ?? autoId
    if (!child.props.id) {
      control = cloneElement(child, { id: controlId })
    }
  }

  return (
    <div className="space-y-1.5">
      <label htmlFor={controlId} className={cn('block text-[10px] font-bold uppercase tracking-widest ml-1', disabled ? 'text-muted-foreground/50' : 'text-muted-foreground')}>
        {label}
      </label>
      {control}
    </div>
  )
}

/** Classe padrão para <input>/<select>/<textarea> estilo técnico pill. */
export function inputClasses(disabled = false) {
  return cn(
    'w-full rounded-full border px-4 py-2 text-sm transition-all focus:outline-none focus:ring-1 focus:ring-ring focus:bg-background',
    disabled
      ? 'border-border bg-muted/30 text-muted-foreground/50 cursor-not-allowed'
      : 'border-border bg-muted/10 text-foreground placeholder:text-muted-foreground/30',
  )
}

/** Variante compacta do input pill — para linhas de itens densas (Compras, PDV). */
export function inputClassesCompact(disabled = false) {
  return cn(
    'w-full rounded-full border px-3 py-1.5 text-xs transition-all focus:outline-none focus:ring-1 focus:ring-ring',
    disabled
      ? 'border-border bg-muted/30 text-muted-foreground/50 cursor-not-allowed'
      : 'border-border bg-card text-foreground placeholder:text-muted-foreground/30',
  )
}

/** Rótulo compacto técnico, reutilizado nas linhas de item (grids densas). */
export const compactLabelClass = 'text-[9px] font-black text-muted-foreground uppercase tracking-tighter mb-1.5 block'
