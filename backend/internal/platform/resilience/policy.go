package resilience

import "context"

// Policy compõe os três mecanismos na ordem:
//
//	Retry → CircuitBreaker → Bulkhead → fn
//
// O Retry (externo) decide novas tentativas; o CircuitBreaker faz fast-fail
// antes de ocupar um slot do Bulkhead; o Bulkhead controla a concorrência
// apenas das chamadas que de fato chegam até a dependência.
type Policy struct {
	retry *Retry
	cb    *CircuitBreaker
	bh    *Bulkhead
}

// NewPolicy monta uma Policy a partir das três configurações.
func NewPolicy(retry RetryConfig, cb CircuitBreakerConfig, bh BulkheadConfig) *Policy {
	return &Policy{
		retry: NewRetry(retry),
		cb:    NewCircuitBreaker(cb),
		bh:    NewBulkhead(bh),
	}
}

// Execute roda fn através de toda a cadeia de resiliência.
func (p *Policy) Execute(ctx context.Context, fn func(context.Context) error) error {
	return p.retry.Execute(ctx, func(ctx context.Context) error {
		return p.cb.Execute(ctx, func(ctx context.Context) error {
			return p.bh.Execute(ctx, fn)
		})
	})
}
