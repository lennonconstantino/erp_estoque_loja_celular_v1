# Checklist de Segurança (pré-deploy)

Gate de revisão a rodar **antes de todo deploy**, adaptado à stack real deste
projeto: **backend Go (chi) + pgx + PostgreSQL puro + JWT/RBAC próprio** e
**frontend React/Vite**. Não há Next.js, Supabase, Prisma nem chamadas de IA —
itens típicos desses contextos aparecem marcados como **N/A** com o equivalente
adotado aqui.

Complementa a referência de autenticação/autorização em [security.md](security.md).

Legenda: ✅ já coberto no repositório · ⚠️ pendente / decidir · ⛔ N/A nesta stack.

---

## Segredos expostos

Chaves, senhas e tokens que nunca deveriam ser visíveis ao usuário final.

### 1. Chaves de API expostas no frontend ⚠️

No Vite, **qualquer variável com prefixo `VITE_` é embutida no bundle** e fica
visível no browser (DevTools → Sources). (O equivalente Next.js é `NEXT_PUBLIC_`,
que **não se aplica** aqui.)

- **Errado:** `VITE_JWT_SECRET=...`, `VITE_DB_PASSWORD=...`, qualquer segredo com `VITE_`.
- **Certo:** segredos só no backend (sem prefixo `VITE_`). O frontend só conhece a
  URL pública da API e o access token do próprio usuário (curta duração).

### 2. `.gitignore` e o arquivo `.env` ✅

Sem `.gitignore`, um `git add .` manda o `.env` com senhas para o repositório;
bots varrem repos públicos em minutos.

- ✅ `.env` está no [.gitignore](../../.gitignore).
- ✅ Existe [backend/.env.example](../../backend/.env.example) com a estrutura (sem valores reais de produção).
- **Regra:** criar/atualizar o `.env.example` ao adicionar qualquer nova config (struct `Config`).

### 3. Secrets em `docker-compose` e configs ✅

- ✅ O [docker-compose.yml](../../docker-compose.yml) usa interpolação obrigatória
  (`${DB_PASSWORD?...}` / `${JWT_SECRET?...}`), então o arquivo **falha cedo** se os
  segredos não estiverem definidos no ambiente.
- ✅ O caminho recomendado para desenvolvimento local é `make up`, que já roda o Compose
  com `--env-file backend/.env`.
- **Se rodar Compose direto:** usar
  `docker compose --env-file backend/.env -f docker-compose.yml up -d --build`.
- **Atenção:** se a senha tiver caracteres especiais (`@`, `!`, `:`, `/`), eles precisam
  ser **URL-encoded** dentro de `DATABASE_URL`.
- **Errado:** `JWT_SECRET: super_secret_123` (texto puro fixo no YAML, versionado).

### 4. GitGuardian e scanners de segredos ✅

Ferramentas como **GitGuardian** analisam o conteúdo versionado e o diff do commit, não
o contexto da equipe. Se um texto parece credencial, ele pode ser sinalizado mesmo que
você saiba que é "só para dev".

- **Não commitar valores literais** que pareçam segredo, mesmo em exemplos. Padrões que
  costumam disparar alertas:
  - DSN com senha embutida: `postgres://usuario:senha@host:5432/db`
  - `JWT_SECRET=algum-valor`
  - `DB_PASSWORD=minha-senha`
- **Usar placeholders vazios ou genéricos** em arquivos versionados:
  - `.env.example`: manter `DB_PASSWORD=` e `JWT_SECRET=` sem valor real
  - docs: preferir `<defina-aqui>`, `${DB_PASSWORD}` ou `postgres://usuario:<senha>@host/db`
- **Não usar defaults sensíveis em código/config**. Mesmo valores de desenvolvimento como
  `erp_secret` ou `troque-este-segredo-em-producao` podem ser classificados como segredo
  porque estão hardcoded.
- **Evitar credenciais em URLs de exemplo** em Markdown, comentários e scripts. Scanners
  também leem `docs/`, não apenas `.env` e código-fonte.
- **Se um segredo real já foi commitado:** considerar o vazamento como ocorrido. Rotacionar
  a credencial, revisar onde ela é usada e, se necessário, limpar o histórico do Git.
- **Antes de abrir PR:** revisar `docker-compose.yml`, `.env.example`, scripts e docs para
  garantir que só existam variáveis ou placeholders, nunca valores concretos.

### 4. Chamando APIs externas direto do browser ⚠️

Integrações com serviços externos (ex.: CEP/ViaCEP, e futuramente cupom/nota fiscal)
devem passar pelo backend, nunca expor credenciais no frontend.

- ✅ A consulta de CEP já é **proxy no backend** (`GET /api/v1/clientes/cep/{cep}` →
  gateway `cep/viacep.go`, com resiliência), não uma chamada direta do browser.
- **Regra:** toda nova integração com chave/segredo entra como **adaptador outbound**
  no backend; o frontend só fala com a nossa API.

---

## Autorização

O sistema precisa verificar **quem é** (autenticação) e **o que pode** (autorização)
em cada rota. Aqui isso é feito pelo middleware `platform/auth` — **não há Supabase/RLS**.

### 1. Controle de acesso a dados (equivalente ao RLS) ⛔→✅

Supabase RLS **não se aplica** (banco é PostgreSQL puro acessado só pelo backend via
pool `pgx`; não há `anon key` nem acesso direto do cliente ao banco).

- ✅ O acesso é mediado 100% pelo backend; nenhuma credencial de banco chega ao cliente.
- ✅ Autorização por rota via `auth.RequirePerm("recurso:acao")` (ver [security.md](security.md)).
- **A fazer por módulo:** quando houver dados "por usuário/loja", filtrar no `WHERE` do
  repositório (escopo do dono), análogo ao que o RLS faria.

### 2. IDs sequenciais e previsíveis (IDOR) ✅

IDs sequenciais (1, 2, 3) em URLs permitem adivinhação e acesso a dados de terceiros.

- ✅ PKs são **UUID** (`gen_random_uuid()` / `uuid` nas migrations), não `SERIAL`.
- **Atenção:** UUID sozinho não basta — a verificação de permissão (`RequirePerm`) e,
  quando aplicável, o filtro de escopo no backend continuam obrigatórios.

### 3. Lógica de autorização só no frontend ✅

Esconder um botão no React não protege a rota — qualquer um chama a URL via `curl`.

- ✅ A autorização é imposta no **backend** (middleware), não no frontend.
- **Regra:** Frontend = UX (esconde o que não pode); Backend = **fonte da verdade**
  (`RequirePerm` → 403). Nunca confie só na UI.

### 4. Admin hardcoded no código ✅

Lista de e-mails/IDs de admin no código vaza quem atacar e é difícil de revogar.

- ✅ Papéis/permissões vivem no **banco** (`iam.papeis`, `iam.permissoes`,
  `iam.papel_permissoes`) e são embutidos no JWT (claim `perms`) no login.
- **Errado:** `const ADMIN_EMAILS = [...]`. **Certo:** papel `ADMIN` no banco (seed).

### 5. Rotas de API sem autenticação ✅ / ⚠️

Autenticação (quem é você) vem antes de autorização (o que pode).

- ✅ As rotas de `clientes` estão todas atrás de `authMgr.Authenticate` + `RequirePerm`
  (router do módulo). `GET /health` é público por design.
- ⚠️ **Regra ao adicionar módulo:** toda rota nova de negócio nasce protegida; só exponha
  rota pública com decisão explícita.

---

## Input e comunicação inseguros

Não confiar no que vem de fora sem validar; comunicação sem proteção.

### 1. Validação de input / mass assignment ✅

Salvar tudo que vem no body permite ao atacante injetar campos (`role:'admin'`,
`ativo:true`, etc.). (Zod/Pydantic **não se aplicam** — é Go.)

- ✅ `httpserver.DecodeJSON` usa `DisallowUnknownFields()` (rejeita campos extras) e
  `MaxBytesReader` de 1 MiB ([httpserver.go](../../backend/internal/platform/httpserver/httpserver.go)).
- ✅ Handlers usam **DTOs próprios** (`clienteRequest`/`clienteResponse`), nunca a
  entidade de domínio direto; invariantes validadas em `domain/`.

### 2. SQL injection ✅

Concatenar input em SQL (`... WHERE email = '${email}'`) permite injeção.

- ✅ Repositórios usam **queries parametrizadas** do `pgx` (`$1`, `$2`, ...); não há
  concatenação de strings em SQL. (ORM Prisma/Drizzle **não se aplica**.)
- **Regra:** todo novo repositório segue o mesmo padrão — placeholders, nunca interpolação.

### 3. CORS aberto ⚠️

`Access-Control-Allow-Origin: *` permite que qualquer site chame a API em nome do
usuário logado.

- ⚠️ Não há middleware de CORS configurado no backend hoje. Decidir:
  - Se o frontend é servido **same-origin** (nginx faz proxy `/api` → backend): CORS
    pode ser desnecessário.
  - Se for **cross-origin**: configurar allowlist explícita de origens, nunca `*`.

### 4. Sem rate limiting ⚠️

Sem limite: brute force em login, esgotamento de recursos, criação massiva de contas.

- ⚠️ Não há rate limiting hoje. **Prioridade quando o `/auth/login` do `iam` for
  implementado** (ex.: 5 tentativas/min por IP); considerar limite global por IP no chi.

### 5. Upload de arquivos sem restrição ⛔

Não há endpoints de upload no escopo atual.

- **Se/quando houver:** validar tipo **real** (magic bytes, não a extensão), limitar
  tamanho, gerar nome aleatório e salvar fora do diretório do código.

### 6. Tokens / JWTs mal implementados ✅ / ⚠️

O payload do JWT é Base64 **legível**, não criptografado — a assinatura é o que garante
integridade.

- ✅ Tokens são emitidos e **verificados no backend** (HS256, `auth.Manager`), com `exp`
  (access TTL curto, default 15min).
- ⚠️ No frontend, evitar `localStorage` para o token (exposto a XSS); preferir cookie
  `httpOnly` quando o fluxo de refresh do `iam` estiver implementado. Ver [security.md](security.md).

### 7. Erros detalhados em produção ✅

Stack traces, paths e SQL em respostas de erro ensinam a estrutura interna ao atacante.

- ✅ O envelope de erro é `{"error":{"code","message"}}` com mensagens genéricas; erros
  internos viram `500 INTERNAL` "erro interno" (`writeDomainError`), sem vazar detalhes.
- **Produção:** manter logs completos só internamente (stdout/observabilidade), nunca no body.

### 8. Webhooks sem verificação de assinatura ⛔

Não há webhooks de entrada no escopo atual. Se/quando houver, validar a assinatura do
serviço (HMAC) antes de processar.

---

## Checklist de deploy

Marque antes de cada deploy:

- [ ] Nenhum segredo tem prefixo `VITE_` (nada sensível embarcado no bundle do frontend)
- [ ] `.env` está no `.gitignore` ✅
- [ ] Existe `.env.example` com a estrutura, sem valores reais ✅
- [ ] `docker-compose.yml` **não** tem senhas hardcoded (usa `${VAR}` interpolado) ✅
- [ ] `.env.example`, scripts e docs usam placeholders, não valores que pareçam segredo
- [ ] `JWT_SECRET` e `DB_PASSWORD` reais existem só em `backend/.env` local ou no secret manager do ambiente
- [ ] Seed `admin@loja.local` / `admin123` desativado ou com senha trocada em produção
- [ ] Toda rota de negócio verifica autenticação (`Authenticate`) ✅
- [ ] Toda rota de negócio verifica autorização (`RequirePerm("recurso:acao")`) ✅
- [ ] IDs públicos são UUID, não `SERIAL` sequencial ✅
- [ ] Papéis/permissões no banco (`iam.*`), não hardcoded no código ✅
- [ ] Input validado: `DecodeJSON` (DisallowUnknownFields) + invariantes de `domain/` ✅
- [ ] Queries usam placeholders `pgx` (`$1`), sem concatenação de SQL ✅
- [ ] CORS decidido: same-origin (nginx) **ou** allowlist explícita, nunca `*`
- [ ] Rate limiting nos endpoints críticos (login, quando o `iam` existir)
- [ ] JWT verificado no backend, com `exp`; token não fica em `localStorage` exposto a XSS
- [ ] Erros de produção não expõem stack trace / SQL / paths (envelope genérico) ✅
- [ ] HTTPS obrigatório no ambiente de produção (Railway/proxy)
