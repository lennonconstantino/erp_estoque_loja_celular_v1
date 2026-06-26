import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Eye, Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button, buttonClasses } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge, type BadgeTone } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'

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
      alert(e instanceof Error ? e.message : 'Erro ao buscar venda')
    }
  }

  async function confirmarVenda(id: string) {
    setConfirmando(true)
    try {
      const v = await api.post<Venda>(`/api/v1/vendas/${id}/confirmar`, {})
      setDetalhe(v)
      await carregarVendas()
      if (v.doc_fiscal_numero) {
        alert(`Venda confirmada!\nDocumento fiscal: ${v.doc_fiscal_numero}`)
      }
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : 'Erro ao confirmar venda')
    } finally {
      setConfirmando(false)
    }
  }

  useEffect(() => {
    carregarVendas()
  }, [])

  const colunas: Column<Venda>[] = [
    { header: 'Data', sortAccessor: (v) => new Date(v.dt_venda).getTime(), cell: (v) => new Date(v.dt_venda).toLocaleDateString('pt-BR') },
    {
      header: 'Cliente',
      sortAccessor: (v) => (v.consumidor_final ? '' : v.cliente_id ?? ''),
      cell: (v) =>
        v.consumidor_final ? (
          <span className="text-gray-400 italic">Consumidor final</span>
        ) : (
          <span className="font-mono text-xs">{v.cliente_id?.slice(0, 8)}…</span>
        ),
    },
    { header: 'Forma Pgto', hideBelow: 'sm', sortAccessor: (v) => FORMA_PGTO_LABEL[v.forma_pgto] ?? v.forma_pgto, cell: (v) => FORMA_PGTO_LABEL[v.forma_pgto] ?? v.forma_pgto },
    { header: 'Doc. Fiscal', hideBelow: 'md', sortAccessor: (v) => v.doc_fiscal, cell: (v) => v.doc_fiscal },
    { header: 'Total', align: 'right', sortAccessor: (v) => v.valor_total, cell: (v) => <span className="font-medium text-gray-900">{brl(v.valor_total)}</span> },
    {
      header: 'Status',
      sortAccessor: (v) => STATUS_LABEL[v.status] ?? v.status,
      cell: (v) => <StatusBadge tone={STATUS_TONE[v.status] ?? 'neutral'}>{STATUS_LABEL[v.status] ?? v.status}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (v) => (
        <button onClick={() => abrirDetalhe(v.id)} className="text-gray-400 hover:text-gray-700" title="Ver detalhe">
          <Eye className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell
      title="Vendas"
      subtitle="Histórico de vendas e PDV"
      maxWidth="max-w-6xl"
      actions={
        <Link to="/vendas/nova" className={buttonClasses('primary')}>
          <Plus className="w-4 h-4" />
          Nova Venda (PDV)
        </Link>
      }
    >
      {erro && <p className="text-sm text-red-600">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={vendas}
        rowKey={(v) => v.id}
        loading={carregando}
        empty="Nenhuma venda registrada."
      />

      {detalhe && (
        <Modal title="Detalhe da Venda" onClose={() => setDetalhe(null)}>
          <div className="px-6 py-4 space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-gray-500">ID</p>
                <p className="font-mono text-xs text-gray-700">{detalhe.id}</p>
              </div>
              <div>
                <p className="text-gray-500">Status</p>
                <StatusBadge tone={STATUS_TONE[detalhe.status] ?? 'neutral'}>
                  {STATUS_LABEL[detalhe.status] ?? detalhe.status}
                </StatusBadge>
              </div>
              <div>
                <p className="text-gray-500">Data</p>
                <p className="text-gray-700">{new Date(detalhe.dt_venda).toLocaleString('pt-BR')}</p>
              </div>
              <div>
                <p className="text-gray-500">Forma Pgto</p>
                <p className="text-gray-700">{FORMA_PGTO_LABEL[detalhe.forma_pgto] ?? detalhe.forma_pgto}</p>
              </div>
              <div>
                <p className="text-gray-500">Doc. Fiscal</p>
                <p className="text-gray-700">{detalhe.doc_fiscal}</p>
              </div>
              <div>
                <p className="text-gray-500">Nº Fiscal</p>
                <p className="text-gray-700">{detalhe.doc_fiscal_numero || '—'}</p>
              </div>
              <div>
                <p className="text-gray-500">Desconto</p>
                <p className="text-gray-700">{brl(detalhe.desconto)}</p>
              </div>
              <div>
                <p className="text-gray-500">Total</p>
                <p className="text-lg font-bold text-gray-900">{brl(detalhe.valor_total)}</p>
              </div>
            </div>

            <div className="border-t border-gray-200 pt-4">
              <h3 className="text-sm font-medium text-gray-700 mb-2">Itens</h3>
              <table className="w-full text-sm">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left font-medium text-gray-600">Produto</th>
                    <th className="px-3 py-2 text-right font-medium text-gray-600">Qtd</th>
                    <th className="px-3 py-2 text-right font-medium text-gray-600">Preço Unit.</th>
                    <th className="px-3 py-2 text-right font-medium text-gray-600">Subtotal</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {detalhe.itens.map((item) => (
                    <tr key={item.id}>
                      <td className="px-3 py-2 font-mono text-xs text-gray-600">{item.produto_id.slice(0, 8)}…</td>
                      <td className="px-3 py-2 text-right text-gray-700">{item.quantidade}</td>
                      <td className="px-3 py-2 text-right text-gray-700">{brl(item.preco_unitario)}</td>
                      <td className="px-3 py-2 text-right font-medium text-gray-900">
                        {brl(item.quantidade * item.preco_unitario)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {detalhe.status === 'RASCUNHO' && (
              <div className="border-t border-gray-200 pt-4 flex justify-end">
                <Button variant="success" onClick={() => confirmarVenda(detalhe.id)} disabled={confirmando}>
                  {confirmando ? 'Confirmando…' : 'Confirmar Venda'}
                </Button>
              </div>
            )}
          </div>
        </Modal>
      )}
    </PageShell>
  )
}
