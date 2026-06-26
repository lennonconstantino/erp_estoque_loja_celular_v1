# Frontend — notas para agentes

Esta é a SPA React do ERP de estoque para loja de acessórios de celular. Este arquivo descreve as convenções específicas do frontend.

## Stack

- **SPA React pura** (Vite + TypeScript strict). **Não é Next.js** — não sugerir Next, SSR, server components ou roteamento por arquivos.
- **Tailwind CSS** para estilização. Sem CSS modules, styled-components, Emotion ou arquivos `.module.css`. Tokens globais do tema ficam em `src/index.css`.
- **shadcn/ui** para primitivos de UI. Adicionar componentes com `pnpm dlx shadcn@latest add <nome>` — não reimplementar o que o shadcn já oferece.
- **React Router** para roteamento.
- **Autenticação JWT** gerenciada pelo backend Go. O frontend armazena o access token e o envia via `Authorization: Bearer`. Não há SDK do Supabase no frontend.

## Gerenciador de pacotes

**Somente `pnpm`.** Não usar `npm install` nem `yarn add`. O lockfile é `pnpm-lock.yaml`. Se `package-lock.json` ou `yarn.lock` aparecerem, é um bug — delete-os.

**Idade mínima de release: 7 dias.** Configurado via `.npmrc` (`minimum-release-age=10080` minutos). O pnpm recusará qualquer versão de pacote publicada há menos de 7 dias. Isso protege contra ataques de typosquat e versões comprometidas que circulam por poucas horas antes de serem removidas.

Se um pacote recém-publicado for genuinamente necessário (ex.: correção urgente de segurança em uma dependência já usada), sobrescreva por instalação e justifique no commit — não reduza o threshold global.

## Política de dependências

- **HTTP:** use a API nativa `fetch` através de um cliente em `src/lib/http.ts` e o singleton `api` em `src/lib/api.ts`. **Sem axios, ky, got, superagent, redaxios.**
- **Datas:** use `Date` nativo e `Intl.DateTimeFormat`. Sem moment, dayjs, date-fns, a menos que seja genuinamente necessário.
- **Utilitários:** use métodos nativos de `Array` / `Object` / `Map`. Sem lodash, ramda.
- **Estado:** `useState` / `useReducer` / `useContext` primeiro. Só recorra a bibliotecas externas de estado quando a dor for real.
- **Formulários:** `<form>` nativo + `FormData` primeiro.
- **Validação:** só adicionar biblioteca de schema quando for necessária validação em runtime nas fronteiras do sistema.
- **Componentes de UI:** primitivos shadcn via `pnpm dlx shadcn@latest add <nome>`. Não reimplementar o que o shadcn já oferece.

Antes de adicionar um pacote, verifique:

1. Existe uma API nativa do browser ou do TS/JS que resolve isso?
2. O shadcn/ui já cobre?
3. É pequeno, bem mantido e vale o custo de manutenção?

Se sim à (3), adicione — mas documente a decisão no commit.

## Layout

```
frontend/
├── src/
│   ├── components/
│   │   └── ui/            # Kit de UI compartilhado (ver seção abaixo) + primitivos shadcn
│   ├── lib/               # Helpers sem dependência de framework (http, api, auth, env, utils)
│   ├── pages/             # Componentes de nível de rota (uma página por tela do dashboard)
│   ├── App.tsx            # Router + PrivateRoute (guarda de sessão)
│   ├── main.tsx
│   └── index.css          # Diretivas Tailwind + tokens globais do tema
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

Manter imports consistentes com o alias `@/*` (ex.: `@/lib/api`, `@/components/ui/button`).

## Kit de UI compartilhado (`@/components/ui`)

Para manter **todas as telas com o mesmo padrão visual** (a divergência anterior
nasceu de cada página reimplementar a própria casca), os primitivos abaixo são a
fonte única de verdade. Páginas novas **devem** compô-los em vez de remontar
cabeçalho/tabela/modal na mão.

| Componente | Papel |
|------------|-------|
| `page-shell.tsx` → `PageShell` | Casca da página: barra de cabeçalho branca (botão voltar + título/subtítulo + `actions`) e `<main>` centralizado (`maxWidth` configurável). |
| `button.tsx` → `Button` / `buttonClasses` | Botão padrão com variantes `primary` (cinza-900), `secondary`, `danger`, `success`. **Nunca** usar azul/índigo ad-hoc. `buttonClasses(...)` aplica o estilo a um `<Link>`. |
| `data-table.tsx` → `DataTable<T>` | Tabela padrão (cartão branco, cabeçalho cinza, hover, estados de loading/vazio) com **ordenação por coluna** embutida. |
| `badge.tsx` → `StatusBadge` | Selo de status com tons `success` / `neutral` / `warning` / `danger`. |
| `modal.tsx` → `Modal` | Janela modal padrão (overlay + cabeçalho + botão fechar). |
| `field.tsx` → `Field` / `inputClasses` | Rótulo de formulário e classe padrão de `input`/`select`/`textarea` (foco em cinza-900). |

### Ordenação de tabelas

`DataTable` ordena no cliente as linhas já carregadas. Uma coluna vira ordenável
ao declarar `sortAccessor: (row) => valor` (string ordena com `localeCompare`
pt-BR; número, numericamente; datas, por timestamp). O cabeçalho cicla
**asc → desc → sem ordenação** ao clicar; colunas de ação (ícones) não recebem
`sortAccessor`.

> Em telas paginadas no servidor (clientes, produtos, categorias, estoque) a
> ordenação atua sobre a **página atual**. Ordenação global exigiria parâmetros
> `sort`/`order` nos endpoints — ainda não implementado.

## Estilo de código

- **TypeScript strict.** Sem `any` a não ser que não haja alternativa; prefira `unknown` com narrowing.
- **Funções e componentes pequenos e composáveis** em vez de abstrações prematuramente genéricas. Três linhas similares > uma abstração genérica prematura.
- **Um componente = um arquivo.** Componentes devem caber em uma tela.
- **Classes Tailwind inline.** Sem CSS modules, styled-components, Emotion ou `.module.css`. Tokens globais ficam em `src/index.css`.

## Configuração

- Todas as leituras de env passam pelo módulo `src/lib/env.ts`, que valida as variáveis obrigatórias na inicialização. Nunca leia `import.meta.env.X` diretamente em componentes.
- Variáveis de env são prefixadas com `VITE_` (convenção Vite). O que não tiver o prefixo não é exposto ao cliente.

Variável obrigatória:

| Variável | Descrição |
|----------|-----------|
| `VITE_API_BASE_URL` | URL base do backend Go (ex.: `http://localhost:8080`) |

## Integração com o backend

- Fala com o backend Go via JSON. A URL vem de `VITE_API_BASE_URL`.
- Sempre use `api.get/post/put/patch/delete` de `@/lib/api` — ele trata base URL, JSON, injeção do Bearer token JWT, timeouts e `ApiError`s tipados.
- **Todo path deve incluir o prefixo `/api/v1`** (ex.: `api.get('/api/v1/produtos')`). O `api` concatena `VITE_API_BASE_URL + path` sem inserir o prefixo. Omiti-lo gera 404 — cujo corpo sem envelope `{error:...}` aparece na UI como o genérico **"Erro desconhecido"**. (Foi exatamente o bug que escondia os dados em Categorias/Produtos/Vendas/Estoque/Relatórios.)
- O login usa o campo **`senha`** (não `password`): `POST /api/v1/auth/login` com `{ email, senha }`.
- Auth é JWT: o login chama `POST /api/v1/auth/login`, recebe `access_token` e `refresh_token`. O access token é armazenado em memória (não em `localStorage`) e injetado automaticamente pelo cliente `api`. Nunca passe tokens via props de componente.
- Renove o access token via `POST /api/v1/auth/refresh` antes do vencimento (TTL padrão: 15 min).

## Testes

**Sem testes de frontend.** Não criar arquivos `*.test.ts` / `*.test.tsx` nem instalar test runner. A verificação do frontend é feita manualmente no browser, mais `pnpm tsc --noEmit` e `pnpm lint`. Se você se pegar pensando em vitest, Playwright ou Cypress — pare. Correção de lógica compartilhada vem de código simples e bem tipado, não de suíte de testes.

## Anti-padrões (rejeitados)

- Ler `import.meta.env.X` diretamente fora de `lib/env.ts`.
- Importar biblioteca HTTP quando `fetch` resolve.
- Misturar bibliotecas de estado (Zustand + Jotai + Redux) no mesmo projeto.
- Anotações `any` para calar o type-checker.
- Arquivos CSS customizados / styled-components junto com Tailwind.
- Reimplementar manualmente um primitivo que o shadcn já oferece.
- Usar Next.js, SSR ou qualquer framework que exija um servidor Node na frente da SPA.
- Usar `@supabase/supabase-js` no frontend — o Supabase é usado apenas como host do PostgreSQL pelo backend.
