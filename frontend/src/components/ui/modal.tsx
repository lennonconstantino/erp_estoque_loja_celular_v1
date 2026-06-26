import type { ReactNode } from 'react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModalProps {
  title: string
  onClose: () => void
  children: ReactNode
  maxWidth?: 'max-w-md' | 'max-w-lg' | 'max-w-xl' | 'max-w-2xl' | 'max-w-3xl' | 'max-w-4xl'
}

/**
 * Janela modal técnica (overlay escuro + cartão com bordas finas e suporte a Dark Mode).
 *
 * Construída sobre o Radix Dialog: garante `role="dialog"` + `aria-modal`,
 * rótulo acessível via `aria-labelledby` (o título), focus trap, foco inicial
 * no conteúdo, restauração do foco ao fechar e fechamento por Escape/overlay
 * (WCAG 2.1.2 / 2.4.3 / 4.1.2). A API pública permanece a mesma.
 */
export function Modal({ title, onClose, children, maxWidth = 'max-w-2xl' }: ModalProps) {
  return (
    <DialogPrimitive.Root open onOpenChange={(aberto) => { if (!aberto) onClose() }}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm animate-in fade-in duration-300" />
        <DialogPrimitive.Content
          aria-describedby={undefined}
          className={cn(
            'fixed left-1/2 top-1/2 z-50 w-full -translate-x-1/2 -translate-y-1/2 bg-card border border-border rounded-2xl shadow-2xl max-h-[90vh] overflow-hidden flex flex-col animate-in zoom-in-95 duration-300 focus:outline-none',
            maxWidth,
          )}
        >
          <div className="flex items-center justify-between px-6 py-4 border-b border-border bg-card sticky top-0 z-10">
            <DialogPrimitive.Title className="text-sm font-bold uppercase tracking-widest text-foreground">{title}</DialogPrimitive.Title>
            <DialogPrimitive.Close
              className="p-1 text-muted-foreground hover:text-foreground transition-all rounded-full hover:bg-muted active:scale-90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              aria-label="Fechar"
            >
              <X className="w-4 h-4" />
            </DialogPrimitive.Close>
          </div>
          <div className="flex-1 overflow-y-auto p-1">
            {children}
          </div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  )
}
