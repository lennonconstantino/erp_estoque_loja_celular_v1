package resilience

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// RetryConfig parametriza o retry com backoff exponencial + full jitter.
type RetryConfig struct {
	MaxAttempts  int                 // total de tentativas (1 = sem retry)
	InitialDelay time.Duration       // espera base antes da 2ª tentativa
	MaxDelay     time.Duration       // teto do backoff
	Multiplier   float64             // fator de crescimento (ex.: 2.0)
	IsRetryable  func(error) bool    // predicado opcional; nil usa o padrão
}

// Retry retenta chamadas que falharam por erros transitórios.
type Retry struct {
	cfg RetryConfig
}

// NewRetry cria um retry, aplicando defaults sensatos.
func NewRetry(cfg RetryConfig) *Retry {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = 100 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 4 * time.Second
	}
	if cfg.Multiplier <= 1 {
		cfg.Multiplier = 2.0
	}
	return &Retry{cfg: cfg}
}

// defaultIsRetryable retenta tudo, exceto erros do próprio stack de resiliência.
func defaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrCircuitOpen) || errors.Is(err, ErrBulkheadFull) {
		return false
	}
	if IsPermanent(err) {
		return false
	}
	return true
}

// Execute roda fn com retentativas conforme a política configurada.
func (r *Retry) Execute(ctx context.Context, fn func(context.Context) error) error {
	isRetryable := r.cfg.IsRetryable
	if isRetryable == nil {
		isRetryable = defaultIsRetryable
	}

	delay := r.cfg.InitialDelay
	var lastErr error

	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		// erro não-retriável ou última tentativa: devolve o erro original
		if !isRetryable(err) || attempt == r.cfg.MaxAttempts {
			if attempt == r.cfg.MaxAttempts && isRetryable(err) {
				return fmt.Errorf("%w: %v", ErrMaxRetriesReached, lastErr)
			}
			return err
		}

		// full jitter: sleep aleatório em [0, delay)
		var sleep time.Duration
		if delay > 0 {
			sleep = time.Duration(rand.Int63n(int64(delay)))
		}
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}

		// próximo passo do backoff exponencial, limitado por MaxDelay
		delay = time.Duration(float64(delay) * r.cfg.Multiplier)
		if delay > r.cfg.MaxDelay {
			delay = r.cfg.MaxDelay
		}
	}

	return fmt.Errorf("%w: %v", ErrMaxRetriesReached, lastErr)
}
