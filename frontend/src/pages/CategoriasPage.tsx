import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

interface Categoria {
  id: string
  descricao: string
}

const LIMITE = 50

export default function CategoriasPage() {
  const [itens, setItens] = useState<Categoria[]>([])
  const [offset, setOffset] = useState(0)
  const [busca, setBusca] = useState('')
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  const [modalAberto, setModalAberto] = useState(false)
  const [editando, setEditando] = useState<Categoria | null>(null)
  const [descricao, setDescricao] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erroModal, setErroModal] = useState('')

  async function carregar(q = busca, off = offset) {
    setCarregando(true)
    setErro('')
    try {
      const res = await api.get<{ items: Categoria[] }>(
        `/categorias?q=${encodeURIComponent(q)}&limit=${LIMITE}&offset=${off}`
      )
      setItens(res.items ?? [])
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar categorias')
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => { carregar() }, [])

  function abrirNova() {
    setEditando(null)
    setDescricao('')
    setErroModal('')
    setModalAberto(true)
  }

  function abrirEditar(c: Categoria) {
    setEditando(c)
    setDescricao(c.descricao)
    setErroModal('')
    setModalAberto(true)
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault()
    setSalvando(true)
    setErroModal('')
    try {
      if (editando) {
        await api.put(`/categorias/${editando.id}`, { descricao })
      } else {
        await api.post('/categorias', { descricao })
      }
      setModalAberto(false)
      carregar(busca, 0)
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
    carregar(busca, 0)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto py-8 px-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Categorias</h1>
          <button
            onClick={abrirNova}
            className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700"
          >
            Nova Categoria
          </button>
        </div>

        <form onSubmit={pesquisar} className="flex gap-2 mb-4">
          <input
            type="text"
            placeholder="Pesquisar por descrição…"
            value={busca}
            onChange={e => setBusca(e.target.value)}
            className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm"
          />
          <button type="submit" className="bg-gray-200 px-4 py-2 rounded-md text-sm hover:bg-gray-300">
            Buscar
          </button>
        </form>

        {erro && <p className="text-red-600 text-sm mb-4">{erro}</p>}

        {carregando ? (
          <p className="text-gray-500 text-sm">Carregando…</p>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Descrição
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Ações
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {itens.length === 0 ? (
                  <tr>
                    <td colSpan={2} className="px-6 py-8 text-center text-gray-400 text-sm">
                      Nenhuma categoria encontrada
                    </td>
                  </tr>
                ) : (
                  itens.map(c => (
                    <tr key={c.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 text-sm text-gray-900">{c.descricao}</td>
                      <td className="px-6 py-4 text-right">
                        <button
                          onClick={() => abrirEditar(c)}
                          className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
                        >
                          Editar
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}

        {(itens.length === LIMITE || offset > 0) && (
          <div className="flex justify-between mt-4">
            <button
              disabled={offset === 0}
              onClick={() => { const off = Math.max(0, offset - LIMITE); setOffset(off); carregar(busca, off) }}
              className="px-4 py-2 text-sm bg-white border rounded-md disabled:opacity-40"
            >
              Anterior
            </button>
            <button
              disabled={itens.length < LIMITE}
              onClick={() => { const off = offset + LIMITE; setOffset(off); carregar(busca, off) }}
              className="px-4 py-2 text-sm bg-white border rounded-md disabled:opacity-40"
            >
              Próxima
            </button>
          </div>
        )}
      </div>

      {modalAberto && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
            <div className="flex items-center justify-between p-6 border-b">
              <h2 className="text-lg font-semibold">
                {editando ? 'Editar Categoria' : 'Nova Categoria'}
              </h2>
              <button onClick={() => setModalAberto(false)} className="text-gray-400 hover:text-gray-600 text-xl">×</button>
            </div>
            <form onSubmit={salvar} className="p-6 space-y-4">
              {erroModal && <p className="text-red-600 text-sm">{erroModal}</p>}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Descrição *</label>
                <input
                  type="text"
                  value={descricao}
                  onChange={e => setDescricao(e.target.value)}
                  required
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="Ex.: Capas e Películas"
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
