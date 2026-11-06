package nfe

import (
	"fmt"
	"time"
)

// RPSRequestV2 representa os dados de um RPS no Leiaute 2 (versão v02 dos schemas de SP).
// Estrutura independente para facilitar futura depreciação do Leiaute 1.
type RPSRequestV2 struct {
	// Identificação do Prestador e do RPS
	Remetente             string    `json:"remetente"`
	IM                    int       `json:"im"`
	SerieRPS              string    `json:"serie_rps,omitempty"`
	NumeroRPS             int       `json:"numero_rps,omitempty"`
	DtEmissao             time.Time `json:"dt_emissao"`
	Tributacao            string    `json:"tributacao"` // T - Trib SP | F - Fora SP | A - Isento | B - Trib Fora e Isento | M - Imune | N - Não listado | X - Exigibilidade Suspensa | P - Exportação
	CodServico            string    `json:"cod_servico"` // 5 dígitos
	Aliquota              float64   `json:"aliquota"`
	ISS                   bool      `json:"iss"`
	Discriminacao         string    `json:"discriminacao"`

	// Identificação e Endereço do Tomador
	Tomador               string    `json:"tomador"`
	TomadorCidade         string    `json:"tomador_cidade,omitempty"` // IBGE 7 dígitos
	TomadorUF             string    `json:"tomador_uf,omitempty"`
	TomadorTipoLogradouro string    `json:"tomador_tipo_logradouro,omitempty"`
	TomadorLogradouro     string    `json:"tomador_logradouro,omitempty"`
	TomadorNumero         string    `json:"tomador_numero,omitempty"`
	TomadorComplemento    string    `json:"tomador_complemento,omitempty"`
	TomadorBairro         string    `json:"tomador_bairro,omitempty"`
	TomadorCEP            string    `json:"tomador_cep,omitempty"`

	// Deduções e Retenções Tradicionais
	ValorDeducoes         float64   `json:"valor_deducoes,omitempty"`
	PIS                   float64   `json:"pis,omitempty"`
	COFINS                float64   `json:"cofins,omitempty"`
	INSS                  float64   `json:"inss,omitempty"`
	IR                    float64   `json:"ir,omitempty"`
	CSLL                  float64   `json:"csll,omitempty"`

	// Valores Específicos do Leiaute 2
	ValorServ             float64   `json:"valor_serv,omitempty"`             // Alias compatível com v1; usado se ValorInicialCobrado for 0
	ValorInicialCobrado   float64   `json:"valor_inicial_cobrado,omitempty"` // Valor cobrado antes de tributos
	ValorFinalCobrado     float64   `json:"valor_final_cobrado,omitempty"`   // Alternativa ao ValorInicialCobrado (choice)
	ValorTotalRecebido    float64   `json:"valor_total_recebido,omitempty"`  // Total recebido (opcional)
	ValorMulta            float64   `json:"valor_multa,omitempty"`
	ValorJuros            float64   `json:"valor_juros,omitempty"`
	ValorIPI              float64   `json:"valor_ipi,omitempty"`             // Obrigatório no v2 (padrão 0.00 se zerado)

	// Regras Fiscais e Enquadramento Leiaute 2
	ExigibilidadeSuspensa int       `json:"exigibilidade_suspensa"`          // 0 - Não, 1 - Sim
	PagamentoParcelado    int       `json:"pagamento_parcelado_antecipado"`  // 0 - Não, 1 - Sim
	NCM                   string    `json:"ncm,omitempty"`                   // 8 dígitos (opcional)
	NBS                   string    `json:"nbs"`                             // 9 dígitos (obrigatório)
	CLocPrestacao         string    `json:"c_loc_prestacao,omitempty"`       // IBGE 7 dígitos
	CPaisPrestacao        string    `json:"c_pais_prestacao,omitempty"`      // Código ISO 2 letras (se exterior)

	// Bloco IBS/CBS (Reforma Tributária)
	FinNFSe               int       `json:"fin_nfse"`                        // 0 = Regular (padrão 0)
	IndFinal              int       `json:"ind_final"`                       // 0 - Não, 1 - Sim
	CIndOp                string    `json:"c_ind_op"`                        // 6 dígitos (obrigatório)
	IndDest               int       `json:"ind_dest"`                        // 0 = Tomador é o destinatário (padrão 0)
	CClassTrib            string    `json:"c_class_trib"`                    // 6 dígitos (obrigatório)
	CClassTribReg         string    `json:"c_class_trib_reg,omitempty"`      // 6 dígitos (opcional)
}

// LoteRPSRequestV2 representa um payload em lote contendo requisições do Leiaute 2.
type LoteRPSRequestV2 struct {
	Requests []RPSRequestV2 `json:"requests"`
}

// GetValorCobrado retorna o valor principal do serviço do RPS (priorizando ValorInicialCobrado, depois ValorFinalCobrado, depois ValorServ).
func (r *RPSRequestV2) GetValorCobrado() float64 {
	if r.ValorInicialCobrado > 0 {
		return r.ValorInicialCobrado
	}
	if r.ValorFinalCobrado > 0 {
		return r.ValorFinalCobrado
	}
	return r.ValorServ
}

// Validate valida campos essenciais do Leiaute 2.
func (r *RPSRequestV2) Validate() error {
	if r.Remetente == "" {
		return fmt.Errorf("remetente obrigatorio")
	}
	if r.Tomador == "" {
		return fmt.Errorf("tomador obrigatorio")
	}
	if r.GetValorCobrado() <= 0 {
		return fmt.Errorf("valor do servico (valor_inicial_cobrado ou valor_serv) obrigatorio")
	}
	if r.NBS == "" {
		return fmt.Errorf("nbs obrigatorio no leiaute 2 (9 digitos)")
	}
	if len(r.NBS) != 9 {
		return fmt.Errorf("nbs deve conter exatamente 9 digitos")
	}
	if r.CIndOp == "" {
		return fmt.Errorf("c_ind_op obrigatorio no leiaute 2 (6 digitos)")
	}
	if len(r.CIndOp) != 6 {
		return fmt.Errorf("c_ind_op deve conter exatamente 6 digitos")
	}
	if r.CClassTrib == "" {
		return fmt.Errorf("c_class_trib obrigatorio no leiaute 2 (6 digitos)")
	}
	if len(r.CClassTrib) != 6 {
		return fmt.Errorf("c_class_trib deve conter exatamente 6 digitos")
	}
	return nil
}

