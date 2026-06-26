import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Pencil, Plus, X } from 'lucide-react'
import { api, ApiError } from '@/lib/api'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge } from '@/components/ui/badge'

// ── tipos ──────────────────────────────────────────────────────────────────────

interface Cliente {
  id: string
  cpf: string
  nome: string
  email: string
  telefone?: string
  cep?: string
  rua?: string
  numero?: string
  complemento?: string
  bairro?: string
  cidade?: string
  uf?: string
  ultima_compra?: string
  ativo: boolean
}

interface EnderecoViaCep {
  CEP: string
  Rua: string
  Bairro: string
  Cidade: string
  UF: string
}

const vazio = {
  cpf: '', nome: '', email: '', telefone: '',
  cep: '', rua: '', numero: '', complemento: '', bairro: '', cidade: '', uf: '',
}

// ── componente principal ───────────────────────────────────────────────────────

export default function ClientesPage() {
  const navigate = useNavigate()

  const [itens, setItens] = useState<Cliente[]>([])
  const [pagina, setPagina] = useState(0)
  const [busca, setBusca] = useState('')
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  const [modalAberto, setModalAberto] = useState(false)
  const [editando, setEditando] = useState<Cliente | null>(null)
  const [form, setForm] = useState({ ...vazio, ativo: true })
  const [erroForm, setErroForm] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [buscandoCep, setBuscandoCep] = useState(false)

  const limite = 20

  async function carregar(q = busca, pg = pagina) {
    setCarregando(true)
    setErro('')
    try {
      const res = await api.get<{ items: Cliente[] }>(
        `/api/v1/clientes?q=${encodeURIComponent(q)}&limit=${limite}&offset=${pg * limite}`,
      )
      setItens(res.items ?? [])
    } catch {
      setErro('Não foi possível carregar os clientes.')
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => { void carregar() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function handleBusca(e: React.FormEvent) {
    e.preventDefault()
    setPagina(0)
    void carregar(busca, 0)
  }

  function abrirCriar() {
    setEditando(null)
    setForm({ ...vazio, ativo: true })
    setErroForm('')
    setModalAberto(true)
  }

  function abrirEditar(c: Cliente) {
    setEditando(c)
    setForm({
      cpf: c.cpf, nome: c.nome, email: c.email, telefone: c.telefone ?? '',
      cep: c.cep ?? '', rua: c.rua ?? '', numero: c.numero ?? '',
      complemento: c.complemento ?? '', bairro: c.bairro ?? '',
      cidade: c.cidade ?? '', uf: c.uf ?? '',
      ativo: c.ativo,
    })
    setErroForm('')
    setModalAberto(true)
  }

  function fecharModal() {
    setModalAberto(false)
    setEditando(null)
  }

  function campo(k: keyof typeof form, v: string | boolean) {
    setForm((prev) => ({ ...prev, [k]: v }))
  }

  async function buscarCep(cep: string) {
    const digitos = cep.replace(/\D/g, '')
    if (digitos.length !== 8) return
    setBuscandoCep(true)
    try {
      const end = await api.get<EnderecoViaCep>(`/api/v1/clientes/cep/${digitos}`)
      setForm((prev) => ({
        ...prev,
        rua: end.Rua || prev.rua,
        bairro: end.Bairro || prev.bairro,
        cidade: end.Cidade || prev.cidade,
        uf: end.UF || prev.uf,
      }))
    } catch {
      // CEP não encontrado — usuário preenche manualmente
    } finally {
      setBuscandoCep(false)
    }
  }

  async function handleSalvar(e: React.FormEvent) {
    e.preventDefault()
    setErroForm('')
    setSalvando(true)
    try {
      const payload = {
        cpf: form.cpf,
        nome: form.nome,
        email: form.email,
        telefone: form.telefone,
        cep: form.cep,
        rua: form.rua,
        numero: form.numero,
        complemento: form.complemento,
        bairro: form.bairro,
        cidade: form.cidade,
        uf: form.uf,
      }
      if (editando) {
        await api.put(`/api/v1/clientes/${editando.id}`, payload)
      } else {
        await api.post('/api/v1/clientes', { ...payload, ativo: form.ativo })
      }
      fecharModal()
      void carregar()
    } catch (err) {
      if (err instanceof ApiError) {
        setErroForm(err.message ?? 'Erro ao salvar cliente.')
      } else {
        setErroForm('Erro inesperado.')
      }
    } finally {
      setSalvando(false)
    }
  }

  function formatarCPF(cpf: string) {
    return cpf.replace(/^(\d{3})(\d{3})(\d{3})(\d{2})$/, '$1.$2.$3-$4')
  }

  const colunas: Column<Cliente>[] = [
    {
      header: 'Nome',
      sortAccessor: (c) => c.nome,
      cell: (c) => (
        <div>
          <p className="font-medium text-gray-900">{c.nome}</p>
          {c.cidade && <p className="text-xs text-gray-500">{c.cidade}{c.uf ? ` / ${c.uf}` : ''}</p>}
        </div>
      ),
    },
    { header: 'CPF', hideBelow: 'sm', sortAccessor: (c) => c.cpf, cell: (c) => <span className="text-gray-600">{formatarCPF(c.cpf)}</span> },
    { header: 'E-mail', hideBelow: 'md', sortAccessor: (c) => c.email, cell: (c) => <span className="text-gray-600">{c.email}</span> },
    { header: 'Telefone', hideBelow: 'md', sortAccessor: (c) => c.telefone ?? '', cell: (c) => <span className="text-gray-600">{c.telefone ?? '—'}</span> },
    {
      header: 'Status',
      sortAccessor: (c) => (c.ativo ? 1 : 0),
      cell: (c) => <StatusBadge tone={c.ativo ? 'success' : 'neutral'}>{c.ativo ? 'Ativo' : 'Inativo'}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (c) => (
        <button onClick={() => abrirEditar(c)} className="text-gray-400 hover:text-gray-700" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  // ── render ──────────────────────────────────────────────────────────────────

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex items-center gap-4">
        <button onClick={() => navigate('/')} className="text-gray-400 hover:text-gray-700">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div>
          <h1 className="text-base font-semibold text-gray-900">Clientes</h1>
          <p className="text-xs text-gray-500">Cadastro e gestão de clientes</p>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-6 py-8 space-y-4">
        <div className="flex items-center gap-3">
          <form onSubmit={handleBusca} className="flex flex-1 gap-2">
            <input
              type="text"
              placeholder="Buscar por nome, CPF ou e-mail…"
              value={busca}
              onChange={(e) => setBusca(e.target.value)}
              className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
            />
            <button
              type="submit"
              className="px-4 py-2 text-sm bg-gray-900 text-white rounded-md hover:bg-gray-800"
            >
              Buscar
            </button>
          </form>
          <button
            onClick={abrirCriar}
            className="flex items-center gap-2 px-4 py-2 text-sm bg-gray-900 text-white rounded-md hover:bg-gray-800"
          >
            <Plus className="w-4 h-4" />
            Novo
          </button>
        </div>

        {erro && <p className="text-sm text-red-600">{erro}</p>}

        <DataTable
          columns={colunas}
          rows={itens}
          rowKey={(c) => c.id}
          loading={carregando}
          empty="Nenhum cliente encontrado."
        />

        {!carregando && itens.length === limite && (
          <div className="flex justify-end gap-2">
            {pagina > 0 && (
              <button
                onClick={() => { const p = pagina - 1; setPagina(p); void carregar(busca, p) }}
                className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50"
              >
                Anterior
              </button>
            )}
            <button
              onClick={() => { const p = pagina + 1; setPagina(p); void carregar(busca, p) }}
              className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50"
            >
              Próxima
            </button>
          </div>
        )}
      </main>

      {modalAberto && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h2 className="text-base font-semibold text-gray-900">
                {editando ? 'Editar Cliente' : 'Novo Cliente'}
              </h2>
              <button onClick={fecharModal} className="text-gray-400 hover:text-gray-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={(e) => { void handleSalvar(e) }} className="px-6 py-4 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Campo label="CPF *" disabled={!!editando}>
                  <input
                    required
                    disabled={!!editando}
                    placeholder="000.000.000-00"
                    value={form.cpf}
                    onChange={(e) => campo('cpf', e.target.value)}
                    className={inputCls(!!editando)}
                  />
                </Campo>
                <Campo label="E-mail *">
                  <input
                    required
                    type="email"
                    value={form.email}
                    onChange={(e) => campo('email', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Campo label="Nome *">
                  <input
                    required
                    value={form.nome}
                    onChange={(e) => campo('nome', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
                <Campo label="Telefone">
                  <input
                    value={form.telefone}
                    onChange={(e) => campo('telefone', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
              </div>

              <p className="text-xs font-medium text-gray-400 uppercase tracking-widest pt-2">Endereço</p>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div className="col-span-2">
                  <Campo label={`CEP${buscandoCep ? ' (buscando…)' : ''}`}>
                    <input
                      value={form.cep}
                      placeholder="00000-000"
                      onChange={(e) => {
                        campo('cep', e.target.value)
                        void buscarCep(e.target.value)
                      }}
                      className={inputCls()}
                    />
                  </Campo>
                </div>
                <Campo label="Número">
                  <input value={form.numero} onChange={(e) => campo('numero', e.target.value)} className={inputCls()} />
                </Campo>
                <Campo label="Complemento">
                  <input value={form.complemento} onChange={(e) => campo('complemento', e.target.value)} className={inputCls()} />
                </Campo>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="sm:col-span-2">
                  <Campo label="Rua">
                    <input value={form.rua} onChange={(e) => campo('rua', e.target.value)} className={inputCls()} />
                  </Campo>
                </div>
                <Campo label="Bairro">
                  <input value={form.bairro} onChange={(e) => campo('bairro', e.target.value)} className={inputCls()} />
                </Campo>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
                <div className="sm:col-span-3">
                  <Campo label="Cidade">
                    <input value={form.cidade} onChange={(e) => campo('cidade', e.target.value)} className={inputCls()} />
                  </Campo>
                </div>
                <Campo label="UF">
                  <input maxLength={2} value={form.uf} onChange={(e) => campo('uf', e.target.value.toUpperCase())} className={inputCls()} />
                </Campo>
              </div>

              {editando && (
                <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={form.ativo}
                    onChange={(e) => campo('ativo', e.target.checked)}
                    className="rounded border-gray-300"
                  />
                  Ativo
                </label>
              )}

              {erroForm && <p className="text-sm text-red-600">{erroForm}</p>}

              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={fecharModal}
                  className="px-4 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={salvando}
                  className="px-4 py-2 text-sm bg-gray-900 text-white rounded-md hover:bg-gray-800 disabled:opacity-50"
                >
                  {salvando ? 'Salvando…' : editando ? 'Salvar' : 'Criar'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

// ── helpers ────────────────────────────────────────────────────────────────────

function Campo({ label, children, disabled }: { label: string; children: React.ReactNode; disabled?: boolean }) {
  return (
    <div>
      <label className={`block text-sm font-medium mb-1 ${disabled ? 'text-gray-400' : 'text-gray-700'}`}>
        {label}
      </label>
      {children}
    </div>
  )
}

function inputCls(disabled = false) {
  return `w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900 ${
    disabled ? 'border-gray-200 bg-gray-50 text-gray-400 cursor-not-allowed' : 'border-gray-300'
  }`
}
