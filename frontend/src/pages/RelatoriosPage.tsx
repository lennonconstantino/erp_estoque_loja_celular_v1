import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/lib/api'

// --- tipos ---

interface ProdutoAbaixoMinimo {
  id: string
  descricao: string
  estoque_atual: number
  estoque_minimo: number
  defasagem: number
}

interface ProdutoVendido {
  produto_id: string
  descricao: string
  total_vendido: number
  total_valor: number
}

interface ResumoVendas {
  total_vendas: number
  valor_total: number
  ticket_medio: number
  de: string
  ate: string
}

interface ResumoCompras {
  total_compras: number
  valor_total: number
  de: string
  ate: string
}

// --- helpers ---

function hojeISO() {
  return new Date().toISOString().slice(0, 10)
}

function primeiroDiaMesISO() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

function brl(v: number) {
  return v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

// --- aba enum ---
type Aba = 'abaixo-minimo' | 'mais-vendidos' | 'menos-vendidos' | 'resumo-vendas' | 'resumo-compras'

const ABAS: { key: Aba; label: string }[] = [
  { key: 'abaixo-minimo', label: 'Abaixo do Mínimo' },
  { key: 'mais-vendidos', label: 'Mais Vendidos' },
  { key: 'menos-vendidos', label: 'Menos Vendidos' },
  { key: 'resumo-vendas', label: 'Resumo de Vendas' },
  { key: 'resumo-compras', label: 'Resumo de Compras' },
]

export default function RelatoriosPage() {
  const [aba, setAba] = useState<Aba>('abaixo-minimo')
  const [de, setDe] = useState(primeiroDiaMesISO())
  const [ate, setAte] = useState(hojeISO())
  const [limite, setLimite] = useState(10)
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  // dados por relatório
  const [abaixoMinimo, setAbaixoMinimo] = useState<ProdutoAbaixoMinimo[]>([])
  const [maisVendidos, setMaisVendidos] = useState<ProdutoVendido[]>([])
  const [menosVendidos, setMenosVendidos] = useState<ProdutoVendido[]>([])
  const [resumoVendas, setResumoVendas] = useState<ResumoVendas | null>(null)
  const [resumoCompras, setResumoCompras] = useState<ResumoCompras | null>(null)

  async function carregar() {
    setCarregando(true)
    setErro('')
    try {
      switch (aba) {
        case 'abaixo-minimo': {
          const d = await api.get<{ items: ProdutoAbaixoMinimo[] }>('/relatorios/produtos/abaixo-do-minimo')
          setAbaixoMinimo(d.items ?? [])
          break
        }
        case 'mais-vendidos': {
          const d = await api.get<{ items: ProdutoVendido[] }>(
            `/relatorios/produtos/mais-vendidos?de=${de}&ate=${ate}&limite=${limite}`,
          )
          setMaisVendidos(d.items ?? [])
          break
        }
        case 'menos-vendidos': {
          const d = await api.get<{ items: ProdutoVendido[] }>(
            `/relatorios/produtos/menos-vendidos?de=${de}&ate=${ate}&limite=${limite}`,
          )
          setMenosVendidos(d.items ?? [])
          break
        }
        case 'resumo-vendas': {
          const d = await api.get<ResumoVendas>(`/relatorios/vendas?de=${de}&ate=${ate}`)
          setResumoVendas(d)
          break
        }
        case 'resumo-compras': {
          const d = await api.get<ResumoCompras>(`/relatorios/compras?de=${de}&ate=${ate}`)
          setResumoCompras(d)
          break
        }
      }
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar relatório')
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => {
    carregar()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [aba])

  const precisaPeriodo = aba !== 'abaixo-minimo'
  const precisaLimite = aba === 'mais-vendidos' || aba === 'menos-vendidos'

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto px-4 py-8">
        {/* cabeçalho */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Relatórios</h1>
            <p className="text-sm text-gray-500 mt-1">Análises de estoque, vendas e compras</p>
          </div>
          <Link to="/" className="px-4 py-2 text-sm font-medium text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
            ← Dashboard
          </Link>
        </div>

        {/* abas */}
        <div className="flex gap-1 mb-5 bg-white border border-gray-200 rounded-xl p-1 overflow-x-auto">
          {ABAS.map((a) => (
            <button
              key={a.key}
              onClick={() => setAba(a.key)}
              className={`flex-1 min-w-max px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
                aba === a.key
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-gray-600 hover:bg-gray-100'
              }`}
            >
              {a.label}
            </button>
          ))}
        </div>

        {/* filtros */}
        {(precisaPeriodo || precisaLimite) && (
          <div className="bg-white rounded-xl border border-gray-200 p-4 mb-5 flex flex-wrap gap-4 items-end">
            {precisaPeriodo && (
              <>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">De</label>
                  <input
                    type="date"
                    value={de}
                    onChange={(e) => setDe(e.target.value)}
                    className="px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Até</label>
                  <input
                    type="date"
                    value={ate}
                    onChange={(e) => setAte(e.target.value)}
                    className="px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </>
            )}
            {precisaLimite && (
              <div>
                <label className="block text-xs text-gray-500 mb-1">Limite</label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={limite}
                  onChange={(e) => setLimite(parseInt(e.target.value) || 10)}
                  className="w-24 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            )}
            <button
              onClick={carregar}
              disabled={carregando}
              className="px-5 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {carregando ? 'Carregando…' : 'Atualizar'}
            </button>
          </div>
        )}

        {erro && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">{erro}</div>
        )}

        {carregando && (
          <div className="p-8 text-center text-gray-500 text-sm">Carregando...</div>
        )}

        {/* conteúdo por aba */}
        {!carregando && (
          <>
            {/* Abaixo do mínimo */}
            {aba === 'abaixo-minimo' && (
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
                  <h2 className="text-sm font-semibold text-gray-700">Produtos abaixo do estoque mínimo</h2>
                  <span className="text-xs text-gray-400">{abaixoMinimo.length} produto(s)</span>
                </div>
                {abaixoMinimo.length === 0 ? (
                  <div className="p-8 text-center text-gray-500 text-sm">Nenhum produto abaixo do mínimo.</div>
                ) : (
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 border-b border-gray-200">
                      <tr>
                        <th className="px-4 py-3 text-left font-medium text-gray-600">Produto</th>
                        <th className="px-4 py-3 text-right font-medium text-gray-600">Estoque Atual</th>
                        <th className="px-4 py-3 text-right font-medium text-gray-600">Mínimo</th>
                        <th className="px-4 py-3 text-right font-medium text-gray-600">Defasagem</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {abaixoMinimo.map((p) => (
                        <tr key={p.id} className="bg-red-50 hover:bg-red-100">
                          <td className="px-4 py-3 font-medium text-gray-900">{p.descricao}</td>
                          <td className="px-4 py-3 text-right text-red-700 font-bold">{p.estoque_atual}</td>
                          <td className="px-4 py-3 text-right text-gray-600">{p.estoque_minimo}</td>
                          <td className="px-4 py-3 text-right text-red-600 font-medium">−{p.defasagem}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            )}

            {/* Mais/Menos vendidos */}
            {(aba === 'mais-vendidos' || aba === 'menos-vendidos') && (
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
                  <h2 className="text-sm font-semibold text-gray-700">
                    {aba === 'mais-vendidos' ? 'Produtos mais vendidos' : 'Produtos menos vendidos'} no período
                  </h2>
                  <span className="text-xs text-gray-400">{(aba === 'mais-vendidos' ? maisVendidos : menosVendidos).length} produto(s)</span>
                </div>
                {(() => {
                  const lista = aba === 'mais-vendidos' ? maisVendidos : menosVendidos
                  return lista.length === 0 ? (
                    <div className="p-8 text-center text-gray-500 text-sm">Nenhuma venda no período.</div>
                  ) : (
                    <table className="w-full text-sm">
                      <thead className="bg-gray-50 border-b border-gray-200">
                        <tr>
                          <th className="px-4 py-3 text-left font-medium text-gray-600">#</th>
                          <th className="px-4 py-3 text-left font-medium text-gray-600">Produto</th>
                          <th className="px-4 py-3 text-right font-medium text-gray-600">Qtd Vendida</th>
                          <th className="px-4 py-3 text-right font-medium text-gray-600">Total (R$)</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {lista.map((p, i) => (
                          <tr key={p.produto_id} className="hover:bg-gray-50">
                            <td className="px-4 py-3 text-gray-400 text-xs">{i + 1}</td>
                            <td className="px-4 py-3 font-medium text-gray-900">{p.descricao}</td>
                            <td className="px-4 py-3 text-right text-gray-700 font-bold">{p.total_vendido}</td>
                            <td className="px-4 py-3 text-right text-gray-700">{brl(p.total_valor)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )
                })()}
              </div>
            )}

            {/* Resumo de Vendas */}
            {aba === 'resumo-vendas' && resumoVendas && (
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <CardMetrica titulo="Total de Vendas" valor={String(resumoVendas.total_vendas)} sufixo="vendas" />
                <CardMetrica titulo="Valor Total" valor={brl(resumoVendas.valor_total)} />
                <CardMetrica titulo="Ticket Médio" valor={brl(resumoVendas.ticket_medio)} />
              </div>
            )}

            {/* Resumo de Compras */}
            {aba === 'resumo-compras' && resumoCompras && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <CardMetrica titulo="Total de Compras" valor={String(resumoCompras.total_compras)} sufixo="compras" />
                <CardMetrica titulo="Valor Total" valor={brl(resumoCompras.valor_total)} />
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

function CardMetrica({ titulo, valor, sufixo }: { titulo: string; valor: string; sufixo?: string }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-5">
      <p className="text-xs text-gray-500 mb-1">{titulo}</p>
      <p className="text-2xl font-bold text-gray-900">
        {valor}
        {sufixo && <span className="text-sm font-normal text-gray-500 ml-1">{sufixo}</span>}
      </p>
    </div>
  )
}
