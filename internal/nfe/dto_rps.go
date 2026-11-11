package nfe

import (
	"encoding/json"
	"fmt"
	"time"
)

type RPSRequest struct {
	Remetente             string    `json:"remetente"`
	Tomador               string    `json:"tomador"`
	TomadorCidade         string    `json:"tomador_cidade,omitempty"`         // opcional (código IBGE 7 dígitos)
	TomadorUF             string    `json:"tomador_uf,omitempty"`             // opcional (sigla estado 2 letras)
	TomadorTipoLogradouro string    `json:"tomador_tipo_logradouro,omitempty"` // opcional (Rua, Av, etc)
	TomadorLogradouro     string    `json:"tomador_logradouro,omitempty"`     // opcional
	TomadorNumero         string    `json:"tomador_numero,omitempty"`         // opcional
	TomadorComplemento    string    `json:"tomador_complemento,omitempty"`    // opcional
	TomadorBairro         string    `json:"tomador_bairro,omitempty"`         // opcional
	TomadorCEP            string    `json:"tomador_cep,omitempty"`            // opcional (8 dígitos)
	IM                    int       `json:"im"`
	SerieRPS              string    `json:"serie_rps,omitempty"`              // opcional, use "" para padrão
	NumeroRPS             int       `json:"numero_rps,omitempty"`
	DtEmissao             time.Time `json:"dt_emissao"`
	Tributacao            string    `json:"tributacao,omitempty"`             // T - Tributado em SP | F - Fora de SP | A - Isento | B - Tributado fora de SP e Isento | M - Imune | N - Serviço não listado no ISS | X - Tributado em SP com Exigibilidade Suspensa | P - Exportação de serviços
	MunicipioPrestacao    string    `json:"municipio_prestacao,omitempty"`    // opcional (código IBGE 7 dígitos - obrigatório se tributação 'F' ou 'B')
	ValorServ             float64   `json:"valor_serv"`
	ValorDeducoes         float64   `json:"valor_deducoes,omitempty"`         // opcional, use 0 para padrão
	PIS                   float64   `json:"pis,omitempty"`                    // opcional, use 0 para padrão
	COFINS                float64   `json:"cofins,omitempty"`                 // opcional, use 0 para padrão
	INSS                  float64   `json:"inss,omitempty"`                   // opcional, use 0 para padrão
	IR                    float64   `json:"ir,omitempty"`                     // opcional, use 0 para padrão
	CSLL                  float64   `json:"csll,omitempty"`                   // opcional, use 0 para padrão
	CodServico            string    `json:"cod_servico"`                      // 5 dígitos
	Aliquota              float64   `json:"aliquota"`
	ISS                   bool      `json:"iss"`
	Discriminacao         string    `json:"discriminacao"`
}

// LoteRPSRequest representa o envelope de lote para Leiaute 1.
type LoteRPSRequest struct {
	Requests []RPSRequest `json:"requests"`
}

// UnmarshalJSON permite desserialização flexível (aceita aliases de valor como valor_servicos ou valor_inicial_cobrado).
func (r *RPSRequest) UnmarshalJSON(data []byte) error {
	type Alias RPSRequest
	aux := struct {
		*Alias
		ValorInicialCobrado float64 `json:"valor_inicial_cobrado"`
		ValorFinalCobrado   float64 `json:"valor_final_cobrado"`
		ValorServicos       float64 `json:"valor_servicos"`
		NumeroDPS           int     `json:"numero_dps"`
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if r.NumeroRPS == 0 && aux.NumeroDPS != 0 {
		r.NumeroRPS = aux.NumeroDPS
	}

	if r.ValorServ == 0 {
		if aux.ValorInicialCobrado > 0 {
			r.ValorServ = aux.ValorInicialCobrado
		} else if aux.ValorServicos > 0 {
			r.ValorServ = aux.ValorServicos
		} else if aux.ValorFinalCobrado > 0 {
			r.ValorServ = aux.ValorFinalCobrado
		}
	}

	return nil
}

// GetMunicipioPrestacao retorna o código IBGE do município onde o serviço foi prestado.
// Para serviços tributados em São Paulo ('T') ou exportação ('P'), retorna vazio (regra 1223).
func (r *RPSRequest) GetMunicipioPrestacao() string {
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
	}
	return ""
}

// Validate valida campos essenciais do Leiaute 1.
func (r *RPSRequest) Validate() error {
	if r.Remetente == "" {
		return fmt.Errorf("remetente obrigatorio")
	}
	if r.Tomador == "" {
		return fmt.Errorf("tomador obrigatorio")
	}
	if r.ValorServ <= 0 {
		return fmt.Errorf("valor do servico (valor_serv) obrigatorio")
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

type InformacoesLote struct {
	NumeroLote          string  `xml:"NumeroLote" json:"numero_lote"`
	InscricaoPrestador  string  `xml:"InscricaoPrestador" json:"inscricao_prestador"`
	CPFCNPJRemetente    CPFCNPJ `xml:"CPFCNPJRemetente" json:"cpfcnpj_remetente"`
	DataEnvioLote       string  `xml:"DataEnvioLote" json:"data_envio_lote"`
	QtdNotasProcessadas int     `xml:"QtdNotasProcessadas" json:"qtd_notas_processadas"`
	TempoProcessamento  int     `xml:"TempoProcessamento" json:"tempo_processamento"`
	ValorTotalServicos  float64 `xml:"ValorTotalServicos" json:"valor_total_servicos"`
}

type CabecalhoLote struct {
	Sucesso         bool            `xml:"Sucesso" json:"sucesso"`
	InformacoesLote InformacoesLote `xml:"InformacoesLote" json:"informacoes_lote"`
}

type Alerta struct {
	Codigo          int      `xml:"Codigo" json:"codigo"`
	Descricao       string   `xml:"Descricao" json:"descricao"`
	NumeroRPS       int      `xml:"-" json:"numero_rps"`
	ChaveRPS        ChaveRPS `xml:"ChaveRPS" json:"chave_rps"`
	DirectNumeroRPS int      `xml:"NumeroRPS" json:"-"`
}

type ChaveNFeRPS struct {
	ChaveNFe ChaveNFe `xml:"ChaveNFe" json:"chave_nfe"`
	ChaveRPS ChaveRPS `xml:"ChaveRPS" json:"chave_rps"`
}

type ErroRetorno struct {
	Codigo          string   `xml:"Codigo" json:"codigo"`
	Descricao       string   `xml:"Descricao" json:"descricao"`
	NumeroRPS       int      `xml:"-" json:"numero_rps"`
	ChaveRPS        ChaveRPS `xml:"ChaveRPS" json:"chave_rps"`
	DirectNumeroRPS int      `xml:"NumeroRPS" json:"-"`
}

type RetornoEnvioLoteRPS struct {
	XMLName      struct{}      `xml:"RetornoEnvioLoteRPS"`
	Cabecalho    CabecalhoLote `xml:"Cabecalho" json:"cabecalho"`
	Alertas      []Alerta      `xml:"Alerta" json:"alertas,omitempty"`
	Erros        []ErroRetorno `xml:"Erro" json:"erros,omitempty"`
	ChavesNFeRPS []ChaveNFeRPS `xml:"ChaveNFeRPS" json:"chaves_nfe_rps,omitempty"`
}

type CabecalhoRPS struct {
	Versao  string `xml:"Versao,attr" json:"versao,omitempty"`
	Sucesso bool   `xml:"Sucesso" json:"sucesso"`
}

type RetornoEnvioRPS struct {
	XMLName     struct{}      `xml:"RetornoEnvioRPS" json:"-"`
	Cabecalho   CabecalhoRPS  `xml:"Cabecalho" json:"cabecalho"`
	Alertas     []Alerta      `xml:"Alerta" json:"alertas,omitempty"`
	Erros       []ErroRetorno `xml:"Erro" json:"erros,omitempty"`
	ChaveNFeRPS *ChaveNFeRPS  `xml:"ChaveNFeRPS" json:"chave_nfe_rps,omitempty"`
}

type ErroAgrupado struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	RPS       []int  `json:"rps"`
}

