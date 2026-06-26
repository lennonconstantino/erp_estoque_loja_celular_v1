import { useEffect, useState } from 'react'
import {
  TrendingUp,
  Package,
  ShoppingCart,
  Users,
  AlertTriangle,
  BarChart2,
  type LucideIcon,
} from 'lucide-react'
import { PageShell } from '@/components/ui/page-shell'
import { cn } from '@/lib/utils'
import { api } from '@/lib/api'

interface MetricCardProps {
  title: string
  value: string | number
  trend?: {
    value: string
    positive: boolean
  }
  icon: LucideIcon
  className?: string
  loading?: boolean
}

function MetricCard({ title, value, trend, icon: Icon, className, loading }: MetricCardProps) {
  return (
    <div className={cn('bg-card border border-border rounded-2xl p-6 shadow-sm group hover:border-primary/50 transition-all duration-300', className)}>
      <div className="flex items-center justify-between mb-4">
        <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest leading-none">{title}</span>
        <div className="p-2 bg-muted/50 rounded-lg group-hover:bg-primary/10 group-hover:text-primary transition-colors">
          <Icon className="w-4 h-4" />
        </div>
      </div>
      <div className="flex items-end justify-between">
        <div className="flex-1">
          {loading ? (
            <div className="h-8 w-24 bg-muted animate-pulse rounded" />
          ) : (
            <h3 className="text-2xl font-extrabold tracking-tight font-mono text-foreground leading-none">{value}</h3>
          )}
          <p className="text-[10px] text-muted-foreground mt-2 font-medium uppercase tracking-tighter opacity-70">Desempenho Geral</p>
        </div>
        {!loading && trend && (
          <div className={cn(
            'flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold border shrink-0',
            trend.positive
              ? 'bg-green-500/10 text-green-600 border-green-500/20 dark:text-green-400'
              : 'bg-destructive/10 text-destructive border-destructive/20'
          )}>
            <TrendingUp className={cn('w-3 h-3', !trend.positive && 'rotate-180')} />
            {trend.value}
          </div>
        )}
      </div>
    </div>
  )
}

export default function DashboardPage() {
  const [metrics, setMetrics] = useState({
    vendasMês: 0,
    clientesTotal: 0,
    produtosTotal: 0,
    estoqueBaixo: 0,
    loading: true
  })

  async function fetchMetrics() {
    try {
      const [vendas, clientes, produtos, baixoMinimo] = await Promise.all([
        api.get<{ valor_total: number }>('/api/v1/relatorios/vendas'),
        api.get<{ total: number }>('/api/v1/clientes?limit=1'),
        api.get<{ total: number }>('/api/v1/produtos?limit=1'),
        api.get<{ total: number }>('/api/v1/relatorios/produtos/abaixo-do-minimo')
      ])

      setMetrics({
        vendasMês: vendas.valor_total || 0,
        clientesTotal: clientes.total || 0,
        produtosTotal: produtos.total || 0,
        estoqueBaixo: baixoMinimo.total || 0,
        loading: false
      })
    } catch (error) {
      console.error('Erro ao buscar métricas:', error)
      setMetrics(m => ({ ...m, loading: false }))
    }
  }

  useEffect(() => {
    fetchMetrics()
  }, [])

  const fmtCurrency = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

  return (
    <PageShell 
      title="Dashboard" 
      subtitle="Resumo operacional e indicadores de performance."
    >
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard 
          title="Faturamento (Mês)" 
          value={fmtCurrency(metrics.vendasMês)} 
          trend={{ value: '+12.5%', positive: true }}
          icon={ShoppingCart}
          loading={metrics.loading}
        />
        <MetricCard 
          title="Total de Clientes" 
          value={metrics.clientesTotal} 
          trend={{ value: '+4', positive: true }}
          icon={Users}
          loading={metrics.loading}
        />
        <MetricCard 
          title="Mix de Produtos" 
          value={metrics.produtosTotal} 
          icon={Package}
          loading={metrics.loading}
        />
        <MetricCard 
          title="Alertas de Estoque" 
          value={metrics.estoqueBaixo} 
          icon={AlertTriangle}
          className={metrics.estoqueBaixo > 0 ? "border-destructive/20 bg-destructive/5 dark:bg-destructive/10" : ""}
          loading={metrics.loading}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-8">
        <div className="bg-card border border-border rounded-2xl p-10 flex flex-col items-center justify-center text-center space-y-4 border-dashed relative group overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          <div className="w-16 h-16 bg-muted/50 rounded-full flex items-center justify-center text-muted-foreground group-hover:scale-110 transition-transform">
            <BarChart2 className="w-8 h-8" />
          </div>
          <div className="space-y-1 relative">
            <h4 className="font-bold text-foreground text-sm uppercase tracking-widest leading-none">Análise de Performance</h4>
            <p className="text-xs text-muted-foreground max-w-xs">Gráficos de vendas e lucratividade serão integrados na próxima versão.</p>
          </div>
        </div>

        <div className="bg-card border border-border rounded-2xl p-10 flex flex-col items-center justify-center text-center space-y-4 border-dashed relative group overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          <div className="w-16 h-16 bg-muted/50 rounded-full flex items-center justify-center text-muted-foreground group-hover:scale-110 transition-transform">
            <TrendingUp className="w-8 h-8" />
          </div>
          <div className="space-y-1 relative">
            <h4 className="font-bold text-foreground text-sm uppercase tracking-widest leading-none">Metas de Vendas</h4>
            <p className="text-xs text-muted-foreground max-w-xs">Acompanhamento de objetivos e comissões em desenvolvimento.</p>
          </div>
        </div>
      </div>
    </PageShell>
  )
}
