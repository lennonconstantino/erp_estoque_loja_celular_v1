import { useEffect, useState } from 'react'
import { Pencil, Plus } from 'lucide-react'
import { api, ApiError } from '@/lib/api'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge } from '@/components/ui/badge'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { Field, inputClasses } from '@/components/ui/field'
import { Modal } from '@/components/ui/modal'

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
      // CEP não encontrado
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

  const colunas: Column<Fornecedor>[] = [
    {
      header: 'Fornecedor',
      sortAccessor: (f) => f.razao_social,
      cell: (f) => (
        <div className="flex flex-col">
          <p className="font-bold text-foreground leading-tight">{f.nome_fantasia || f.razao_social}</p>
          <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-wider mt-1">{f.razao_social}</p>
        </div>
      ),
    },
    {
      header: 'CNPJ',
      hideBelow: 'sm',
      sortAccessor: (f) => f.cnpj,
      cell: (f) => <span className="text-muted-foreground font-mono">{f.cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5')}</span>,
      isTechnical: true,
    },
    { header: 'Email', hideBelow: 'md', sortAccessor: (f) => f.email, cell: (f) => <span className="text-muted-foreground">{f.email}</span> },
    { header: 'Telefone', hideBelow: 'md', sortAccessor: (f) => f.telefone1, cell: (f) => <span className="text-muted-foreground">{f.telefone1}</span> },
    {
      header: 'Status',
      sortAccessor: (f) => (f.ativo ? 1 : 0),
      cell: (f) => <StatusBadge tone={f.ativo ? 'success' : 'neutral'}>{f.ativo ? 'Ativo' : 'Inativo'}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (f) => (
        <button onClick={() => abrirEditar(f)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted active:scale-90" aria-label="Editar" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell 
      title="Fornecedores" 
      subtitle="Cadastro e gestão de fornecedores" 
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirCriar}>
          <Plus className="w-4 h-4" />
          Novo Fornecedor
        </Button>
      }
    >
      <div className="flex items-center gap-3 bg-card p-4 rounded-2xl border border-border shadow-sm animate-in fade-in duration-500">
        <form onSubmit={handleBusca} className="flex flex-1 gap-2">
          <input
            type="text"
            placeholder="Buscar por razão social, nome fantasia ou CNPJ…"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            className={inputClasses() + ' flex-1'}
          />
          <Button type="submit" variant="secondary" className="px-8 h-10">Pesquisar</Button>
        </form>
      </div>

      {erro && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={itens}
        rowKey={(f) => f.id}
        loading={carregando}
        empty="Nenhum fornecedor encontrado."
      />

      {(itens.length === limite || pagina > 0) && (
        <div className="flex justify-between items-center bg-card p-3 rounded-xl border border-border shadow-sm">
          <Button
            variant="secondary"
            size="sm"
            disabled={pagina === 0}
            onClick={() => { const p = pagina - 1; setPagina(p); void carregar(busca, p) }}
          >
            Anterior
          </Button>
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em]">Página {pagina + 1}</span>
          <Button
            variant="secondary"
            size="sm"
            disabled={itens.length < limite}
            onClick={() => { const p = pagina + 1; setPagina(p); void carregar(busca, p) }}
          >
            Próxima
          </Button>
        </div>
      )}

      {modalAberto && (
        <Modal title={editando ? 'Editar Fornecedor' : 'Novo Fornecedor'} onClose={fecharModal} maxWidth="max-w-2xl">
          <form onSubmit={(e) => { void handleSalvar(e) }} className="px-6 py-6 space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <Field label="CNPJ *" disabled={!!editando}>
                <input
                  required
                  disabled={!!editando}
                  placeholder="00.000.000/0000-00"
                  value={form.cnpj}
                  onChange={(e) => campo('cnpj', e.target.value)}
                  className={inputClasses(!!editando)}
                />
              </Field>
              <Field label="E-mail *">
                <input
                  required
                  type="email"
                  value={form.email}
                  onChange={(e) => campo('email', e.target.value)}
                  className={inputClasses()}
                  placeholder="contato@fornecedor.com"
                />
              </Field>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <Field label="Razão Social *">
                <input
                  required
                  value={form.razao_social}
                  onChange={(e) => campo('razao_social', e.target.value)}
                  className={inputClasses()}
                  placeholder="Nome jurídico"
                />
              </Field>
              <Field label="Nome Fantasia *">
                <input
                  required
                  value={form.nome_fantasia}
                  onChange={(e) => campo('nome_fantasia', e.target.value)}
                  className={inputClasses()}
                  placeholder="Nome comercial"
                />
              </Field>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <Field label="Telefone Principal *">
                <input
                  required
                  value={form.telefone1}
                  onChange={(e) => campo('telefone1', e.target.value)}
                  className={inputClasses()}
                  placeholder="(00) 0000-0000"
                />
              </Field>
              <Field label="Telefone Secundário">
                <input
                  value={form.telefone2}
                  onChange={(e) => campo('telefone2', e.target.value)}
                  className={inputClasses()}
                  placeholder="(00) 0000-0000"
                />
              </Field>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <Field label="Contato Comercial *">
                <input
                  required
                  value={form.comercial}
                  onChange={(e) => campo('comercial', e.target.value)}
                  className={inputClasses()}
                  placeholder="Nome do vendedor/atendente"
                />
              </Field>
              <Field label="Contato Financeiro">
                <input
                  value={form.financeiro}
                  onChange={(e) => campo('financeiro', e.target.value)}
                  className={inputClasses()}
                  placeholder="Nome do responsável financeiro"
                />
              </Field>
            </div>

            <div className="pt-2">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] mb-4 border-b border-border pb-2">Endereço de Sede</p>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-6">
                <div className="col-span-2">
                  <Field label={`CEP${buscandoCep ? ' (Buscando…)' : ''}`}>
                    <input
                      value={form.cep}
                      placeholder="00000-000"
                      onChange={(e) => {
                        campo('cep', e.target.value)
                        void buscarCep(e.target.value)
                      }}
                      className={inputClasses()}
                    />
                  </Field>
                </div>
                <Field label="Número">
                  <input value={form.numero} onChange={(e) => campo('numero', e.target.value)} className={inputClasses()} placeholder="123" />
                </Field>
                <Field label="Complemento">
                  <input value={form.complemento} onChange={(e) => campo('complemento', e.target.value)} className={inputClasses()} placeholder="Sl 101" />
                </Field>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mt-6">
                <div className="sm:col-span-2">
                  <Field label="Rua / Avenida">
                    <input value={form.rua} onChange={(e) => campo('rua', e.target.value)} className={inputClasses()} placeholder="Logradouro" />
                  </Field>
                </div>
                <Field label="Bairro">
                  <input value={form.bairro} onChange={(e) => campo('bairro', e.target.value)} className={inputClasses()} placeholder="Bairro" />
                </Field>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-4 gap-6 mt-6">
                <div className="sm:col-span-3">
                  <Field label="Cidade">
                    <input value={form.cidade} onChange={(e) => campo('cidade', e.target.value)} className={inputClasses()} placeholder="Cidade" />
                  </Field>
                </div>
                <Field label="UF">
                  <input maxLength={2} value={form.uf} onChange={(e) => campo('uf', e.target.value.toUpperCase())} className={inputClasses()} placeholder="UF" />
                </Field>
              </div>
            </div>

            {editando && (
              <label className="flex items-center gap-3 text-xs font-bold text-muted-foreground uppercase tracking-widest cursor-pointer hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={form.ativo}
                  onChange={(e) => campo('ativo', e.target.checked)}
                  className="w-4 h-4 rounded border-border bg-muted/20 text-primary focus:ring-primary"
                />
                Fornecedor ativo para compras
              </label>
            )}

            {erroForm && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erroForm}</p>}

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button type="button" variant="secondary" onClick={fecharModal}>Cancelar</Button>
              <Button type="submit" disabled={salvando} className="min-w-32">
                {salvando ? 'Salvando…' : editando ? 'Salvar Alterações' : 'Cadastrar Fornecedor'}
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </PageShell>
  )
}
