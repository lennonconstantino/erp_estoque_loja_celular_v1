# Resilience Stack

Políticas de resiliência aplicadas nos adaptadores de saída (*outbound adapters*)
que se comunicam com dependências externas: `CepGateway`, `FiscalGateway` e
quaisquer integrações HTTP futuras.

O domínio e os casos de uso **não conhecem** essas políticas — elas ficam
exclusivamente na camada de adaptadores, dentro de `internal/platform/resilience`.

---

## Onde se aplica

```
Core (casos de uso)
       │
       ▼ porta (interface)
  Adapter outbound
       │
       ▼ platform/resilience.Policy
  Dependência externa  ← CepGateway, FiscalGateway, ...
```

Cada adaptador de saída instancia uma `Policy` e envolve toda chamada HTTP nela.
O domínio continua recebendo apenas sucesso ou erro de domínio.

---

## Os três mecanismos

### 1 — Circuit Breaker

Evita chamadas em cascata a uma dependência que já está falhando.

**Estados:**

```
             falhas >= FailureThreshold
  Closed ──────────────────────────────► Open
    ▲                                      │
    │                                      │ após Timeout
    │      sucessos >= SuccessThreshold    │
    └──────────────────── HalfOpen ◄───────┘
                              │
                              │ falha
                              └──────────► Open  (reinicia o timer)
```

| Parâmetro | Descrição |
|-----------|-----------|
| `FailureThreshold` | Falhas consecutivas em `Closed` para abrir |
| `SuccessThreshold` | Sucessos consecutivos em `HalfOpen` para fechar |
| `Timeout` | Tempo em `Open` antes de tentar `HalfOpen` |

**Comportamento por estado:**

- **Closed:** chamadas passam normalmente; falha consecutiva incrementa contador.
  Sucesso zera o contador.
- **Open:** chamadas retornam `ErrCircuitOpen` imediatamente (fast-fail).
  Após `Timeout`, transiciona para `HalfOpen`.
- **HalfOpen:** permite **uma única sonda** por vez. Demais chamadas recebem
  `ErrCircuitOpen`. Sonda com sucesso decrementa a contagem de faltas;
  ao atingir `SuccessThreshold`, fecha. Sonda com falha reabre e reinicia o timer.

**Valores sugeridos para gateways externos:**

```
FailureThreshold : 5
SuccessThreshold : 2
Timeout          : 30s
```

---

### 2 — Bulkhead (semáforo de concorrência)

Limita quantas goroutines podem chamar uma dependência simultaneamente,
impedindo que um serviço lento esgote o pool de workers de toda a aplicação.

Implementado como **canal bufferizado** (semáforo counting):

```
┌──── sem (capacidade = MaxConcurrency) ────┐
│  slot  │  slot  │  slot  │ ...            │
└───────────────────────────────────────────┘
        ▲                        ▲
  adquire slot               libera slot
  (antes de chamar)          (defer, após retorno)
```

- Slot disponível → chamada prossegue; slot liberado no `defer`.
- Sem slot livre → retorna `ErrBulkheadFull` imediatamente (sem enfileirar).

| Parâmetro | Descrição |
|-----------|-----------|
| `MaxConcurrency` | Máximo de chamadas simultâneas permitidas |

**Valores sugeridos:**

```
CepGateway    MaxConcurrency: 10
FiscalGateway MaxConcurrency: 5
```

---

### 3 — Retry com backoff exponencial + full jitter

Retenta chamadas que falharam por erros transitórios, com espera crescente
e aleatoriedade para evitar thundering herd.

**Fórmula (full jitter):**

```
delay₀  = InitialDelay
delayₙ  = min(delayₙ₋₁ × Multiplier, MaxDelay)
sleep   = random[0, delayₙ)          ← full jitter
```

| Parâmetro | Descrição |
|-----------|-----------|
| `MaxAttempts` | Total de tentativas (1 = sem retry) |
| `InitialDelay` | Espera base antes da 2ª tentativa |
| `MaxDelay` | Teto do backoff |
| `Multiplier` | Fator de crescimento (ex.: `2.0` → dobra a cada ciclo) |
| `IsRetryable` | Predicado opcional; `nil` usa o padrão abaixo |

**Padrão de `IsRetryable`:** retenta tudo, **exceto**:
- `ErrCircuitOpen` — CB aberto; retry imediato não faz sentido.
- `ErrBulkheadFull` — bulkhead lotado; retry imediato não liberaria slot.
- Erros permanentes que o chamador marcar como não-retriáveis (ex.: 400, 404).

**Exemplo de progressão (InitialDelay=100ms, Multiplier=2, MaxDelay=4s):**

| Tentativa | Janela de jitter |
|-----------|-----------------|
| 1 → 2 | 0 – 100 ms |
| 2 → 3 | 0 – 200 ms |
| 3 → 4 | 0 – 400 ms |
| 4 → 5 | 0 – 800 ms |
| 5 (última) | — |

**Valores sugeridos:**

```
MaxAttempts  : 5
InitialDelay : 100ms
MaxDelay     : 4s
Multiplier   : 2.0
```

---

## Composição em `Policy`

Os três mecanismos são compostos em uma `Policy` única. A ordem importa:

```
Retry
  └──► CircuitBreaker
            └──► Bulkhead
                    └──► fn (chamada real)
```

**Por que essa ordem:**

| Camada | Papel |
|--------|-------|
| **Retry** (externo) | Decide se tenta de novo após o resultado de todas as camadas internas |
| **CircuitBreaker** | Fast-fail antes de ocupar um slot do bulkhead quando a dependência está down |
| **Bulkhead** (interno) | Controla concorrência apenas das chamadas que de fato chegam até aqui |

`ErrCircuitOpen` e `ErrBulkheadFull` são marcados como não-retriáveis, então o
Retry para imediatamente nesses casos.

---

## Localização no código

```
internal/platform/resilience/
├── errors.go           # ErrCircuitOpen, ErrBulkheadFull, ErrMaxRetriesReached
├── circuit_breaker.go  # CircuitBreakerConfig, CircuitBreaker
├── bulkhead.go         # BulkheadConfig, Bulkhead
├── retry.go            # RetryConfig, Retry
└── policy.go           # Policy — composição dos três
```

A `Policy` é instanciada no `module.go` de cada módulo que usa gateways externos
e injetada no adaptador de saída correspondente.

---

## Erros sentinela

| Erro | Quando ocorre |
|------|---------------|
| `ErrCircuitOpen` | CB em estado `Open` ou sonda já em andamento em `HalfOpen` |
| `ErrBulkheadFull` | `MaxConcurrency` atingido no momento da chamada |
| `ErrMaxRetriesReached` | Todas as tentativas esgotadas; encapsula o último erro |

Os adaptadores convertem esses erros para erros de domínio ou HTTP adequados
antes de retorná-los ao caso de uso.
