// Package domain define os tipos de valor retornados pelos relatórios.
// Não há invariantes de negócio: os dados já vêm validados do banco.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProdutoAbaixoMinimo representa um produto cujo estoque atual ficou
// abaixo do estoque mínimo configurado.
type ProdutoAbaixoMinimo struct {
	ID            uuid.UUID
	Descricao     string
	EstoqueAtual  int
	EstoqueMinimo int
	Defasagem     int // EstoqueMinimo - EstoqueAtual
}

// ProdutoVendido agrega as vendas de um produto num período.
type ProdutoVendido struct {
	ProdutoID    uuid.UUID
	Descricao    string
	TotalVendido int
	TotalValor   float64
}

// ResumoVendas consolida as vendas confirmadas num intervalo de datas.
type ResumoVendas struct {
	TotalVendas int
	ValorTotal  float64
	TicketMedio float64
	De          time.Time
	Ate         time.Time
}

// ResumoCompras consolida as compras confirmadas num intervalo de datas.
type ResumoCompras struct {
	TotalCompras int
	ValorTotal   float64
	De           time.Time
	Ate          time.Time
}
