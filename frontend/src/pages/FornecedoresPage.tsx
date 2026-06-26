import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Pencil, Plus, X } from 'lucide-react'
import { api, ApiError } from '@/lib/api'

// ── tipos ──────────────────────────────────────────────────────────────────────

interface Fornecedor {
  id: string
  cnpj: string
  razao_social: string
  nome_fantasia: string
  email: string
  telefone1: string
  telefone2?: string
  cep?: string
  rua?: string
  numero?: string
  complemento?: string
  bairro?: string
  cidade?: string
  uf?: string
  comercial: string
  financeiro?: string
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

const vazio: Omit<Fornecedor, 'id' | 'ativo'> = {
  cnpj: '', razao_social: '', nome_fantasia: '', email: '',
  telefone1: '', telefone2: '', cep: '', rua: '', numero: '',
  complemento: '', bairro: '', cidade: '', uf: '',
  comercial: '', financeiro: '',
}

// ── componente principal ───────────────────────────────────────────────────────

export default function FornecedoresPage() {
  const navigate = useNavigate()

  const [itens, setItens] = useState<Fornecedor[]>([])
  const [pagina, setPagina] = useState(0)
  const [busca, setBusca] = useState('')
  const [carregando, setCarregando] = useState(false)
  const [erro, setErro] = useState('')

  const [modalAberto, setModalAberto] = useState(false)
  const [editando, setEditando] = useState<Fornecedor | null>(null)
  const [form, setForm] = useState({ ...vazio, ativo: true })
  const [erroForm, setErroForm] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [buscandoCep, setBuscandoCep] = useState(false)

  const limite = 20

  async function carregar(q = busca, pg = pagina) {
    setCarregando(true)
    setErro('')
    try {
      const res = await api.get<{ items: Fornecedor[] }>(
        `/api/v1/fornecedores?q=${encodeURIComponent(q)}&limit=${limite}&offset=${pg * limite}`,
      )
      setItens(res.items ?? [])
    } catch {
      setErro('Não foi possível carregar os fornecedores.')
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

  function abrirEditar(f: Fornecedor) {
    setEditando(f)
    setForm({
      cnpj: f.cnpj, razao_social: f.razao_social, nome_fantasia: f.nome_fantasia,
      email: f.email, telefone1: f.telefone1, telefone2: f.telefone2 ?? '',
      cep: f.cep ?? '', rua: f.rua ?? '', numero: f.numero ?? '',
      complemento: f.complemento ?? '', bairro: f.bairro ?? '',
      cidade: f.cidade ?? '', uf: f.uf ?? '',
      comercial: f.comercial, financeiro: f.financeiro ?? '',
      ativo: f.ativo,
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
      const end = await api.get<EnderecoViaCep>(`/api/v1/fornecedores/cep/${digitos}`)
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
        cnpj: form.cnpj,
        razao_social: form.razao_social,
        nome_fantasia: form.nome_fantasia,
        email: form.email,
        telefone1: form.telefone1,
        telefone2: form.telefone2,
        cep: form.cep,
        rua: form.rua,
        numero: form.numero,
        complemento: form.complemento,
        bairro: form.bairro,
        cidade: form.cidade,
        uf: form.uf,
        comercial: form.comercial,
        financeiro: form.financeiro,
        ativo: form.ativo,
      }
      if (editando) {
        await api.put(`/api/v1/fornecedores/${editando.id}`, payload)
      } else {
        await api.post('/api/v1/fornecedores', payload)
      }
      fecharModal()
      void carregar()
    } catch (err) {
      if (err instanceof ApiError) {
        setErroForm(err.message ?? 'Erro ao salvar fornecedor.')
      } else {
        setErroForm('Erro inesperado.')
      }
    } finally {
      setSalvando(false)
    }
  }

  // ── render ──────────────────────────────────────────────────────────────────

  return (
    <div className="min-h-screen bg-gray-50">
      {/* cabeçalho */}
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex items-center gap-4">
        <button onClick={() => navigate('/')} className="text-gray-400 hover:text-gray-700">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div>
          <h1 className="text-base font-semibold text-gray-900">Fornecedores</h1>
          <p className="text-xs text-gray-500">Cadastro e gestão de fornecedores</p>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-6 py-8 space-y-4">
        {/* barra de ações */}
        <div className="flex items-center gap-3">
          <form onSubmit={handleBusca} className="flex flex-1 gap-2">
            <input
              type="text"
              placeholder="Buscar por razão social, nome fantasia ou CNPJ…"
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

        {/* feedback */}
        {erro && <p className="text-sm text-red-600">{erro}</p>}

        {/* tabela */}
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          {carregando ? (
            <p className="p-6 text-sm text-gray-500 text-center">Carregando…</p>
          ) : itens.length === 0 ? (
            <p className="p-6 text-sm text-gray-500 text-center">Nenhum fornecedor encontrado.</p>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Razão Social</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600 hidden sm:table-cell">CNPJ</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600 hidden md:table-cell">Email</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600 hidden md:table-cell">Telefone</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {itens.map((f) => (
                  <tr key={f.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <p className="font-medium text-gray-900">{f.razao_social}</p>
                      <p className="text-xs text-gray-500">{f.nome_fantasia}</p>
                    </td>
                    <td className="px-4 py-3 text-gray-600 hidden sm:table-cell">
                      {f.cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5')}
                    </td>
                    <td className="px-4 py-3 text-gray-600 hidden md:table-cell">{f.email}</td>
                    <td className="px-4 py-3 text-gray-600 hidden md:table-cell">{f.telefone1}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                          f.ativo ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'
                        }`}
                      >
                        {f.ativo ? 'Ativo' : 'Inativo'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => abrirEditar(f)}
                        className="text-gray-400 hover:text-gray-700"
                        title="Editar"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* paginação */}
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

      {/* modal de cadastro / edição */}
      {modalAberto && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            {/* cabeçalho do modal */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h2 className="text-base font-semibold text-gray-900">
                {editando ? 'Editar Fornecedor' : 'Novo Fornecedor'}
              </h2>
              <button onClick={fecharModal} className="text-gray-400 hover:text-gray-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* formulário */}
            <form onSubmit={(e) => { void handleSalvar(e) }} className="px-6 py-4 space-y-4">
              {/* CNPJ */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Campo label="CNPJ *" disabled={!!editando}>
                  <input
                    required
                    disabled={!!editando}
                    placeholder="00.000.000/0000-00"
                    value={form.cnpj}
                    onChange={(e) => campo('cnpj', e.target.value)}
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

              {/* razão social + nome fantasia */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Campo label="Razão Social *">
                  <input
                    required
                    value={form.razao_social}
                    onChange={(e) => campo('razao_social', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
                <Campo label="Nome Fantasia *">
                  <input
                    required
                    value={form.nome_fantasia}
                    onChange={(e) => campo('nome_fantasia', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
              </div>

              {/* telefones */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Campo label="Telefone 1 *">
                  <input
                    required
                    value={form.telefone1}
                    onChange={(e) => campo('telefone1', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
                <Campo label="Telefone 2">
                  <input
                    value={form.telefone2}
                    onChange={(e) => campo('telefone2', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
              </div>

              {/* contatos */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Campo label="Contato Comercial *">
                  <input
                    required
                    value={form.comercial}
                    onChange={(e) => campo('comercial', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
                <Campo label="Contato Financeiro">
                  <input
                    value={form.financeiro}
                    onChange={(e) => campo('financeiro', e.target.value)}
                    className={inputCls()}
                  />
                </Campo>
              </div>

              {/* endereço */}
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

              {/* ativo (somente edição) */}
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
