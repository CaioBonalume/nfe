package nfe

import "time"

type ConsultaNFERequest struct {
	CNPJ_REMETENTE string    // 14 dígitos
	IM             string    // V.1 8 dígitos / V.2 12 dígitos
	CNPJ           string    // 14 dígitos
	DTInicio       time.Time // Formato YYYY-MM-DD
	DTFim          time.Time // Formato YYYY-MM-DD
	Pagina         int
}

type CancelarNFeRequest struct {
	Remetente         string `json:"remetente"`           // CPF/CNPJ
	Inscricao         string `json:"inscricao_prestador"` // CCM (8 dígitos)
	NumeroNFe         int    `json:"numero_nfe"`          // (12 dígitos)
	CodigoVerificacao string `json:"codigo_verificacao"`  // (8 chars)
}
type Endereco struct {
	TipoLogradouro string `xml:"TipoLogradouro" json:"tipo_logradouro"`
	Logradouro     string `xml:"Logradouro" json:"logradouro"`
	NumeroEndereco string `xml:"NumeroEndereco" json:"numero_endereco"`
	Bairro         string `xml:"Bairro,omitempty" json:"bairro,omitempty"`
	Cidade         string `xml:"Cidade" json:"cidade"`
	UF             string `xml:"UF" json:"uf"`
	CEP            string `xml:"CEP" json:"cep"`
}

type CPFCNPJ struct {
	CNPJ string `xml:"CNPJ,omitempty" json:"cnpj,omitempty"`
	CPF  string `xml:"CPF,omitempty" json:"cpf,omitempty"`
}

type ChaveNFe struct {
	InscricaoPrestador string `xml:"InscricaoPrestador" json:"inscricao_prestador"`
	NumeroNFe          int    `xml:"NumeroNFe" json:"numero_nfe"`
	CodigoVerificacao  string `xml:"CodigoVerificacao" json:"codigo_verificacao"`
	ChaveNotaNacional  string `xml:"ChaveNotaNacional" json:"chave_nota_nacional"`
}

type ChaveRPS struct {
	InscricaoPrestador string `xml:"InscricaoPrestador" json:"inscricao_prestador"`
	SerieRPS           string `xml:"SerieRPS" json:"serie_rps"`
	NumeroRPS          int    `xml:"NumeroRPS" json:"numero_rps"`
}

type ConsultaNFEResponse struct {
	XMLName                   struct{} `xml:"NFe"`
	StatusOperacao            bool     `xml:"Sucesso" json:"status_operacao"`
	Assinatura                string   `xml:"Assinatura" json:"assinatura"`
	ChaveNFe                  ChaveNFe `xml:"ChaveNFe" json:"chave_nfe"`
	DataEmissaoNFe            string   `xml:"DataEmissaoNFe" json:"data_emissao_nfe"`
	DataFatoGeradorNFe        string   `xml:"DataFatoGeradorNFe" json:"data_fato_gerador_nfe"`
	ChaveRPS                  ChaveRPS `xml:"ChaveRPS" json:"chave_rps"`
	TipoRPS                   string   `xml:"TipoRPS" json:"tipo_rps"`
	DataEmissaoRPS            string   `xml:"DataEmissaoRPS" json:"data_emissao_rps"`
	CPFCNPJPrestador          CPFCNPJ  `xml:"CPFCNPJPrestador" json:"cpfcnpj_prestador"`
	RazaoSocialPrestador      string   `xml:"RazaoSocialPrestador" json:"razao_social_prestador"`
	InscricaoMunicipalTomador string   `xml:"InscricaoMunicipalTomador,omitempty" json:"inscricao_municipal_tomador,omitempty"`
	EnderecoPrestador         Endereco `xml:"EnderecoPrestador" json:"endereco_prestador"`
	StatusNFe                 string   `xml:"StatusNFe" json:"status_nfe"`
	DataCancelamento          string   `xml:"DataCancelamento,omitempty" json:"data_cancelamento,omitempty"`
	TributacaoNFe             string   `xml:"TributacaoNFe" json:"tributacao_nfe"`
	OpcaoSimples              string   `xml:"OpcaoSimples" json:"opcao_simples"`
	ValorServicos             float64  `xml:"ValorServicos" json:"valor_servicos"`
	CodigoServico             string   `xml:"CodigoServico" json:"codigo_servico"`
	AliquotaServicos          float64  `xml:"AliquotaServicos" json:"aliquota_servicos"`
	ValorISS                  float64  `xml:"ValorISS" json:"valor_iss"`
	ValorCredito              float64  `xml:"ValorCredito" json:"valor_credito"`
	ISSRetido                 bool     `xml:"ISSRetido" json:"iss_retido"`
	CPFCNPJTomador            CPFCNPJ  `xml:"CPFCNPJTomador" json:"cpfcnpj_tomador"`
	RazaoSocialTomador        string   `xml:"RazaoSocialTomador" json:"razao_social_tomador"`
	EnderecoTomador           Endereco `xml:"EnderecoTomador" json:"endereco_tomador"`
	Discriminacao             string   `xml:"Discriminacao" json:"discriminacao"`
}

type RetornoConsulta struct {
	NFes []ConsultaNFEResponse `xml:"NFe"`
}
