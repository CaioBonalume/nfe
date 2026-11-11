package nfe

import (
	"encoding/json"
	"fmt"
	"time"
)

// RPSRequestV2 representa os dados de um RPS no Leiaute 2 (versão v02 dos schemas de SP).
// Estrutura independente para facilitar futura depreciação do Leiaute 1.
type RPSRequestV2 struct {
	// Identificação do Prestador e do RPS
	Remetente     string    `json:"remetente"`
	IM            int       `json:"im"`
	SerieRPS      string    `json:"serie_rps,omitempty"`
	NumeroRPS     int       `json:"numero_rps,omitempty"`
	DtEmissao     time.Time `json:"dt_emissao"`
	Tributacao    string    `json:"tributacao"`  // T - Trib SP | F - Fora SP | A - Isento | B - Trib Fora e Isento | M - Imune | N - Não listado | X - Exigibilidade Suspensa | P - Exportação
	CodServico    string    `json:"cod_servico"` // 5 dígitos
	Aliquota      float64   `json:"aliquota"`
	ISS           bool      `json:"iss"`
	Discriminacao string    `json:"discriminacao"`

	// Identificação e Endereço do Tomador
	Tomador               string `json:"tomador"`
	TomadorCidade         string `json:"tomador_cidade,omitempty"` // IBGE 7 dígitos
	TomadorUF             string `json:"tomador_uf,omitempty"`
	TomadorTipoLogradouro string `json:"tomador_tipo_logradouro,omitempty"`
	TomadorLogradouro     string `json:"tomador_logradouro,omitempty"`
	TomadorNumero         string `json:"tomador_numero,omitempty"`
	TomadorComplemento    string `json:"tomador_complemento,omitempty"`
	TomadorBairro         string `json:"tomador_bairro,omitempty"`
	TomadorCEP            string `json:"tomador_cep,omitempty"`

	// Deduções e Retenções Tradicionais
	ValorDeducoes     float64 `json:"valor_deducoes,omitempty"`
	PIS               float64 `json:"pis,omitempty"`
	COFINS            float64 `json:"cofins,omitempty"`
	INSS              float64 `json:"inss,omitempty"`
	IR                float64 `json:"ir,omitempty"`
	CSLL              float64 `json:"csll,omitempty"`
	RetencaoPisCofins string  `json:"retencao_pis_cofins,omitempty"` // 0, 3, 4, 5, 6, 7, 8, 9 (conforme tpRetencaoPisCofins)

	// Valores Específicos do Leiaute 2
	ValorServ           float64 `json:"valor_serv,omitempty"`            // Alias compatível com v1; usado se ValorInicialCobrado for 0
	ValorInicialCobrado float64 `json:"valor_inicial_cobrado,omitempty"` // (padrão 0.00 se zerado)
	ValorFinalCobrado   float64 `json:"valor_final_cobrado,omitempty"`   // Alternativa ao ValorInicialCobrado (choice)
	ValorTotalRecebido  float64 `json:"valor_total_recebido,omitempty"`  // Total recebido (opcional)
	ValorMulta          float64 `json:"valor_multa,omitempty"`
	ValorJuros          float64 `json:"valor_juros,omitempty"`
	ValorIPI            float64 `json:"valor_ipi,omitempty"` // Obrigatório no v2 (padrão 0.00 se zerado)

	// Regras Fiscais e Enquadramento Leiaute 2
	ExigibilidadeSuspensa int    `json:"exigibilidade_suspensa"`                   // 0 - Não, 1 - Sim
	PagamentoParcelado    int    `json:"pagamento_parcelado_antecipado,omitempty"` // Deprecated: descontinuado pela Prefeitura (Alerta 648)
	NCM                   string `json:"ncm,omitempty"`                            // 8 dígitos (opcional)
	NBS                   string `json:"nbs"`                            // 9 dígitos (obrigatório)
	MunicipioPrestacao    string `json:"municipio_prestacao,omitempty"`  // IBGE 7 dígitos (obrigatório se tributação 'F' ou 'B')
	CLocPrestacao         string `json:"c_loc_prestacao,omitempty"`      // IBGE 7 dígitos
	CPaisPrestacao        string `json:"c_pais_prestacao,omitempty"`     // Código ISO 2 letras (se exterior)

	// Bloco IBS/CBS (Reforma Tributária)
	FinNFSe       int    `json:"fin_nfse"`                   // 0 = Regular (padrão 0)
	IndFinal      int    `json:"ind_final"`                  // 0 - Não, 1 - Sim
	CIndOp        string `json:"c_ind_op"`                   // 6 dígitos (obrigatório)
	IndDest       int    `json:"ind_dest"`                   // 0 = Tomador é o destinatário (padrão 0)
	CClassTrib    string `json:"c_class_trib"`               // 6 dígitos (obrigatório)
	CClassTribReg string `json:"c_class_trib_reg,omitempty"` // 6 dígitos (opcional)
}

// LoteRPSRequestV2 representa um payload em lote contendo requisições do Leiaute 2.
type LoteRPSRequestV2 struct {
	Requests []RPSRequestV2 `json:"requests"`
}

// UnmarshalJSON permite desserialização flexível para o Leiaute 2, aceitando aliases como numero_dps.
func (r *RPSRequestV2) UnmarshalJSON(data []byte) error {
	type Alias RPSRequestV2
	aux := struct {
		*Alias
		NumeroDPS int `json:"numero_dps"`
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if r.NumeroRPS == 0 && aux.NumeroDPS != 0 {
		r.NumeroRPS = aux.NumeroDPS
	}

	return nil
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

// GetMunicipioPrestacao retorna o código IBGE do município onde o serviço foi prestado.
// Regras da Prefeitura de SP:
// - Para 'T' (São Paulo) ou 'P' (Exportação): NÃO deve ser informado (Erro 1223).
// - Para 'F' (Fora de SP) ou 'B' (Fora de SP e Isento): OBRIGATÓRIO informar (Erro 1222).
// Fallback: se não informado explicitamente em MunicipioPrestacao, resolve a partir de
// TomadorCidade ou CLocPrestacao (desde que diferente de SP 3550308).
func (r *RPSRequestV2) GetMunicipioPrestacao() string {
	if r.Tributacao == "T" || r.Tributacao == "P" {
		return ""
	}
	if r.MunicipioPrestacao != "" {
		return r.MunicipioPrestacao
	}
	if r.Tributacao == "F" || r.Tributacao == "B" {
		if r.TomadorCidade != "" && r.TomadorCidade != "3550308" {
			return r.TomadorCidade
		}
		if r.CLocPrestacao != "" && r.CLocPrestacao != "3550308" {
			return r.CLocPrestacao
		}
	}
	return ""
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
	if r.RetencaoPisCofins != "" {
		switch r.RetencaoPisCofins {
		case "0", "3", "4", "5", "6", "7", "8", "9":
		default:
			return fmt.Errorf("retencao_pis_cofins invalido: deve ser 0, 3, 4, 5, 6, 7, 8 ou 9")
		}
	}
	if r.Tributacao == "F" || r.Tributacao == "B" {
		mun := r.GetMunicipioPrestacao()
		if mun == "" {
			return fmt.Errorf("municipio_prestacao obrigatorio quando tributacao for fora de Sao Paulo ('F' ou 'B')")
		}
		if mun == "3550308" {
			return fmt.Errorf("municipio_prestacao nao pode ser Sao Paulo (3550308) quando tributacao for fora de Sao Paulo ('F' ou 'B')")
		}
		if len(mun) != 7 {
			return fmt.Errorf("municipio_prestacao deve conter exatamente 7 digitos (codigo IBGE)")
		}
	}
	return nil
}
