import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

interface Categoria {
  id: string
  descricao: string
}

interface Produto {
  id: string
  categoria_id: string
  descricao: string
  preco_custo: number
  preco_venda: number
  margem_pct: number
  estoque_minimo: number
  estoque_atual: number
  garantia_meses: number
  modelo: string
  disponivel: boolean
  ativo: boolean
}

const LIMITE = 20

function fmt(v: number) {
  return v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

export default function ProdutosPage() {
  const [itens, setItens] = useState<Produto[]>([])
  const [categorias, setCategorias] = useState<Categoria[]>([])
  const [offset, setOffset] = useState(0)
  const [busca, setBusca] = useState('')
  const [filtroCategoria, setFiltroCategoria] = useState('')
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  const [modalAberto, setModalAberto] = useState(false)
  const [editando, setEditando] = useState<Produto | null>(null)
  const [salvando, setSalvando] = useState(false)
  const [erroModal, setErroModal] = useState('')

  const [form, setForm] = useState({
    categoria_id: '',
    descricao: '',
    preco_custo: '',
    preco_venda: '',
    estoque_minimo: '0',
    garantia_meses: '0',
    modelo: '',
    ativo: true,
  })

  const margemCalculada = (() => {
    const custo = parseFloat(form.preco_custo) || 0
    const venda = parseFloat(form.preco_venda) || 0
    if (custo <= 0) return 0
    return ((venda - custo) / custo) * 100
  })()

  async function carregarCategorias() {
    try {
      const res = await api.get<{ items: Categoria[] }>('/categorias?limit=100')
      setCategorias(res.items ?? [])
    } catch {
      // silencia erro de categorias — não bloqueia a listagem
    }
  }

  async function carregar(q = busca, cat = filtroCategoria, off = offset) {
    setCarregando(true)
    setErro('')
    try {
      let url = `/produtos?q=${encodeURIComponent(q)}&limit=${LIMITE}&offset=${off}`
      if (cat) url += `&categoria_id=${encodeURIComponent(cat)}`
      const res = await api.get<{ items: Produto[] }>(url)
      setItens(res.items ?? [])
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar produtos')
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => {
    carregarCategorias()
    carregar()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function abrirNovo() {
    setEditando(null)
    setForm({ categoria_id: '', descricao: '', preco_custo: '', preco_venda: '', estoque_minimo: '0', garantia_meses: '0', modelo: '', ativo: true })
    setErroModal('')
    setModalAberto(true)
  }

  function abrirEditar(p: Produto) {
    setEditando(p)
    setForm({
      categoria_id: p.categoria_id,
      descricao: p.descricao,
      preco_custo: String(p.preco_custo),
      preco_venda: String(p.preco_venda),
      estoque_minimo: String(p.estoque_minimo),
      garantia_meses: String(p.garantia_meses),
      modelo: p.modelo ?? '',
      ativo: p.ativo,
    })
    setErroModal('')
    setModalAberto(true)
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault()
    setSalvando(true)
    setErroModal('')
    const payload = {
      categoria_id: form.categoria_id,
      descricao: form.descricao,
      preco_custo: parseFloat(form.preco_custo),
      preco_venda: parseFloat(form.preco_venda),
      estoque_minimo: parseInt(form.estoque_minimo),
      garantia_meses: parseInt(form.garantia_meses),
      modelo: form.modelo,
      ativo: form.ativo,
    }
    try {
      if (editando) {
        await api.put(`/produtos/${editando.id}`, payload)
      } else {
        await api.post('/produtos', payload)
      }
      setModalAberto(false)
      carregar(busca, filtroCategoria, 0)
      setOffset(0)
    } catch (err: unknown) {
      setErroModal(err instanceof Error ? err.message : 'Erro ao salvar')
    } finally {
      setSalvando(false)
    }
  }

  function pesquisar(e: React.FormEvent) {
    e.preventDefault()
    setOffset(0)
    carregar(busca, filtroCategoria, 0)
  }

  function nomeCategoria(id: string) {
    return categorias.find(c => c.id === id)?.descricao ?? '—'
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto py-8 px-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Produtos</h1>
          <button
            onClick={abrirNovo}
            className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700"
          >
            Novo Produto
          </button>
        </div>

        <form onSubmit={pesquisar} className="flex gap-2 mb-4 flex-wrap">
          <input
            type="text"
            placeholder="Pesquisar por descrição…"
            value={busca}
            onChange={e => setBusca(e.target.value)}
            className="flex-1 min-w-48 border border-gray-300 rounded-md px-3 py-2 text-sm"
          />
          <select
            value={filtroCategoria}
            onChange={e => setFiltroCategoria(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="">Todas as categorias</option>
            {categorias.map(c => <option key={c.id} value={c.id}>{c.descricao}</option>)}
          </select>
          <button type="submit" className="bg-gray-200 px-4 py-2 rounded-md text-sm hover:bg-gray-300">
            Buscar
          </button>
        </form>

        {erro && <p className="text-red-600 text-sm mb-4">{erro}</p>}

        {carregando ? (
          <p className="text-gray-500 text-sm">Carregando…</p>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-500 uppercase tracking-wider text-xs">Descrição</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-500 uppercase tracking-wider text-xs">Categoria</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500 uppercase tracking-wider text-xs">Custo</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500 uppercase tracking-wider text-xs">Venda</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500 uppercase tracking-wider text-xs">Margem</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500 uppercase tracking-wider text-xs">Estoque</th>
                  <th className="px-4 py-3 text-center font-medium text-gray-500 uppercase tracking-wider text-xs">Status</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-500 uppercase tracking-wider text-xs">Ações</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {itens.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="px-6 py-8 text-center text-gray-400 text-sm">
                      Nenhum produto encontrado
                    </td>
                  </tr>
                ) : (
                  itens.map(p => {
                    const abaixoMin = p.estoque_atual < p.estoque_minimo
                    return (
                      <tr key={p.id} className="hover:bg-gray-50">
                        <td className="px-4 py-3 text-gray-900">
                          <div className="font-medium">{p.descricao}</div>
                          {p.modelo && <div className="text-xs text-gray-400">{p.modelo}</div>}
                        </td>
                        <td className="px-4 py-3 text-gray-600">{nomeCategoria(p.categoria_id)}</td>
                        <td className="px-4 py-3 text-right text-gray-700">{fmt(p.preco_custo)}</td>
                        <td className="px-4 py-3 text-right text-gray-700">{fmt(p.preco_venda)}</td>
                        <td className="px-4 py-3 text-right text-green-600 font-medium">
                          {p.margem_pct.toFixed(1)}%
                        </td>
                        <td className="px-4 py-3 text-right">
                          <span className={abaixoMin ? 'text-red-600 font-semibold' : 'text-gray-700'}>
                            {p.estoque_atual}
                          </span>
                          <span className="text-gray-400 text-xs"> / {p.estoque_minimo} mín</span>
                          {abaixoMin && (
                            <span className="ml-1 text-xs bg-red-100 text-red-700 px-1 rounded">⚠</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-center">
                          <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${p.ativo ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-500'}`}>
                            {p.ativo ? 'Ativo' : 'Inativo'}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <button
                            onClick={() => abrirEditar(p)}
                            className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
                          >
                            Editar
                          </button>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>
        )}

        {(itens.length === LIMITE || offset > 0) && (
          <div className="flex justify-between mt-4">
            <button
              disabled={offset === 0}
              onClick={() => { const off = Math.max(0, offset - LIMITE); setOffset(off); carregar(busca, filtroCategoria, off) }}
              className="px-4 py-2 text-sm bg-white border rounded-md disabled:opacity-40"
            >
              Anterior
            </button>
            <button
              disabled={itens.length < LIMITE}
              onClick={() => { const off = offset + LIMITE; setOffset(off); carregar(busca, filtroCategoria, off) }}
              className="px-4 py-2 text-sm bg-white border rounded-md disabled:opacity-40"
            >
              Próxima
            </button>
          </div>
        )}
      </div>

      {modalAberto && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between p-6 border-b sticky top-0 bg-white">
              <h2 className="text-lg font-semibold">{editando ? 'Editar Produto' : 'Novo Produto'}</h2>
              <button onClick={() => setModalAberto(false)} className="text-gray-400 hover:text-gray-600 text-xl">×</button>
            </div>
            <form onSubmit={salvar} className="p-6 space-y-4">
              {erroModal && <p className="text-red-600 text-sm">{erroModal}</p>}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Categoria *</label>
                <select
                  value={form.categoria_id}
                  onChange={e => setForm(f => ({ ...f, categoria_id: e.target.value }))}
                  required
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                >
                  <option value="">Selecione uma categoria</option>
                  {categorias.map(c => <option key={c.id} value={c.id}>{c.descricao}</option>)}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Descrição *</label>
                <input
                  type="text"
                  value={form.descricao}
                  onChange={e => setForm(f => ({ ...f, descricao: e.target.value }))}
                  required
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="Ex.: Capa iPhone 15 Pro"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Preço de Custo (R$) *</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={form.preco_custo}
                    onChange={e => setForm(f => ({ ...f, preco_custo: e.target.value }))}
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                    placeholder="0,00"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Preço de Venda (R$) *</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={form.preco_venda}
                    onChange={e => setForm(f => ({ ...f, preco_venda: e.target.value }))}
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                    placeholder="0,00"
                  />
                </div>
              </div>

              {(parseFloat(form.preco_custo) > 0 || parseFloat(form.preco_venda) > 0) && (
                <div className={`text-sm rounded-md px-3 py-2 ${margemCalculada > 0 ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                  Margem: <strong>{margemCalculada.toFixed(2)}%</strong>
                  {margemCalculada <= 0 && ' — preço de venda deve ser maior que o custo'}
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Estoque Mínimo</label>
                  <input
                    type="number"
                    min="0"
                    value={form.estoque_minimo}
                    onChange={e => setForm(f => ({ ...f, estoque_minimo: e.target.value }))}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Garantia (meses)</label>
                  <input
                    type="number"
                    min="0"
                    value={form.garantia_meses}
                    onChange={e => setForm(f => ({ ...f, garantia_meses: e.target.value }))}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Modelo de Celular</label>
                <input
                  type="text"
                  value={form.modelo}
                  onChange={e => setForm(f => ({ ...f, modelo: e.target.value }))}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="Ex.: iPhone 15, Samsung S24"
                />
              </div>

              {editando && (
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="ativo"
                    checked={form.ativo}
                    onChange={e => setForm(f => ({ ...f, ativo: e.target.checked }))}
                    className="h-4 w-4 rounded border-gray-300 text-indigo-600"
                  />
                  <label htmlFor="ativo" className="text-sm text-gray-700">Produto ativo</label>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setModalAberto(false)}
                  className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={salvando}
                  className="px-4 py-2 text-sm text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50"
                >
                  {salvando ? 'Salvando…' : 'Salvar'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
