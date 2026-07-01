import { useEffect, useState } from 'react'
import { Pencil, Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Field, inputClasses } from '@/components/ui/field'
import { cn } from '@/lib/utils'

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
      // ignore
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
        <div className="flex flex-col">
          <p className="font-bold text-foreground leading-tight">{p.descricao}</p>
          {p.modelo && <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-wider mt-1">{p.modelo}</p>}
        </div>
      ),
    },
    { header: 'Categoria', sortAccessor: (p) => nomeCategoria(p.categoria_id), cell: (p) => <span className="text-muted-foreground font-medium">{nomeCategoria(p.categoria_id)}</span>, hideBelow: 'md' },
    { header: 'Custo', align: 'right', hideBelow: 'sm', sortAccessor: (p) => p.preco_custo, cell: (p) => fmt(p.preco_custo), isTechnical: true },
    { header: 'Venda', align: 'right', sortAccessor: (p) => p.preco_venda, cell: (p) => fmt(p.preco_venda), isTechnical: true },
    { 
      header: 'Margem', 
      align: 'right', 
      hideBelow: 'sm', 
      sortAccessor: (p) => p.margem_pct, 
      cell: (p) => (
        <span className={cn('font-black font-mono text-xs', p.margem_pct > 0 ? 'text-green-700 dark:text-green-400' : 'text-destructive')}>
          {p.margem_pct.toFixed(1)}%
        </span>
      ),
      isTechnical: true 
    },
    {
      header: 'Estoque',
      align: 'right',
      sortAccessor: (p) => p.estoque_atual,
      cell: (p) => {
        const abaixoMin = p.estoque_atual < p.estoque_minimo
        return (
          <div className="flex flex-col items-end gap-1">
            <span className={cn('font-mono font-black', abaixoMin ? 'text-destructive' : 'text-foreground')}>
              {p.estoque_atual}
            </span>
            <span className="text-[9px] text-muted-foreground uppercase font-black tracking-tighter">
              mín: {p.estoque_minimo}
            </span>
          </div>
        )
      },
      isTechnical: true
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
        <button onClick={() => abrirEditar(p)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted active:scale-90" aria-label="Editar" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell
      title="Produtos"
      subtitle="Gerenciamento de estoque e catálogo técnico."
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirNovo}>
          <Plus className="w-4 h-4" />
          Novo Produto
        </Button>
      }
    >
      <div className="flex items-center gap-3 bg-card p-4 rounded-2xl border border-border shadow-sm animate-in fade-in duration-500">
        <form onSubmit={pesquisar} className="flex flex-1 flex-wrap gap-3">
          <input
            type="text"
            placeholder="Pesquisar por descrição ou modelo…"
            value={busca}
            onChange={e => setBusca(e.target.value)}
            className={cn(inputClasses(), 'flex-1 min-w-[12rem]')}
          />
          <select
            aria-label="Filtrar por categoria"
            value={filtroCategoria}
            onChange={e => setFiltroCategoria(e.target.value)}
            className={cn(inputClasses(), 'w-auto hidden sm:block')}
          >
            <option value="">Todas as categorias</option>
            {categorias.map(c => <option key={c.id} value={c.id}>{c.descricao}</option>)}
          </select>
          <Button type="submit" variant="secondary" className="px-8 h-10">Buscar</Button>
        </form>
      </div>

      {erro && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={itens}
        rowKey={(p) => p.id}
        loading={carregando}
        empty="Nenhum produto encontrado no catálogo."
      />

      {(itens.length === LIMITE || offset > 0) && (
        <div className="flex justify-between items-center bg-card p-3 rounded-xl border border-border shadow-sm">
          <Button
            variant="secondary"
            size="sm"
            disabled={offset === 0}
            onClick={() => { const off = Math.max(0, offset - LIMITE); setOffset(off); carregar(busca, filtroCategoria, off) }}
          >
            Anterior
          </Button>
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em]">Exibindo {itens.length} resultados</span>
          <Button
            variant="secondary"
            size="sm"
            disabled={itens.length < LIMITE}
            onClick={() => { const off = offset + LIMITE; setOffset(off); carregar(busca, filtroCategoria, off) }}
          >
            Próxima
          </Button>
        </div>
      )}

      {modalAberto && (
        <Modal title={editando ? 'Editar Especificações' : 'Cadastrar Novo Produto'} onClose={() => setModalAberto(false)} maxWidth="max-w-2xl">
          <form onSubmit={salvar} className="px-8 py-8 space-y-8 animate-in fade-in duration-300">
            {erroModal && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erroModal}</p>}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-8">
              <Field label="Categoria do Produto *">
                <select
                  value={form.categoria_id}
                  onChange={e => setForm(f => ({ ...f, categoria_id: e.target.value }))}
                  required
                  className={inputClasses()}
                >
                  <option value="">Selecione uma categoria…</option>
                  {categorias.map(c => <option key={c.id} value={c.id}>{c.descricao}</option>)}
                </select>
              </Field>

              <Field label="Modelo de Referência">
                <input
                  type="text"
                  value={form.modelo}
                  onChange={e => setForm(f => ({ ...f, modelo: e.target.value }))}
                  className={inputClasses()}
                  placeholder="Ex: iPhone 15 Pro, S24 Ultra"
                />
              </Field>
            </div>

            <Field label="Descrição Completa *">
              <input
                type="text"
                value={form.descricao}
                onChange={e => setForm(f => ({ ...f, descricao: e.target.value }))}
                required
                className={inputClasses()}
                placeholder="Ex: Capa de Silicone com MagSafe"
              />
            </Field>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-8">
              <Field label="Preço de Custo (BRL) *">
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={form.preco_custo}
                  onChange={e => setForm(f => ({ ...f, preco_custo: e.target.value }))}
                  required
                  className={inputClasses() + ' font-mono'}
                  placeholder="0,00"
                />
              </Field>
              <Field label="Preço de Venda (BRL) *">
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={form.preco_venda}
                  onChange={e => setForm(f => ({ ...f, preco_venda: e.target.value }))}
                  required
                  className={inputClasses() + ' font-mono'}
                  placeholder="0,00"
                />
              </Field>
            </div>

            {(parseFloat(form.preco_custo) > 0 || parseFloat(form.preco_venda) > 0) && (
              <div className={cn(
                "p-4 rounded-2xl border flex justify-between items-center",
                margemCalculada > 0 ? "bg-green-500/5 border-green-500/20 text-green-700 dark:text-green-400" : "bg-destructive/5 border-destructive/20 text-destructive"
              )}>
                <span className="text-[10px] font-black uppercase tracking-widest leading-none">Margem de Lucro Estimada</span>
                <strong className="text-xl font-black font-mono tracking-tighter">{margemCalculada.toFixed(2)}%</strong>
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-8">
              <Field label="Estoque Mínimo (Alerta)">
                <input
                  type="number"
                  min="0"
                  value={form.estoque_minimo}
                  onChange={e => setForm(f => ({ ...f, estoque_minimo: e.target.value }))}
                  className={inputClasses() + ' font-mono'}
                />
              </Field>
              <Field label="Garantia Técnica (Meses)">
                <input
                  type="number"
                  min="0"
                  value={form.garantia_meses}
                  onChange={e => setForm(f => ({ ...f, garantia_meses: e.target.value }))}
                  className={inputClasses() + ' font-mono'}
                />
              </Field>
            </div>

            {editando && (
              <label className="flex items-center gap-3 text-xs font-bold text-muted-foreground uppercase tracking-widest cursor-pointer hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={form.ativo}
                  onChange={e => setForm(f => ({ ...f, ativo: e.target.checked }))}
                  className="w-4 h-4 rounded border-border bg-muted/20 text-primary focus:ring-primary"
                />
                Produto disponível para venda
              </label>
            )}

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button type="button" variant="secondary" onClick={() => setModalAberto(false)}>Descartar</Button>
              <Button type="submit" disabled={salvando} className="min-w-32">
                {salvando ? 'Salvando…' : editando ? 'Atualizar Dados' : 'Criar Produto'}
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </PageShell>
  )
}
