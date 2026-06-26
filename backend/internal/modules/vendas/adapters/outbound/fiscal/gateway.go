// Package fiscal implementa ports.FiscalGateway como stub para emissão de
// cupons e notas fiscais. Em produção, substituir pela API fiscal real.
package fiscal

import (
	"context"
	"fmt"
	"time"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/vendas/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/resilience"
)

// Gateway emite documentos fiscais via API externa (stub em desenvolvimento).
type Gateway struct {
	policy *resilience.Policy
}

// NewGateway cria o gateway com a política de resiliência injetada.
func NewGateway(policy *resilience.Policy) *Gateway {
	return &Gateway{policy: policy}
}

// EmitirDocumento despacha para EmitirCupom ou EmitirNF de acordo com o tipo
// da venda, envolvendo a chamada na Policy de resiliência.
func (g *Gateway) EmitirDocumento(ctx context.Context, v *domain.Venda) (string, error) {
	var numDoc string
	err := g.policy.Execute(ctx, func(ctx context.Context) error {
		var n string
		var err error
		switch v.DocFiscal {
		case domain.DocNF:
			n, err = g.emitirNF(ctx, v)
		default:
			n, err = g.emitirCupom(ctx, v)
		}
		if err != nil {
			return err
		}
		numDoc = n
		return nil
	})
	return numDoc, err
}

// emitirCupom gera um número de cupom fiscal (stub).
func (g *Gateway) emitirCupom(_ context.Context, v *domain.Venda) (string, error) {
	return fmt.Sprintf("CUP-%s-%04d", time.Now().UTC().Format("20060102"), sequencial(v.ID.String())), nil
}

// emitirNF gera um número de nota fiscal (stub).
func (g *Gateway) emitirNF(_ context.Context, v *domain.Venda) (string, error) {
	return fmt.Sprintf("NF-%s-%04d", time.Now().UTC().Format("20060102"), sequencial(v.ID.String())), nil
}

// sequencial deriva um número de sequência a partir do UUID (últimos 4 dígitos em hex → decimal).
func sequencial(id string) int {
	if len(id) < 4 {
		return 1
	}
	last := id[len(id)-4:]
	var n int
	for _, c := range last {
		n = n*16
		switch {
		case c >= '0' && c <= '9':
			n += int(c - '0')
		case c >= 'a' && c <= 'f':
			n += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			n += int(c-'A') + 10
		}
	}
	return (n % 9999) + 1
}
