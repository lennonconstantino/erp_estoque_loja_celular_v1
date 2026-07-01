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
        // Na edição o checkbox "Cliente ativo" é enviado e honrado pelo backend.
        await api.put(`/api/v1/clientes/${editando.id}`, { ...payload, ativo: form.ativo })
      } else {
        // Na criação o backend cria todo cliente como ativo (domain.NovoCliente);
        // não enviamos `ativo` (o form de criação nem exibe o controle).
        await api.post('/api/v1/clientes', payload)
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
        <div className="flex flex-col">
          <p className="font-bold text-foreground leading-tight">{c.nome}</p>
          {c.cidade && <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-wider mt-1">{c.cidade}{c.uf ? ` / ${c.uf}` : ''}</p>}
        </div>
      ),
    },
    { header: 'CPF', hideBelow: 'sm', sortAccessor: (c) => c.cpf, cell: (c) => <span className="text-muted-foreground font-mono">{formatarCPF(c.cpf)}</span>, isTechnical: true },
    { header: 'E-mail', hideBelow: 'md', sortAccessor: (c) => c.email, cell: (c) => <span className="text-muted-foreground">{c.email}</span> },
    { header: 'Telefone', hideBelow: 'md', sortAccessor: (c) => c.telefone ?? '', cell: (c) => <span className="text-muted-foreground">{c.telefone ?? '—'}</span> },
    {
      header: 'Status',
      sortAccessor: (c) => (c.ativo ? 1 : 0),
      cell: (c) => <StatusBadge tone={c.ativo ? 'success' : 'neutral'}>{c.ativo ? 'Ativo' : 'Inativo'}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (c) => (
        <button onClick={() => abrirEditar(c)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted" aria-label="Editar" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <PageShell 
      title="Clientes" 
      subtitle="Cadastro e gestão de clientes" 
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirCriar}>
          <Plus className="w-4 h-4" />
          Novo Cliente
        </Button>
      }
    >
      <div className="flex items-center gap-3 bg-card p-4 rounded-2xl border border-border shadow-sm animate-in fade-in duration-500">
        <form onSubmit={handleBusca} className="flex flex-1 gap-2">
          <input
            type="text"
            placeholder="Buscar por nome, CPF ou e-mail…"
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
        rowKey={(c) => c.id}
        loading={carregando}
        empty="Nenhum cliente encontrado."
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
        <Modal title={editando ? 'Editar Cliente' : 'Novo Cliente'} onClose={fecharModal} maxWidth="max-w-2xl">
          <form onSubmit={(e) => { void handleSalvar(e) }} className="px-6 py-6 space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <Field label="CPF *" disabled={!!editando}>
                <input
                  required
                  disabled={!!editando}
                  inputMode="numeric"
                  placeholder="000.000.000-00"
                  value={form.cpf}
                  onChange={(e) => campo('cpf', e.target.value)}
                  className={inputClasses(!!editando)}
                />
              </Field>
              <Field label="E-mail *">
                <input
                  required
                  type="email"
                  autoComplete="email"
                  inputMode="email"
                  value={form.email}
                  onChange={(e) => campo('email', e.target.value)}
                  className={inputClasses()}
                  placeholder="cliente@exemplo.com"
                />
              </Field>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <Field label="Nome Completo *">
                <input
                  required
                  value={form.nome}
                  onChange={(e) => campo('nome', e.target.value)}
                  className={inputClasses()}
                  placeholder="Nome do cliente"
                />
              </Field>
              <Field label="Telefone">
                <input
                  type="tel"
                  autoComplete="tel"
                  inputMode="tel"
                  value={form.telefone}
                  onChange={(e) => campo('telefone', e.target.value)}
                  className={inputClasses()}
                  placeholder="(00) 00000-0000"
                />
              </Field>
            </div>

            <div className="pt-2">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] mb-4 border-b border-border pb-2">Endereço de Entrega</p>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-6">
                <div className="col-span-2">
                  <Field label={`CEP${buscandoCep ? ' (Buscando…)' : ''}`}>
                    <input
                      value={form.cep}
                      autoComplete="postal-code"
                      inputMode="numeric"
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
                  <input value={form.numero} onChange={(e) => campo('numero', e.target.value)} inputMode="numeric" className={inputClasses()} placeholder="123" />
                </Field>
                <Field label="Complemento">
                  <input value={form.complemento} onChange={(e) => campo('complemento', e.target.value)} autoComplete="address-line2" className={inputClasses()} placeholder="Apt 42" />
                </Field>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mt-6">
                <div className="sm:col-span-2">
                  <Field label="Rua / Logradouro">
                    <input value={form.rua} onChange={(e) => campo('rua', e.target.value)} autoComplete="address-line1" className={inputClasses()} placeholder="Nome da rua" />
                  </Field>
                </div>
                <Field label="Bairro">
                  <input value={form.bairro} onChange={(e) => campo('bairro', e.target.value)} className={inputClasses()} placeholder="Nome do bairro" />
                </Field>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-4 gap-6 mt-6">
                <div className="sm:col-span-3">
                  <Field label="Cidade">
                    <input value={form.cidade} onChange={(e) => campo('cidade', e.target.value)} autoComplete="address-level2" className={inputClasses()} placeholder="Ex: São Paulo" />
                  </Field>
                </div>
                <Field label="UF">
                  <input maxLength={2} value={form.uf} onChange={(e) => campo('uf', e.target.value.toUpperCase())} autoComplete="address-level1" className={inputClasses()} placeholder="SP" />
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
                Cliente ativo no sistema
              </label>
            )}

            {erroForm && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erroForm}</p>}

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button type="button" variant="secondary" onClick={fecharModal}>Cancelar</Button>
              <Button type="submit" disabled={salvando} className="min-w-32">
                {salvando ? 'Salvando…' : editando ? 'Salvar Alterações' : 'Cadastrar Cliente'}
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </PageShell>
  )
}
