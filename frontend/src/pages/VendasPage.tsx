import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/lib/api'

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

const STATUS_CLASS: Record<string, string> = {
  RASCUNHO: 'bg-yellow-100 text-yellow-800',
  CONFIRMADA: 'bg-green-100 text-green-800',
  CANCELADA: 'bg-red-100 text-red-800',
}

const FORMA_PGTO_LABEL: Record<string, string> = {
  DINHEIRO: 'Dinheiro',
  PIX: 'PIX',
  DEBITO: 'Débito',
  CREDITO: 'Crédito',
  OUTRO: 'Outro',
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
      const data = await api.get<{ items: Venda[] }>('/vendas?limit=50')
      setVendas(data.items ?? [])
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar vendas')
    } finally {
      setCarregando(false)
    }
  }

  async function abrirDetalhe(id: string) {
    try {
      const v = await api.get<Venda>(`/vendas/${id}`)
      setDetalhe(v)
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : 'Erro ao buscar venda')
    }
  }

  async function confirmarVenda(id: string) {
    setConfirmando(true)
    try {
      const v = await api.post<Venda>(`/vendas/${id}/confirmar`, {})
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

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Vendas</h1>
            <p className="text-sm text-gray-500 mt-1">Histórico de vendas e PDV</p>
          </div>
          <div className="flex gap-3">
            <Link
              to="/"
              className="px-4 py-2 text-sm font-medium text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              ← Dashboard
            </Link>
            <Link
              to="/vendas/nova"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700"
            >
              Nova Venda (PDV)
            </Link>
          </div>
        </div>

        {erro && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
            {erro}
          </div>
        )}

        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          {carregando ? (
            <div className="p-8 text-center text-gray-500 text-sm">Carregando...</div>
          ) : vendas.length === 0 ? (
            <div className="p-8 text-center text-gray-500 text-sm">Nenhuma venda registrada.</div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">Data</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">Cliente</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">Forma Pgto</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">Doc. Fiscal</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-600">Total</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">Status</th>
                  <th className="px-4 py-3 text-center font-medium text-gray-600">Ações</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {vendas.map((v) => (
                  <tr key={v.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-700">
                      {new Date(v.dt_venda).toLocaleDateString('pt-BR')}
                    </td>
                    <td className="px-4 py-3 text-gray-700">
                      {v.consumidor_final ? (
                        <span className="text-gray-400 italic">Consumidor final</span>
                      ) : (
                        <span className="font-mono text-xs">{v.cliente_id?.slice(0, 8)}…</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-700">
                      {FORMA_PGTO_LABEL[v.forma_pgto] ?? v.forma_pgto}
                    </td>
                    <td className="px-4 py-3 text-gray-700">{v.doc_fiscal}</td>
                    <td className="px-4 py-3 text-right font-medium text-gray-900">
                      {v.valor_total.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${STATUS_CLASS[v.status] ?? 'bg-gray-100 text-gray-700'}`}>
                        {STATUS_LABEL[v.status] ?? v.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <button
                        onClick={() => abrirDetalhe(v.id)}
                        className="px-3 py-1 text-xs text-blue-600 border border-blue-300 rounded hover:bg-blue-50"
                      >
                        Ver
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Modal de Detalhe */}
      {detalhe && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between p-5 border-b">
              <h2 className="text-lg font-semibold text-gray-900">Detalhe da Venda</h2>
              <button onClick={() => setDetalhe(null)} className="text-gray-400 hover:text-gray-600 text-xl">×</button>
            </div>
            <div className="p-5 space-y-4">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-gray-500">ID</p>
                  <p className="font-mono text-xs text-gray-700">{detalhe.id}</p>
                </div>
                <div>
                  <p className="text-gray-500">Status</p>
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${STATUS_CLASS[detalhe.status] ?? ''}`}>
                    {STATUS_LABEL[detalhe.status] ?? detalhe.status}
                  </span>
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
                  <p className="text-gray-700">
                    {detalhe.desconto.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                  </p>
                </div>
                <div>
                  <p className="text-gray-500">Total</p>
                  <p className="text-lg font-bold text-gray-900">
                    {detalhe.valor_total.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                  </p>
                </div>
              </div>

              <div className="border-t pt-4">
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
                        <td className="px-3 py-2 text-right text-gray-700">
                          {item.preco_unitario.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                        </td>
                        <td className="px-3 py-2 text-right font-medium text-gray-900">
                          {(item.quantidade * item.preco_unitario).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {detalhe.status === 'RASCUNHO' && (
                <div className="border-t pt-4 flex justify-end">
                  <button
                    onClick={() => confirmarVenda(detalhe.id)}
                    disabled={confirmando}
                    className="px-5 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50"
                  >
                    {confirmando ? 'Confirmando…' : 'Confirmar Venda'}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
