import { useEffect, useState } from 'react'
import { Pencil, Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { DataTable, type Column } from '@/components/ui/data-table'
import { Modal } from '@/components/ui/modal'
import { Field, inputClasses } from '@/components/ui/field'

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
        `/api/v1/categorias?q=${encodeURIComponent(q)}&limit=${LIMITE}&offset=${off}`
      )
      setItens(res.items ?? [])
    } catch (e: unknown) {
      setErro(e instanceof Error ? e.message : 'Erro ao carregar categorias')
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => { carregar() }, []) // eslint-disable-line react-hooks/exhaustive-deps

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
        await api.put(`/api/v1/categorias/${editando.id}`, { descricao })
      } else {
        await api.post('/api/v1/categorias', { descricao })
      }
      setModalAberto(false)
      setOffset(0)
      carregar(busca, 0)
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

  const colunas: Column<Categoria>[] = [
    { header: 'Descrição', sortAccessor: (c) => c.descricao, cell: (c) => <span className="font-bold text-foreground">{c.descricao}</span> },
    {
      header: '',
      align: 'right',
      cell: (c) => (
        <button onClick={() => abrirEditar(c)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted active:scale-90" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell
      title="Categorias"
      subtitle="Cadastro e gestão de categorias"
      maxWidth="max-w-4xl"
      actions={
        <Button onClick={abrirNova}>
          <Plus className="w-4 h-4" />
          Nova Categoria
        </Button>
      }
    >
      <form onSubmit={pesquisar} className="flex gap-3 bg-card p-4 rounded-2xl border border-border shadow-sm animate-in fade-in duration-500">
        <input
          type="text"
          placeholder="Pesquisar categorias por descrição…"
          value={busca}
          onChange={e => setBusca(e.target.value)}
          className={inputClasses() + ' flex-1'}
        />
        <Button type="submit" variant="secondary" className="px-8 h-10">Pesquisar</Button>
      </form>

      {erro && <p className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={itens}
        rowKey={(c) => c.id}
        loading={carregando}
        empty="Nenhuma categoria encontrada."
      />

      {(itens.length === LIMITE || offset > 0) && (
        <div className="flex justify-between items-center bg-card p-3 rounded-xl border border-border shadow-sm">
          <Button
            variant="secondary"
            size="sm"
            disabled={offset === 0}
            onClick={() => { const off = Math.max(0, offset - LIMITE); setOffset(off); carregar(busca, off) }}
          >
            Anterior
          </Button>
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em]">Exibindo {itens.length} resultados</span>
          <Button
            variant="secondary"
            size="sm"
            disabled={itens.length < LIMITE}
            onClick={() => { const off = offset + LIMITE; setOffset(off); carregar(busca, off) }}
          >
            Próxima
          </Button>
        </div>
      )}

      {modalAberto && (
        <Modal title={editando ? 'Editar Categoria' : 'Nova Categoria'} onClose={() => setModalAberto(false)} maxWidth="max-w-md">
          <form onSubmit={salvar} className="px-6 py-4 space-y-4">
            {erroModal && <p className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erroModal}</p>}
            <Field label="Descrição *">
              <input
                type="text"
                value={descricao}
                onChange={e => setDescricao(e.target.value)}
                required
                className={inputClasses()}
                placeholder="Ex.: Capas e Películas"
              />
            </Field>
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
