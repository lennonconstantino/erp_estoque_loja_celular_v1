import { cn } from '@/lib/utils'

export type BadgeTone = 'success' | 'neutral' | 'warning' | 'danger'

const tones: Record<BadgeTone, string> = {
  success: 'bg-green-500/10 text-green-800 border-green-500/20 dark:text-green-400 dark:bg-green-400/10',
  neutral: 'bg-muted/50 text-muted-foreground border-border',
  warning: 'bg-yellow-500/10 text-yellow-600 border-yellow-500/20 dark:text-yellow-400 dark:bg-yellow-400/10',
  danger: 'bg-destructive/10 text-destructive border-destructive/20 dark:bg-destructive/10',
}

/** Selo de status com tons semânticos e estilo pill técnico. */
export function StatusBadge({
  tone = 'neutral',
  children,
  className,
}: {
  tone?: BadgeTone
  children: React.ReactNode
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold border tracking-wider uppercase leading-none',
        tones[tone],
        className,
      )}
    >
      <span className="w-1 h-1 rounded-full bg-current mr-1.5 opacity-60" />
      {children}
    </span>
  )
}
