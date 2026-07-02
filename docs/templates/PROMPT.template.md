# PROMPT.md — <Nome do Projeto>

<!--
  Esqueleto preenchível de especificação executável. Copie para a raiz do novo
  projeto como PROMPT.md e preencha os <placeholders>. O método por trás deste
  formato está em docs/playbook-planejamento.md. Apague os comentários "guia"
  conforme preenche.
-->

> Documento de especificação executável. Leia §0 inteiro antes de escrever qualquer
> linha de código. Implemente uma fase de cada vez, na ordem indicada. Todo critério
> de aceitação deve dar `PASS` antes de avançar.

---

## §0 — Arquitetura (LEIA, NÃO IMPLEMENTE)

### Contexto de negócio

<!-- guia: 2–4 frases. Quem usa, que dor resolve, o que substitui. -->

<descreva o produto e os usuários>

**Critério de aceitação do cliente:** <uma frase testável — o "dia completo" do
sistema. Ex.: "operar um dia inteiro sem planilha, sem saldo negativo, sem erro fiscal".>

### Stack

| Camada | Tecnologia |
|--------|------------|
| Backend | <linguagem, framework, libs principais> |
| Banco | <SGBD + host local/produção> |
| Frontend | <framework + gerenciador de pacotes> |
| Deploy | <plataforma(s)> |

### Leis arquiteturais (L — nunca violar)

<!-- guia: invariantes que NUNCA quebram. Curtas, numeradas, verificáveis pelo juiz.
     Exemplos reais: "módulo não importa o interno de outro"; "dependências apontam
     pra dentro"; "sem FK entre schemas"; "toda rota protegida tem RBAC";
     "saldo nunca negativo"; "ledger append-only"; "código em português". -->

```
L1  <invariante>
L2  <invariante>
...
```

### Módulos e dependências entre contextos

<!-- guia: o grafo que ORDENA as fases. Quem é consumido vem antes de quem consome. -->

```
<moduloA>   (sem deps de outro módulo)
<moduloB>   → <PortaX> (implementada por <moduloA>)
...
```

### Fluxos-chave

<!-- guia: descreva os 1–3 fluxos críticos (ex.: autenticação, transação central):
     estado, invariantes e passos. É onde mora o risco. -->

<fluxo 1>

### Estado atual do repositório

**Pronto — não reimplementar:** <o que já existe: platform, migrations, módulo de referência>

**Falta implementar (nesta ordem):** <lista dos módulos/fases>

---

## §0.5 — Definition of Done por tarefa (D — obrigatório)

> Toda tarefa (uma fase ou um item do checklist) só conclui com D1, D2 e D3 em `PASS`.
> Detalhe vinculante e protocolo do juiz em [docs/mandates.md](../mandates.md).

```
D1  TESTES NO MESMO PASSO — testes junto da lógica de domínio/caso de uso.
    Metas: <domain/ ≥ X%, application/ ≥ Y%>. Frontend: <verificação, ex. tsc + lint>.
D2  CHECKLIST VIVO        — derive o checklist antes de codar; marque [x] em docs/todos.md.
D3  AGENTE JUIZ           — subagent INDEPENDENTE julga o diff vs spec + Leis + Regras
    → CONFORME | NÃO CONFORME. NÃO CONFORME bloqueia o avanço de fase.

Ordem:  implementar → D1 → D2 → critério PASS da fase → D3 (CONFORME)
```

---

## §1 — Regras de execução (R — siga sempre)

<!-- guia: COMO escrever o código (procedimentos, não invariantes). Exemplos:
     "leia antes de editar"; "compile após cada arquivo"; "sem stub/TODO/panic";
     "DTO no handler, nunca a entidade"; "replique o módulo de referência". -->

```
R1  <regra>
R2  <regra>
...
```

---

## §2 — Estrutura de diretórios esperada

<!-- guia: a árvore-alvo do repo + a "anatomia obrigatória" de um módulo, para
     replicar em cada fase. -->

```
<árvore do repositório>
```

Anatomia obrigatória de cada módulo:

```
<estrutura de camadas de um módulo>
```

---

## §3+ — Fases

<!-- guia: uma seção por fase, NA ORDEM do grafo de dependências. Repita este bloco.
     Cada fase termina com um critério de aceitação que é um comando que imprime PASS. -->

### Fase N: <nome do módulo/capacidade>

**Objetivo:** <uma frase>

**A implementar:**

- Domínio: <entidades, invariantes, erros sentinela>
- Portas: <inbound (o que oferece) / outbound (o que exige)>
- Aplicação: <casos de uso>
- Adapters: <http, persistência, gateways externos>
- Fronteira/DI: <como monta e onde registra (composition root)>
- Frontend (se houver): <telas>

**Critério de aceitação FN (executável):**

```bash
<comando que termina em: && echo "PASS FN.<algo>">
```

<!-- Repita "### Fase N+1: ..." até cobrir todo o grafo de dependências. -->

---

## §Final — Critério final de aceitação do cliente

<!-- guia: o roteiro end-to-end que exercita o sistema inteiro uma vez, sem atalhos.
     É o "dia completo" do contexto de negócio, com papéis/perfis reais. -->

```
1. <passo>
2. <passo>
...
PASS FINAL: <condição observável — ex.: nenhum saldo negativo, nenhum erro fiscal, histórico completo>
```
