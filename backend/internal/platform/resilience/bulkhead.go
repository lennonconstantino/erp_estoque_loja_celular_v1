package resilience

import "context"

// BulkheadConfig parametriza o semáforo de concorrência.
type BulkheadConfig struct {
	MaxConcurrency int // máximo de chamadas simultâneas
}

// Bulkhead limita a concorrência de chamadas a uma dependência usando um
// semáforo (canal bufferizado). Não enfileira: sem slot livre, falha rápido.
type Bulkhead struct {
	sem chan struct{}
}

// NewBulkhead cria um bulkhead com a capacidade informada (mínimo 1).
func NewBulkhead(cfg BulkheadConfig) *Bulkhead {
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 1
	}
	return &Bulkhead{sem: make(chan struct{}, cfg.MaxConcurrency)}
}

// Execute roda fn se houver slot livre; caso contrário retorna ErrBulkheadFull.
func (b *Bulkhead) Execute(ctx context.Context, fn func(context.Context) error) error {
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return fn(ctx)
	default:
		return ErrBulkheadFull
	}
}
