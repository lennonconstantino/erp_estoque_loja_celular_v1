import { useEffect, useState } from 'react'
import { CheckCircle, Eye, Plus, Trash2 } from 'lucide-react'
import { api, ApiError } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge, type BadgeTone } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Field, inputClasses, inputClassesCompact, compactLabelClass } from '@/components/ui/field'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

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

const STATUS_TONE: Record<string, BadgeTone> = {
  RASCUNHO: 'warning',
  CONFIRMADA: 'success',
  CANCELADA: 'danger',
}

const STATUS_LABEL: Record<string, string> = {
  RASCUNHO: 'Rascunho',
  CONFIRMADA: 'Confirmada',
  CANCELADA: 'Cancelada',
}

function brl(v: number) {
  return v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

// ── componente principal ───────────────────────────────────────────────────────

export default function ComprasPage() {
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
  }, [])

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
    setConfirmando(id)
    try {
      await api.post(`/api/v1/compras/${id}/confirmar`, {})
      toast.success('Compra confirmada e estoque atualizado!')
      void carregarCompras()
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'Erro ao confirmar compra.')
    } finally {
      setConfirmando(null)
    }
  }

  async function verDetalhe(id: string) {
    try {
      const c = await api.get<Compra>(`/api/v1/compras/${id}`)
      setDetalhe(c)
    } catch {
      toast.error('Não foi possível carregar o detalhe da compra.')
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

  const colunas: Column<Compra>[] = [
    { header: 'Data', sortAccessor: (c) => c.dt_compra, cell: (c) => <span className="font-bold text-foreground">{c.dt_compra}</span> },
    { header: 'Fornecedor', sortAccessor: (c) => nomeFornecedor(c.fornecedor_id), cell: (c) => <span className="text-muted-foreground">{nomeFornecedor(c.fornecedor_id)}</span> },
    { header: 'NF', hideBelow: 'sm', sortAccessor: (c) => c.nf, cell: (c) => <span className="text-muted-foreground font-mono text-xs">{c.nf || '—'}</span>, isTechnical: true },
    { header: 'Total', align: 'right', sortAccessor: (c) => c.valor_total, cell: (c) => <span className="font-black text-foreground font-mono">{brl(c.valor_total)}</span>, isTechnical: true },
    { header: 'Status', sortAccessor: (c) => STATUS_LABEL[c.status], cell: (c) => <StatusBadge tone={STATUS_TONE[c.status] ?? 'neutral'}>{STATUS_LABEL[c.status]}</StatusBadge> },
    {
      header: '',
      align: 'right',
      cell: (c) => (
        <div className="flex items-center justify-end gap-2">
          <button aria-label="Ver detalhes" title="Ver detalhes" onClick={() => verDetalhe(c.id)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted">
            <Eye className="h-4 w-4" />
          </button>
          {c.status === 'RASCUNHO' && (
            <button
              title="Confirmar compra"
              disabled={confirmando === c.id}
              onClick={() => confirmarCompra(c.id)}
              className="text-muted-foreground hover:text-green-500 transition-colors p-2 rounded-full hover:bg-green-500/10 disabled:opacity-50"
            >
              <CheckCircle className="h-4 w-4" />
            </button>
          )}
        </div>
      ),
    },
  ]

  return (
    <PageShell
      title="Compras"
      subtitle="Entrada de mercadorias no catálogo."
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirModal}>
          <Plus className="h-4 w-4" />
          Registrar Compra
        </Button>
      }
    >
      {erro && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={compras}
        rowKey={(c) => c.id}
        loading={carregando}
        empty="Nenhuma compra cadastrada ainda."
      />

      {/* ── Modal: nova compra ─────────────────────────────────────────────── */}
      {modalAberto && (
        <Modal title="Nova Entrada de Mercadoria" onClose={() => setModalAberto(false)} maxWidth="max-w-4xl">
          <div className="px-8 py-8 space-y-8 animate-in fade-in duration-300">
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
              <div className="sm:col-span-2">
                <Field label="Fornecedor Selecionado *">
                  <select value={fornecedorID} onChange={e => setFornecedorID(e.target.value)} className={inputClasses()}>
                    <option value="">Selecione um fornecedor cadastrado…</option>
                    {fornecedores.map(f => (
                      <option key={f.id} value={f.id}>{f.nome_fantasia || f.razao_social}</option>
                    ))}
                  </select>
                </Field>
              </div>
              <Field label="Data de Emissão *">
                <input type="date" value={dtCompra} onChange={e => setDtCompra(e.target.value)} className={inputClasses()} />
              </Field>
              <div className="sm:col-span-3">
                <Field label="Número da Nota Fiscal (NF-e / NFC-e)">
                  <input
                    type="text"
                    value={nf}
                    onChange={e => setNf(e.target.value)}
                    placeholder="Chave de acesso ou número da nota"
                    className={inputClasses()}
                  />
                </Field>
              </div>
            </div>

            {/* itens */}
            <div className="space-y-4">
              <div className="flex items-center justify-between border-b border-border pb-2">
                <span className="text-[10px] font-black text-muted-foreground uppercase tracking-widest leading-none">Itens da Remessa</span>
                <button
                  onClick={() => setItensForm(prev => [...prev, itemVazio()])}
                  className="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-primary hover:opacity-70 transition-opacity"
                >
                  <Plus className="h-3 w-3" /> Incluir Produto
                </button>
              </div>

              <div className="space-y-4">
                {itensForm.map((it, idx) => (
                  <div key={idx} className="grid grid-cols-12 gap-4 bg-muted/5 rounded-2xl border border-border/50 p-4 group hover:bg-muted/10 transition-colors">
                    <div className="col-span-12 lg:col-span-4">
                      <label className={compactLabelClass}>Produto</label>
                      <select
                        value={it.produto_id}
                        onChange={e => atualizarItem(idx, 'produto_id', e.target.value)}
                        className={inputClassesCompact()}
                      >
                        <option value="">Selecione…</option>
                        {produtos.map(p => (
                          <option key={p.id} value={p.id}>{p.descricao}{p.modelo ? ` (${p.modelo})` : ''}</option>
                        ))}
                      </select>
                    </div>
                    <div className="col-span-3 lg:col-span-2">
                      <label className={compactLabelClass}>Qtd</label>
                      <input
                        type="number" min="1"
                        value={it.quantidade}
                        onChange={e => atualizarItem(idx, 'quantidade', Number(e.target.value))}
                        className={cn(inputClassesCompact(), 'font-mono')}
                      />
                    </div>
                    <div className="col-span-4 lg:col-span-2">
                      <label className={compactLabelClass}>Custo (Unit)</label>
                      <input
                        type="number" min="0" step="0.01"
                        value={it.preco_compra}
                        onChange={e => atualizarItem(idx, 'preco_compra', Number(e.target.value))}
                        className={cn(inputClassesCompact(), 'font-mono')}
                      />
                    </div>
                    <div className="col-span-4 lg:col-span-2">
                      <label className={compactLabelClass}>Venda (Sugerido)</label>
                      <input
                        type="number" min="0" step="0.01"
                        value={it.preco_venda}
                        onChange={e => atualizarItem(idx, 'preco_venda', Number(e.target.value))}
                        className={cn(inputClassesCompact(), 'font-mono')}
                      />
                    </div>
                    <div className="col-span-1 flex flex-col items-center justify-center">
                      <label className={compactLabelClass}>Margem</label>
                      <span className={cn('text-[10px] font-black font-mono', margemItem(it) > 0 ? 'text-green-700 dark:text-green-400' : 'text-muted-foreground/30')}>
                        {margemItem(it).toFixed(0)}%
                      </span>
                    </div>
                    <div className="col-span-1 flex items-end justify-center pb-1">
                      <button
                        disabled={itensForm.length === 1}
                        onClick={() => removerItem(idx)}
                        className="text-muted-foreground hover:text-destructive transition-colors p-1.5 rounded-full hover:bg-destructive/10 disabled:opacity-30 active:scale-90"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              <div className="bg-muted/30 rounded-2xl p-6 flex justify-between items-end border border-border shadow-inner">
                <span className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em]">Total Estimado</span>
                <span className="text-3xl font-black text-foreground font-mono tracking-tighter">{brl(totalForm())}</span>
              </div>
            </div>

            {erroForm && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erroForm}</p>}

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button variant="secondary" onClick={() => setModalAberto(false)}>Cancelar</Button>
              <Button onClick={salvarCompra} disabled={salvando} className="min-w-40">
                {salvando ? 'Salvando Remessa…' : 'Salvar como Rascunho'}
              </Button>
            </div>
          </div>
        </Modal>
      )}

      {/* ── Modal: detalhe da compra ──────────────────────────────────────── */}
      {detalhe && (
        <Modal title="Comprovante de Entrada" onClose={() => setDetalhe(null)} maxWidth="max-w-3xl">
          <div className="px-8 py-8 space-y-8 animate-in fade-in duration-300">
            <div className="grid grid-cols-2 lg:grid-cols-3 gap-8 text-sm">
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Fornecedor</p>
                <p className="text-foreground font-bold">{nomeFornecedor(detalhe.fornecedor_id)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Data Entrada</p>
                <p className="text-foreground font-bold">{detalhe.dt_compra}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Nota Fiscal</p>
                <p className="text-foreground font-bold font-mono uppercase text-xs">{detalhe.nf || 'NÃO INFORMADA'}</p>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Status</p>
                <StatusBadge tone={STATUS_TONE[detalhe.status] ?? 'neutral'}>{STATUS_LABEL[detalhe.status]}</StatusBadge>
              </div>
              <div className="space-y-1 col-span-2">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Identificador Interno</p>
                <p className="text-muted-foreground font-mono text-[10px] truncate">{detalhe.id}</p>
              </div>
            </div>

            <div className="space-y-4">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest border-b border-border pb-2">Produtos Recebidos</p>
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-muted-foreground font-black uppercase tracking-tighter">
                    <th className="py-2 text-left">Descrição</th>
                    <th className="py-2 text-right">Qtd</th>
                    <th className="py-2 text-right">Custo</th>
                    <th className="py-2 text-right">Venda</th>
                    <th className="py-2 text-right">Margem</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/50">
                  {(detalhe.itens ?? []).map(it => (
                    <tr key={it.id} className="group hover:bg-muted/5 transition-colors">
                      <td className="py-3 text-foreground font-bold">{nomeProduto(it.produto_id)}</td>
                      <td className="py-3 text-right font-black font-mono">{it.quantidade}</td>
                      <td className="py-3 text-right text-muted-foreground font-mono">{brl(it.preco_compra)}</td>
                      <td className="py-3 text-right text-muted-foreground font-mono">{brl(it.preco_venda)}</td>
                      <td className="py-3 text-right font-black font-mono text-green-700 dark:text-green-400">{it.margem.toFixed(1)}%</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="bg-muted/30 rounded-2xl p-8 flex justify-between items-end border border-border shadow-inner">
               <span className="text-xs font-black text-foreground uppercase tracking-[0.2em]">Total da Remessa</span>
               <span className="text-4xl font-black text-primary font-mono tracking-tighter">{brl(detalhe.valor_total)}</span>
            </div>

            <div className="flex justify-end gap-3 border-t border-border pt-6">
              <Button variant="secondary" onClick={() => setDetalhe(null)}>Fechar</Button>
              {detalhe.status === 'RASCUNHO' && (
                <Button
                  variant="primary"
                  disabled={confirmando === detalhe.id}
                  onClick={async () => {
                    await confirmarCompra(detalhe.id)
                    setDetalhe(null)
                  }}
                  className="min-w-48"
                >
                  <CheckCircle className="h-4 w-4" /> Confirmar e Dar Entrada
                </Button>
              )}
            </div>
          </div>
        </Modal>
      )}
    </PageShell>
  )
}
