import { useMemo, useState, type ReactNode } from 'react'
import { ChevronDown, ChevronsUpDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'

export type SortValue = string | number | null | undefined

export interface Column<T> {
  header: ReactNode
  cell: (row: T) => ReactNode
  align?: 'left' | 'right' | 'center'
  hideBelow?: 'sm' | 'md'
  thClassName?: string
  tdClassName?: string
  /** Torna a coluna ordenável: retorna o valor usado na comparação. */
  sortAccessor?: (row: T) => SortValue
  /** Aplica fonte mono e tamanho reduzido para dados técnicos (SKU, preços). */
  isTechnical?: boolean
}

interface DataTableProps<T> {
  columns: Column<T>[]
  rows: T[]
  rowKey: (row: T) => string
  loading?: boolean
  empty?: string
  rowClassName?: (row: T) => string | undefined
}

const alignCls = { left: 'text-left', right: 'text-right', center: 'text-center' } as const
const hideCls = { sm: 'hidden sm:table-cell', md: 'hidden md:table-cell' } as const

type SortDir = 'asc' | 'desc'

function comparar(a: SortValue, b: SortValue): number {
  if (a == null && b == null) return 0
  if (a == null) return -1
  if (b == null) return 1
  if (typeof a === 'number' && typeof b === 'number') return a - b
  return String(a).localeCompare(String(b), 'pt-BR', { numeric: true, sensitivity: 'base' })
}

/** Tabela técnica (cartão escuro/claro adaptável, cabeçalho minimalista, tipografia mono). */
export function DataTable<T>({
  columns,
  rows,
  rowKey,
  loading,
  empty = 'Nenhum registro encontrado.',
  rowClassName,
}: DataTableProps<T>) {
  const [sort, setSort] = useState<{ index: number; dir: SortDir } | null>(null)

  const rowsOrdenadas = useMemo(() => {
    if (!sort) return rows
    const acessar = columns[sort.index]?.sortAccessor
    if (!acessar) return rows
    const fator = sort.dir === 'asc' ? 1 : -1
    return [...rows].sort((a, b) => fator * comparar(acessar(a), acessar(b)))
  }, [rows, sort, columns])

  function alternarOrdenacao(index: number) {
    setSort((prev) => {
      if (!prev || prev.index !== index) return { index, dir: 'asc' }
      if (prev.dir === 'asc') return { index, dir: 'desc' }
      return null
    })
  }

  return (
    <div className="bg-card rounded-xl border border-border overflow-hidden overflow-x-auto shadow-sm">
      {loading ? (
        <p className="p-6 text-sm text-muted-foreground text-center">Carregando…</p>
      ) : rowsOrdenadas.length === 0 ? (
        <p className="p-6 text-sm text-muted-foreground text-center">{empty}</p>
      ) : (
        <table className="w-full text-sm">
          <thead className="bg-muted/10 border-b border-border">
            <tr>
              {columns.map((c, i) => {
                const ordenavel = !!c.sortAccessor
                const ativa = sort?.index === i
                const baseTh = cn(
                  'px-4 py-3 font-semibold text-muted-foreground tracking-tight uppercase text-[10px]',
                  alignCls[c.align ?? 'left'],
                  c.hideBelow && hideCls[c.hideBelow],
                  c.thClassName,
                )
                if (!ordenavel) {
                  return <th key={i} className={baseTh}>{c.header === '' ? <span className="sr-only">Ações</span> : c.header}</th>
                }
                return (
                  <th key={i} className={baseTh} aria-sort={ativa ? (sort!.dir === 'asc' ? 'ascending' : 'descending') : 'none'}>
                    <button
                      type="button"
                      onClick={() => alternarOrdenacao(i)}
                      className={cn(
                        'inline-flex items-center gap-1 hover:text-foreground transition-colors outline-none',
                        c.align === 'right' && 'flex-row-reverse',
                        c.align === 'center' && 'mx-auto',
                        ativa ? 'text-foreground' : 'text-muted-foreground',
                      )}
                    >
                      {c.header}
                      {ativa ? (
                        sort!.dir === 'asc'
                          ? <ChevronUp className="w-3 h-3" />
                          : <ChevronDown className="w-3 h-3" />
                      ) : (
                        <ChevronsUpDown className="w-3 h-3 opacity-20" />
                      )}
                    </button>
                  </th>
                )
              })}
            </tr>
          </thead>
          <tbody className="divide-y divide-border/50">
            {rowsOrdenadas.map((row) => (
              <tr key={rowKey(row)} className={cn('hover:bg-muted/5 transition-colors group', rowClassName?.(row))}>
                {columns.map((c, i) => (
                  <td
                    key={i}
                    className={cn(
                      'px-4 py-3 text-foreground/80',
                      c.isTechnical && 'font-mono text-xs text-foreground/60 tracking-tight',
                      alignCls[c.align ?? 'left'],
                      c.hideBelow && hideCls[c.hideBelow],
                      c.tdClassName,
                    )}
                  >
                    {c.cell(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
