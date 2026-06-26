import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PageShellProps {
  title: string
  subtitle?: string
  /** Ações exibidas à direita do cabeçalho (ex.: botão "Novo"). */
  actions?: ReactNode
  /** Largura máxima do conteúdo. */
  maxWidth?: 'max-w-4xl' | 'max-w-5xl' | 'max-w-6xl'
  /** Destino do botão voltar. */
  back?: string
  children: ReactNode
}

/**
 * Casca padrão das páginas internas: barra de cabeçalho branca com botão
 * voltar + título/subtítulo, e contêiner principal centralizado.
 */
export function PageShell({
  title,
  subtitle,
  actions,
  maxWidth = 'max-w-5xl',
  back = '/',
  children,
}: PageShellProps) {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex items-center gap-4">
        <button
          onClick={() => navigate(back)}
          className="text-gray-400 hover:text-gray-700"
          aria-label="Voltar"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div className="flex-1">
          <h1 className="text-base font-semibold text-gray-900">{title}</h1>
          {subtitle && <p className="text-xs text-gray-500">{subtitle}</p>}
        </div>
        {actions && <div className="flex items-center gap-3">{actions}</div>}
      </header>

      <main className={cn('mx-auto px-6 py-8 space-y-4', maxWidth)}>{children}</main>
    </div>
  )
}
