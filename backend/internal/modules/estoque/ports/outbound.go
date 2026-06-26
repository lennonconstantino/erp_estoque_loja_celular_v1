package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
)

// MovimentacaoRepository persiste e consulta o razão de movimentações (append-only).
type MovimentacaoRepository interface {
	Create(ctx context.Context, m *domain.Movimentacao) error
	ListByProduto(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Movimentacao, error)
	// ConsultarSaldoAtual retorna o saldo_atu_mov da última entrada do produto,
	// ou 0 se ainda não houver movimentações.
	ConsultarSaldoAtual(ctx context.Context, produtoID uuid.UUID) (int, error)
}

// AjusteRepository persiste e consulta os documentos de ajuste (append-only).
type AjusteRepository interface {
	Create(ctx context.Context, a *domain.Ajuste) error
	ListByProduto(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Ajuste, error)
}

// CatalogoWriter é a porta de saída para atualizar o saldo materializado
// em catalogo.produtos (estoque_a_pro e disp_pro).
// Implementada pelo módulo catálogo e injetada via module.go.
type CatalogoWriter interface {
	AtualizarSaldo(ctx context.Context, produtoID uuid.UUID, novoSaldo int) error
}
