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

### Paleta de cores (tokens)

Espelho de [`src/index.css`](../../frontend/src/index.css) — valores em **HSL** (`H S% L%`,
o formato consumido pelo Tailwind via `hsl(var(--token))`). A fonte da verdade continua sendo
o `index.css`; ao alterar um token lá, atualize a linha correspondente aqui. É um sistema
**neutro/monocromático**: no tema claro `primary` é **preto puro**; no escuro vira **índigo**.

| Token (`--`) | Utilitário Tailwind | Claro (`:root`) | Escuro (`.dark`) | Papel |
|---|---|---|---|---|
| `background` | `bg-background` | `0 0% 100%` | `240 17% 4%` | Fundo geral da aplicação. |
| `foreground` | `text-foreground` | `225 12% 7%` | `180 6% 97%` | Texto principal sobre o fundo. |
| `card` / `card-foreground` | `bg-card` / `text-card-foreground` | `220 14% 98%` / `225 12% 7%` | `240 6% 6%` / `180 6% 97%` | Cartões, tabelas e blocos de conteúdo. |
| `popover` / `popover-foreground` | `bg-popover` / `text-popover-foreground` | `0 0% 100%` / `225 12% 7%` | `240 6% 6%` / `180 6% 97%` | Menus, paleta de comandos, dropdowns. |
| `primary` / `primary-foreground` | `bg-primary` / `text-primary-foreground` | `0 0% 0%` / `0 0% 100%` | `234 58% 60%` / `0 0% 100%` | CTAs, estado selecionado, `Button.primary`. |
| `secondary` / `secondary-foreground` | `bg-secondary` / `text-secondary-foreground` | `220 14% 96%` / `225 12% 7%` | `240 4% 12%` / `180 6% 97%` | Botões/fundos secundários. |
| `muted` / `muted-foreground` | `bg-muted` / `text-muted-foreground` | `220 14% 96%` / `215 18% 42%` | `240 4% 12%` / `225 5% 57%` | Fundo sutil; texto secundário/legendas (contraste AA — ver invariantes). |
| `accent` / `accent-foreground` | `bg-accent` / `text-accent-foreground` | `220 14% 96%` / `225 12% 7%` | `240 4% 12%` / `180 6% 97%` | Realce de hover/itens ativos. |
| `destructive` / `destructive-foreground` | `bg-destructive` / `text-destructive` | `0 84% 60%` / `0 0% 98%` | `0 63% 31%` / `180 6% 97%` | Erros, exclusão, ações perigosas. |
| `border` | `border-border` | `225 10% 93%` | `0 0% 100% / 0.05` | Bordas de cartões, tabelas e divisórias. |
| `input` | `border-input` | `225 10% 93%` | `0 0% 100% / 0.05` | Borda de campos de formulário. |
| `ring` | `ring-ring` | `0 0% 0%` | `234 58% 60%` | Anel de foco (teclado). |
| `radius` | `rounded-*` | `0.5rem` (8px) | — | Raio base; `Button` é pill (`rounded-full`). |

**Tipografia** (importada no topo do `index.css`): **Inter** (`font-sans`, corpo e títulos) e
**JetBrains Mono** (`.font-mono`, com `tabular-nums` — usada em números/valores para alinhamento).

**Verde (success):** não é token semântico — é cor crua Tailwind por convenção (`Button.success`
usa `green-600`; como texto siga as regras de contraste abaixo).

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
> o nginx serve o `dist` buildado, não o source — auditar sem rebuildar lê o bundle antigo. O
> **dev server (Vite)** aplica HMR na hora; o **container Docker** só reflete a mudança após o
> rebuild. Desde então, o `index.html` é servido com **`Cache-Control: no-cache`** (ver
> [`nginx.conf`](../../frontend/nginx.conf)) — os assets têm hash e cache longo, mas o HTML de
> entrada revalida sempre, então **um refresh normal já pega o build novo** (não precisa mais de
> hard refresh). Se validar no browser, mire a porta que você realmente usa (`:80` do Docker, não
> só o dev server) — foi assim que um bug de layout "não corrigido" acabou sendo só bundle em cache.

## Pendências e próximos passos (UI)

> **Todos os itens de polimento de a11y/UX foram resolvidos** (ver "Concluído" abaixo).
> Resta apenas uma pendência opcional **de backend**, fora do escopo deste doc:
> ordenação global no servidor (parâmetros `sort`/`order` nos endpoints paginados) —
> hoje a ordenação do `DataTable` é client-side sobre a página carregada. Registrada em
> [docs/todos.md](../todos.md) (Fase 10 → Ordenação de tabelas).

**Concluído — polimento de a11y/UX (2ª leva):**

- **Toasts no tema certo:** `sonner.tsx` agora importa `useTheme` de `@/lib/theme` (não mais
  `next-themes`) → o toast segue o tema real (light/dark) em vez de ficar em `system`.
- **`autocomplete`/`inputMode` nos formulários:** campos de e-mail, telefone, CEP e numéricos
  declaram `type`/`inputMode`/`autocomplete` adequados. Em **Clientes** (dados do próprio
  titular) usa-se os tokens `autocomplete` completos (`email`, `tel`, `postal-code`,
  `address-*`); em **Fornecedores** (dados de terceiros) só `type`/`inputMode` + `autocomplete="off"`,
  para não vazar autofill pessoal do operador num cadastro alheio.
- **Tom `warning` do `StatusBadge`:** subido para `text-yellow-800` sobre `bg-yellow-500/10`
  (mesma regra do `success`), eliminando o risco de contraste latente.
- **Navegação semântica:** o botão "voltar" do `PageShell` virou `<Link to={back}>` — suporta
  Cmd/middle-click e abertura em nova aba.
- **`color-scheme`/`theme-color`:** `color-scheme` declarado em `:root`/`.dark` (`index.css`)
  e `<meta name="color-scheme">` + `<meta name="theme-color">` (light/dark) no `index.html` —
  corrige scrollbars e controles nativos por tema.

**Concluído** (auditoria de acessibilidade, axe-core limpo): associação rótulo↔controle no
`Field`; `Modal` sobre Radix Dialog (focus trap/Escape/aria); skip link + `<main>` focável e
`<nav aria-label>`; `role="alert"` nas mensagens de erro; `aria-label` nos botões só-ícone;
`prefers-reduced-motion`; contraste WCAG AA (token `--muted-foreground` recalibrado, verdes
700/800, fim do `opacity-*` em texto muted); `select` solto com `aria-label`; hierarquia de
headings (KPI deixou de ser heading, seções em `<h2>`); `sr-only` na coluna de ações da tabela.
