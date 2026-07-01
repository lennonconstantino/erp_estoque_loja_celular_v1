import * as React from 'react'
import { useNavigate } from 'react-router-dom'
import {
  CreditCard,
  Package,
  ShoppingCart,
  Users,
  Tag,
  Truck,
  ArrowUpDown,
  BarChart2,
  LayoutDashboard,
  Plus,
  UserCog,
} from 'lucide-react'

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import { hasPerm } from '@/lib/auth'

export function CommandPalette() {
  const [open, setOpen] = React.useState(false)
  const navigate = useNavigate()

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  const runCommand = React.useCallback((command: () => void) => {
    setOpen(false)
    command()
  }, [])

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="relative hidden lg:flex items-center gap-2 px-4 py-1.5 rounded-full border border-border bg-muted/30 text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-all text-xs w-64 text-left group"
      >
        <span className="opacity-50 group-hover:opacity-100 transition-opacity italic">Buscar comando (⌘K)</span>
        <kbd className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 hidden h-5 select-none items-center gap-1 rounded border border-border bg-card px-1.5 font-mono text-[9px] font-bold opacity-100 sm:flex shadow-sm">
          <span className="text-[10px]">⌘</span>K
        </kbd>
      </button>

      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Digite um comando ou navegue..." />
        <CommandList>
          <CommandEmpty>Nenhum resultado encontrado.</CommandEmpty>
          <CommandGroup heading="Sugestões">
            <CommandItem onSelect={() => runCommand(() => navigate('/'))} className="cursor-pointer">
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Ir para Dashboard</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => navigate('/vendas/nova'))} className="cursor-pointer">
              <Plus className="mr-2 h-4 w-4" />
              <span>Iniciar Nova Venda</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Cadastros">
            <CommandItem onSelect={() => runCommand(() => navigate('/produtos'))} className="cursor-pointer">
              <Package className="mr-2 h-4 w-4" />
              <span>Produtos</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => navigate('/clientes'))} className="cursor-pointer">
              <Users className="mr-2 h-4 w-4" />
              <span>Clientes</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => navigate('/fornecedores'))} className="cursor-pointer">
              <Truck className="mr-2 h-4 w-4" />
              <span>Fornecedores</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => navigate('/categorias'))} className="cursor-pointer">
              <Tag className="mr-2 h-4 w-4" />
              <span>Categorias</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Movimentações">
            <CommandItem onSelect={() => runCommand(() => navigate('/vendas'))} className="cursor-pointer">
              <ShoppingCart className="mr-2 h-4 w-4" />
              <span>Vendas</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => navigate('/compras'))} className="cursor-pointer">
              <CreditCard className="mr-2 h-4 w-4" />
              <span>Compras</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => navigate('/estoque/ajustes'))} className="cursor-pointer">
              <ArrowUpDown className="mr-2 h-4 w-4" />
              <span>Ajustes de Estoque</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Inteligência">
            <CommandItem onSelect={() => runCommand(() => navigate('/relatorios'))} className="cursor-pointer">
              <BarChart2 className="mr-2 h-4 w-4" />
              <span>Relatórios e BI</span>
            </CommandItem>
          </CommandGroup>
          {hasPerm('iam:admin') && (
            <>
              <CommandSeparator />
              <CommandGroup heading="Administração">
                <CommandItem onSelect={() => runCommand(() => navigate('/usuarios'))} className="cursor-pointer">
                  <UserCog className="mr-2 h-4 w-4" />
                  <span>Usuários</span>
                </CommandItem>
              </CommandGroup>
            </>
          )}
        </CommandList>
      </CommandDialog>
    </>
  )
}
