# Mandates — Regras Inegociáveis de Execução

> **Fonte canônica do mandate de processo.** Este documento define o que toda
> tarefa proveniente de um plano (uma fase do [PROMPT.md](../PROMPT.md) ou um item
> de [todos.md](todos.md)) deve cumprir, independentemente da fase. O `curl`/PASS
> de aceitação de cada fase é o **piso**, não o teto.

## Hierarquia de documentos

| Documento | Papel |
|-----------|-------|
| [PROMPT.md](../PROMPT.md) | Spec **executável**: fases, arquitetura, Leis (L1–L8), Regras (R1–R12), critérios de aceitação |
| **mandates.md** (este) | **Mandate de processo**: Definition of Done, protocolo do juiz, templates, política de cobertura |
| [CLAUDE.md](../CLAUDE.md) | Guia operacional para o agente (comandos, anatomia de módulo) |
| [todos.md](todos.md) | Checklist vivo da implementação, por fase |

> As **Leis L1–L8** (invariantes de arquitetura) e as **Regras R1–R12** (regras de
> execução) permanecem no PROMPT.md — esta página as referencia, não as duplica.
> Se uma regra desta página conflitar com o PROMPT.md, o PROMPT.md vence e esta
> página deve ser corrigida.

---

## Definition of Done por tarefa (D1–D3)

Uma tarefa só está concluída quando **D1, D2 e D3** dão `PASS`. Ordem:

```
implementar → D1 (testes) → D2 (checklist) → critério curl/PASS da fase → D3 (juiz CONFORME)
```

### D1 — Testes unitários no mesmo passo

Os testes acompanham a implementação **na mesma tarefa/commit** — nunca são
adiados para a Fase 8.

- **Escopo mínimo:** cada invariante de domínio e cada erro sentinela tem teste
  (caminho feliz + cada erro).
- **Domínio puro:** roda em memória, sem banco nem HTTP.
- **Casos de uso:** portas externas são substituídas por fakes/stubs em memória
  (não mock de rede).
- **Cobertura:** `domain/` ≥ **80%**, `application/` ≥ **70%** (ver
  [política de cobertura](#política-de-cobertura)).
- **Verificação:**
  ```bash
  cd backend && go test -cover ./internal/modules/<dominio>/... \
    && echo "PASS D1.<dominio>"
  ```
- **Frontend:** não tem testes (ver [frontend-setup.md](setup/frontend-setup.md)).
  A verificação é `pnpm tsc --noEmit` + `pnpm lint`.

### D2 — Checklist vivo da tarefa

- Antes de codar, derive o checklist da tarefa a partir da fase no PROMPT.md e do
  bloco correspondente em [todos.md](todos.md) (template em
  [Template de checklist](#template-de-checklist-de-tarefa)).
- Marque `[x]` **somente** quando o item estiver feito **e** verificado.
- Mantenha [todos.md](todos.md) atualizado: os itens da fase concluída ficam `[x]`.
- Nada é "feito" sem um item correspondente marcado.

### D3 — Agente juiz (validação independente)

Ao terminar a tarefa, **antes** de declarar a fase como PASS, um subagent **juiz
independente** (não o que implementou) revisa o trabalho contra a spec.

- **Entrada:** o diff da tarefa + a seção da fase no PROMPT.md + Leis L1–L8 +
  Regras R1–R12.
- **O juiz não corrige código** — apenas julga e justifica.
- **Veredito obrigatório:**
  - `CONFORME` → todos os critérios atendidos; pode declarar PASS e avançar.
  - `NÃO CONFORME` → lista cada desvio (lei/regra/critério violado + `arquivo:linha`).
- **NÃO CONFORME bloqueia o avanço de fase.** Corrija e reenvie ao juiz até obter
  `CONFORME` — equivale a um critério de aceitação que falhou.
- O juiz sempre confere: D1 presente (testes + cobertura), D2 completo (checklist),
  Leis L1–L8 respeitadas, e o critério `curl`/PASS da fase reproduzido.

---

## Protocolo do agente juiz

Abra um subagent dedicado com o papel de **revisor de conformidade**. Ele responde
apenas com o veredito e a lista de desvios — não edita arquivos.

### Prompt template do juiz

```text
Você é o AGENTE JUIZ de conformidade do ERP Estoque. Não escreva nem corrija
código — apenas julgue.

CONTEXTO:
- Spec da fase: PROMPT.md §<N> (<título da fase>).
- Leis inegociáveis: PROMPT.md §0 (L1–L8).
- Regras de execução: PROMPT.md §1 (R1–R12).
- Definition of Done: docs/mandates.md (D1–D3).

ENTRADA A JULGAR:
<diff da tarefa OU lista de arquivos alterados>

AVALIE, item a item:
1. D1 — Há testes unitários no mesmo passo? Cobertura domain/ ≥ 80% e
   application/ ≥ 70%? (Cite a saída de `go test -cover`.)
2. D2 — O checklist da fase em docs/todos.md está atualizado e coerente com o diff?
3. Leis L1–L8 — Alguma foi violada? (ex.: import cross-module, FK entre schemas,
   rota protegida sem RBAC, saldo podendo ficar negativo, append-only burlado.)
4. Regras R1–R12 — Alguma foi violada? (ex.: stub/TODO/panic, entidade de domínio
   serializada direto, env lido fora de lib/env.ts.)
5. Critério de aceitação da fase — o `curl`/PASS descrito na spec foi reproduzido?

SAÍDA (exatamente neste formato):
VEREDITO: CONFORME | NÃO CONFORME
DESVIOS:
- <lei/regra/critério>: <descrição> — <arquivo:linha>
  (omita esta lista se CONFORME)
JUSTIFICATIVA: <2–4 linhas resumindo a decisão>
```

> **Como acionar:** instancie o juiz como subagent separado do implementador,
> passando o diff e a seção da spec. Trate `NÃO CONFORME` como bloqueio rígido de
> avanço de fase.

---

## Template de checklist de tarefa

Copie para o bloco da fase em [todos.md](todos.md) (ou para a descrição da tarefa)
e marque conforme avança:

```markdown
### Tarefa: <módulo/fase> — <objetivo>

Implementação
- [ ] <camada/arquivo 1>
- [ ] <camada/arquivo 2>

D1 — Testes
- [ ] Teste de cada invariante de domínio (caminho feliz + erros)
- [ ] `go test -cover ./internal/modules/<dominio>/...` → domain/ ≥ 80%, application/ ≥ 70%

D2 — Checklist
- [ ] Itens da fase marcados [x] em docs/todos.md

Aceitação da fase
- [ ] Critério `curl`/PASS da fase reproduzido (cole a linha PASS)

D3 — Juiz
- [ ] Veredito CONFORME do agente juiz
```

---

## Política de cobertura

- **Metas:** `domain/` ≥ 80%, `application/` ≥ 70%. São pisos por pacote de módulo,
  medidos com `go test -cover`.
- **O que conta:** invariantes de negócio, erros sentinela e ramos dos casos de uso.
- **O que não exigir cobertura:** adaptadores triviais (mapeamento DTO↔SQL sem
  lógica), `main.go`/wiring, código gerado.
- **Frontend:** sem suíte de testes por decisão de projeto
  ([frontend-setup.md](setup/frontend-setup.md)). Qualidade do FE = TypeScript
  strict + `pnpm tsc --noEmit` + `pnpm lint`.
- **Não burle a meta** com testes vazios ou asserções triviais — o juiz (D3) deve
  rejeitar cobertura inflada artificialmente.
