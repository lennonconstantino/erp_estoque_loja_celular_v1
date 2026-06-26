package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Categoria é a raiz de agregação do subcontexto de categorias de produto.
type Categoria struct {
	ID        uuid.UUID
	Descricao string
	CriadoEm time.Time
}

// NovaCategoria cria uma categoria válida.
func NovaCategoria(descricao string) (*Categoria, error) {
	c := &Categoria{
		ID:        uuid.New(),
		Descricao: strings.TrimSpace(descricao),
		CriadoEm: time.Now().UTC(),
	}
	if err := c.Validar(); err != nil {
		return nil, err
	}
	return c, nil
}

// Atualizar altera a descrição da categoria.
func (c *Categoria) Atualizar(descricao string) error {
	c.Descricao = strings.TrimSpace(descricao)
	return c.Validar()
}

// Validar verifica as invariantes da categoria.
func (c *Categoria) Validar() error {
	if c.Descricao == "" {
		return ErrDescricaoCatObrigatoria
	}
	return nil
}
