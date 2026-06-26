import { cn } from '@/lib/utils'

interface TabsProps<T extends string> {
  tabs: { key: T; label: string }[]
  activeTab: T
  onTabChange: (key: T) => void
  className?: string
}

/** Componente de Tabs técnico no estilo "Linear" (Pill, minimalista, alto contraste). */
export function Tabs<T extends string>({ tabs, activeTab, onTabChange, className }: TabsProps<T>) {
  return (
    <div className={cn('flex gap-1 bg-muted/20 border border-border rounded-full p-1 w-fit overflow-x-auto scrollbar-none', className)}>
      {tabs.map((tab) => {
        const isActive = activeTab === tab.key
        return (
          <button
            key={tab.key}
            onClick={() => onTabChange(tab.key)}
            className={cn(
              'px-4 py-1.5 text-[10px] font-bold uppercase tracking-widest rounded-full transition-all duration-300 whitespace-nowrap outline-none',
              isActive
                ? 'bg-primary text-primary-foreground shadow-md shadow-primary/10'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
            )}
          >
            {tab.label}
          </button>
        )
      })}
    </div>
  )
}
