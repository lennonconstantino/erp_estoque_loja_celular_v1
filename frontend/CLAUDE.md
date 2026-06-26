# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

SPA do ERP: React 18 + Vite + TypeScript + Tailwind + shadcn/ui. Textos de UI em **português**. Visão geral do repositório em [../CLAUDE.md](../CLAUDE.md).

## Commands

Gerenciado com **pnpm** (não use npm/yarn).

```bash
pnpm install
pnpm dev      # Vite dev server
pnpm build    # tsc --noEmit && vite build   (type-check é parte do build)
pnpm lint     # eslint .
pnpm preview  # serve o build
```

Não há suíte de testes configurada. `pnpm build` é o portão de qualidade: ele roda `tsc --noEmit`, então erros de tipo quebram o build.

## Convenções

- Alias `@` → `src/` (definido em `vite.config.ts` e `tsconfig.json`). Importe sempre via `@/...`, não com caminhos relativos profundos.
- **shadcn/ui** configurado em `components.json`: componentes vão em `@/components/ui`, estilo `default`, base `slate`, ícones `lucide`. Adicione componentes com o CLI do shadcn, não à mão.
- **Kit de UI compartilhado** em `@/components/ui` é a fonte única do padrão visual: `PageShell` (casca), `DataTable<T>` (tabela com ordenação por `sortAccessor`), `Button`/`buttonClasses` (variantes; sem azul/índigo ad-hoc), `StatusBadge`, `Modal`, `Field`/`inputClasses`. Páginas novas **compõem** esses primitivos em vez de remontar cabeçalho/tabela/modal na mão — foi a ausência dessa camada que fez as telas divergirem.
- `cn()` em `@/lib/utils` (clsx + tailwind-merge) é o helper padrão para compor classes.
- Variáveis de ambiente passam por `@/lib/env` via `required(...)` — adicione novas lá (prefixo `VITE_`) em vez de ler `import.meta.env` direto. Obrigatória: `VITE_API_BASE_URL`.

## Camada de rede e autenticação

Three camadas, use sempre a de cima:

1. `@/lib/http` — `request<T>()`: wrapper de `fetch` tipado. Infra de baixo nível, raramente chamado direto.
2. `@/lib/api` — objeto `api.get/post/put/patch/delete`. **É esta a API a usar nas páginas.** Cuida de: base URL, header `Authorization: Bearer`, timeout (15s), envelope de erro do backend (`{error:{code,message}}` → `ApiError`), e **refresh automático de token** em 401 (re-tenta a chamada uma vez; refresh concorrente é deduplicado).
3. `@/lib/auth` — armazenamento de tokens. **Access token vive só em memória**; apenas o refresh token persiste em `localStorage` (`erp_refresh_token`). Logo, após reload não há access token até o primeiro refresh.

Implicação para rotas protegidas: `PrivateRoute` (em `App.tsx`) considera autenticado quem tem access token **ou** refresh token — porque após reload só o refresh existe e o access é obtido na primeira chamada via 401→refresh.

O backend expõe tudo sob `/api/v1`; endpoints retornam erros no formato `{"error":{"code","message"}}` e listas como `{"items":[...]}`. **Inclua sempre o prefixo `/api/v1` no path** passado ao `api` — ele só concatena `VITE_API_BASE_URL + path`. Omitir o prefixo gera 404, e como a resposta não traz o envelope `{error:...}`, a UI exibe o genérico **"Erro desconhecido"**. O login usa o campo `senha` (não `password`).

## Deploy

Build estático servido por nginx (`Dockerfile` + `nginx.conf`): SPA fallback para `/index.html` e cache longo de assets. O `VITE_API_BASE_URL` é injetado em build time.
