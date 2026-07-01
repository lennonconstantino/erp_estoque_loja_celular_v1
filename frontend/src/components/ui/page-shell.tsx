import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, ChevronRight, Home } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Sidebar } from './sidebar'
import { ThemeToggle } from './theme-toggle'
import { CommandPalette } from './command-palette'

interface PageShellProps {
  title: string
  subtitle?: string
  /** Ações exibidas à direita do cabeçalho (ex.: botão "Novo"). */
  actions?: ReactNode
  /** Largura máxima do conteúdo. */
  maxWidth?: 'max-w-4xl' | 'max-w-5xl' | 'max-w-6xl' | 'max-w-7xl' | 'max-w-none'
  /** Destino do botão voltar. */
  back?: string
  children: ReactNode
}

/**
 * PageShell refinado: integra Sidebar fixa, Header técnico com CommandPalette e ThemeToggle.
 */
export function PageShell({
  title,
  subtitle,
  actions,
  maxWidth = 'max-w-6xl',
  back,
  children,
}: PageShellProps) {
  return (
    <div className="flex min-h-screen bg-background text-foreground selection:bg-primary/20">
      <a
        href="#conteudo-principal"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[100] focus:rounded-full focus:bg-primary focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-foreground focus:shadow-lg"
      >
        Pular para o conteúdo
      </a>
      <Sidebar />

      <div className="flex-1 flex flex-col min-w-0">
        <header className="h-16 border-b border-border bg-background/80 backdrop-blur-sm sticky top-0 z-10 px-6 flex items-center justify-between gap-4">
          <div className="flex items-center gap-4 flex-1">
            {back ? (
              <Link
                to={back}
                className="p-2 -ml-2 text-muted-foreground hover:text-foreground transition-all rounded-full hover:bg-accent active:scale-95"
                aria-label="Voltar"
              >
                <ArrowLeft className="w-4 h-4" />
              </Link>
            ) : (
              <div className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider shrink-0">
                <Home className="w-3 h-3" />
                <ChevronRight className="w-3 h-3 opacity-20" />
                <span>ERP</span>
                <ChevronRight className="w-3 h-3 opacity-20" />
                <span className="text-foreground">{title}</span>
              </div>
            )}
            
            <div className="hidden lg:flex flex-1 max-w-sm ml-4">
              <CommandPalette />
            </div>
          </div>

          <div className="flex items-center gap-2">
            {actions && <div className="flex items-center gap-2 mr-2">{actions}</div>}
            <div className="w-px h-4 bg-border mx-1" />
            <ThemeToggle />
          </div>
        </header>

        <main id="conteudo-principal" tabIndex={-1} className="flex-1 p-6 lg:p-10 overflow-y-auto scrollbar-thin scrollbar-thumb-border scrollbar-track-transparent focus:outline-none">
          <div className={cn('mx-auto space-y-8', maxWidth)}>
            <div className="space-y-1 animate-in fade-in slide-in-from-top-4 duration-700">
              <h1 className="text-3xl font-extrabold tracking-tight text-foreground">{title}</h1>
              {subtitle && <p className="text-base text-muted-foreground max-w-2xl leading-relaxed">{subtitle}</p>}
            </div>
            
            <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 delay-150">
              {children}
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
