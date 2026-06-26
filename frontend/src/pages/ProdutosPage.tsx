import { useEffect, useState } from 'react'
import { Pencil, Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Field, inputClasses } from '@/components/ui/field'

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
      const res = await api.get<{ items: Categoria[] }>('/api/v1/categorias?limit=100')
      setCategorias(res.items ?? [])
    } catch {
      // silencia erro de categorias — não bloqueia a listagem
    }
  }

  async function carregar(q = busca, cat = filtroCategoria, off = offset) {
    setCarregando(true)
    setErro('')
    try {
      let url = `/api/v1/produtos?q=${encodeURIComponent(q)}&limit=${LIMITE}&offset=${off}`
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
        await api.put(`/api/v1/produtos/${editando.id}`, payload)
      } else {
        await api.post('/api/v1/produtos', payload)
      }
      setModalAberto(false)
      setOffset(0)
      carregar(busca, filtroCategoria, 0)
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

  const colunas: Column<Produto>[] = [
    {
      header: 'Descrição',
      sortAccessor: (p) => p.descricao,
      cell: (p) => (
        <div>
          <p className="font-medium text-gray-900">{p.descricao}</p>
          {p.modelo && <p className="text-xs text-gray-400">{p.modelo}</p>}
        </div>
      ),
    },
    { header: 'Categoria', sortAccessor: (p) => nomeCategoria(p.categoria_id), cell: (p) => <span className="text-gray-600">{nomeCategoria(p.categoria_id)}</span>, hideBelow: 'md' },
    { header: 'Custo', align: 'right', hideBelow: 'sm', sortAccessor: (p) => p.preco_custo, cell: (p) => fmt(p.preco_custo) },
    { header: 'Venda', align: 'right', sortAccessor: (p) => p.preco_venda, cell: (p) => fmt(p.preco_venda) },
    { header: 'Margem', align: 'right', hideBelow: 'sm', sortAccessor: (p) => p.margem_pct, cell: (p) => <span className="text-green-600 font-medium">{p.margem_pct.toFixed(1)}%</span> },
    {
      header: 'Estoque',
      align: 'right',
      sortAccessor: (p) => p.estoque_atual,
      cell: (p) => {
        const abaixoMin = p.estoque_atual < p.estoque_minimo
        return (
          <span>
            <span className={abaixoMin ? 'text-red-600 font-semibold' : 'text-gray-700'}>{p.estoque_atual}</span>
            <span className="text-gray-400 text-xs"> / {p.estoque_minimo} mín</span>
            {abaixoMin && <span className="ml-1 text-xs bg-red-100 text-red-700 px-1 rounded">⚠</span>}
          </span>
        )
      },
    },
    {
      header: 'Status',
      align: 'center',
      hideBelow: 'md',
      sortAccessor: (p) => (p.ativo ? 1 : 0),
      cell: (p) => <StatusBadge tone={p.ativo ? 'success' : 'neutral'}>{p.ativo ? 'Ativo' : 'Inativo'}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (p) => (
        <button onClick={() => abrirEditar(p)} className="text-gray-400 hover:text-gray-700" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell
      title="Produtos"
      subtitle="Catálogo de produtos"
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirNovo}>
          <Plus className="w-4 h-4" />
          Novo Produto
        </Button>
      }
    >
      <form onSubmit={pesquisar} className="flex gap-2 flex-wrap">
        <input
          type="text"
          placeholder="Pesquisar por descrição…"
          value={busca}
          onChange={e => setBusca(e.target.value)}
          className={inputClasses() + ' flex-1 min-w-48'}
        />
        <select
          value={filtroCategoria}
          onChange={e => setFiltroCategoria(e.target.value)}
          className={inputClasses() + ' w-auto'}
        >
          <option value="">Todas as categorias</option>
          {categorias.map(c => <option key={c.id} value={c.id}>{c.descricao}</option>)}
        </select>
        <Button type="submit" variant="secondary">Buscar</Button>
      </form>

      {erro && <p className="text-sm text-red-600">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={itens}
        rowKey={(p) => p.id}
        loading={carregando}
        empty="Nenhum produto encontrado."
      />

      {(itens.length === LIMITE || offset > 0) && (
        <div className="flex justify-between">
          <Button
            variant="secondary"
            disabled={offset === 0}
            onClick={() => { const off = Math.max(0, offset - LIMITE); setOffset(off); carregar(busca, filtroCategoria, off) }}
          >
            Anterior
          </Button>
          <Button
            variant="secondary"
            disabled={itens.length < LIMITE}
            onClick={() => { const off = offset + LIMITE; setOffset(off); carregar(busca, filtroCategoria, off) }}
          >
            Próxima
          </Button>
        </div>
      )}

      {modalAberto && (
        <Modal title={editando ? 'Editar Produto' : 'Novo Produto'} onClose={() => setModalAberto(false)} maxWidth="max-w-lg">
          <form onSubmit={salvar} className="px-6 py-4 space-y-4">
            {erroModal && <p className="text-sm text-red-600">{erroModal}</p>}

            <Field label="Categoria *">
              <select
                value={form.categoria_id}
                onChange={e => setForm(f => ({ ...f, categoria_id: e.target.value }))}
                required
                className={inputClasses()}
              >
                <option value="">Selecione uma categoria</option>
                {categorias.map(c => <option key={c.id} value={c.id}>{c.descricao}</option>)}
              </select>
            </Field>

            <Field label="Descrição *">
              <input
                type="text"
                value={form.descricao}
                onChange={e => setForm(f => ({ ...f, descricao: e.target.value }))}
                required
                className={inputClasses()}
                placeholder="Ex.: Capa iPhone 15 Pro"
              />
            </Field>

            <div className="grid grid-cols-2 gap-4">
              <Field label="Preço de Custo (R$) *">
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={form.preco_custo}
                  onChange={e => setForm(f => ({ ...f, preco_custo: e.target.value }))}
                  required
                  className={inputClasses()}
                  placeholder="0,00"
                />
              </Field>
              <Field label="Preço de Venda (R$) *">
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={form.preco_venda}
                  onChange={e => setForm(f => ({ ...f, preco_venda: e.target.value }))}
                  required
                  className={inputClasses()}
                  placeholder="0,00"
                />
              </Field>
            </div>

            {(parseFloat(form.preco_custo) > 0 || parseFloat(form.preco_venda) > 0) && (
              <div className={`text-sm rounded-md px-3 py-2 ${margemCalculada > 0 ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                Margem: <strong>{margemCalculada.toFixed(2)}%</strong>
                {margemCalculada <= 0 && ' — preço de venda deve ser maior que o custo'}
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <Field label="Estoque Mínimo">
                <input
                  type="number"
                  min="0"
                  value={form.estoque_minimo}
                  onChange={e => setForm(f => ({ ...f, estoque_minimo: e.target.value }))}
                  className={inputClasses()}
                />
              </Field>
              <Field label="Garantia (meses)">
                <input
                  type="number"
                  min="0"
                  value={form.garantia_meses}
                  onChange={e => setForm(f => ({ ...f, garantia_meses: e.target.value }))}
                  className={inputClasses()}
                />
              </Field>
            </div>

            <Field label="Modelo de Celular">
              <input
                type="text"
                value={form.modelo}
                onChange={e => setForm(f => ({ ...f, modelo: e.target.value }))}
                className={inputClasses()}
                placeholder="Ex.: iPhone 15, Samsung S24"
              />
            </Field>

            {editando && (
              <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.ativo}
                  onChange={e => setForm(f => ({ ...f, ativo: e.target.checked }))}
                  className="rounded border-gray-300"
                />
                Produto ativo
              </label>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <Button type="button" variant="secondary" onClick={() => setModalAberto(false)}>Cancelar</Button>
              <Button type="submit" disabled={salvando}>{salvando ? 'Salvando…' : 'Salvar'}</Button>
            </div>
          </form>
        </Modal>
      )}
    </PageShell>
  )
}
