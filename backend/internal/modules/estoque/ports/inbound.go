package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/estoque/domain"
)

// LancarAjusteInput dados necessários para lançar um ajuste manual de estoque.
type LancarAjusteInput struct {
	ProdutoID     uuid.UUID
	QtdEntrada    int
	QtdSaida      int
	Motivo        string
	ResponsavelID uuid.UUID
}

// EstoqueWriter é a interface exposta para outros módulos lançarem movimentações
// de entrada (compras) ou saída (vendas) sem passar pelos casos de uso de ajuste.
type EstoqueWriter interface {
	RegistrarEntradaCompra(ctx context.Context, produtoID, compraID, responsavelID uuid.UUID, quantidade int) error
	RegistrarSaidaVenda(ctx context.Context, produtoID, vendaID, responsavelID uuid.UUID, quantidade int) error
}

// EstoqueService é a porta de entrada do módulo estoque.
type EstoqueService interface {
	// LancarAjuste registra um ajuste manual, gera movimentações no razão e
	// atualiza o saldo materializado no catálogo.
	LancarAjuste(ctx context.Context, in LancarAjusteInput) (*domain.Ajuste, error)

	// ConsultarMovimentacoes retorna o razão (ledger) de um produto, do mais
	// recente para o mais antigo.
	ConsultarMovimentacoes(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Movimentacao, error)

	// ConsultarAjustes retorna os ajustes manuais de um produto, do mais recente
	// para o mais antigo.
	ConsultarAjustes(ctx context.Context, produtoID uuid.UUID, limit, offset int) ([]domain.Ajuste, error)
}
