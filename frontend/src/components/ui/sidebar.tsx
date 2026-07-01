import { NavLink, useNavigate } from 'react-router-dom'
import {
  ArrowUpDown,
  BarChart2,
  ClipboardList,
  LayoutDashboard,
  LogOut,
  Package,
  ShoppingCart,
  Tag,
  Truck,
  UserCog,
  Users,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { api } from '@/lib/api'
import { clearTokens, hasPerm } from '@/lib/auth'

interface MenuItem {
  label: string
  icon: typeof LayoutDashboard
  path: string
  /** Permissão exigida para exibir o item (gating só de UI). */
  perm?: string
}

const menuItems: MenuItem[] = [
  { label: 'Dashboard', icon: LayoutDashboard, path: '/' },
  { label: 'Clientes', icon: Users, path: '/clientes' },
  { label: 'Fornecedores', icon: Truck, path: '/fornecedores' },
  { label: 'Categorias', icon: Tag, path: '/categorias' },
  { label: 'Produtos', icon: Package, path: '/produtos' },
  { label: 'Compras', icon: ShoppingCart, path: '/compras' },
  { label: 'Vendas', icon: ClipboardList, path: '/vendas' },
  { label: 'Ajuste Estoque', icon: ArrowUpDown, path: '/estoque/ajustes' },
  { label: 'Relatórios', icon: BarChart2, path: '/relatorios' },
  { label: 'Usuários', icon: UserCog, path: '/usuarios', perm: 'iam:admin' },
]

export function Sidebar() {
  const navigate = useNavigate()
  const itensVisiveis = menuItems.filter((item) => !item.perm || hasPerm(item.perm))
  const principal = itensVisiveis.slice(0, 1)
  const gestao = itensVisiveis.slice(1)

  async function handleLogout() {
    try {
      await api.post('/api/v1/auth/logout')
    } catch {
      // ignore
    }
    clearTokens()
    navigate('/login')
  }

  return (
    <aside className="w-64 h-screen flex flex-col bg-card border-r border-border sticky top-0 shrink-0">
      <div className="p-6">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-primary-foreground font-bold shadow-lg shadow-primary/20">
            E
          </div>
          <div>
            <h2 className="text-sm font-bold tracking-tight text-foreground leading-none">ERP Estoque</h2>
            <p className="text-[10px] text-muted-foreground mt-1 font-medium uppercase tracking-wider">Acessórios Celular</p>
          </div>
        </div>
      </div>

      <nav aria-label="Navegação principal" className="flex-1 px-4 space-y-1 overflow-y-auto py-2">
        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest px-3 mb-2">Principal</div>
        {principal.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-full transition-all group',
                isActive
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )
            }
          >
            <item.icon className="w-4 h-4" />
            {item.label}
          </NavLink>
        ))}

        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest px-3 mt-6 mb-2">Gestão</div>
        {gestao.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-full transition-all group',
                isActive
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )
            }
          >
            <item.icon className="w-4 h-4" />
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-border mt-auto">
        <button
          onClick={() => { void handleLogout() }}
          className="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium text-muted-foreground hover:text-destructive transition-colors rounded-full hover:bg-destructive/10"
        >
          <LogOut className="w-4 h-4" />
          Sair da conta
        </button>
      </div>
    </aside>
  )
}
