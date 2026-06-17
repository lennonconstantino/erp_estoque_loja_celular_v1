# Frontend — ERP Estoque

SPA do ERP de estoque em **React 18 + Vite + TypeScript**, estilizada com
**Tailwind CSS** e componentes **shadcn/ui**. Consome a API REST do backend
(`/api/v1`) com autenticação JWT (access em memória + refresh automático).

- **Build:** [Vite](https://vitejs.dev) 5
- **Roteamento:** [react-router-dom](https://reactrouter.com) 6
- **UI:** Tailwind + [shadcn/ui](https://ui.shadcn.com) (ícones [lucide](https://lucide.dev))
- **Gerenciador de pacotes:** **pnpm** (não use npm/yarn)

## Pré-requisitos

- Node 20+
- pnpm 9+

## Início rápido

```bash
pnpm install
echo "VITE_API_BASE_URL=http://localhost:8080" > .env   # aponta para a API
pnpm dev                                                 # http://localhost:5173
```

> `VITE_API_BASE_URL` é **obrigatória** — a app lança erro no boot se ausente
> (ver `src/lib/env.ts`). Em build de produção ela é injetada em build time.

## Comandos

| Comando | Descrição |
|---------|-----------|
| `pnpm dev` | Vite dev server (HMR) |
| `pnpm build` | `tsc --noEmit && vite build` — o type-check faz parte do build |
| `pnpm lint` | ESLint |
| `pnpm preview` | serve o build de produção localmente |

Não há suíte de testes. **`pnpm build` é o portão de qualidade**: como roda
`tsc --noEmit`, qualquer erro de tipo quebra o build.

## Estrutura

```
src/
├── main.tsx          # entrypoint (monta <App/>)
├── App.tsx           # rotas + PrivateRoute (guarda de sessão)
├── index.css         # Tailwind + tokens de tema (CSS variables)
├── lib/
│   ├── http.ts       # request<T>(): wrapper tipado de fetch (baixo nível)
│   ├── api.ts        # api.get/post/...: camada usada pelas páginas
│   ├── auth.ts       # armazenamento de tokens
│   ├── env.ts        # variáveis de ambiente validadas
│   └── utils.ts      # cn() (clsx + tailwind-merge)
└── pages/
    ├── LoginPage.tsx
    └── DashboardPage.tsx
```

## Convenções

- **Alias `@` → `src/`** (configurado em `vite.config.ts` e `tsconfig.json`).
  Importe via `@/...`, não com caminhos relativos profundos.
- **shadcn/ui** configurado em `components.json` (estilo `default`, base `slate`,
  ícones `lucide`). Componentes vão para `@/components/ui` — adicione-os pelo CLI
  do shadcn, não à mão.
- Componha classes Tailwind com `cn()` de `@/lib/utils`.
- Novas variáveis de ambiente entram em `@/lib/env` via `required(...)` (prefixo
  `VITE_`); não leia `import.meta.env` direto nas páginas.

## Rede e autenticação

Três camadas — nas páginas, use sempre o objeto `api`:

1. **`lib/http`** — `request<T>()`: wrapper de `fetch` tipado, baixo nível.
2. **`lib/api`** — `api.get/post/put/patch/delete`. Cuida de: base URL, header
   `Authorization: Bearer`, timeout (15s), tradução do envelope de erro do
   backend (`{error:{code,message}}` → `ApiError`) e **refresh automático de
   token** em respostas 401 (re-tenta a chamada uma vez; refresh concorrente é
   deduplicado).
3. **`lib/auth`** — tokens. O **access token vive apenas em memória**; só o
   refresh token persiste em `localStorage` (`erp_refresh_token`).

Por isso `PrivateRoute` (em `App.tsx`) considera autenticado quem tiver access
token **ou** refresh token: após um reload, só o refresh sobrevive e o access é
obtido na primeira chamada via fluxo 401 → refresh.

## Build e deploy (Docker)

`Dockerfile` faz o build estático e o serve com **nginx** (`nginx.conf`): SPA
fallback para `/index.html` e cache longo de assets. No `docker compose` da raiz,
o frontend sobe na porta `:80`, atrás da API.

Mais convenções para o Claude em [`CLAUDE.md`](CLAUDE.md).
