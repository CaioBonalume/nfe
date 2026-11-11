package nfe

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var nonDigitsRegex = regexp.MustCompile(`\D`)

// DPSRequest representa os dados recebidos via JSON para emissão pelo Padrão Nacional (Sefin Nacional).
type DPSRequest struct {
	// Identificação e Prestador
	Remetente             string    `json:"remetente"`                        // CNPJ do Prestador (14 dígitos)
	IM                    string    `json:"im,omitempty"`                    // Inscrição Municipal do Prestador (opcional)
	CLocEmi               string    `json:"c_loc_emi,omitempty"`             // Código IBGE do município emissor (default: "3550308")
	SerieDPS              string    `json:"serie_dps,omitempty"`             // Série da DPS (default: "1")
	NumeroDPS             int       `json:"numero_dps,omitempty"`            // Número da DPS (se 0, pode ser auto-incrementado)
	DtEmissao             time.Time `json:"dt_emissao"`                      // Data e hora de emissão da DPS
	Competencia           string    `json:"competencia,omitempty"`           // Data de competência AAAA-MM-DD (se vazio, usa DtEmissao)
	OpSimplesNacional     int       `json:"op_simples_nacional,omitempty"`   // 1 - Não optante, 2 - Optante MEI, 3 - Optante ME/EPP (default)
	RegEspTrib            int       `json:"reg_esp_trib,omitempty"`          // 0 - Nenhum (default)

	// Tomador
	Tomador               string `json:"tomador"`                          // CPF (11) ou CNPJ (14)
	TomadorNome           string `json:"tomador_nome,omitempty"`           // Razão Social ou Nome
	TomadorCidade         string `json:"tomador_cidade,omitempty"`         // IBGE 7 dígitos
	TomadorUF             string `json:"tomador_uf,omitempty"`             // Sigla UF 2 letras
	TomadorLogradouro     string `json:"tomador_logradouro,omitempty"`
	TomadorNumero         string `json:"tomador_numero,omitempty"`
	TomadorComplemento    string `json:"tomador_complemento,omitempty"`
	TomadorBairro         string `json:"tomador_bairro,omitempty"`
	TomadorCEP            string `json:"tomador_cep,omitempty"`            // 8 dígitos
	TomadorEmail          string `json:"tomador_email,omitempty"`
	TomadorTelefone       string `json:"tomador_telefone,omitempty"`

	// Serviço
	CodServico            string `json:"cod_servico"`                      // Código de Tributação Nacional (cTribNac - 6 dígitos)
	CTribMun              string `json:"c_trib_mun,omitempty"`             // Código de Tributação Municipal (opcional)
	MunicipioPrestacao    string `json:"municipio_prestacao,omitempty"`    // IBGE 7 dígitos (se vazio, usa CLocEmi)
	Discriminacao         string `json:"discriminacao"`                    // Discriminação dos serviços prestados

	// Valores e ISS
	ValorServ             float64 `json:"valor_serv"`                        // Valor total do serviço
	ValorDeducoes         float64 `json:"valor_deducoes,omitempty"`          // Deduções (opcional)
	DescontoIncondicionado float64 `json:"desconto_incondicionado,omitempty"` // Desconto incondicionado (opcional)
	DescontoCondicionado   float64 `json:"desconto_condicionado,omitempty"`   // Desconto condicionado (opcional)
	Aliquota              float64 `json:"aliquota"`                          // Alíquota percentual do ISS (ex: 2.00 para 2%)
	PTotTribSN            float64 `json:"p_tot_trib_sn,omitempty"`           // Percentual aproximado tributos Simples Nacional (se 0, usa Aliquota)
	ISS                   bool    `json:"iss"`                               // false = Não retido (tpRetISSQN=1), true = Retido pelo Tomador (tpRetISSQN=2)
	TribISSQN             int     `json:"trib_issqn,omitempty"`              // 1 - Tributável (default), 2 - Imunidade, 3 - Exportação, 4 - Não Incidência

	// Tributos Federais (opcionais para optantes do Simples Nacional)
	PIS                   float64 `json:"pis,omitempty"`
	COFINS                float64 `json:"cofins,omitempty"`
	INSS                  float64 `json:"inss,omitempty"`
	IR                    float64 `json:"ir,omitempty"`
	CSLL                  float64 `json:"csll,omitempty"`
}

// UnmarshalJSON oferece suporte a aliases comuns vindos de sistemas existentes (ex: RPSRequest).
func (r *DPSRequest) UnmarshalJSON(data []byte) error {
	type Alias DPSRequest
	aux := struct {
		*Alias
		// Aliases
		PrestadorCNPJ       string  `json:"prestador_cnpj"`
		NumeroRPS           int     `json:"numero_rps"`
		SerieRPS            string  `json:"serie_rps"`
		ValorServicos       float64 `json:"valor_servicos"`
		ValorInicialCobrado float64 `json:"valor_inicial_cobrado"`
		ValorFinalCobrado   float64 `json:"valor_final_cobrado"`
		CTribNac            string  `json:"c_trib_nac"`
		CLocPrestacao       string  `json:"c_loc_prestacao"`
		RazaoSocialTomador  string  `json:"razao_social_tomador"`
		ISSRetido           *bool   `json:"iss_retido"`
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if r.Remetente == "" && aux.PrestadorCNPJ != "" {
		r.Remetente = aux.PrestadorCNPJ
	}
	if r.NumeroDPS == 0 && aux.NumeroRPS != 0 {
		r.NumeroDPS = aux.NumeroRPS
	}
	if r.SerieDPS == "" && aux.SerieRPS != "" {
		r.SerieDPS = aux.SerieRPS
	}
	if r.ValorServ == 0 {
		if aux.ValorServicos > 0 {
			r.ValorServ = aux.ValorServicos
		} else if aux.ValorInicialCobrado > 0 {
			r.ValorServ = aux.ValorInicialCobrado
		} else if aux.ValorFinalCobrado > 0 {
			r.ValorServ = aux.ValorFinalCobrado
		}
	}
	if r.CodServico == "" && aux.CTribNac != "" {
		r.CodServico = aux.CTribNac
	}
	if r.MunicipioPrestacao == "" && aux.CLocPrestacao != "" {
		r.MunicipioPrestacao = aux.CLocPrestacao
	}
	if r.TomadorNome == "" && aux.RazaoSocialTomador != "" {
		r.TomadorNome = aux.RazaoSocialTomador
	}
	if aux.ISSRetido != nil {
		r.ISS = *aux.ISSRetido
	}

	return nil
}

// Sanitize aplica defaults e formatações limpas aos campos.
func (r *DPSRequest) Sanitize() {
	r.Remetente = nonDigitsRegex.ReplaceAllString(r.Remetente, "")
	r.Tomador = nonDigitsRegex.ReplaceAllString(r.Tomador, "")
	r.TomadorCEP = nonDigitsRegex.ReplaceAllString(r.TomadorCEP, "")
	r.CodServico = nonDigitsRegex.ReplaceAllString(r.CodServico, "")

	// Série da DPS no padrão nacional deve ser numérica de 1 a 5 dígitos (tipo TSSerieDPS).
	// Séries legadas com letras (ex: "A", "NFBON") devem ser convertidas/normalizadas para numérico.
	digitsOnlySerie := nonDigitsRegex.ReplaceAllString(r.SerieDPS, "")
	if sNum, err := strconv.Atoi(digitsOnlySerie); err == nil && sNum > 0 {
		r.SerieDPS = strconv.Itoa(sNum)
	} else {
		r.SerieDPS = "1"
	}

	// Alíquota: se enviada em formato decimal fracionário (ex: 0.02 para 2%), converte para percentual (2.00)
	if r.Aliquota > 0 && r.Aliquota < 1.0 {
		r.Aliquota = r.Aliquota * 100
	}

	// Mapeamento automático de códigos municipais de serviços conhecidos de SP para cTribNac (6 dígitos)
	switch r.CodServico {
	case "7498", "07498":
		if r.CTribMun == "" {
			r.CTribMun = "07498"
		}
		r.CodServico = "140101"
	case "2800", "02800":
		if r.CTribMun == "" {
			r.CTribMun = "02800"
		}
		r.CodServico = "010701"
	case "2798", "02798":
		if r.CTribMun == "" {
			r.CTribMun = "02798"
		}
		r.CodServico = "010501"
	}

	if r.CLocEmi == "" {
		r.CLocEmi = "3550308" // default São Paulo - SP
	}
	if r.DtEmissao.IsZero() {
		r.DtEmissao = time.Now()
	}
	if r.Competencia == "" {
		r.Competencia = r.DtEmissao.Format("2006-01-02")
	}
	if r.OpSimplesNacional == 0 {
		r.OpSimplesNacional = OpSimplesOptanteMEEPP // 3 = Simples Nacional ME/EPP
	}
	if r.TribISSQN == 0 {
		r.TribISSQN = TribISSQNOperacaoTributavel // 1 = Tributável
	}
	if r.MunicipioPrestacao == "" {
		r.MunicipioPrestacao = r.CLocEmi
	}
}

// Validate valida a consistência dos dados para o Padrão Nacional.
func (r *DPSRequest) Validate() error {
	r.Sanitize()

	if len(r.Remetente) != 14 {
		return fmt.Errorf("remetente deve ser um CNPJ válido com 14 dígitos (recebido: %s)", r.Remetente)
	}
	if r.NumeroDPS <= 0 {
		return fmt.Errorf("numero_dps (ou numero_rps) é obrigatório e deve ser maior que zero para emissão no padrão nacional Sefin")
	}
	if len(r.Tomador) != 11 && len(r.Tomador) != 14 {
		return fmt.Errorf("tomador deve ser CPF (11 dígitos) ou CNPJ (14 dígitos) (recebido: %s)", r.Tomador)
	}
	if r.ValorServ <= 0 {
		return fmt.Errorf("valor_serv deve ser maior que zero (recebido: %.2f)", r.ValorServ)
	}
	if r.CodServico == "" {
		return fmt.Errorf("cod_servico (cTribNac) é obrigatório no padrão nacional")
	}
	if len(r.CodServico) != 6 {
		return fmt.Errorf("cod_servico (cTribNac) deve conter exatamente 6 dígitos no padrão nacional (recebido: '%s')", r.CodServico)
	}
	if r.ISS && r.Aliquota <= 0 {
		return fmt.Errorf("quando o ISS é retido pelo tomador (iss: true), a aliquota deve ser informada e maior que zero")
	}
	if strings.TrimSpace(r.Discriminacao) == "" {
		return fmt.Errorf("discriminacao dos serviços é obrigatória")
	}
	if len(r.CLocEmi) != 7 {
		return fmt.Errorf("c_loc_emi deve ser o código IBGE com 7 dígitos (recebido: %s)", r.CLocEmi)
	}

	return nil
}

// GenerateDPSID gera o identificador de 45 caracteres do nó <infDPS>.
// Regra: "DPS" + cLocEmi (7 dígitos) + tpInsc (1 dígito: 1-CPF, 2-CNPJ) + Inscricao (14 dígitos com zeros à esquerda) + serie (5 dígitos) + nDPS (15 dígitos)
func (r *DPSRequest) GenerateDPSID() string {
	r.Sanitize()

	tpInsc := "2"
	insc := r.Remetente
	if len(insc) == 11 {
		tpInsc = "1"
		insc = fmt.Sprintf("%014s", insc)
	} else if len(insc) < 14 {
		insc = fmt.Sprintf("%014s", insc)
	}

	serieNum, _ := strconv.Atoi(nonDigitsRegex.ReplaceAllString(r.SerieDPS, ""))
	if serieNum <= 0 {
		serieNum = 1
	}
	serieStr := fmt.Sprintf("%05d", serieNum)
	nDPSStr := fmt.Sprintf("%015d", r.NumeroDPS)

	return fmt.Sprintf("DPS%s%s%s%s%s", r.CLocEmi, tpInsc, insc, serieStr, nDPSStr)
}

// --- Estruturas de Envio e Retorno da API Sefin Nacional (REST JSON) ---

// SefinNFSeRequest é o payload JSON enviado ao endpoint POST https://sefin.nfse.gov.br/SefinNacional/nfse.
type SefinNFSeRequest struct {
	DPSXmlGZipB64 string `json:"dpsXmlGZipB64"`
}

// SefinMensagem representa avisos, erros ou alertas retornados pela Sefin.
type SefinMensagem struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
}

// SefinNFSeResponse representa o retorno da API da Sefin Nacional.
type SefinNFSeResponse struct {
	Sucesso        bool            `json:"sucesso"`
	ChaveAcesso    string          `json:"chaveAcesso,omitempty"`
	NFSeXmlGZipB64 string          `json:"nfseXmlGZipB64,omitempty"`
	NFSeXML        string          `json:"nfseXml,omitempty"` // XML descompactado para conveniência
	Alertas        []SefinMensagem `json:"alertas,omitempty"`
	Erros          []SefinMensagem `json:"erros,omitempty"`
}

// --- Estruturas XML da DPS (Padrão Nacional NFS-e) ---

type DPSXML struct {
	XMLName xml.Name  `xml:"http://www.sped.fazenda.gov.br/nfse DPS"`
	Versao  string    `xml:"versao,attr"`
	InfDPS  InfDPSXML `xml:"infDPS"`
}

type InfDPSXML struct {
	Id       string       `xml:"Id,attr"`
	TpAmb    int          `xml:"tpAmb"`
	DhEmi    string       `xml:"dhEmi"`
	VerAplic string       `xml:"verAplic"`
	Serie    string       `xml:"serie"`
	NDPS     int          `xml:"nDPS"`
	DCompet  string       `xml:"dCompet"`
	TpEmit   int          `xml:"tpEmit"`
	CLocEmi  string       `xml:"cLocEmi"`
	Prest    PrestadorXML `xml:"prest"`
	Toma     TomadorXML   `xml:"toma"`
	Serv     ServicoXML   `xml:"serv"`
	Valores  ValoresXML   `xml:"valores"`
}

type PrestadorXML struct {
	CNPJ    string      `xml:"CNPJ,omitempty"`
	CPF     string      `xml:"CPF,omitempty"`
	IM      string      `xml:"IM,omitempty"`
	RegTrib RegTribXML  `xml:"regTrib"`
}

type RegTribXML struct {
	OpSimpNac   int  `xml:"opSimpNac"`
	RegApTribSN *int `xml:"regApTribSN,omitempty"`
	RegEspTrib  int  `xml:"regEspTrib"`
}

type TomadorXML struct {
	CNPJ  string       `xml:"CNPJ,omitempty"`
	CPF   string       `xml:"CPF,omitempty"`
	XNome string       `xml:"xNome,omitempty"`
	End   *EnderecoXML `xml:"end,omitempty"`
	Fone  string       `xml:"fone,omitempty"`
	Email string       `xml:"email,omitempty"`
}

type EnderecoXML struct {
	EndNac  EndNacXML `xml:"endNac"`
	XLgr    string    `xml:"xLgr,omitempty"`
	Nro     string    `xml:"nro,omitempty"`
	XCpl    string    `xml:"xCpl,omitempty"`
	XBairro string    `xml:"xBairro,omitempty"`
}

type EndNacXML struct {
	CMun string `xml:"cMun"`
	CEP  string `xml:"CEP"`
}

type ServicoXML struct {
	LocPrest LocPrestXML `xml:"locPrest"`
	CServ    CServXML    `xml:"cServ"`
}

type LocPrestXML struct {
	CLocPrestacao string `xml:"cLocPrestacao"`
}

type CServXML struct {
	CTribNac  string `xml:"cTribNac"`
	CTribMun  string `xml:"cTribMun,omitempty"`
	XDescServ string `xml:"xDescServ"`
}

type ValoresXML struct {
	VServPrest      VServPrestXML       `xml:"vServPrest"`
	VDescCondIncond *VDescCondIncondXML `xml:"vDescCondIncond,omitempty"`
	VDedRed         *VDedRedXML         `xml:"vDedRed,omitempty"`
	Trib            TribXML             `xml:"trib"`
}

type VServPrestXML struct {
	VServ  string `xml:"vServ"`
	VReceb string `xml:"vReceb,omitempty"`
}

type VDescCondIncondXML struct {
	VDescIncond string `xml:"vDescIncond,omitempty"`
	VDescCond   string `xml:"vDescCond,omitempty"`
}

type VDedRedXML struct {
	VDed string `xml:"vDed,omitempty"`
}

type TribXML struct {
	TribMun TribMunXML `xml:"tribMun"`
	TotTrib TotTribXML `xml:"totTrib"`
}

type TribMunXML struct {
	TribISSQN  int    `xml:"tribISSQN"`
	TpRetISSQN int    `xml:"tpRetISSQN"`
	PAliq      string `xml:"pAliq,omitempty"`
	VISSQN     string `xml:"vISSQN,omitempty"`
}

type TotTribXML struct {
	IndTotTrib *int   `xml:"indTotTrib,omitempty"`
	PTotTribSN string `xml:"pTotTribSN,omitempty"`
}

