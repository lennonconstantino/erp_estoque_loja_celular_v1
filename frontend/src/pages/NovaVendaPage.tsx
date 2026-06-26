import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '@/lib/api'

interface Cliente {
  id: string
  nome: string
  cpf: string
  email: string
}

interface Produto {
  id: string
  descricao: string
  preco_venda: number
  estoque_atual: number
  disponivel: boolean
  ativo: boolean
}

interface ItemForm {
  produto_id: string
  quantidade: number
  preco_unitario: number
}

const FORMAS_PGTO = ['DINHEIRO', 'PIX', 'DEBITO', 'CREDITO', 'OUTRO'] as const
const DOCS_FISCAIS = ['CUPOM', 'NF'] as const

const FORMA_LABEL: Record<string, string> = {
  DINHEIRO: 'Dinheiro',
  PIX: 'PIX',
  DEBITO: 'Débito',
  CREDITO: 'Crédito',
  OUTRO: 'Outro',
}

export default function NovaVendaPage() {
  const navigate = useNavigate()
  const [produtos, setProdutos] = useState<Produto[]>([])
  const [cliente, setCliente] = useState<Cliente | null>(null)
  const [cpfBusca, setCpfBusca] = useState('')
  const [buscandoCliente, setBuscandoCliente] = useState(false)
  const [consumidorFinal, setConsumidorFinal] = useState(true)
  const [formaPgto, setFormaPgto] = useState<string>('DINHEIRO')
  const [docFiscal, setDocFiscal] = useState<string>('CUPOM')
  const [desconto, setDesconto] = useState(0)
  const [itens, setItens] = useState<ItemForm[]>([])
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  const totalItens = itens.reduce((s, i) => s + i.quantidade * i.preco_unitario, 0)
  const total = Math.max(0, totalItens - desconto)

  useEffect(() => {
    api.get<{ items: Produto[] }>('/api/v1/produtos?limit=200').then((d) => {
      setProdutos((d.items ?? []).filter((p) => p.ativo && p.disponivel))
    })
  }, [])

  async function buscarCliente() {
    if (!cpfBusca.trim()) return
    setBuscandoCliente(true)
    setErro('')
    try {
      const cpf = cpfBusca.replace(/\D/g, '')
      const data = await api.get<{ items: Cliente[] }>(`/api/v1/clientes?q=${cpf}&limit=1`)
      const encontrado = (data.items ?? [])[0]
      if (encontrado) {
        setCliente(encontrado)
      } else {
        setErro('Cliente não encontrado para este CPF.')
        setCliente(null)
      }
    } catch {
      setErro('Erro ao buscar cliente.')
      setCliente(null)
    } finally {
      setBuscandoCliente(false)
    }
  }

  function adicionarItem() {
    setItens((prev) => [...prev, { produto_id: '', quantidade: 1, preco_unitario: 0 }])
  }

  function removerItem(idx: number) {
    setItens((prev) => prev.filter((_, i) => i !== idx))
  }

  function atualizarItem(idx: number, campo: keyof ItemForm, valor: string | number) {
    setItens((prev) =>
      prev.map((item, i) => {
        if (i !== idx) return item
        if (campo === 'produto_id') {
          const prod = produtos.find((p) => p.id === valor)
          return { ...item, produto_id: String(valor), preco_unitario: prod?.preco_venda ?? item.preco_unitario }
        }
        return { ...item, [campo]: Number(valor) }
      }),
    )
  }

  function saldoDisponivel(produtoId: string): number {
    const prod = produtos.find((p) => p.id === produtoId)
    return prod?.estoque_atual ?? 0
  }

  function validarItens(): string | null {
    for (const item of itens) {
      if (!item.produto_id) return 'Selecione um produto em todos os itens.'
      const saldo = saldoDisponivel(item.produto_id)
      if (item.quantidade > saldo) {
        const prod = produtos.find((p) => p.id === item.produto_id)
        return `Saldo insuficiente para "${prod?.descricao}": disponível ${saldo}, solicitado ${item.quantidade}.`
      }
      if (item.preco_unitario <= 0) return 'Preço unitário deve ser positivo.'
    }
    return null
  }

  async function salvarVenda(confirmar: boolean) {
    if (itens.length === 0) { setErro('Adicione pelo menos um item.'); return }
    const validacao = validarItens()
    if (validacao) { setErro(validacao); return }
    if (!consumidorFinal && !cliente) { setErro('Selecione um cliente ou marque como consumidor final.'); return }
    if (desconto > totalItens) { setErro('Desconto maior que o total dos itens.'); return }

    setSalvando(true)
    setErro('')
    try {
      const body = {
        consumidor_final: consumidorFinal,
        cliente_id: consumidorFinal ? '' : (cliente?.id ?? ''),
        forma_pgto: formaPgto,
        doc_fiscal: docFiscal,
        desconto,
        itens: itens.map((i) => ({
          produto_id: i.produto_id,
          quantidade: i.quantidade,
          preco_unitario: i.preco_unitario,
        })),
      }

      interface Venda { id: string; doc_fiscal_numero?: string }
      const venda = await api.post<Venda>('/api/v1/vendas', body)

      if (confirmar) {
        const confirmada = await api.post<Venda>(`/api/v1/vendas/${venda.id}/confirmar`, {})
        if (confirmada.doc_fiscal_numero) {
          alert(`Venda confirmada!\nDocumento fiscal: ${confirmada.doc_fiscal_numero}`)
        }
      }
      navigate('/vendas')
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao salvar venda')
    } finally {
      setSalvando(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-3xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Nova Venda (PDV)</h1>
            <p className="text-sm text-gray-500 mt-1">Registre uma venda e emita o documento fiscal</p>
          </div>
          <Link to="/vendas" className="px-4 py-2 text-sm font-medium text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
            ← Voltar
          </Link>
        </div>

        {erro && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">{erro}</div>
        )}

        <div className="space-y-5">
          {/* Cliente */}
          <div className="bg-white rounded-xl border border-gray-200 p-5">
            <h2 className="text-sm font-semibold text-gray-700 mb-3">Cliente</h2>
            <div className="flex items-center gap-3 mb-3">
              <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={consumidorFinal}
                  onChange={(e) => { setConsumidorFinal(e.target.checked); setCliente(null); setCpfBusca('') }}
                  className="rounded"
                />
                Consumidor final
              </label>
            </div>
            {!consumidorFinal && (
              <div className="space-y-2">
                <div className="flex gap-2">
                  <input
                    type="text"
                    placeholder="CPF do cliente"
                    value={cpfBusca}
                    onChange={(e) => setCpfBusca(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && buscarCliente()}
                    className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    onClick={buscarCliente}
                    disabled={buscandoCliente}
                    className="px-4 py-2 text-sm text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {buscandoCliente ? '…' : 'Buscar'}
                  </button>
                </div>
                {cliente && (
                  <div className="p-3 bg-green-50 border border-green-200 rounded-lg text-sm">
                    <p className="font-medium text-green-800">{cliente.nome}</p>
                    <p className="text-green-600">{cliente.cpf} · {cliente.email}</p>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Pagamento */}
          <div className="bg-white rounded-xl border border-gray-200 p-5">
            <h2 className="text-sm font-semibold text-gray-700 mb-3">Pagamento</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs text-gray-500 mb-1">Forma de pagamento</label>
                <select
                  value={formaPgto}
                  onChange={(e) => setFormaPgto(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {FORMAS_PGTO.map((f) => (
                    <option key={f} value={f}>{FORMA_LABEL[f]}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-gray-500 mb-1">Documento fiscal</label>
                <select
                  value={docFiscal}
                  onChange={(e) => setDocFiscal(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {DOCS_FISCAIS.map((d) => (
                    <option key={d} value={d}>{d}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-gray-500 mb-1">Desconto (R$)</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  value={desconto}
                  onChange={(e) => setDesconto(parseFloat(e.target.value) || 0)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
          </div>

          {/* Itens */}
          <div className="bg-white rounded-xl border border-gray-200 p-5">
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-sm font-semibold text-gray-700">Itens</h2>
              <button
                onClick={adicionarItem}
                className="px-3 py-1 text-xs font-medium text-blue-600 border border-blue-300 rounded-lg hover:bg-blue-50"
              >
                + Adicionar item
              </button>
            </div>

            {itens.length === 0 && (
              <p className="text-sm text-gray-400 text-center py-4">Nenhum item adicionado.</p>
            )}

            <div className="space-y-3">
              {itens.map((item, idx) => {
                const saldo = item.produto_id ? saldoDisponivel(item.produto_id) : null
                const semEstoque = saldo !== null && item.quantidade > saldo
                return (
                  <div key={idx} className={`p-3 border rounded-lg ${semEstoque ? 'border-red-300 bg-red-50' : 'border-gray-200'}`}>
                    <div className="grid grid-cols-12 gap-2 items-end">
                      <div className="col-span-5">
                        <label className="block text-xs text-gray-500 mb-1">Produto</label>
                        <select
                          value={item.produto_id}
                          onChange={(e) => atualizarItem(idx, 'produto_id', e.target.value)}
                          className="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                        >
                          <option value="">Selecione…</option>
                          {produtos.map((p) => (
                            <option key={p.id} value={p.id}>
                              {p.descricao} (estoque: {p.estoque_atual})
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="col-span-2">
                        <label className="block text-xs text-gray-500 mb-1">Qtd</label>
                        <input
                          type="number"
                          min={1}
                          value={item.quantidade}
                          onChange={(e) => atualizarItem(idx, 'quantidade', e.target.value)}
                          className={`w-full px-2 py-1.5 text-sm border rounded focus:outline-none focus:ring-1 focus:ring-blue-500 ${semEstoque ? 'border-red-400' : 'border-gray-300'}`}
                        />
                        {semEstoque && (
                          <p className="text-xs text-red-600 mt-0.5">Máx: {saldo}</p>
                        )}
                      </div>
                      <div className="col-span-3">
                        <label className="block text-xs text-gray-500 mb-1">Preço Unit. (R$)</label>
                        <input
                          type="number"
                          min={0}
                          step={0.01}
                          value={item.preco_unitario}
                          onChange={(e) => atualizarItem(idx, 'preco_unitario', e.target.value)}
                          className="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                        />
                      </div>
                      <div className="col-span-2 flex justify-end">
                        <button
                          onClick={() => removerItem(idx)}
                          className="px-2 py-1.5 text-xs text-red-500 border border-red-200 rounded hover:bg-red-50"
                        >
                          Remover
                        </button>
                      </div>
                    </div>
                    {item.produto_id && item.preco_unitario > 0 && (
                      <p className="text-xs text-gray-500 mt-1 text-right">
                        Subtotal: {(item.quantidade * item.preco_unitario).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                      </p>
                    )}
                  </div>
                )
              })}
            </div>

            {itens.length > 0 && (
              <div className="mt-4 pt-3 border-t border-gray-100 text-sm space-y-1">
                <div className="flex justify-between text-gray-600">
                  <span>Subtotal</span>
                  <span>{totalItens.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</span>
                </div>
                {desconto > 0 && (
                  <div className="flex justify-between text-red-600">
                    <span>Desconto</span>
                    <span>− {desconto.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</span>
                  </div>
                )}
                <div className="flex justify-between font-bold text-gray-900 text-base pt-1">
                  <span>Total</span>
                  <span>{total.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</span>
                </div>
              </div>
            )}
          </div>

          {/* Ações */}
          <div className="flex justify-end gap-3">
            <button
              onClick={() => salvarVenda(false)}
              disabled={salvando}
              className="px-5 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
            >
              {salvando ? 'Salvando…' : 'Salvar rascunho'}
            </button>
            <button
              onClick={() => salvarVenda(true)}
              disabled={salvando}
              className="px-5 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50"
            >
              {salvando ? 'Processando…' : 'Confirmar e emitir fiscal'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
