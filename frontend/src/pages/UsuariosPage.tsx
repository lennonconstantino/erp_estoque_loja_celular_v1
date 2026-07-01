import { useEffect, useState } from 'react'
import { Pencil, Plus, ShieldAlert } from 'lucide-react'
import { toast } from 'sonner'
import { api, ApiError } from '@/lib/api'
import { getUserId } from '@/lib/auth'
import { cn } from '@/lib/utils'
import { DataTable, type Column } from '@/components/ui/data-table'
import { StatusBadge } from '@/components/ui/badge'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { Field, inputClasses } from '@/components/ui/field'
import { Modal } from '@/components/ui/modal'

// ── tipos ──────────────────────────────────────────────────────────────────────

interface Usuario {
  id: string
  nome: string
  email: string
  ativo: boolean
  ult_acesso?: string
  criado_em: string
}

const vazio = { nome: '', email: '', senha: '', admin: false, ativo: true }

// ── helpers ─────────────────────────────────────────────────────────────────────

function formatarData(iso?: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

// ── componente principal ───────────────────────────────────────────────────────

export default function UsuariosPage() {
  const [itens, setItens] = useState<Usuario[]>([])
  const [pagina, setPagina] = useState(0)
  const [carregando, setCarregando] = useState(true)
  const [erro, setErro] = useState('')
  const [restrito, setRestrito] = useState(false)

  const [modalAberto, setModalAberto] = useState(false)
  const [editando, setEditando] = useState<Usuario | null>(null)
  const [form, setForm] = useState({ ...vazio })
  const [erroForm, setErroForm] = useState('')
  const [salvando, setSalvando] = useState(false)

  const limite = 20

  async function carregar(pg = pagina) {
    setCarregando(true)
    setErro('')
    try {
      const res = await api.get<{ items: Usuario[] }>(
        `/api/v1/usuarios?limit=${limite}&offset=${pg * limite}`,
      )
      setItens(res.items ?? [])
    } catch (err) {
      // O backend é a autoridade sobre a permissão: 403 ⇒ não é admin. Assim,
      // um admin cujas perms ainda não estão no localStorage (sessão antiga)
      // não é barrado por engano — a própria chamada dispara o refresh.
      if (err instanceof ApiError && err.status === 403) {
        setRestrito(true)
      } else {
        setErro('Não foi possível carregar os usuários.')
      }
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => { void carregar() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function abrirCriar() {
    setEditando(null)
    setForm({ ...vazio })
    setErroForm('')
    setModalAberto(true)
  }

  function abrirEditar(u: Usuario) {
    setEditando(u)
    setForm({ nome: u.nome, email: u.email, senha: '', admin: false, ativo: u.ativo })
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

  // Impede o admin logado de desativar a própria conta — como `ADMIN` é o único
  // papel semeado, isso o trancaria para fora sem caminho de recuperação in-app.
  const souEuMesmo = !!editando && getUserId() === editando.id

  async function handleSalvar(e: React.FormEvent) {
    e.preventDefault()
    setErroForm('')
    if (souEuMesmo && !form.ativo) {
      setErroForm('Você não pode desativar a própria conta.')
      return
    }
    setSalvando(true)
    try {
      if (editando) {
        const payload: Record<string, unknown> = {
          nome: form.nome,
          email: form.email,
          ativo: form.ativo,
        }
        if (form.senha) payload.senha = form.senha
        await api.patch(`/api/v1/usuarios/${editando.id}`, payload)
        toast.success('Usuário atualizado.')
      } else {
        await api.post('/api/v1/usuarios', {
          nome: form.nome,
          email: form.email,
          senha: form.senha,
          papeis: form.admin ? ['ADMIN'] : [],
        })
        toast.success('Usuário criado.')
      }
      fecharModal()
      void carregar()
    } catch (err) {
      let msg = err instanceof ApiError ? err.message : 'Erro inesperado.'
      // O backend só checa e-mail duplicado na criação; na edição a colisão cai
      // em 500 genérico ("erro interno"). Dá um recado útil se o e-mail mudou.
      if (err instanceof ApiError && err.status >= 500 && editando && form.email !== editando.email) {
        msg = 'Não foi possível salvar. O e-mail informado pode já estar em uso por outro usuário.'
      }
      setErroForm(msg || 'Erro ao salvar usuário.')
    } finally {
      setSalvando(false)
    }
  }

  const colunas: Column<Usuario>[] = [
    {
      header: 'Nome',
      sortAccessor: (u) => u.nome,
      cell: (u) => <p className="font-bold text-foreground leading-tight">{u.nome}</p>,
    },
    {
      header: 'E-mail',
      hideBelow: 'sm',
      sortAccessor: (u) => u.email,
      cell: (u) => <span className="text-muted-foreground">{u.email}</span>,
    },
    {
      header: 'Últ. acesso',
      hideBelow: 'md',
      sortAccessor: (u) => u.ult_acesso ?? '',
      cell: (u) => <span className="text-muted-foreground">{formatarData(u.ult_acesso)}</span>,
      isTechnical: true,
    },
    {
      header: 'Criado em',
      hideBelow: 'md',
      sortAccessor: (u) => u.criado_em,
      cell: (u) => <span className="text-muted-foreground">{formatarData(u.criado_em)}</span>,
      isTechnical: true,
    },
    {
      header: 'Status',
      sortAccessor: (u) => (u.ativo ? 1 : 0),
      cell: (u) => <StatusBadge tone={u.ativo ? 'success' : 'neutral'}>{u.ativo ? 'Ativo' : 'Inativo'}</StatusBadge>,
    },
    {
      header: '',
      align: 'right',
      cell: (u) => (
        <button onClick={() => abrirEditar(u)} className="text-muted-foreground hover:text-foreground transition-colors p-2 rounded-full hover:bg-muted" aria-label="Editar" title="Editar">
          <Pencil className="w-4 h-4" />
        </button>
      ),
    },
  ]

  // Guarda de UI: o backend respondeu 403 (sem `iam:admin`). A autorização real
  // é sempre do backend — isto só troca o erro cru por um estado amigável.
  if (restrito) {
    return (
      <PageShell title="Usuários" subtitle="Gestão de contas de acesso" maxWidth="max-w-6xl">
        <div className="flex flex-col items-center justify-center gap-3 bg-card border border-border rounded-2xl p-12 text-center shadow-sm">
          <ShieldAlert className="w-8 h-8 text-muted-foreground" />
          <h2 className="text-sm font-bold uppercase tracking-widest text-foreground">Acesso restrito</h2>
          <p className="text-sm text-muted-foreground max-w-md">
            Somente administradores podem gerenciar usuários. Fale com um administrador se precisar de acesso.
          </p>
        </div>
      </PageShell>
    )
  }

  return (
    <PageShell
      title="Usuários"
      subtitle="Gestão de contas de acesso ao sistema"
      maxWidth="max-w-6xl"
      actions={
        <Button onClick={abrirCriar}>
          <Plus className="w-4 h-4" />
          Novo Usuário
        </Button>
      }
    >
      {erro && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erro}</p>}

      <DataTable
        columns={colunas}
        rows={itens}
        rowKey={(u) => u.id}
        loading={carregando}
        empty="Nenhum usuário encontrado."
      />

      {(itens.length === limite || pagina > 0) && (
        <div className="flex justify-between items-center bg-card p-3 rounded-xl border border-border shadow-sm">
          <Button
            variant="secondary"
            size="sm"
            disabled={pagina === 0}
            onClick={() => { const p = pagina - 1; setPagina(p); void carregar(p) }}
          >
            Anterior
          </Button>
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em]">Página {pagina + 1}</span>
          <Button
            variant="secondary"
            size="sm"
            disabled={itens.length < limite}
            onClick={() => { const p = pagina + 1; setPagina(p); void carregar(p) }}
          >
            Próxima
          </Button>
        </div>
      )}

      {modalAberto && (
        <Modal title={editando ? 'Editar Usuário' : 'Novo Usuário'} onClose={fecharModal} maxWidth="max-w-lg">
          <form onSubmit={(e) => { void handleSalvar(e) }} className="px-6 py-6 space-y-6">
            <Field label="Nome *">
              <input
                required
                value={form.nome}
                onChange={(e) => campo('nome', e.target.value)}
                className={inputClasses()}
                placeholder="Nome do usuário"
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
                placeholder="usuario@loja.local"
              />
            </Field>

            <Field label={editando ? 'Nova senha (deixe em branco para manter)' : 'Senha * (mín. 8 caracteres)'}>
              <input
                required={!editando}
                type="password"
                autoComplete="new-password"
                minLength={8}
                value={form.senha}
                onChange={(e) => campo('senha', e.target.value)}
                className={inputClasses()}
                placeholder="••••••••"
              />
            </Field>

            {editando ? (
              <div className="space-y-1.5">
                <label className={cn('flex items-center gap-3 text-xs font-bold uppercase tracking-widest transition-colors', souEuMesmo ? 'text-muted-foreground/60 cursor-not-allowed' : 'text-muted-foreground cursor-pointer hover:text-foreground')}>
                  <input
                    type="checkbox"
                    checked={form.ativo}
                    disabled={souEuMesmo}
                    onChange={(e) => campo('ativo', e.target.checked)}
                    className="w-4 h-4 rounded border-border bg-muted/20 text-primary focus:ring-primary disabled:cursor-not-allowed"
                  />
                  Usuário ativo no sistema
                </label>
                {souEuMesmo && (
                  <p className="text-[10px] text-muted-foreground ml-1">
                    Você não pode desativar a própria conta.
                  </p>
                )}
              </div>
            ) : (
              <label className="flex items-center gap-3 text-xs font-bold text-muted-foreground uppercase tracking-widest cursor-pointer hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={form.admin}
                  onChange={(e) => campo('admin', e.target.checked)}
                  className="w-4 h-4 rounded border-border bg-muted/20 text-primary focus:ring-primary"
                />
                Administrador (acesso total)
              </label>
            )}

            {erroForm && <p role="alert" className="text-xs text-destructive font-bold bg-destructive/10 border border-destructive/20 rounded-full px-4 py-2 w-fit uppercase tracking-wider">{erroForm}</p>}

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button type="button" variant="secondary" onClick={fecharModal}>Cancelar</Button>
              <Button type="submit" disabled={salvando} className="min-w-32">
                {salvando ? 'Salvando…' : editando ? 'Salvar Alterações' : 'Cadastrar Usuário'}
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </PageShell>
  )
}
