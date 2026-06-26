import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { cn } from '@/lib/utils'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge, type BadgeTone } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Field, inputClasses } from '@/components/ui/field'

interface Produto {
  id: string
  descricao: string
  modelo: string
  estoque_atual: number
  disponivel: boolean
}

interface Movimentacao {
  id: string
  produto_id: string
  tipo: string
  quantidade: number
  saldo_antes: number
  saldo_depois: number
  origem_tipo: string
  criado_em: string
}

interface Ajuste {
  id: string
  produto_id: string
  qtd_entrada: number
  qtd_saida: number
  motivo: string
  responsavel_id: string
  criado_em: string
}

const LIMITE = 50

const rotulaTipo: Record<string, string> = {
  COMPRA: 'Compra',
  VENDA: 'Venda',
  AJUSTE_ENTRADA: 'Ajuste Entrada',
  AJUSTE_SAIDA: 'Ajuste Saída',
}

const toneTipo: Record<string, BadgeTone> = {
  COMPRA: 'neutral',
  VENDA: 'danger',
  AJUSTE_ENTRADA: 'success',
  AJUSTE_SAIDA: 'warning',
}

export default function AjustesEstoquePage() {
  const [produtos, setProdutos] = useState<Produto[]>([])
  const [produtoSel, setProdutoSel] = useState<string>('')
  const [aba, setAba] = useState<'movimentacoes' | 'ajustes'>('movimentacoes')
  const [movimentacoes, setMovimentacoes] = useState<Movimentacao[]>([])
  const [ajustes, setAjustes] = useState<Ajuste[]>([])
  const [offset, setOffset] = useState(0)
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  // Modal de novo ajuste
  const [modalAberto, setModalAberto] = useState(false)
  const [tipoAjuste, setTipoAjuste] = useState<'entrada' | 'saida'>('entrada')
  const [quantidade, setQuantidade] = useState('')
  const [motivo, setMotivo] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erroModal, setErroModal] = useState('')

  useEffect(() => {
    api.get<{ items: Produto[] }>('/api/v1/produtos?limit=200').then(r => setProdutos(r.items ?? []))
  }, [])

  useEffect(() => {
    if (!produtoSel) return
    setOffset(0)
    carregarHistorico(produtoSel, aba, 0)
  }, [produtoSel, aba])

  async function carregarHistorico(id: string, tipo: 'movimentacoes' | 'ajustes', off: number) {
    setCarregando(true)
    setErro('')
    try {
      const sufixo = `?limit=${LIMITE}&offset=${off}`
      if (tipo === 'movimentacoes') {
        const res = await api.get<{ items: Movimentacao[] }>(`/api/v1/estoque/${id}${sufixo}`)
        setMovimentacoes(res.items ?? [])
      } else {
        const res = await api.get<{ items: Ajuste[] }>(`/api/v1/estoque/${id}/ajustes${sufixo}`)
        setAjustes(res.items ?? [])
      }
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar histórico')
    } finally {
      setCarregando(false)
    }
  }

  function abrirModal() {
    setTipoAjuste('entrada')
    setQuantidade('')
    setMotivo('')
    setErroModal('')
    setModalAberto(true)
  }

  async function salvarAjuste(e: React.FormEvent) {
    e.preventDefault()
    if (!produtoSel) return
    const qtd = parseInt(quantidade, 10)
    if (!qtd || qtd <= 0) {
      setErroModal('Quantidade deve ser um número positivo')
      return
    }
    setSalvando(true)
    setErroModal('')
    try {
      await api.post('/api/v1/estoque/ajustes', {
        produto_id: produtoSel,
        qtd_entrada: tipoAjuste === 'entrada' ? qtd : 0,
        qtd_saida:   tipoAjuste === 'saida'   ? qtd : 0,
        motivo,
      })
      setModalAberto(false)
      // Recarrega histórico e atualiza saldo na lista de produtos
      const prodRes = await api.get<{ items: Produto[] }>('/api/v1/produtos?limit=200')
      setProdutos(prodRes.items ?? [])
      setOffset(0)
      carregarHistorico(produtoSel, aba, 0)
    } catch (err: unknown) {
      setErroModal(err instanceof Error ? err.message : 'Erro ao salvar ajuste')
    } finally {
      setSalvando(false)
    }
  }

  const produtoAtual = produtos.find(p => p.id === produtoSel)
  const listaLen = aba === 'movimentacoes' ? movimentacoes.length : ajustes.length

  const colunasMov: Column<Movimentacao>[] = [
    { header: 'Tipo', sortAccessor: (m) => rotulaTipo[m.tipo] ?? m.tipo, cell: (m) => <StatusBadge tone={toneTipo[m.tipo] ?? 'neutral'}>{rotulaTipo[m.tipo] ?? m.tipo}</StatusBadge> },
    { header: 'Qtd', align: 'right', sortAccessor: (m) => m.quantidade, cell: (m) => <span className="font-mono">{m.quantidade}</span> },
    { header: 'Saldo Antes', align: 'right', hideBelow: 'sm', sortAccessor: (m) => m.saldo_antes, cell: (m) => <span className="font-mono text-gray-500">{m.saldo_antes}</span> },
    { header: 'Saldo Depois', align: 'right', sortAccessor: (m) => m.saldo_depois, cell: (m) => <span className="font-mono font-semibold">{m.saldo_depois}</span> },
    { header: 'Origem', hideBelow: 'md', sortAccessor: (m) => m.origem_tipo, cell: (m) => <span className="text-gray-500">{m.origem_tipo || '—'}</span> },
    { header: 'Data', hideBelow: 'sm', sortAccessor: (m) => new Date(m.criado_em).getTime(), cell: (m) => <span className="text-gray-500 whitespace-nowrap">{new Date(m.criado_em).toLocaleString('pt-BR')}</span> },
  ]

  const colunasAjuste: Column<Ajuste>[] = [
    { header: 'Entrada', align: 'right', sortAccessor: (a) => a.qtd_entrada, cell: (a) => a.qtd_entrada > 0 ? <span className="text-green-700 font-semibold font-mono">+{a.qtd_entrada}</span> : <span className="text-gray-300">—</span> },
    { header: 'Saída', align: 'right', sortAccessor: (a) => a.qtd_saida, cell: (a) => a.qtd_saida > 0 ? <span className="text-red-700 font-semibold font-mono">-{a.qtd_saida}</span> : <span className="text-gray-300">—</span> },
    { header: 'Motivo', sortAccessor: (a) => a.motivo, cell: (a) => <span className="text-gray-700">{a.motivo}</span> },
    { header: 'Data', hideBelow: 'sm', sortAccessor: (a) => new Date(a.criado_em).getTime(), cell: (a) => <span className="text-gray-500 whitespace-nowrap">{new Date(a.criado_em).toLocaleString('pt-BR')}</span> },
  ]

  return (
    <PageShell
      title="Estoque — Ajustes e Razão"
      subtitle="Movimentações e ajustes manuais de estoque"
      actions={
        produtoSel ? (
          <Button onClick={abrirModal}>
            <Plus className="w-4 h-4" />
            Novo Ajuste
          </Button>
        ) : undefined
      }
    >
      {/* Seleção de produto */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <Field label="Produto">
          <select value={produtoSel} onChange={e => setProdutoSel(e.target.value)} className={inputClasses()}>
            <option value="">— selecione um produto —</option>
            {produtos.map(p => (
              <option key={p.id} value={p.id}>
                {p.descricao}{p.modelo ? ` — ${p.modelo}` : ''}
              </option>
            ))}
          </select>
        </Field>

        {produtoAtual && (
          <div className="mt-3 flex items-center gap-4 text-sm text-gray-600">
            <span>
              Saldo atual:{' '}
              <strong className={produtoAtual.estoque_atual === 0 ? 'text-red-600' : 'text-gray-900'}>
                {produtoAtual.estoque_atual}
              </strong>
            </span>
            <StatusBadge tone={produtoAtual.disponivel ? 'success' : 'danger'}>
              {produtoAtual.disponivel ? 'Disponível' : 'Indisponível'}
            </StatusBadge>
          </div>
        )}
      </div>

      {produtoSel ? (
        <>
          {/* Abas */}
          <div className="flex gap-1 border-b border-gray-200">
            {(['movimentacoes', 'ajustes'] as const).map(tab => (
              <button
                key={tab}
                onClick={() => setAba(tab)}
                className={cn(
                  'px-4 py-2 text-sm font-medium -mb-px border-b-2 transition-colors',
                  aba === tab ? 'border-gray-900 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700',
                )}
              >
                {tab === 'movimentacoes' ? 'Razão / Movimentações' : 'Ajustes Manuais'}
              </button>
            ))}
          </div>

          {erro && <p className="text-sm text-red-600">{erro}</p>}

          {aba === 'movimentacoes' ? (
            <DataTable
              columns={colunasMov}
              rows={movimentacoes}
              rowKey={(m) => m.id}
              loading={carregando}
              empty="Nenhuma movimentação registrada para este produto."
            />
          ) : (
            <DataTable
              columns={colunasAjuste}
              rows={ajustes}
              rowKey={(a) => a.id}
              loading={carregando}
              empty="Nenhum ajuste manual registrado para este produto."
            />
          )}

          {(listaLen === LIMITE || offset > 0) && (
            <div className="flex justify-between">
              <Button
                variant="secondary"
                disabled={offset === 0}
                onClick={() => { const off = Math.max(0, offset - LIMITE); setOffset(off); carregarHistorico(produtoSel, aba, off) }}
              >
                Anterior
              </Button>
              <Button
                variant="secondary"
                disabled={listaLen < LIMITE}
                onClick={() => { const off = offset + LIMITE; setOffset(off); carregarHistorico(produtoSel, aba, off) }}
              >
                Próxima
              </Button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center text-gray-400 text-sm py-16">
          Selecione um produto para ver o histórico ou lançar um ajuste.
        </div>
      )}

      {modalAberto && (
        <Modal title="Novo Ajuste de Estoque" onClose={() => setModalAberto(false)} maxWidth="max-w-md">
          <form onSubmit={salvarAjuste} className="px-6 py-4 space-y-4">
            {erroModal && <p className="text-sm text-red-600">{erroModal}</p>}

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Produto</label>
              <p className="text-sm text-gray-900 font-medium">
                {produtoAtual?.descricao}{produtoAtual?.modelo ? ` — ${produtoAtual.modelo}` : ''}
              </p>
              <p className="text-xs text-gray-500 mt-0.5">Saldo atual: {produtoAtual?.estoque_atual ?? 0}</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Tipo de Ajuste *</label>
              <div className="flex gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="tipo" value="entrada" checked={tipoAjuste === 'entrada'} onChange={() => setTipoAjuste('entrada')} />
                  <span className="text-sm text-green-700 font-medium">Entrada (acrescenta)</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="tipo" value="saida" checked={tipoAjuste === 'saida'} onChange={() => setTipoAjuste('saida')} />
                  <span className="text-sm text-red-700 font-medium">Saída (diminui)</span>
                </label>
              </div>
            </div>

            <Field label="Quantidade *">
              <input
                type="number"
                min={1}
                value={quantidade}
                onChange={e => setQuantidade(e.target.value)}
                required
                className={inputClasses()}
                placeholder="Ex.: 10"
              />
              {quantidade && parseInt(quantidade, 10) > 0 && (
                <p className={`text-xs mt-1 font-medium ${tipoAjuste === 'entrada' ? 'text-green-600' : 'text-red-600'}`}>
                  Novo saldo: {(produtoAtual?.estoque_atual ?? 0) + (tipoAjuste === 'entrada' ? 1 : -1) * parseInt(quantidade, 10)}
                </p>
              )}
            </Field>

            <Field label="Motivo *">
              <textarea
                value={motivo}
                onChange={e => setMotivo(e.target.value)}
                required
                rows={2}
                className={inputClasses() + ' resize-none'}
                placeholder="Ex.: Inventário físico, quebra, devolução de cliente…"
              />
            </Field>

            <div className="flex justify-end gap-3 pt-2">
              <Button type="button" variant="secondary" onClick={() => setModalAberto(false)}>Cancelar</Button>
              <Button type="submit" disabled={salvando}>{salvando ? 'Salvando…' : 'Confirmar Ajuste'}</Button>
            </div>
          </form>
        </Modal>
      )}
    </PageShell>
  )
}
