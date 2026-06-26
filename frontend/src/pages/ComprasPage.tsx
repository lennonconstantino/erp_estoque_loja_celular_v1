import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, CheckCircle, Eye, Plus, Trash2 } from 'lucide-react'
import { api, ApiError } from '@/lib/api'

// ── tipos ──────────────────────────────────────────────────────────────────────

interface Fornecedor {
  id: string
  razao_social: string
  nome_fantasia: string
}

interface Produto {
  id: string
  descricao: string
  modelo: string
  preco_custo: number
  preco_venda: number
}

interface ItemCompra {
  id: string
  produto_id: string
  quantidade: number
  preco_compra: number
  preco_venda: number
  margem: number
}

interface Compra {
  id: string
  fornecedor_id: string
  nf: string
  dt_compra: string
  valor_total: number
  status: 'RASCUNHO' | 'CONFIRMADA' | 'CANCELADA'
  itens: ItemCompra[]
  criado_em: string
  atualizado_em: string
}

interface ItemForm {
  produto_id: string
  quantidade: number
  preco_compra: number
  preco_venda: number
}

const itemVazio = (): ItemForm => ({
  produto_id: '',
  quantidade: 1,
  preco_compra: 0,
  preco_venda: 0,
})

const STATUS_BADGE: Record<string, string> = {
  RASCUNHO: 'bg-yellow-100 text-yellow-800',
  CONFIRMADA: 'bg-green-100 text-green-800',
  CANCELADA: 'bg-red-100 text-red-800',
}

const STATUS_LABEL: Record<string, string> = {
  RASCUNHO: 'Rascunho',
  CONFIRMADA: 'Confirmada',
  CANCELADA: 'Cancelada',
}

// ── componente principal ───────────────────────────────────────────────────────

export default function ComprasPage() {
  const navigate = useNavigate()

  const [compras, setCompras] = useState<Compra[]>([])
  const [fornecedores, setFornecedores] = useState<Fornecedor[]>([])
  const [produtos, setProdutos] = useState<Produto[]>([])
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  // modal nova compra
  const [modalAberto, setModalAberto] = useState(false)
  const [fornecedorID, setFornecedorID] = useState('')
  const [nf, setNf] = useState('')
  const [dtCompra, setDtCompra] = useState(new Date().toISOString().slice(0, 10))
  const [itensForm, setItensForm] = useState<ItemForm[]>([itemVazio()])
  const [erroForm, setErroForm] = useState('')
  const [salvando, setSalvando] = useState(false)

  // modal detalhe
  const [detalhe, setDetalhe] = useState<Compra | null>(null)
  const [confirmando, setConfirmando] = useState<string | null>(null)

  async function carregarCompras() {
    setCarregando(true)
    setErro('')
    try {
      const res = await api.get<{ items: Compra[] }>('/api/v1/compras?limit=50')
      setCompras(res.items ?? [])
    } catch {
      setErro('Não foi possível carregar as compras.')
    } finally {
      setCarregando(false)
    }
  }

  async function carregarDependencias() {
    try {
      const [fRes, pRes] = await Promise.all([
        api.get<{ items: Fornecedor[] }>('/api/v1/fornecedores?limit=200'),
        api.get<{ items: Produto[] }>('/api/v1/produtos?limit=200'),
      ])
      setFornecedores(fRes.items ?? [])
      setProdutos(pRes.items ?? [])
    } catch {
      // não bloqueia a listagem
    }
  }

  useEffect(() => {
    void carregarCompras()
    void carregarDependencias()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // ── helpers de formulário ────────────────────────────────────────────────────

  function abrirModal() {
    setFornecedorID('')
    setNf('')
    setDtCompra(new Date().toISOString().slice(0, 10))
    setItensForm([itemVazio()])
    setErroForm('')
    setModalAberto(true)
  }

  function atualizarItem(idx: number, campo: keyof ItemForm, valor: string | number) {
    setItensForm(prev => prev.map((it, i) => {
      if (i !== idx) return it
      const novo = { ...it, [campo]: valor }
      if (campo === 'produto_id') {
        const prod = produtos.find(p => p.id === valor)
        if (prod) {
          novo.preco_compra = prod.preco_custo
          novo.preco_venda = prod.preco_venda
        }
      }
      return novo
    }))
  }

  function removerItem(idx: number) {
    setItensForm(prev => prev.filter((_, i) => i !== idx))
  }

  function margemItem(it: ItemForm): number {
    if (it.preco_compra <= 0) return 0
    return ((it.preco_venda - it.preco_compra) / it.preco_compra) * 100
  }

  function totalForm(): number {
    return itensForm.reduce((acc, it) => acc + it.quantidade * it.preco_compra, 0)
  }

  async function salvarCompra() {
    setErroForm('')
    if (!fornecedorID) { setErroForm('Selecione um fornecedor.'); return }
    if (!dtCompra) { setErroForm('Informe a data da compra.'); return }
    if (itensForm.length === 0) { setErroForm('Adicione pelo menos um item.'); return }
    for (const it of itensForm) {
      if (!it.produto_id) { setErroForm('Selecione o produto de todos os itens.'); return }
      if (it.quantidade <= 0) { setErroForm('Quantidade deve ser maior que zero.'); return }
      if (it.preco_compra <= 0 || it.preco_venda <= 0) { setErroForm('Preços devem ser maiores que zero.'); return }
      if (it.preco_compra >= it.preco_venda) { setErroForm('Preço de compra deve ser menor que o de venda.'); return }
    }
    setSalvando(true)
    try {
      await api.post('/api/v1/compras', {
        fornecedor_id: fornecedorID,
        nf,
        dt_compra: dtCompra,
        itens: itensForm.map(it => ({
          produto_id: it.produto_id,
          quantidade: Number(it.quantidade),
          preco_compra: Number(it.preco_compra),
          preco_venda: Number(it.preco_venda),
        })),
      })
      setModalAberto(false)
      void carregarCompras()
    } catch (e) {
      setErroForm(e instanceof ApiError ? e.message : 'Erro ao salvar compra.')
    } finally {
      setSalvando(false)
    }
  }

  async function confirmarCompra(id: string) {
    if (!confirm('Confirmar esta compra? O estoque será atualizado imediatamente.')) return
    setConfirmando(id)
    try {
      await api.post(`/api/v1/compras/${id}/confirmar`, {})
      void carregarCompras()
    } catch (e) {
      alert(e instanceof ApiError ? e.message : 'Erro ao confirmar compra.')
    } finally {
      setConfirmando(null)
    }
  }

  async function verDetalhe(id: string) {
    try {
      const c = await api.get<Compra>(`/api/v1/compras/${id}`)
      setDetalhe(c)
    } catch {
      alert('Não foi possível carregar o detalhe da compra.')
    }
  }

  function nomeProduto(id: string): string {
    const p = produtos.find(p => p.id === id)
    return p ? `${p.descricao}${p.modelo ? ` (${p.modelo})` : ''}` : id.slice(0, 8) + '…'
  }

  function nomeFornecedor(id: string): string {
    const f = fornecedores.find(f => f.id === id)
    return f ? (f.nome_fantasia || f.razao_social) : id.slice(0, 8) + '…'
  }

  // ── render ───────────────────────────────────────────────────────────────────

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-6xl px-4 py-8">
        {/* cabeçalho */}
        <div className="mb-6 flex items-center gap-4">
          <button onClick={() => navigate('/')} className="text-gray-500 hover:text-gray-700">
            <ArrowLeft className="h-5 w-5" />
          </button>
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-gray-900">Compras</h1>
            <p className="text-sm text-gray-500">Entrada de mercadorias</p>
          </div>
          <button
            onClick={abrirModal}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            <Plus className="h-4 w-4" /> Nova Compra
          </button>
        </div>

        {/* erros */}
        {erro && <div className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{erro}</div>}

        {/* tabela de compras */}
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
          {carregando ? (
            <div className="py-12 text-center text-gray-500">Carregando…</div>
          ) : compras.length === 0 ? (
            <div className="py-12 text-center text-gray-500">Nenhuma compra cadastrada ainda.</div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  {['Data', 'Fornecedor', 'NF', 'Total', 'Status', 'Ações'].map(h => (
                    <th key={h} className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {compras.map(c => (
                  <tr key={c.id} className="hover:bg-gray-50">
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">
                      {c.dt_compra}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-900">
                      {nomeFornecedor(c.fornecedor_id)}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">{c.nf || '—'}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm font-medium text-gray-900">
                      {c.valor_total.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_BADGE[c.status]}`}>
                        {STATUS_LABEL[c.status]}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3">
                      <div className="flex items-center gap-2">
                        <button
                          title="Ver detalhes"
                          onClick={() => verDetalhe(c.id)}
                          className="text-gray-400 hover:text-blue-600"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        {c.status === 'RASCUNHO' && (
                          <button
                            title="Confirmar compra"
                            disabled={confirmando === c.id}
                            onClick={() => confirmarCompra(c.id)}
                            className="text-gray-400 hover:text-green-600 disabled:opacity-50"
                          >
                            <CheckCircle className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* ── Modal: nova compra ─────────────────────────────────────────────── */}
      {modalAberto && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4 pt-12">
          <div className="w-full max-w-3xl rounded-xl bg-white p-6 shadow-xl">
            <h2 className="mb-4 text-lg font-semibold text-gray-900">Nova Compra</h2>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              {/* fornecedor */}
              <div className="sm:col-span-2">
                <label className="mb-1 block text-sm font-medium text-gray-700">Fornecedor *</label>
                <select
                  value={fornecedorID}
                  onChange={e => setFornecedorID(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                >
                  <option value="">Selecione…</option>
                  {fornecedores.map(f => (
                    <option key={f.id} value={f.id}>
                      {f.nome_fantasia || f.razao_social}
                    </option>
                  ))}
                </select>
              </div>

              {/* data */}
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Data *</label>
                <input
                  type="date"
                  value={dtCompra}
                  onChange={e => setDtCompra(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                />
              </div>

              {/* NF */}
              <div className="sm:col-span-3">
                <label className="mb-1 block text-sm font-medium text-gray-700">Nota Fiscal</label>
                <input
                  type="text"
                  value={nf}
                  onChange={e => setNf(e.target.value)}
                  placeholder="Número da NF (opcional)"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                />
              </div>
            </div>

            {/* itens */}
            <div className="mt-5">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">Itens</span>
                <button
                  onClick={() => setItensForm(prev => [...prev, itemVazio()])}
                  className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800"
                >
                  <Plus className="h-3 w-3" /> Adicionar item
                </button>
              </div>

              <div className="space-y-3">
                {itensForm.map((it, idx) => (
                  <div key={idx} className="grid grid-cols-12 gap-2 rounded-lg border border-gray-200 p-3">
                    {/* produto */}
                    <div className="col-span-4">
                      <label className="mb-0.5 block text-xs text-gray-500">Produto</label>
                      <select
                        value={it.produto_id}
                        onChange={e => atualizarItem(idx, 'produto_id', e.target.value)}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none"
                      >
                        <option value="">Selecione…</option>
                        {produtos.map(p => (
                          <option key={p.id} value={p.id}>
                            {p.descricao}{p.modelo ? ` (${p.modelo})` : ''}
                          </option>
                        ))}
                      </select>
                    </div>
                    {/* qtd */}
                    <div className="col-span-2">
                      <label className="mb-0.5 block text-xs text-gray-500">Qtd</label>
                      <input
                        type="number" min="1"
                        value={it.quantidade}
                        onChange={e => atualizarItem(idx, 'quantidade', Number(e.target.value))}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none"
                      />
                    </div>
                    {/* preço custo */}
                    <div className="col-span-2">
                      <label className="mb-0.5 block text-xs text-gray-500">Custo (R$)</label>
                      <input
                        type="number" min="0" step="0.01"
                        value={it.preco_compra}
                        onChange={e => atualizarItem(idx, 'preco_compra', Number(e.target.value))}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none"
                      />
                    </div>
                    {/* preço venda */}
                    <div className="col-span-2">
                      <label className="mb-0.5 block text-xs text-gray-500">Venda (R$)</label>
                      <input
                        type="number" min="0" step="0.01"
                        value={it.preco_venda}
                        onChange={e => atualizarItem(idx, 'preco_venda', Number(e.target.value))}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none"
                      />
                    </div>
                    {/* margem */}
                    <div className="col-span-1">
                      <label className="mb-0.5 block text-xs text-gray-500">Margem</label>
                      <span className={`block py-1 text-xs font-medium ${margemItem(it) > 0 ? 'text-green-700' : 'text-gray-400'}`}>
                        {margemItem(it).toFixed(0)}%
                      </span>
                    </div>
                    {/* remover */}
                    <div className="col-span-1 flex items-end justify-center pb-1">
                      <button
                        disabled={itensForm.length === 1}
                        onClick={() => removerItem(idx)}
                        className="text-gray-400 hover:text-red-600 disabled:opacity-30"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              {/* total */}
              <div className="mt-3 text-right text-sm font-medium text-gray-900">
                Total:{' '}
                {totalForm().toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
              </div>
            </div>

            {erroForm && (
              <div className="mt-3 rounded-lg bg-red-50 p-2 text-sm text-red-700">{erroForm}</div>
            )}

            <div className="mt-5 flex justify-end gap-3">
              <button
                onClick={() => setModalAberto(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
              >
                Cancelar
              </button>
              <button
                onClick={salvarCompra}
                disabled={salvando}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {salvando ? 'Salvando…' : 'Criar Rascunho'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Modal: detalhe da compra ──────────────────────────────────────── */}
      {detalhe && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
          onClick={() => setDetalhe(null)}
        >
          <div
            className="w-full max-w-2xl rounded-xl bg-white p-6 shadow-xl"
            onClick={e => e.stopPropagation()}
          >
            <div className="mb-4 flex items-start justify-between">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">Detalhe da Compra</h2>
                <p className="text-xs text-gray-500">{detalhe.id}</p>
              </div>
              <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_BADGE[detalhe.status]}`}>
                {STATUS_LABEL[detalhe.status]}
              </span>
            </div>

            <dl className="mb-4 grid grid-cols-3 gap-3 text-sm">
              <div>
                <dt className="text-gray-500">Fornecedor</dt>
                <dd className="font-medium text-gray-900">{nomeFornecedor(detalhe.fornecedor_id)}</dd>
              </div>
              <div>
                <dt className="text-gray-500">Data</dt>
                <dd className="font-medium text-gray-900">{detalhe.dt_compra}</dd>
              </div>
              <div>
                <dt className="text-gray-500">NF</dt>
                <dd className="font-medium text-gray-900">{detalhe.nf || '—'}</dd>
              </div>
            </dl>

            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead>
                <tr>
                  {['Produto', 'Qtd', 'Custo', 'Venda', 'Margem'].map(h => (
                    <th key={h} className="pb-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {(detalhe.itens ?? []).map(it => (
                  <tr key={it.id}>
                    <td className="py-2 text-gray-900">{nomeProduto(it.produto_id)}</td>
                    <td className="py-2 text-gray-900">{it.quantidade}</td>
                    <td className="py-2 text-gray-900">
                      {it.preco_compra.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                    </td>
                    <td className="py-2 text-gray-900">
                      {it.preco_venda.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                    </td>
                    <td className="py-2 text-green-700">{it.margem.toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr>
                  <td colSpan={5} className="pt-3 text-right text-sm font-semibold text-gray-900">
                    Total:{' '}
                    {detalhe.valor_total.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                  </td>
                </tr>
              </tfoot>
            </table>

            <div className="mt-5 flex justify-between">
              {detalhe.status === 'RASCUNHO' && (
                <button
                  disabled={confirmando === detalhe.id}
                  onClick={async () => {
                    await confirmarCompra(detalhe.id)
                    setDetalhe(null)
                  }}
                  className="flex items-center gap-2 rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50"
                >
                  <CheckCircle className="h-4 w-4" /> Confirmar Compra
                </button>
              )}
              <button
                onClick={() => setDetalhe(null)}
                className="ml-auto rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
              >
                Fechar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
