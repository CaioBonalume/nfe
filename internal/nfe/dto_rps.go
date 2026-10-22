package nfe

import "time"

type RPSRequest struct {
	Remetente     string
	Tomador       string
	TomadorCidade       string // opcional (código IBGE 7 dígitos)
	TomadorUF           string // opcional (sigla estado 2 letras)
	TomadorTipoLogradouro string // opcional (Rua, Av, etc)
	TomadorLogradouro   string // opcional
	TomadorNumero       string // opcional
	TomadorComplemento  string // opcional
	TomadorBairro       string // opcional
	TomadorCEP          string // opcional (8 dígitos)
	IM                  int
	SerieRPS      string // opcional, use "" para padrão
	NumeroRPS     int
	DtEmissao     time.Time
	Tributacao    string // T - Tributado em SP | F - Fora de SP | A - Isento | B - Tributado fora de SP e Isento | M - Imune | N - Serviço não listado no ISS | X - Tributado em SP com Exigibilidade Suspensa | P - Exportação de serviços
	ValorServ     float64
	ValorDeducoes float64 // opcional, use 0 para padrão
	PIS           float64 // opcional, use 0 para padrão
	COFINS        float64 // opcional, use 0 para padrão
	INSS          float64 // opcional, use 0 para padrão
	IR            float64 // opcional, use 0 para padrão
	CSLL          float64 // opcional, use 0 para padrão
	CodServico    string  // 5 dígitos
	Aliquota      float64
	ISS           bool
	Discriminacao string
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

type ErroAgrupado struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	RPS       []int  `json:"rps"`
}

