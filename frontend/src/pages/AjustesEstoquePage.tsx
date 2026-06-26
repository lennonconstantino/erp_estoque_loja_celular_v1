import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

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

const badgeTipo: Record<string, string> = {
  COMPRA: 'bg-blue-100 text-blue-800',
  VENDA: 'bg-red-100 text-red-800',
  AJUSTE_ENTRADA: 'bg-green-100 text-green-800',
  AJUSTE_SAIDA: 'bg-orange-100 text-orange-800',
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
    api.get<{ items: Produto[] }>('/produtos?limit=200').then(r => setProdutos(r.items ?? []))
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
      if (tipo === 'movimentacoes') {
        const sufixo = `?limit=${LIMITE}&offset=${off}`
        const res = await api.get<{ items: Movimentacao[] }>(`/estoque/${id}${sufixo}`)
        setMovimentacoes(res.items ?? [])
      } else {
        const sufixo = `?limit=${LIMITE}&offset=${off}`
        const res = await api.get<{ items: Ajuste[] }>(`/estoque/${id}/ajustes${sufixo}`)
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
      await api.post('/estoque/ajustes', {
        produto_id: produtoSel,
        qtd_entrada: tipoAjuste === 'entrada' ? qtd : 0,
        qtd_saida:   tipoAjuste === 'saida'   ? qtd : 0,
        motivo,
      })
      setModalAberto(false)
      // Recarrega histórico e atualiza saldo na lista de produtos
      const [prodRes] = await Promise.all([
        api.get<{ items: Produto[] }>('/produtos?limit=200'),
      ])
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

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-5xl mx-auto py-8 px-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Estoque — Ajustes e Razão</h1>
          {produtoSel && (
            <button
              onClick={abrirModal}
              className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700"
            >
              Novo Ajuste
            </button>
          )}
        </div>

        {/* Seleção de produto */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-2">Produto</label>
          <select
            value={produtoSel}
            onChange={e => setProdutoSel(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="">— selecione um produto —</option>
            {produtos.map(p => (
              <option key={p.id} value={p.id}>
                {p.descricao}{p.modelo ? ` — ${p.modelo}` : ''}
              </option>
            ))}
          </select>

          {produtoAtual && (
            <div className="mt-3 flex items-center gap-4 text-sm text-gray-600">
              <span>
                Saldo atual:{' '}
                <strong className={produtoAtual.estoque_atual === 0 ? 'text-red-600' : 'text-gray-900'}>
                  {produtoAtual.estoque_atual}
                </strong>
              </span>
              <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${produtoAtual.disponivel ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                {produtoAtual.disponivel ? 'Disponível' : 'Indisponível'}
              </span>
            </div>
          )}
        </div>

        {produtoSel && (
          <>
            {/* Abas */}
            <div className="flex gap-1 mb-4 border-b">
              {(['movimentacoes', 'ajustes'] as const).map(tab => (
                <button
                  key={tab}
                  onClick={() => setAba(tab)}
                  className={`px-4 py-2 text-sm font-medium -mb-px border-b-2 transition-colors ${
                    aba === tab
                      ? 'border-indigo-600 text-indigo-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  {tab === 'movimentacoes' ? 'Razão / Movimentações' : 'Ajustes Manuais'}
                </button>
              ))}
            </div>

            {erro && <p className="text-red-600 text-sm mb-4">{erro}</p>}

            {carregando ? (
              <p className="text-gray-500 text-sm">Carregando…</p>
            ) : aba === 'movimentacoes' ? (
              <TabelaMovimentacoes itens={movimentacoes} />
            ) : (
              <TabelaAjustes itens={ajustes} />
            )}

            {/* Paginação */}
            {((aba === 'movimentacoes' ? movimentacoes.length : ajustes.length) === LIMITE || offset > 0) && (
              <div className="flex justify-between mt-4">
                <button
                  disabled={offset === 0}
                  onClick={() => {
                    const off = Math.max(0, offset - LIMITE)
                    setOffset(off)
                    carregarHistorico(produtoSel, aba, off)
                  }}
                  className="px-4 py-2 text-sm bg-white border rounded-md disabled:opacity-40"
                >
                  Anterior
                </button>
                <button
                  disabled={(aba === 'movimentacoes' ? movimentacoes.length : ajustes.length) < LIMITE}
                  onClick={() => {
                    const off = offset + LIMITE
                    setOffset(off)
                    carregarHistorico(produtoSel, aba, off)
                  }}
                  className="px-4 py-2 text-sm bg-white border rounded-md disabled:opacity-40"
                >
                  Próxima
                </button>
              </div>
            )}
          </>
        )}

        {!produtoSel && (
          <div className="text-center text-gray-400 text-sm py-16">
            Selecione um produto para ver o histórico ou lançar um ajuste.
          </div>
        )}
      </div>

      {/* Modal de Ajuste */}
      {modalAberto && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
            <div className="flex items-center justify-between p-6 border-b">
              <h2 className="text-lg font-semibold">Novo Ajuste de Estoque</h2>
              <button onClick={() => setModalAberto(false)} className="text-gray-400 hover:text-gray-600 text-xl">×</button>
            </div>
            <form onSubmit={salvarAjuste} className="p-6 space-y-4">
              {erroModal && <p className="text-red-600 text-sm">{erroModal}</p>}

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
                    <input
                      type="radio"
                      name="tipo"
                      value="entrada"
                      checked={tipoAjuste === 'entrada'}
                      onChange={() => setTipoAjuste('entrada')}
                      className="text-indigo-600"
                    />
                    <span className="text-sm text-green-700 font-medium">Entrada (acrescenta)</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="tipo"
                      value="saida"
                      checked={tipoAjuste === 'saida'}
                      onChange={() => setTipoAjuste('saida')}
                      className="text-indigo-600"
                    />
                    <span className="text-sm text-red-700 font-medium">Saída (diminui)</span>
                  </label>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Quantidade *</label>
                <input
                  type="number"
                  min={1}
                  value={quantidade}
                  onChange={e => setQuantidade(e.target.value)}
                  required
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="Ex.: 10"
                />
                {quantidade && parseInt(quantidade, 10) > 0 && (
                  <p className={`text-xs mt-1 font-medium ${tipoAjuste === 'entrada' ? 'text-green-600' : 'text-red-600'}`}>
                    Novo saldo: {(produtoAtual?.estoque_atual ?? 0) + (tipoAjuste === 'entrada' ? 1 : -1) * parseInt(quantidade, 10)}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Motivo *</label>
                <textarea
                  value={motivo}
                  onChange={e => setMotivo(e.target.value)}
                  required
                  rows={2}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm resize-none"
                  placeholder="Ex.: Inventário físico, quebra, devolução de cliente…"
                />
              </div>

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
                  {salvando ? 'Salvando…' : 'Confirmar Ajuste'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Subcomponentes de tabela ────────────────────────────────────────────────

function TabelaMovimentacoes({ itens }: { itens: Movimentacao[] }) {
  if (itens.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400 text-sm">
        Nenhuma movimentação registrada para este produto.
      </div>
    )
  }
  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {['Tipo', 'Qtd', 'Saldo Antes', 'Saldo Depois', 'Origem', 'Data'].map(h => (
              <th key={h} className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {itens.map(m => (
            <tr key={m.id} className="hover:bg-gray-50">
              <td className="px-4 py-3">
                <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${badgeTipo[m.tipo] ?? 'bg-gray-100 text-gray-700'}`}>
                  {rotulaTipo[m.tipo] ?? m.tipo}
                </span>
              </td>
              <td className="px-4 py-3 text-sm font-mono">{m.quantidade}</td>
              <td className="px-4 py-3 text-sm font-mono text-gray-500">{m.saldo_antes}</td>
              <td className="px-4 py-3 text-sm font-mono font-semibold">{m.saldo_depois}</td>
              <td className="px-4 py-3 text-sm text-gray-500">{m.origem_tipo || '—'}</td>
              <td className="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">
                {new Date(m.criado_em).toLocaleString('pt-BR')}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function TabelaAjustes({ itens }: { itens: Ajuste[] }) {
  if (itens.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400 text-sm">
        Nenhum ajuste manual registrado para este produto.
      </div>
    )
  }
  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {['Entrada', 'Saída', 'Motivo', 'Data'].map(h => (
              <th key={h} className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {itens.map(a => (
            <tr key={a.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 text-sm font-mono">
                {a.qtd_entrada > 0 ? (
                  <span className="text-green-700 font-semibold">+{a.qtd_entrada}</span>
                ) : (
                  <span className="text-gray-300">—</span>
                )}
              </td>
              <td className="px-4 py-3 text-sm font-mono">
                {a.qtd_saida > 0 ? (
                  <span className="text-red-700 font-semibold">-{a.qtd_saida}</span>
                ) : (
                  <span className="text-gray-300">—</span>
                )}
              </td>
              <td className="px-4 py-3 text-sm text-gray-700">{a.motivo}</td>
              <td className="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">
                {new Date(a.criado_em).toLocaleString('pt-BR')}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
