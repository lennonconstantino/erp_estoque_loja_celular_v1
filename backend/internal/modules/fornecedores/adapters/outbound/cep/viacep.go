// Package cep implementa ports.CepGateway consultando a API ViaCEP, com as
// políticas de resiliência aplicadas no adaptador (fora do domínio).
package cep

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/modules/fornecedores/domain"
	"github.com/lennonconstantino/erp_estoque_loja_celular/backend/internal/platform/resilience"
)

// Gateway consulta CEPs em uma API HTTP (ViaCEP por padrão).
type Gateway struct {
	baseURL string
	client  *http.Client
	policy  *resilience.Policy
}

// NewGateway cria o adaptador. baseURL ex.: "https://viacep.com.br/ws".
func NewGateway(baseURL string, policy *resilience.Policy) *Gateway {
	return &Gateway{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
		policy:  policy,
	}
}

type resposta struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"`
}

// Lookup busca o endereço do CEP, envolvendo a chamada na Policy de resiliência.
func (g *Gateway) Lookup(ctx context.Context, cep string) (domain.Endereco, error) {
	cep = domain.NormalizarDigitos(cep)
	if len(cep) != 8 {
		return domain.Endereco{}, resilience.Permanent(fmt.Errorf("cep inválido"))
	}

	var out domain.Endereco
	err := g.policy.Execute(ctx, func(ctx context.Context) error {
		end, err := g.fetch(ctx, cep)
		if err != nil {
			return err
		}
		out = end
		return nil
	})
	return out, err
}

func (g *Gateway) fetch(ctx context.Context, cep string) (domain.Endereco, error) {
	url := fmt.Sprintf("%s/%s/json/", g.baseURL, cep)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Endereco{}, resilience.Permanent(err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return domain.Endereco{}, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		// segue
	case resp.StatusCode >= 500:
		return domain.Endereco{}, fmt.Errorf("viacep: status %d", resp.StatusCode)
	default:
		return domain.Endereco{}, resilience.Permanent(fmt.Errorf("viacep: status %d", resp.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return domain.Endereco{}, err
	}
	var r resposta
	if err := json.Unmarshal(body, &r); err != nil {
		return domain.Endereco{}, resilience.Permanent(fmt.Errorf("viacep: json inválido"))
	}
	if r.Erro {
		return domain.Endereco{}, resilience.Permanent(fmt.Errorf("cep não encontrado"))
	}

	return domain.Endereco{
		CEP:    cep,
		Rua:    r.Logradouro,
		Bairro: r.Bairro,
		Cidade: r.Localidade,
		UF:     r.UF,
	}, nil
}
