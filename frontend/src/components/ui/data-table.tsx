import { useMemo, useState, type ReactNode } from 'react'
import { ChevronDown, ChevronsUpDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'

export type SortValue = string | number | null | undefined

export interface Column<T> {
  header: ReactNode
  cell: (row: T) => ReactNode
  align?: 'left' | 'right' | 'center'
  /** Esconde a coluna abaixo do breakpoint informado. */
  hideBelow?: 'sm' | 'md'
  thClassName?: string
  tdClassName?: string
  /** Torna a coluna ordenável: retorna o valor usado na comparação. */
  sortAccessor?: (row: T) => SortValue
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

/** Tabela padrão (cartão branco, cabeçalho cinza, linhas com hover, ordenação opcional). */
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
      return null // terceiro clique remove a ordenação
    })
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden overflow-x-auto">
      {loading ? (
        <p className="p-6 text-sm text-gray-500 text-center">Carregando…</p>
      ) : rowsOrdenadas.length === 0 ? (
        <p className="p-6 text-sm text-gray-500 text-center">{empty}</p>
      ) : (
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              {columns.map((c, i) => {
                const ordenavel = !!c.sortAccessor
                const ativa = sort?.index === i
                const baseTh = cn(
                  'px-4 py-3 font-medium text-gray-600',
                  alignCls[c.align ?? 'left'],
                  c.hideBelow && hideCls[c.hideBelow],
                  c.thClassName,
                )
                if (!ordenavel) {
                  return <th key={i} className={baseTh}>{c.header}</th>
                }
                return (
                  <th key={i} className={baseTh} aria-sort={ativa ? (sort!.dir === 'asc' ? 'ascending' : 'descending') : 'none'}>
                    <button
                      type="button"
                      onClick={() => alternarOrdenacao(i)}
                      className={cn(
                        'inline-flex items-center gap-1 font-medium hover:text-gray-900',
                        c.align === 'right' && 'flex-row-reverse',
                        c.align === 'center' && 'mx-auto',
                        ativa ? 'text-gray-900' : 'text-gray-600',
                      )}
                    >
                      {c.header}
                      {ativa ? (
                        sort!.dir === 'asc'
                          ? <ChevronUp className="w-3.5 h-3.5" />
                          : <ChevronDown className="w-3.5 h-3.5" />
                      ) : (
                        <ChevronsUpDown className="w-3.5 h-3.5 text-gray-300" />
                      )}
                    </button>
                  </th>
                )
              })}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {rowsOrdenadas.map((row) => (
              <tr key={rowKey(row)} className={cn('hover:bg-gray-50', rowClassName?.(row))}>
                {columns.map((c, i) => (
                  <td
                    key={i}
                    className={cn(
                      'px-4 py-3 text-gray-700',
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
