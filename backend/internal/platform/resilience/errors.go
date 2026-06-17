// Package resilience fornece políticas de resiliência (Circuit Breaker,
// Bulkhead e Retry) aplicadas exclusivamente nos adaptadores de saída que
// se comunicam com dependências externas. O domínio e os casos de uso não
// conhecem este pacote.
package resilience

import "errors"

var (
	// ErrCircuitOpen indica que o circuit breaker está aberto (ou que já há
	// uma sonda em andamento no estado HalfOpen).
	ErrCircuitOpen = errors.New("resilience: circuit breaker aberto")

	// ErrBulkheadFull indica que o limite de concorrência foi atingido.
	ErrBulkheadFull = errors.New("resilience: bulkhead lotado")

	// ErrMaxRetriesReached indica que todas as tentativas foram esgotadas.
	// Encapsula (via %w) o último erro observado.
	ErrMaxRetriesReached = errors.New("resilience: máximo de tentativas atingido")
)

// permanentErr marca um erro como não-retriável (ex.: 400/404 de uma API).
type permanentErr struct{ err error }

func (e permanentErr) Error() string { return e.err.Error() }
func (e permanentErr) Unwrap() error { return e.err }

// Permanent envolve err sinalizando ao Retry que NÃO deve retentá-lo.
// O adaptador de saída usa isto para erros definitivos da dependência.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return permanentErr{err}
}

// IsPermanent reporta se err (ou algo na sua cadeia) foi marcado como permanente.
func IsPermanent(err error) bool {
	var p permanentErr
	return errors.As(err, &p)
}
