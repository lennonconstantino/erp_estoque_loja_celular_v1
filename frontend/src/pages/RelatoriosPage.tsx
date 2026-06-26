import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { Field, inputClasses } from '@/components/ui/field'
import { Tabs } from '@/components/ui/tabs'

// --- tipos ---

interface ProdutoAbaixoMinimo {
  id: string
  descricao: string
  estoque_atual: number
  estoque_minimo: number
  defasagem: number
}

interface ProdutoVendido {
  produto_id: string
  descricao: string
  total_vendido: number
  total_valor: number
}

interface ResumoVendas {
  total_vendas: number
  valor_total: number
  ticket_medio: number
  de: string
  ate: string
}

interface ResumoCompras {
  total_compras: number
  valor_total: number
  de: string
  ate: string
}

// --- helpers ---

function hojeISO() {
  return new Date().toISOString().slice(0, 10)
}

function primeiroDiaMesISO() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

function brl(v: number) {
  return v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

// --- aba enum ---
type Aba = 'abaixo-minimo' | 'mais-vendidos' | 'menos-vendidos' | 'resumo-vendas' | 'resumo-compras'

const ABAS: { key: Aba; label: string }[] = [
  { key: 'abaixo-minimo', label: 'Abaixo do Mínimo' },
  { key: 'mais-vendidos', label: 'Mais Vendidos' },
  { key: 'menos-vendidos', label: 'Menos Vendidos' },
  { key: 'resumo-vendas', label: 'Resumo de Vendas' },
  { key: 'resumo-compras', label: 'Resumo de Compras' },
]

export default function RelatoriosPage() {
  const [aba, setAba] = useState<Aba>('abaixo-minimo')
  const [de, setDe] = useState(primeiroDiaMesISO())
  const [ate, setAte] = useState(hojeISO())
  const [limite, setLimite] = useState(10)
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  // dados por relatório
  const [abaixoMinimo, setAbaixoMinimo] = useState<ProdutoAbaixoMinimo[]>([])
  const [maisVendidos, setMaisVendidos] = useState<ProdutoVendido[]>([])
  const [menosVendidos, setMenosVendidos] = useState<ProdutoVendido[]>([])
  const [resumoVendas, setResumoVendas] = useState<ResumoVendas | null>(null)
  const [resumoCompras, setResumoCompras] = useState<ResumoCompras | null>(null)

  async function carregar() {
    setCarregando(true)
    setErro('')
    try {
      switch (aba) {
        case 'abaixo-minimo': {
          const d = await api.get<{ items: ProdutoAbaixoMinimo[] }>('/api/v1/relatorios/produtos/abaixo-do-minimo')
          setAbaixoMinimo(d.items ?? [])
          break
        }
        case 'mais-vendidos': {
          const d = await api.get<{ items: ProdutoVendido[] }>(
            `/api/v1/relatorios/produtos/mais-vendidos?de=${de}&ate=${ate}&limite=${limite}`,
          )
          setMaisVendidos(d.items ?? [])
          break
        }
        case 'menos-vendidos': {
          const d = await api.get<{ items: ProdutoVendido[] }>(
            `/api/v1/relatorios/produtos/menos-vendidos?de=${de}&ate=${ate}&limite=${limite}`,
          )
          setMenosVendidos(d.items ?? [])
          break
        }
        case 'resumo-vendas': {
          const d = await api.get<ResumoVendas>(`/api/v1/relatorios/vendas?de=${de}&ate=${ate}`)
          setResumoVendas(d)
          break
        }
        case 'resumo-compras': {
          const d = await api.get<ResumoCompras>(`/api/v1/relatorios/compras?de=${de}&ate=${ate}`)
          setResumoCompras(d)
          break
        }
      }
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar relatório')
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => {
    carregar()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [aba])

  const precisaPeriodo = aba !== 'abaixo-minimo'
  const precisaLimite = aba === 'mais-vendidos' || aba === 'menos-vendidos'

  const colunasAbaixo: Column<ProdutoAbaixoMinimo>[] = [
    { header: 'Produto', sortAccessor: (p) => p.descricao, cell: (p) => <span className="font-bold text-foreground">{p.descricao}</span> },
    { header: 'Estoque Atual', align: 'right', sortAccessor: (p) => p.estoque_atual, cell: (p) => <span className="text-destructive font-bold">{p.estoque_atual}</span>, isTechnical: true },
    { header: 'Mínimo', align: 'right', sortAccessor: (p) => p.estoque_minimo, cell: (p) => <span className="text-muted-foreground">{p.estoque_minimo}</span>, isTechnical: true },
    { header: 'Defasagem', align: 'right', sortAccessor: (p) => p.defasagem, cell: (p) => <span className="text-destructive font-bold">−{p.defasagem}</span>, isTechnical: true },
  ]

  const listaVendidos = aba === 'mais-vendidos' ? maisVendidos : menosVendidos
  const colunasVendidos: Column<ProdutoVendido>[] = [
    { header: 'Produto', sortAccessor: (p) => p.descricao, cell: (p) => <span className="font-bold text-foreground">{p.descricao}</span> },
    { header: 'Qtd Vendida', align: 'right', sortAccessor: (p) => p.total_vendido, cell: (p) => <span className="text-foreground font-bold">{p.total_vendido}</span>, isTechnical: true },
    { header: 'Total (R$)', align: 'right', sortAccessor: (p) => p.total_valor, cell: (p) => <span className="text-muted-foreground font-mono">{brl(p.total_valor)}</span>, isTechnical: true },
  ]

  return (
    <PageShell title="Relatórios" subtitle="Análises de estoque, vendas e compras" maxWidth="max-w-6xl">
      <Tabs tabs={ABAS} activeTab={aba} onTabChange={setAba} />

      {/* filtros */}
      {(precisaPeriodo || precisaLimite) && (
        <div className="bg-card rounded-2xl border border-border p-6 flex flex-wrap gap-6 items-end shadow-sm animate-in fade-in duration-500">
          {precisaPeriodo && (
            <>
              <Field label="Período: De">
                <input type="date" value={de} onChange={(e) => setDe(e.target.value)} className={inputClasses() + ' w-full sm:w-44'} />
              </Field>
              <Field label="Até">
                <input type="date" value={ate} onChange={(e) => setAte(e.target.value)} className={inputClasses() + ' w-full sm:w-44'} />
              </Field>
            </>
          )}
          {precisaLimite && (
            <Field label="Limite">
              <input
                type="number"
                min={1}
                max={100}
                value={limite}
                onChange={(e) => setLimite(parseInt(e.target.value) || 10)}
                className={inputClasses() + ' w-full sm:w-28'}
              />
            </Field>
          )}
          <Button onClick={carregar} disabled={carregando} className="min-w-32 h-10">
            {carregando ? 'Carregando…' : 'Atualizar Dados'}
          </Button>
        </div>
      )}

      {erro && <p className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      {/* conteúdo por aba */}
      {aba === 'abaixo-minimo' && (
        <DataTable
          columns={colunasAbaixo}
          rows={abaixoMinimo}
          rowKey={(p) => p.id}
          loading={carregando}
          empty="Nenhum produto abaixo do mínimo."
          rowClassName={() => 'bg-destructive/5 hover:bg-destructive/10 transition-colors'}
        />
      )}

      {(aba === 'mais-vendidos' || aba === 'menos-vendidos') && (
        <DataTable
          columns={colunasVendidos}
          rows={listaVendidos}
          rowKey={(p) => p.produto_id}
          loading={carregando}
          empty="Nenhuma venda no período."
        />
      )}

      {aba === 'resumo-vendas' && resumoVendas && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 animate-in slide-in-from-bottom-4 duration-500">
          <CardMetrica titulo="Volume de Vendas" valor={String(resumoVendas.total_vendas)} sufixo="vendas" />
          <CardMetrica titulo="Faturamento Total" valor={brl(resumoVendas.valor_total)} />
          <CardMetrica titulo="Ticket Médio" valor={brl(resumoVendas.ticket_medio)} />
        </div>
      )}

      {aba === 'resumo-compras' && resumoCompras && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 animate-in slide-in-from-bottom-4 duration-500">
          <CardMetrica titulo="Volume de Compras" valor={String(resumoCompras.total_compras)} sufixo="compras" />
          <CardMetrica titulo="Investimento Total" valor={brl(resumoCompras.valor_total)} />
        </div>
      )}
    </PageShell>
  )
}

function CardMetrica({ titulo, valor, sufixo }: { titulo: string; valor: string; sufixo?: string }) {
  return (
    <div className="bg-card rounded-2xl border border-border p-8 shadow-sm group hover:border-primary/40 hover:shadow-md transition-all duration-300">
      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] mb-6 leading-none opacity-60 group-hover:opacity-100 transition-opacity">{titulo}</p>
      <div className="flex items-baseline gap-2">
        <span className="text-3xl font-black tracking-tighter text-foreground font-mono leading-none">{valor}</span>
        {sufixo && <span className="text-xs font-black text-muted-foreground uppercase tracking-widest opacity-40">{sufixo}</span>}
      </div>
    </div>
  )
}
