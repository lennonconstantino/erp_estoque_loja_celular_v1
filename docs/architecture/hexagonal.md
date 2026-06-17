# 02 — Arquitetura

## Princípios

1. **Arquitetura Hexagonal (Ports & Adapters).** O núcleo de cada domínio
   (entidades + regras) não conhece HTTP, banco ou bibliotecas externas. Tudo
   entra e sai por **portas** (interfaces Go), implementadas por **adaptadores**.
2. **Domínios totalmente isolados.** Cada bounded context é um pacote Go
   autossuficiente, com seu próprio domínio, casos de uso, portas, adaptadores e
   schema de banco. Um domínio **nunca importa o pacote interno de outro** — a
   comunicação acontece por contratos públicos (interfaces/eventos).
3. **Preparado para microsserviços.** Como cada módulo já é isolado e fala com
   os demais apenas por contratos, extrair um módulo para um serviço próprio é
   mover um pacote + seu schema, sem reescrever regra de negócio.

## As três camadas (por domínio)

```
        Driving / Inbound                         Driven / Outbound
        (quem aciona o app)                       (o que o app aciona)

   HTTP handler ─┐                          ┌─ PostgreSQL repository
   CLI / cron  ──┼──►  PORTA  ──►  CORE  ──►  PORTA  ──►  CEP/Fiscal client
   gRPC (futuro)─┘    (entrada)   (domínio)  (saída)      Event publisher
                                  + casos
                                   de uso
```

- **Core (domínio):** entidades, value objects e regras puras. Sem dependências
  de infraestrutura. Ex.: `Produto` garante `custo < venda`.
- **Application (casos de uso):** orquestra o domínio e as portas de saída.
  Ex.: `ConfirmarVenda` valida saldo, grava venda e dispara baixa de estoque.
- **Ports (portas):** interfaces. *Inbound* = o que o app oferece (services);
  *Outbound* = o que o app precisa (repositórios, gateways).
- **Adapters (adaptadores):** implementações concretas das portas. *Inbound* =
  handlers HTTP/chi; *Outbound* = repositórios pgx, clientes HTTP de CEP/fiscal.

## Regra de dependência

As dependências apontam **sempre para dentro** (em direção ao domínio):

```
adapters  ──►  application  ──►  domain
   ▲                                 │
   └──────── implementa ports ◄──────┘   (inversão de dependência)
```

O domínio define a interface (porta); o adaptador a implementa. O domínio nunca
importa o adaptador — recebe-o por injeção de dependência no `main`.

## Comunicação entre domínios

- **Síncrona, hoje (monólito modular):** via interface pública do outro módulo,
  injetada no caso de uso. Ex.: `vendas` precisa baixar estoque ⇒ depende de uma
  porta `EstoqueWriter` implementada pelo módulo `estoque`.
- **Assíncrona, alvo (microsserviços):** via **eventos de domínio**
  (`CompraConfirmada`, `VendaConfirmada`, `EstoqueAjustado`). Um barramento
  (NATS/Kafka/RabbitMQ) substitui a chamada direta sem alterar o domínio.

> **Sem foreign keys entre schemas.** IDs de outros contextos (cliente,
> produto, fornecedor) são guardados como UUID "solto". A consistência é
> garantida pela aplicação (e, no futuro, por sagas). Isso é o que permite
> separar os bancos depois.

## Estoque: razão (ledger) + saldo materializado

O saldo de cada produto (`catalogo.produtos.estoque_a_pro`) é um **cache**. A
**fonte da verdade** é `estoque.movimentacoes` — um livro-razão *append-only*.
Compras, Vendas e Ajustes:

1. gravam seu documento no próprio schema;
2. emitem uma **movimentação** no contexto `estoque`;
3. o contexto `estoque` atualiza o saldo materializado do produto.

Isso dá auditabilidade total e é o padrão correto para quando estoque virar um
serviço independente (event sourcing simplificado).

## Stack técnica

| Preocupação | Escolha | Observação |
|-------------|---------|------------|
| Roteamento HTTP | `go-chi/chi` | leve, idiomático |
| Banco | PostgreSQL + `pgx/v5` | sem ORM pesado |
| Migrations | `golang-migrate` | arquivos `.sql` versionados |
| Auth | `golang-jwt/jwt` + `bcrypt` | JWT + RBAC |
| Config | `joho/godotenv` | `.env` |
| IDs | `google/uuid` | UUID v4 |

> A stack acima é a referência sugerida; o que é arquitetural é a separação em
> ports & adapters, não a biblioteca específica.
