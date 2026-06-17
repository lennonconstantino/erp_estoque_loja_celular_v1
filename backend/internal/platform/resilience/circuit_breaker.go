package resilience

import (
	"context"
	"sync"
	"time"
)

// CircuitBreakerConfig parametriza o circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int           // falhas consecutivas em Closed para abrir
	SuccessThreshold int           // sucessos consecutivos em HalfOpen para fechar
	Timeout          time.Duration // tempo em Open antes de tentar HalfOpen
}

type cbState int

const (
	stateClosed cbState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker implementa o padrão homônimo com os estados
// Closed → Open → HalfOpen. É seguro para uso concorrente.
type CircuitBreaker struct {
	cfg CircuitBreakerConfig

	mu        sync.Mutex
	state     cbState
	failures  int
	successes int
	openedAt  time.Time
	probing   bool // true enquanto uma sonda HalfOpen está em andamento
}

// NewCircuitBreaker cria um breaker, aplicando defaults sensatos.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &CircuitBreaker{cfg: cfg, state: stateClosed}
}

// Execute roda fn protegida pelo breaker.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if err := cb.before(); err != nil {
		return err
	}
	err := fn(ctx)
	cb.after(err)
	return err
}

// before decide se a chamada pode prosseguir e atualiza o estado de transição.
func (cb *CircuitBreaker) before() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.cfg.Timeout {
			// janela de sonda: transiciona para HalfOpen e libera UMA chamada
			cb.state = stateHalfOpen
			cb.successes = 0
			cb.probing = true
			return nil
		}
		return ErrCircuitOpen
	case stateHalfOpen:
		if cb.probing {
			return ErrCircuitOpen // já existe uma sonda em andamento
		}
		cb.probing = true
		return nil
	default: // stateClosed
		return nil
	}
}

// after contabiliza o resultado e aplica as transições de estado.
func (cb *CircuitBreaker) after(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == stateHalfOpen {
		cb.probing = false
	}

	// Erros do próprio stack interno não contam como falha da dependência.
	if err == ErrCircuitOpen || err == ErrBulkheadFull {
		return
	}

	if err != nil {
		switch cb.state {
		case stateHalfOpen:
			cb.trip() // sonda falhou: reabre e reinicia o timer
		case stateClosed:
			cb.failures++
			if cb.failures >= cb.cfg.FailureThreshold {
				cb.trip()
			}
		}
		return
	}

	// sucesso
	switch cb.state {
	case stateHalfOpen:
		cb.successes++
		if cb.successes >= cb.cfg.SuccessThreshold {
			cb.closeCircuit()
		}
	case stateClosed:
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) trip() {
	cb.state = stateOpen
	cb.openedAt = time.Now()
	cb.failures = 0
	cb.successes = 0
}

func (cb *CircuitBreaker) closeCircuit() {
	cb.state = stateClosed
	cb.failures = 0
	cb.successes = 0
	cb.probing = false
}
