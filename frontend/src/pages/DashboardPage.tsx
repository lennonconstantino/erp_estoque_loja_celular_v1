import { useNavigate } from 'react-router-dom'
import {
  ArrowUpDown,
  BarChart2,
  ClipboardList,
  LogOut,
  Package,
  ShoppingCart,
  Tag,
  Truck,
  Users,
} from 'lucide-react'
import { api } from '@/lib/api'
import { clearTokens } from '@/lib/auth'

const modulos = [
  { label: 'Clientes',           icon: Users,        path: '/clientes' },
  { label: 'Fornecedores',       icon: Truck,        path: '/fornecedores' },
  { label: 'Categorias',         icon: Tag,          path: '/categorias' },
  { label: 'Produtos',           icon: Package,      path: '/produtos' },
  { label: 'Compras',            icon: ShoppingCart, path: '/compras' },
  { label: 'Vendas',             icon: ClipboardList, path: '/vendas' },
  { label: 'Ajuste de Estoque',  icon: ArrowUpDown,  path: '/estoque/ajustes' },
  { label: 'Relatórios',         icon: BarChart2,    path: '/relatorios' },
]

export default function DashboardPage() {
  const navigate = useNavigate()

  async function handleLogout() {
    try {
      await api.post('/api/v1/auth/logout')
    } catch {
      // ignora erros de rede no logout
    }
    clearTokens()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
        <div>
          <h1 className="text-base font-semibold text-gray-900">ERP Estoque</h1>
          <p className="text-xs text-gray-500">Loja de Acessórios de Celular</p>
        </div>
        <button
          onClick={() => { void handleLogout() }}
          className="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-900 transition-colors"
        >
          <LogOut className="w-4 h-4" />
          Sair
        </button>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-10">
        <p className="text-xs font-medium text-gray-400 uppercase tracking-widest mb-4">
          Módulos
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          {modulos.map((mod) => (
            <button
              key={mod.path}
              onClick={() => navigate(mod.path)}
              className="flex flex-col items-center gap-3 bg-white rounded-lg border border-gray-200 p-6 hover:border-gray-400 hover:shadow-sm transition-all text-center"
            >
              <mod.icon className="w-6 h-6 text-gray-600" />
              <span className="text-sm font-medium text-gray-800">{mod.label}</span>
            </button>
          ))}
        </div>
      </main>
    </div>
  )
}
