# 09 — Roadmap para Microsserviços

A arquitetura nasce como **monólito modular**: um único processo/binário, mas com
módulos já isolados (1 pacote + 1 schema cada) e comunicação por contratos. Isso
dá a produtividade do monólito agora e o caminho de extração depois — sem
"big bang".

## Por que já está pronto para extrair

| Decisão tomada agora | Benefício na extração |
|----------------------|-----------------------|
| 1 schema por domínio | vira 1 banco por serviço (copiar migrations do módulo) |
| Sem FK entre schemas | nenhuma constraint física cruza o limite do serviço |
| Comunicação por **ports** | trocar chamada in-process por HTTP/gRPC sem tocar no domínio |
| Eventos de domínio (`*Confirmada`) | barramento substitui a chamada direta |
| Hexagonal | só os adaptadores mudam; o core é reaproveitado |

## Fases de evolução

### Fase 1 — Monólito modular (atual)
Um binário (`cmd/api`). Módulos conversam por interface in-process; eventos via
barramento em memória (`platform/events`).

### Fase 2 — Modular + assíncrono
Trocar o barramento em memória por um real (NATS/Kafka/RabbitMQ), ainda no mesmo
processo. Compras/Vendas passam a baixar estoque por **evento**, não por chamada
direta — desacoplamento total antes mesmo de separar processos.

### Fase 3 — Extração do primeiro serviço
Candidato natural: **estoque** (já é um ledger com fronteira clara) ou **iam**.
Passos:
1. Criar `cmd/estoque-svc/main.go` montando só o módulo `estoque`.
2. Mover o schema `estoque` para um banco próprio.
3. Substituir a porta `EstoqueWriter` (in-process) por um adaptador HTTP/gRPC ou
   por publish/subscribe de eventos.
4. Compras/Vendas passam a chamar o serviço/evento — domínio inalterado.

### Fase 4 — Malha de serviços
Cada bounded context vira um serviço com seu banco:

```
[iam-svc] [clientes-svc] [fornecedores-svc] [catalogo-svc]
[compras-svc] [vendas-svc] [estoque-svc] [relatorios-svc]
        │            │            │
        └──────── event bus (NATS/Kafka) ────────┘
                        │
                  API Gateway  ◄── front-end
```

## Sagas (consistência distribuída)

Quando os contextos estiverem separados, operações que cruzam fronteiras viram
**sagas** orquestradas por eventos. Exemplo — confirmar venda:

```
vendas: VendaConfirmada ─► estoque: tenta baixar saldo
   ├─ EstoqueBaixado    ─► vendas: marca CONFIRMADA + emite doc. fiscal
   └─ EstoqueInsuficiente ─► vendas: marca FALHA (compensa)
```

## O que NÃO fazer agora

- Não criar 7 repositórios/deploys hoje — custo operacional sem retorno.
- Não compartilhar tabelas entre módulos (mataria o isolamento).
- Não introduzir FK cross-schema "porque é prático" — quebra a Fase 3.

> Regra de ouro: **manter os limites limpos no monólito**. Se os imports e os
> schemas permanecerem isolados, a migração para microsserviços é mecânica.
