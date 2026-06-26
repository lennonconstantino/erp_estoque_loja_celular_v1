import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Search, Trash2, CheckCircle, Save } from 'lucide-react'
import { api } from '@/lib/api'
import { PageShell } from '@/components/ui/page-shell'
import { Button } from '@/components/ui/button'
import { Field, inputClasses, inputClassesCompact, compactLabelClass } from '@/components/ui/field'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

interface Cliente {
  id: string
  nome: string
  cpf: string
  email: string
}

interface Produto {
  id: string
  descricao: string
  preco_venda: number
  estoque_atual: number
  disponivel: boolean
  ativo: boolean
}

interface ItemForm {
  produto_id: string
  quantidade: number
  preco_unitario: number
}

const FORMAS_PGTO = ['DINHEIRO', 'PIX', 'DEBITO', 'CREDITO', 'OUTRO'] as const
const DOCS_FISCAIS = ['CUPOM', 'NF'] as const

const FORMA_LABEL: Record<string, string> = {
  DINHEIRO: 'Dinheiro',
  PIX: 'PIX',
  DEBITO: 'Débito',
  CREDITO: 'Crédito',
  OUTRO: 'Outro',
}

export default function NovaVendaPage() {
  const navigate = useNavigate()
  const [produtos, setProdutos] = useState<Produto[]>([])
  const [cliente, setCliente] = useState<Cliente | null>(null)
  const [cpfBusca, setCpfBusca] = useState('')
  const [buscandoCliente, setBuscandoCliente] = useState(false)
  const [consumidorFinal, setConsumidorFinal] = useState(true)
  const [formaPgto, setFormaPgto] = useState<string>('DINHEIRO')
  const [docFiscal, setDocFiscal] = useState<string>('CUPOM')
  const [desconto, setDesconto] = useState(0)
  const [itens, setItens] = useState<ItemForm[]>([])
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  const totalItens = itens.reduce((s, i) => s + i.quantidade * i.preco_unitario, 0)
  const total = Math.max(0, totalItens - desconto)

  useEffect(() => {
    api.get<{ items: Produto[] }>('/api/v1/produtos?limit=200').then((d) => {
      setProdutos((d.items ?? []).filter((p) => p.ativo && p.disponivel))
    })
  }, [])

  async function buscarCliente() {
    if (!cpfBusca.trim()) return
    setBuscandoCliente(true)
    setErro('')
    try {
      const cpf = cpfBusca.replace(/\D/g, '')
      const data = await api.get<{ items: Cliente[] }>(`/api/v1/clientes?q=${cpf}&limit=1`)
      const encontrado = (data.items ?? [])[0]
      if (encontrado) {
        setCliente(encontrado)
        toast.success(`Cliente ${encontrado.nome} selecionado`)
      } else {
        setErro('Cliente não encontrado para este CPF.')
        setCliente(null)
        toast.error('CPF não localizado')
      }
    } catch {
      setErro('Erro ao buscar cliente.')
      setCliente(null)
    } finally {
      setBuscandoCliente(false)
    }
  }

  function adicionarItem() {
    setItens((prev) => [...prev, { produto_id: '', quantidade: 1, preco_unitario: 0 }])
  }

  function removerItem(idx: number) {
    setItens((prev) => prev.filter((_, i) => i !== idx))
  }

  function atualizarItem(idx: number, campo: keyof ItemForm, valor: string | number) {
    setItens((prev) =>
      prev.map((item, i) => {
        if (i !== idx) return item
        if (campo === 'produto_id') {
          const prod = produtos.find((p) => p.id === valor)
          return { ...item, produto_id: String(valor), preco_unitario: prod?.preco_venda ?? item.preco_unitario }
        }
        return { ...item, [campo]: Number(valor) }
      }),
    )
  }

  function saldoDisponivel(produtoId: string): number {
    const prod = produtos.find((p) => p.id === produtoId)
    return prod?.estoque_atual ?? 0
  }

  function validarItens(): string | null {
    for (const item of itens) {
      if (!item.produto_id) return 'Selecione um produto em todos os itens.'
      const saldo = saldoDisponivel(item.produto_id)
      if (item.quantidade > saldo) {
        const prod = produtos.find((p) => p.id === item.produto_id)
        return `Saldo insuficiente para "${prod?.descricao}": disponível ${saldo}, solicitado ${item.quantidade}.`
      }
      if (item.preco_unitario <= 0) return 'Preço unitário deve ser positivo.'
    }
    return null
  }

  async function salvarVenda(confirmar: boolean) {
    if (itens.length === 0) { setErro('Adicione pelo menos um item.'); return }
    const validacao = validarItens()
    if (validacao) { setErro(validacao); return }
    if (!consumidorFinal && !cliente) { setErro('Selecione um cliente ou marque como consumidor final.'); return }
    if (desconto > totalItens) { setErro('Desconto maior que o total dos itens.'); return }

    setSalvando(true)
    setErro('')
    try {
      const body = {
        consumidor_final: consumidorFinal,
        cliente_id: consumidorFinal ? '' : (cliente?.id ?? ''),
        forma_pgto: formaPgto,
        doc_fiscal: docFiscal,
        desconto,
        itens: itens.map((i) => ({
          produto_id: i.produto_id,
          quantidade: i.quantidade,
          preco_unitario: i.preco_unitario,
        })),
      }

      interface Venda { id: string; doc_fiscal_numero?: string }
      const venda = await api.post<Venda>('/api/v1/vendas', body)

      if (confirmar) {
        const confirmada = await api.post<Venda>(`/api/v1/vendas/${venda.id}/confirmar`, {})
        toast.success(`Venda confirmada! Documento: ${confirmada.doc_fiscal_numero || 'Emitido'}`)
      } else {
        toast.success('Rascunho de venda salvo')
      }
      navigate('/vendas')
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Erro ao salvar venda'
      setErro(msg)
      toast.error(msg)
    } finally {
      setSalvando(false)
    }
  }

  const fmtBrl = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

  return (
    <PageShell 
      title="Nova Venda" 
      subtitle="Ponto de venda e emissão de cupom/NF." 
      maxWidth="max-w-4xl"
      back="/vendas"
    >
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
        {/* Esquerda: Itens e Cliente */}
        <div className="lg:col-span-2 space-y-6">
          {/* Card Cliente */}
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] leading-none">Dados do Cliente</h2>
              <label className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider cursor-pointer hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={consumidorFinal}
                  onChange={(e) => { setConsumidorFinal(e.target.checked); setCliente(null); setCpfBusca('') }}
                  className="w-3.5 h-3.5 rounded border-border bg-muted/20 text-primary focus:ring-primary"
                />
                Consumidor Final
              </label>
            </div>
            
            {!consumidorFinal && (
              <div className="space-y-4 animate-in slide-in-from-top-2 duration-300">
                <div className="flex gap-2">
                  <div className="relative flex-1 group">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground group-focus-within:text-primary transition-colors" />
                    <input
                      type="text"
                      placeholder="CPF do cliente..."
                      value={cpfBusca}
                      onChange={(e) => setCpfBusca(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && buscarCliente()}
                      className={cn(inputClasses(), "pl-9")}
                    />
                  </div>
                  <Button 
                    onClick={buscarCliente} 
                    disabled={buscandoCliente}
                    variant="secondary"
                    className="h-10 px-6"
                  >
                    {buscandoCliente ? '...' : 'Buscar'}
                  </Button>
                </div>
                {cliente && (
                  <div className="p-4 bg-primary/5 border border-primary/20 rounded-xl flex justify-between items-center animate-in zoom-in-95 duration-200">
                    <div>
                      <p className="text-sm font-black text-foreground uppercase tracking-tight">{cliente.nome}</p>
                      <p className="text-[10px] text-muted-foreground font-mono mt-1">{cliente.cpf} · {cliente.email}</p>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Card Itens */}
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm">
            <div className="flex items-center justify-between mb-6 border-b border-border pb-4">
              <h2 className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] leading-none">Itens da Venda</h2>
              <button
                onClick={adicionarItem}
                className="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-primary hover:opacity-70 transition-opacity"
              >
                <Plus className="w-3 h-3" /> Incluir Produto
              </button>
            </div>

            {itens.length === 0 ? (
              <div className="text-center py-12 border-2 border-dashed border-border rounded-2xl bg-muted/5">
                <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">Nenhum item adicionado</p>
                <p className="text-[10px] text-muted-foreground/50 mt-2">Clique em "Incluir Produto" para começar</p>
              </div>
            ) : (
              <div className="space-y-4">
                {itens.map((item, idx) => {
                  const saldo = item.produto_id ? saldoDisponivel(item.produto_id) : null
                  const semEstoque = saldo !== null && item.quantidade > saldo
                  return (
                    <div key={idx} className={cn(
                      "p-4 border rounded-2xl transition-all group",
                      semEstoque ? "border-destructive/30 bg-destructive/5" : "border-border/50 bg-muted/5 hover:bg-muted/10"
                    )}>
                      <div className="grid grid-cols-12 gap-4 items-end">
                        <div className="col-span-12 md:col-span-6">
                          <label className={compactLabelClass}>Produto</label>
                          <select
                            value={item.produto_id}
                            onChange={(e) => atualizarItem(idx, 'produto_id', e.target.value)}
                            className={inputClassesCompact()}
                          >
                            <option value="">Selecione…</option>
                            {produtos.map((p) => (
                              <option key={p.id} value={p.id}>
                                {p.descricao} (Disponível: {p.estoque_atual})
                              </option>
                            ))}
                          </select>
                        </div>
                        <div className="col-span-4 md:col-span-2">
                          <label className={compactLabelClass}>Qtd</label>
                          <input
                            type="number"
                            min={1}
                            value={item.quantidade}
                            onChange={(e) => atualizarItem(idx, 'quantidade', e.target.value)}
                            className={cn(inputClassesCompact(), 'font-mono')}
                          />
                        </div>
                        <div className="col-span-5 md:col-span-3">
                          <label className={compactLabelClass}>Unitário</label>
                          <div className="relative">
                            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-muted-foreground opacity-50">R$</span>
                            <input
                              type="number"
                              min={0}
                              step={0.01}
                              value={item.preco_unitario}
                              onChange={(e) => atualizarItem(idx, 'preco_unitario', e.target.value)}
                              className={cn(inputClassesCompact(), 'font-mono pl-8')}
                            />
                          </div>
                        </div>
                        <div className="col-span-3 md:col-span-1 flex justify-center pb-1">
                          <button
                            onClick={() => removerItem(idx)}
                            className="p-2 text-muted-foreground hover:text-destructive transition-all rounded-full hover:bg-destructive/10 active:scale-90"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                      {item.produto_id && (
                        <div className="mt-3 flex justify-between items-center">
                          <p className={cn("text-[9px] font-black uppercase tracking-widest", semEstoque ? "text-destructive" : "text-green-500/50")}>
                            {semEstoque ? `SALDO INSUFICIENTE: APENAS ${saldo}` : "DISPONÍVEL EM ESTOQUE"}
                          </p>
                          <p className="text-xs font-black text-foreground font-mono">
                            {fmtBrl(item.quantidade * item.preco_unitario)}
                          </p>
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </div>

        {/* Direita: Pagamento e Total */}
        <div className="space-y-6 lg:sticky lg:top-24">
          <div className="bg-card border border-border rounded-2xl p-6 shadow-sm space-y-6">
            <h2 className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] leading-none mb-6">Resumo da Venda</h2>
            
            <div className="space-y-4">
              <Field label="Meio de Pagamento">
                <select
                  value={formaPgto}
                  onChange={(e) => setFormaPgto(e.target.value)}
                  className={inputClasses()}
                >
                  {FORMAS_PGTO.map((f) => (
                    <option key={f} value={f}>{FORMA_LABEL[f]}</option>
                  ))}
                </select>
              </Field>

              <Field label="Tipo de Emissão">
                <select
                  value={docFiscal}
                  onChange={(e) => setDocFiscal(e.target.value)}
                  className={inputClasses()}
                >
                  {DOCS_FISCAIS.map((d) => (
                    <option key={d} value={d}>{d}</option>
                  ))}
                </select>
              </Field>

              <Field label="Desconto Aplicado (R$)">
                <div className="relative">
                  <span className="absolute left-4 top-1/2 -translate-y-1/2 text-xs font-bold text-muted-foreground opacity-50">R$</span>
                  <input
                    type="number"
                    min={0}
                    step={0.01}
                    value={desconto}
                    onChange={(e) => setDesconto(parseFloat(e.target.value) || 0)}
                    className={cn(inputClasses(), "pl-10 font-mono")}
                  />
                </div>
              </Field>
            </div>

            <div className="bg-muted/30 rounded-2xl p-6 space-y-4 border border-border shadow-inner">
               <div className="flex justify-between text-[10px] text-muted-foreground font-black uppercase tracking-[0.1em]">
                  <span>Subtotal</span>
                  <span className="font-mono">{fmtBrl(totalItens)}</span>
               </div>
               <div className="flex justify-between text-[10px] text-destructive font-black uppercase tracking-[0.1em]">
                  <span>Descontos</span>
                  <span className="font-mono">− {fmtBrl(desconto)}</span>
               </div>
               <div className="pt-2 border-t border-border/50">
                 <div className="flex justify-between items-end">
                    <span className="text-xs font-black text-foreground uppercase tracking-[0.2em]">Total</span>
                    <span className="text-3xl font-black text-primary font-mono tracking-tighter">{fmtBrl(total)}</span>
                 </div>
               </div>
            </div>

            <div className="space-y-3 pt-2">
              <Button 
                onClick={() => salvarVenda(true)} 
                disabled={salvando} 
                className="w-full h-12 text-sm font-black uppercase tracking-widest"
              >
                {salvando ? '...' : <><CheckCircle className="w-4 h-4 mr-2" /> Confirmar Venda</>}
              </Button>
              <Button 
                onClick={() => salvarVenda(false)} 
                disabled={salvando} 
                variant="secondary"
                className="w-full h-10 text-[10px] font-black uppercase tracking-widest"
              >
                <Save className="w-3.5 h-3.5 mr-2 opacity-50" /> Salvar Rascunho
              </Button>
            </div>

            {erro && (
              <div className="p-3 bg-destructive/10 border border-destructive/20 text-destructive rounded-xl text-[10px] font-bold uppercase tracking-wider animate-in fade-in">
                {erro}
              </div>
            )}
          </div>
        </div>
      </div>
    </PageShell>
  )
}
