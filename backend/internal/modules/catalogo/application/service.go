// Package application contém os casos de uso do contexto catálogo. Orquestra o
// domínio e as portas de saída; não conhece HTTP nem detalhes de banco.
package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/catalogo/ports"
)

// Service implementa CategoriaService, ProdutoService, CatalogoReader e CatalogoWriter.
type Service struct {
	cats  ports.CategoriaRepository
	prods ports.ProdutoRepository
}

// NewService injeta as portas de saída.
func NewService(cats ports.CategoriaRepository, prods ports.ProdutoRepository) *Service {
	return &Service{cats: cats, prods: prods}
}

// Conformidade em tempo de compilação.
var _ ports.CategoriaService = (*Service)(nil)
var _ ports.ProdutoService   = (*Service)(nil)
var _ ports.CatalogoReader   = (*Service)(nil) // inclui ExisteProduto e ConsultarSaldoProduto
var _ ports.CatalogoWriter   = (*Service)(nil) // inclui DecrementarSaldo

// ─── CategoriaService ────────────────────────────────────────────────────────

func (s *Service) CriarCategoria(ctx context.Context, in ports.CriarCategoriaInput) (*domain.Categoria, error) {
	cat, err := domain.NovaCategoria(in.Descricao)
	if err != nil {
		return nil, err
	}
	if ex, err := s.cats.FindByDescricao(ctx, cat.Descricao); err != nil {
		if !errors.Is(err, domain.ErrCategoriaNaoEncontrada) {
			return nil, err
		}
	} else if ex != nil {
		return nil, domain.ErrCategoriaJaCadastrada
	}
	if err := s.cats.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) AtualizarCategoria(ctx context.Context, id uuid.UUID, in ports.AtualizarCategoriaInput) (*domain.Categoria, error) {
	cat, err := s.cats.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := cat.Atualizar(in.Descricao); err != nil {
		return nil, err
	}
	if err := s.cats.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) RemoverCategoria(ctx context.Context, id uuid.UUID) error {
	tem, err := s.cats.HasProdutos(ctx, id)
	if err != nil {
		return err
	}
	if tem {
		return domain.ErrCategoriaComProdutos
	}
	return s.cats.Delete(ctx, id)
}

func (s *Service) BuscarCategoriaPorID(ctx context.Context, id uuid.UUID) (*domain.Categoria, error) {
	return s.cats.FindByID(ctx, id)
}

func (s *Service) ListarCategorias(ctx context.Context, q string, limit, offset int) ([]domain.Categoria, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.cats.List(ctx, q, limit, offset)
}

// ─── ProdutoService ──────────────────────────────────────────────────────────

func (s *Service) CriarProduto(ctx context.Context, in ports.CriarProdutoInput) (*domain.Produto, error) {
	if _, err := s.cats.FindByID(ctx, in.CategoriaID); err != nil {
		return nil, err
	}
	p, err := domain.NovoProduto(in.CategoriaID, in.Descricao, in.PrecoCusto, in.PrecoVenda, in.EstoqueMinimo, in.GarantiaMeses, in.Modelo)
	if err != nil {
		return nil, err
	}
	if err := s.prods.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) AtualizarProduto(ctx context.Context, id uuid.UUID, in ports.AtualizarProdutoInput) (*domain.Produto, error) {
	p, err := s.prods.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ativo := p.Ativo
	if in.Ativo != nil {
		ativo = *in.Ativo
	}
	if err := p.AtualizarDados(in.Descricao, in.CategoriaID, in.PrecoCusto, in.PrecoVenda, in.EstoqueMinimo, in.GarantiaMeses, in.Modelo, ativo); err != nil {
		return nil, err
	}
	if err := s.prods.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) BuscarProdutoPorID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	return s.prods.FindByID(ctx, id)
}

func (s *Service) ListarProdutos(ctx context.Context, q string, categoriaID *uuid.UUID, limit, offset int) ([]domain.Produto, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.prods.List(ctx, q, categoriaID, limit, offset)
}

// ─── CatalogoReader ──────────────────────────────────────────────────────────

func (s *Service) BuscarProduto(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	return s.prods.FindByID(ctx, id)
}

// ─── CatalogoReader (cross-module) ───────────────────────────────────────────

// ExisteProduto retorna nil se o produto existir, ou o erro de domínio caso contrário.
// Consumida como porta de saída por compras e vendas (duck typing sem importar catalogo/domain).
func (s *Service) ExisteProduto(ctx context.Context, id uuid.UUID) error {
	_, err := s.prods.FindByID(ctx, id)
	return err
}

// ConsultarSaldoProduto retorna o saldo materializado atual do produto.
// Consumida por vendas para validação pré-confirmação.
func (s *Service) ConsultarSaldoProduto(ctx context.Context, id uuid.UUID) (int, error) {
	p, err := s.prods.FindByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return p.EstoqueAtual, nil
}

// ─── CatalogoWriter ──────────────────────────────────────────────────────────

func (s *Service) AtualizarSaldo(ctx context.Context, produtoID uuid.UUID, novoSaldo int) error {
	p, err := s.prods.FindByID(ctx, produtoID)
	if err != nil {
		return err
	}
	if err := p.AtualizarSaldo(novoSaldo); err != nil {
		return err
	}
	return s.prods.UpdateSaldo(ctx, produtoID, p.EstoqueAtual, p.Disponivel)
}

// DecrementarSaldo atomicamente decrementa o saldo materializado se houver estoque.
// Usado por vendas para garantir que saldo negativo é impossível sob concorrência.
func (s *Service) DecrementarSaldo(ctx context.Context, produtoID uuid.UUID, quantidade int) (int, error) {
	return s.prods.DecrementarSaldo(ctx, produtoID, quantidade)
}
