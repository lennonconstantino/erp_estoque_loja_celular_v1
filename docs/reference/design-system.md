# Design System — referência da linguagem visual

Referência do **sistema de design** da SPA: princípios, catálogo de componentes, tema/tokens,
paleta com restrições de contraste e baseline de acessibilidade. O foco operacional (stack,
comandos, deps, rede, build) fica em [frontend-setup.md](../setup/frontend-setup.md).

> **Fonte da verdade é o código.** Este documento aponta para os arquivos — não duplica
> valores. Tokens vivem em [`src/index.css`](../../frontend/src/index.css); componentes em
> [`@/components/ui`](../../frontend/src/components/ui). Ao mudar um, ajuste a descrição aqui,
> nunca copie valores que vão divergir.

## Princípios

1. **Composição sobre remontagem.** Telas novas **compõem** os primitivos do kit
   (`@/components/ui`) — não remontam cabeçalho/tabela/modal na mão. Foi a ausência dessa
   camada que fez as telas divergirem.
2. **Só tokens semânticos.** Estilize com `bg-card`, `text-foreground`, `border-border`,
   `text-muted-foreground`, `text-primary`, `text-destructive`… **nunca** cor crua
   (`text-red-600`, `bg-white`, azul/índigo ad-hoc). Cor crua não acompanha Dark/Light.
3. **Acessível por padrão.** A baseline (foco, labels, contraste, movimento) está embutida no
   kit; compor já entrega um app que passa limpo no axe. Ver [Acessibilidade](#acessibilidade-a11y).

## Catálogo de componentes (`@/components/ui`)

| Componente | Papel |
|------------|-------|
| `page-shell.tsx` → `PageShell` | Casca da página: **`Sidebar` fixa** + cabeçalho técnico (breadcrumb/voltar + `CommandPalette` + `ThemeToggle` + `actions`) e `<main id="conteudo-principal" tabIndex={-1}>` centralizado (`maxWidth` configurável). Inclui **skip link** "Pular para o conteúdo" como primeiro elemento focável. |
| `sidebar.tsx` → `Sidebar` | Navegação lateral fixa (grupos Principal/Gestão via `NavLink`) + logout. `<nav aria-label>`. |
| `button.tsx` → `Button` / `buttonClasses` | Botão pill com variantes `primary`, `secondary`, `danger`, `success`, `ghost` e tamanhos `sm`/`md`/`icon`. **Nunca** usar azul/índigo ad-hoc. `buttonClasses(...)` aplica o estilo a um `<Link>`. |
| `data-table.tsx` → `DataTable<T>` | Tabela padrão (cartão, cabeçalho minimalista, hover, estados de loading/vazio) com **ordenação por coluna** embutida (ver abaixo). |
| `tabs.tsx` → `Tabs<T>` | Abas técnicas estilo pill (controladas via `activeTab`/`onTabChange`). Em uso na tela de Relatórios. |
| `badge.tsx` → `StatusBadge` | Selo de status com tons `success` / `neutral` / `warning` / `danger`. |
| `modal.tsx` → `Modal` | Janela modal padrão (overlay + cabeçalho + botão fechar; `maxWidth` até `max-w-4xl`). Construído sobre **Radix Dialog**: `role="dialog"` + `aria-modal`, `aria-labelledby` no título, focus trap, foco inicial/restaurado e fechar por **Escape**/overlay. API `{title,onClose,children,maxWidth}` inalterada. |
| `field.tsx` → `Field` / `inputClasses` / `inputClassesCompact` / `compactLabelClass` | Rótulo de formulário + classe pill de `input`/`select`/`textarea`. Use `inputClasses()` no padrão e **`inputClassesCompact()` + `compactLabelClass`** nas grades densas de itens (linhas de Compras e do PDV). `Field` **associa rótulo↔controle** automaticamente (`useId` + `cloneElement` injetam `htmlFor`/`id` casados) — basta passar um único controle como filho. |
| `command-palette.tsx` / `command.tsx` / `dialog.tsx` | Paleta de comandos (⌘K) para navegação rápida, sobre os primitivos shadcn `Command`/`Dialog` (Radix Dialog + `cmdk`). |
| `theme-toggle.tsx` → `ThemeToggle` | Alternância Dark/Light (consome `useTheme` de `@/lib/theme`). |
| `sonner.tsx` → `Toaster` + `toast` | Notificações (sucesso/erro). `<Toaster>` é montado uma vez em `App.tsx`; páginas chamam `toast.success/error`. |

### Ordenação de tabelas

`DataTable` ordena no cliente as linhas já carregadas. Uma coluna vira ordenável
ao declarar `sortAccessor: (row) => valor` (string ordena com `localeCompare`
pt-BR; número, numericamente; datas, por timestamp). O cabeçalho cicla
**asc → desc → sem ordenação** ao clicar; colunas de ação (ícones) não recebem
`sortAccessor`.

> Em telas paginadas no servidor (clientes, produtos, categorias, estoque) a
> ordenação atua sobre a **página atual**. Ordenação global exigiria parâmetros
> `sort`/`order` nos endpoints — ainda não implementado.

## Tema, tokens e contraste (Dark/Light)

O tema é gerido por `@/lib/theme` (`ThemeProvider` + hook `useTheme`), montado no topo
da árvore em `App.tsx`. Ele alterna a classe `.dark` no `<html>` e persiste a escolha
em `localStorage` (chave `theme`), com fallback para `prefers-color-scheme` na primeira
visita. O `ThemeToggle` no cabeçalho do `PageShell` dispara `toggleTheme()`.

Os **tokens** (cores, raio) ficam em `src/index.css`: o bloco `:root` define o tema
claro e `.dark` o escuro, ambos como variáveis HSL consumidas pelo Tailwind
(`bg-background`, `text-foreground`, `bg-card`, `border-border`, `text-primary`,
`text-destructive`, etc.). **Regra prática:** estilize componentes só com esses tokens —
qualquer cor crua (`text-red-600`, `bg-white`, `bg-indigo-500`) não acompanha a troca de
tema e quebra a consistência. Para status/erros use os tokens `destructive`/`success` (ou
o `StatusBadge`), não `red-600`/`green-600` soltos.

> **Contraste calibrado (WCAG AA) — invariantes que não devem ser revertidas sem reauditar:**
> - `--muted-foreground` (claro) está em `215 18% 42%` — foi escurecido de `47%` porque o
>   valor antigo reprovava por 0,01 (4.49:1). **Não clarear de volta.**
> - Evite `text-muted-foreground` combinado com `opacity-{40..70}` em texto real (o modificador
>   derruba o contraste — use o token cheio).
> - **Verde como texto:** `text-green-700 dark:text-green-400` sobre fundo claro;
>   `text-green-800` sobre tinta verde (`bg-green-500/10`). `green-500/600` reprovam como texto.
> - Não existe token semântico de *success* — verde é cru por convenção (o próprio
>   `Button.success` usa `green-600`).

## Acessibilidade (a11y)

A SPA passa **limpa no axe-core** (regra `color-contrast` + demais) em todas as telas, nos
temas claro e escuro, incluindo modais abertos. A baseline está embutida no kit — componha os
primitivos e você herda o comportamento acessível:

- **Formulários:** `Field` associa rótulo↔controle automaticamente. Para um `<select>`/`<input>`
  **fora** de um `Field` (ex.: filtro solto), passe `aria-label`.
- **Modais:** use `Modal` (Radix Dialog) — focus trap, Escape e `aria-*` vêm de graça.
- **Botões só-ícone:** sempre `aria-label` (não confie só em `title`).
- **Erros de formulário:** envolva a mensagem com `role="alert"` para anúncio em leitores de tela
  (toasts do `sonner` já têm live region própria).
- **Tabelas:** `DataTable` já emite `aria-sort` nos cabeçalhos ordenáveis e rótulo `sr-only` na
  coluna de ações (header vazio).
- **Hierarquia de headings:** o `PageShell` emite o `<h1>` da página; seções internas começam em
  `<h2>`. Não use heading para dado puro (ex.: valor de KPI → `<div>`).
- **Movimento:** animações respeitam `prefers-reduced-motion` (bloco global em `src/index.css`).
- **Contraste:** ver as invariantes em [Tema, tokens e contraste](#tema-tokens-e-contraste-darklight).

**Como reauditar** (não há suíte de frontend; o axe roda contra o app de pé):

```bash
make up                      # sobe db + api + frontend (http://localhost)
# Playwright + @axe-core/playwright num dir scratch (fora do package.json):
#   login admin@loja.local / admin123 → varre as rotas nos 2 temas → filtra color-contrast.
```

> Após mexer no `frontend/`, rebuild o container antes de auditar (`docker compose build frontend`):
> o nginx serve o `dist` buildado, não o source — auditar sem rebuildar lê o bundle antigo.

## Pendências e próximos passos (UI)

Itens de polimento levantados na padronização visual/refino do kit, ainda em aberto. **Não
bloqueiam** o uso atual:

- **Toasts no tema certo:** `sonner.tsx` importa `useTheme` de `next-themes`, mas o
  provider é o `@/lib/theme` → toasts ficam em `system`. Trocar a importação.
- **`autocomplete`/`inputMode` nos formulários:** a associação rótulo↔controle já existe
  (via `Field`), mas falta declarar `autocomplete`/`inputMode` adequados nos campos
  (email, tel, postal-code, numeric) para melhorar preenchimento e teclado mobile.
- **Tom `warning` do `StatusBadge`:** `text-yellow-600` sobre `bg-yellow-500/10` é um risco
  de contraste latente (não apareceu no axe por falta de dados de demo com aviso) — quando
  houver badge de aviso real, subir para `yellow-700/800`.
- **Navegação semântica:** o botão "voltar" do `PageShell` usa `<button onClick={navigate}>`;
  trocar por `<Link>` para suportar Cmd/middle-click.
- **`color-scheme`/`theme-color`:** declarar `color-scheme` no `.dark` e
  `<meta name="theme-color">` no `index.html` (corrige scrollbars/controles nativos).

**Concluído** (auditoria de acessibilidade, axe-core limpo): associação rótulo↔controle no
`Field`; `Modal` sobre Radix Dialog (focus trap/Escape/aria); skip link + `<main>` focável e
`<nav aria-label>`; `role="alert"` nas mensagens de erro; `aria-label` nos botões só-ícone;
`prefers-reduced-motion`; contraste WCAG AA (token `--muted-foreground` recalibrado, verdes
700/800, fim do `opacity-*` em texto muted); `select` solto com `aria-label`; hierarquia de
headings (KPI deixou de ser heading, seções em `<h2>`); `sr-only` na coluna de ações da tabela.
