import { useEffect, useState } from 'react'
import { CheckCircle, Eye, Plus, Trash2 } from 'lucide-react'
import { api, ApiError } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge, type BadgeTone } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Field, inputClasses } from '@/components/ui/field'

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

  const colunas: Column<Compra>[] = [
    { header: 'Data', sortAccessor: (c) => c.dt_compra, cell: (c) => <span className="whitespace-nowrap text-gray-900">{c.dt_compra}</span> },
    { header: 'Fornecedor', sortAccessor: (c) => nomeFornecedor(c.fornecedor_id), cell: (c) => <span className="text-gray-900">{nomeFornecedor(c.fornecedor_id)}</span> },
    { header: 'NF', hideBelow: 'sm', sortAccessor: (c) => c.nf, cell: (c) => <span className="text-gray-500">{c.nf || '—'}</span> },
    { header: 'Total', align: 'right', sortAccessor: (c) => c.valor_total, cell: (c) => <span className="font-medium text-gray-900 whitespace-nowrap">{brl(c.valor_total)}</span> },
    { header: 'Status', sortAccessor: (c) => STATUS_LABEL[c.status], cell: (c) => <StatusBadge tone={STATUS_TONE[c.status] ?? 'neutral'}>{STATUS_LABEL[c.status]}</StatusBadge> },
    {
      header: '',
      align: 'right',
      cell: (c) => (
        <div className="flex items-center justify-end gap-2">
          <button title="Ver detalhes" onClick={() => verDetalhe(c.id)} className="text-gray-400 hover:text-gray-700">
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
      ),
    },
  ]

  return (
    <PageShell
      title="Compras"
      subtitle="Entrada de mercadorias"
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirModal}>
          <Plus className="h-4 w-4" />
          Nova Compra
        </Button>
      }
    >
      {erro && <p className="text-sm text-red-600">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={compras}
        rowKey={(c) => c.id}
        loading={carregando}
        empty="Nenhuma compra cadastrada ainda."
      />

      {/* ── Modal: nova compra ─────────────────────────────────────────────── */}
      {modalAberto && (
        <Modal title="Nova Compra" onClose={() => setModalAberto(false)} maxWidth="max-w-3xl">
          <div className="px-6 py-4 space-y-5">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="sm:col-span-2">
                <Field label="Fornecedor *">
                  <select value={fornecedorID} onChange={e => setFornecedorID(e.target.value)} className={inputClasses()}>
                    <option value="">Selecione…</option>
                    {fornecedores.map(f => (
                      <option key={f.id} value={f.id}>{f.nome_fantasia || f.razao_social}</option>
                    ))}
                  </select>
                </Field>
              </div>
              <Field label="Data *">
                <input type="date" value={dtCompra} onChange={e => setDtCompra(e.target.value)} className={inputClasses()} />
              </Field>
              <div className="sm:col-span-3">
                <Field label="Nota Fiscal">
                  <input
                    type="text"
                    value={nf}
                    onChange={e => setNf(e.target.value)}
                    placeholder="Número da NF (opcional)"
                    className={inputClasses()}
                  />
                </Field>
              </div>
            </div>

            {/* itens */}
            <div>
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">Itens</span>
                <button
                  onClick={() => setItensForm(prev => [...prev, itemVazio()])}
                  className="flex items-center gap-1 text-xs text-gray-700 hover:text-gray-900"
                >
                  <Plus className="h-3 w-3" /> Adicionar item
                </button>
              </div>

              <div className="space-y-3">
                {itensForm.map((it, idx) => (
                  <div key={idx} className="grid grid-cols-12 gap-2 rounded-lg border border-gray-200 p-3">
                    <div className="col-span-4">
                      <label className="mb-0.5 block text-xs text-gray-500">Produto</label>
                      <select
                        value={it.produto_id}
                        onChange={e => atualizarItem(idx, 'produto_id', e.target.value)}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-gray-900 focus:outline-none"
                      >
                        <option value="">Selecione…</option>
                        {produtos.map(p => (
                          <option key={p.id} value={p.id}>{p.descricao}{p.modelo ? ` (${p.modelo})` : ''}</option>
                        ))}
                      </select>
                    </div>
                    <div className="col-span-2">
                      <label className="mb-0.5 block text-xs text-gray-500">Qtd</label>
                      <input
                        type="number" min="1"
                        value={it.quantidade}
                        onChange={e => atualizarItem(idx, 'quantidade', Number(e.target.value))}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-gray-900 focus:outline-none"
                      />
                    </div>
                    <div className="col-span-2">
                      <label className="mb-0.5 block text-xs text-gray-500">Custo (R$)</label>
                      <input
                        type="number" min="0" step="0.01"
                        value={it.preco_compra}
                        onChange={e => atualizarItem(idx, 'preco_compra', Number(e.target.value))}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-gray-900 focus:outline-none"
                      />
                    </div>
                    <div className="col-span-2">
                      <label className="mb-0.5 block text-xs text-gray-500">Venda (R$)</label>
                      <input
                        type="number" min="0" step="0.01"
                        value={it.preco_venda}
                        onChange={e => atualizarItem(idx, 'preco_venda', Number(e.target.value))}
                        className="w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-gray-900 focus:outline-none"
                      />
                    </div>
                    <div className="col-span-1">
                      <label className="mb-0.5 block text-xs text-gray-500">Margem</label>
                      <span className={`block py-1 text-xs font-medium ${margemItem(it) > 0 ? 'text-green-700' : 'text-gray-400'}`}>
                        {margemItem(it).toFixed(0)}%
                      </span>
                    </div>
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

              <div className="mt-3 text-right text-sm font-medium text-gray-900">
                Total: {brl(totalForm())}
              </div>
            </div>

            {erroForm && <p className="text-sm text-red-600">{erroForm}</p>}

            <div className="flex justify-end gap-3 pt-1">
              <Button variant="secondary" onClick={() => setModalAberto(false)}>Cancelar</Button>
              <Button onClick={salvarCompra} disabled={salvando}>{salvando ? 'Salvando…' : 'Criar Rascunho'}</Button>
            </div>
          </div>
        </Modal>
      )}

      {/* ── Modal: detalhe da compra ──────────────────────────────────────── */}
      {detalhe && (
        <Modal title="Detalhe da Compra" onClose={() => setDetalhe(null)}>
          <div className="px-6 py-4 space-y-4">
            <div className="flex items-center justify-between">
              <p className="font-mono text-xs text-gray-500">{detalhe.id}</p>
              <StatusBadge tone={STATUS_TONE[detalhe.status] ?? 'neutral'}>{STATUS_LABEL[detalhe.status]}</StatusBadge>
            </div>

            <dl className="grid grid-cols-3 gap-3 text-sm">
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
              <thead className="bg-gray-50">
                <tr>
                  {['Produto', 'Qtd', 'Custo', 'Venda', 'Margem'].map(h => (
                    <th key={h} className="px-3 py-2 text-left text-xs font-medium text-gray-600">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {(detalhe.itens ?? []).map(it => (
                  <tr key={it.id}>
                    <td className="px-3 py-2 text-gray-900">{nomeProduto(it.produto_id)}</td>
                    <td className="px-3 py-2 text-gray-900">{it.quantidade}</td>
                    <td className="px-3 py-2 text-gray-900">{brl(it.preco_compra)}</td>
                    <td className="px-3 py-2 text-gray-900">{brl(it.preco_venda)}</td>
                    <td className="px-3 py-2 text-green-700">{it.margem.toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr>
                  <td colSpan={5} className="px-3 pt-3 text-right text-sm font-semibold text-gray-900">
                    Total: {brl(detalhe.valor_total)}
                  </td>
                </tr>
              </tfoot>
            </table>

            <div className="flex justify-between pt-1">
              {detalhe.status === 'RASCUNHO' && (
                <Button
                  variant="success"
                  disabled={confirmando === detalhe.id}
                  onClick={async () => {
                    await confirmarCompra(detalhe.id)
                    setDetalhe(null)
                  }}
                >
                  <CheckCircle className="h-4 w-4" /> Confirmar Compra
                </Button>
              )}
              <Button variant="secondary" className="ml-auto" onClick={() => setDetalhe(null)}>Fechar</Button>
            </div>
          </div>
        </Modal>
      )}
    </PageShell>
  )
}
