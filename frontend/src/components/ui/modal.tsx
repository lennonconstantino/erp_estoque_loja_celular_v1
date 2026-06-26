import type { ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModalProps {
  title: string
  onClose: () => void
  children: ReactNode
  maxWidth?: 'max-w-md' | 'max-w-lg' | 'max-w-xl' | 'max-w-2xl' | 'max-w-3xl' | 'max-w-4xl'
}

/** Janela modal técnica (overlay escuro + cartão com bordas finas e suporte a Dark Mode). */
export function Modal({ title, onClose, children, maxWidth = 'max-w-2xl' }: ModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4 animate-in fade-in duration-300">
      <div className={cn('bg-card border border-border rounded-2xl shadow-2xl w-full max-h-[90vh] overflow-hidden flex flex-col animate-in zoom-in-95 duration-300', maxWidth)}>
        <div className="flex items-center justify-between px-6 py-4 border-b border-border bg-card sticky top-0 z-10">
          <h2 className="text-sm font-bold uppercase tracking-widest text-foreground">{title}</h2>
          <button 
            onClick={onClose} 
            className="p-1 text-muted-foreground hover:text-foreground transition-all rounded-full hover:bg-muted active:scale-90" 
            aria-label="Fechar"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-1">
          {children}
        </div>
      </div>
    </div>
  )
}
