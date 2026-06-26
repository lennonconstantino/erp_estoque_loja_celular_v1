import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Eye, Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button, buttonClasses } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge, type BadgeTone } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { toast } from 'sonner'

interface DetalheVenda {
  id: string
  produto_id: string
  quantidade: number
  preco_unitario: number
}

interface Venda {
  id: string
  dt_venda: string
  valor_total: number
  desconto: number
  forma_pgto: string
  cliente_id: string | null
  consumidor_final: boolean
  doc_fiscal: string
  status: string
  doc_fiscal_numero?: string
  itens: DetalheVenda[]
  criado_em: string
  atualizado_em: string
}

const STATUS_LABEL: Record<string, string> = {
  RASCUNHO: 'Rascunho',
  CONFIRMADA: 'Confirmada',
  CANCELADA: 'Cancelada',
}

const STATUS_TONE: Record<string, BadgeTone> = {
  RASCUNHO: 'warning',
  CONFIRMADA: 'success',
  CANCELADA: 'danger',
}

const FORMA_PGTO_LABEL: Record<string, string> = {
  DINHEIRO: 'Dinheiro',
  PIX: 'PIX',
  DEBITO: 'Débito',
  CREDITO: 'Crédito',
  OUTRO: 'Outro',
}

function brl(v: number) {
  return v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

export default function VendasPage() {
  const [vendas, setVendas] = useState<Venda[]>([])
  const [carregando, setCarregando] = useState(true)
  const [erro, setErro] = useState('')
  const [detalhe, setDetalhe] = useState<Venda | null>(null)
  const [confirmando, setConfirmando] = useState(false)

  async function carregarVendas() {
    setCarregando(true)
    setErro('')
    try {
      const data = await api.get<{ items: Venda[] }>('/api/v1/vendas?limit=50')
      setVendas(data.items ?? [])
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar vendas')
    } finally {
      setCarregando(false)
    }
  }

  async function abrirDetalhe(id: string) {
    try {
      const v = await api.get<Venda>(`/api/v1/vendas/${id}`)
      setDetalhe(v)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Erro ao buscar venda')
    }
  }

  async function confirmarVenda(id: string) {
    setConfirmando(true)
    try {
      const v = await api.post<Venda>(`/api/v1/vendas/${id}/confirmar`, {})
      setDetalhe(v)
      await carregarVendas()
      toast.success('Venda confirmada com sucesso!')
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Erro ao confirmar venda')
    } finally {
      setConfirmando(false)
    }
  }

  useEffect(() => {
    carregarVendas()
  }, [])

  const colunas: Column<Venda>[] = [
    { 
      header: 'Data', 
      sortAccessor: (v) => new Date(v.dt_venda).getTime(), 
      cell: (v) => <span className="font-bold text-foreground">{new Date(v.dt_venda).toLocaleDateString('pt-BR')}</span> 
    },
    {
      header: 'Cliente',
      sortAccessor: (v) => (v.consumidor_final ? '' : v.cliente_id ?? ''),
      cell: (v) =>
        v.consumidor_final ? (
          <span className="text-muted-foreground italic text-xs uppercase tracking-widest font-bold">Consumidor final</span>
        ) : (
          <span className="font-mono text-xs text-muted-foreground">{v.cliente_id?.slice(0, 8)}…</span>
        ),
      isTechnical: true,
    },
    { header: 'Pagamento', hideBelow: 'sm', sortAccessor: (v) => FORMA_PGTO_LABEL[v.forma_pgto] ?? v.forma_pgto, cell: (v) => <span className="text-muted-foreground">{FORMA_PGTO_LABEL[v.forma_pgto] ?? v.forma_pgto}</span> },
    { header: 'Documento', hideBelow: 'md', sortAccessor: (v) => v.doc_fiscal, cell: (v) => <span className="text-muted-foreground font-mono text-[10px]">{v.doc_fiscal}</span>, isTechnical: true },
    { header: 'Total', align: 'right', sortAccessor: (v) => v.valor_total, cell: (v) => <span className="font-black text-foreground font-mono">{brl(v.valor_total)}</span>, isTechnical: true },
    {
      header: 'Status',
      sortAccessor: (v) => STATUS_LABEL[v.status] ?? v.status,
      cell: (v) => <StatusBadge tone={STATUS_TONE[v.status] ?? 'neutral'}>{STATUS_LABEL[v.status] ?? v.status}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (v) => (
        <button onClick={() => abrirDetalhe(v.id)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted" aria-label="Ver detalhe" title="Ver detalhe">
          <Eye className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell
      title="Vendas"
      subtitle="Histórico de transações e frente de caixa."
      maxWidth="max-w-6xl"
      actions={
        <Link to="/vendas/nova" className={buttonClasses('primary')}>
          <Plus className="w-4 h-4" />
          Nova Venda (PDV)
        </Link>
      }
    >
      {erro && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={vendas}
        rowKey={(v) => v.id}
        loading={carregando}
        empty="Nenhuma venda registrada."
      />

      {detalhe && (
        <Modal title="Comprovante de Venda" onClose={() => setDetalhe(null)} maxWidth="max-w-xl">
          <div className="px-8 py-8 space-y-8 animate-in fade-in duration-300">
            <div className="grid grid-cols-2 gap-8 text-sm">
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Identificador</p>
                <p className="font-mono text-xs text-foreground truncate" title={detalhe.id}>{detalhe.id}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Status Atual</p>
                <StatusBadge tone={STATUS_TONE[detalhe.status] ?? 'neutral'}>
                  {STATUS_LABEL[detalhe.status] ?? detalhe.status}
                </StatusBadge>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Data e Hora</p>
                <p className="text-foreground font-bold">{new Date(detalhe.dt_venda).toLocaleString('pt-BR')}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Forma de Pagamento</p>
                <p className="text-foreground font-bold">{FORMA_PGTO_LABEL[detalhe.forma_pgto] ?? detalhe.forma_pgto}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Tipo de Documento</p>
                <p className="text-foreground font-bold">{detalhe.doc_fiscal}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Nº Documento</p>
                <p className="text-foreground font-bold font-mono">{detalhe.doc_fiscal_numero || 'PENDENTE'}</p>
              </div>
            </div>

            <div className="space-y-4">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest border-b border-border pb-2">Itens da Transação</p>
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-muted-foreground font-black uppercase tracking-tighter">
                    <th className="py-2 text-left">Referência</th>
                    <th className="py-2 text-right">Qtd</th>
                    <th className="py-2 text-right">Unitário</th>
                    <th className="py-2 text-right">Subtotal</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/50">
                  {detalhe.itens.map((item) => (
                    <tr key={item.id} className="group hover:bg-muted/5 transition-colors">
                      <td className="py-3 font-mono text-muted-foreground">{item.produto_id.slice(0, 12)}…</td>
                      <td className="py-3 text-right font-bold">{item.quantidade}</td>
                      <td className="py-3 text-right text-muted-foreground font-mono">{brl(item.preco_unitario)}</td>
                      <td className="py-3 text-right font-black font-mono">
                        {brl(item.quantidade * item.preco_unitario)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="bg-muted/30 rounded-2xl p-6 space-y-2 border border-border shadow-inner">
               <div className="flex justify-between text-xs text-muted-foreground font-bold uppercase tracking-widest">
                  <span>Descontos Aplicados</span>
                  <span className="font-mono">{brl(detalhe.desconto)}</span>
               </div>
               <div className="flex justify-between items-end">
                  <span className="text-xs font-black text-foreground uppercase tracking-[0.2em]">Valor Final</span>
                  <span className="text-3xl font-black text-primary font-mono tracking-tighter">{brl(detalhe.valor_total)}</span>
               </div>
            </div>

            <div className="flex justify-end gap-3 border-t border-border pt-6">
              <Button variant="secondary" onClick={() => setDetalhe(null)}>Fechar</Button>
              {detalhe.status === 'RASCUNHO' && (
                <Button variant="primary" onClick={() => confirmarVenda(detalhe.id)} disabled={confirmando} className="min-w-40">
                  {confirmando ? 'Processando…' : 'Finalizar Venda'}
                </Button>
              )}
            </div>
          </div>
        </Modal>
      )}
    </PageShell>
  )
}
