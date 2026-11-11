package nfe

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"nfe/config"
	"nfe/internal/security"
	"time"
)

// GzipCompressBase64 compacta os bytes com Gzip e codifica em Base64.
func GzipCompressBase64(data []byte) (string, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return "", fmt.Errorf("erro ao compactar com gzip: %v", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("erro ao fechar gzip writer: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GzipDecompressBase64 decodifica Base64 e descompacta Gzip.
func GzipDecompressBase64(b64Str string) ([]byte, error) {
	if b64Str == "" {
		return nil, nil
	}
	rawBytes, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar base64: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(rawBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir gzip reader: %v", err)
	}
	defer zr.Close()

	decompressed, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler dados descompactados: %v", err)
	}
	return decompressed, nil
}

// BuildDPSXML converte um DPSRequest no struct oficial DPSXML pronto para serialização.
func BuildDPSXML(req DPSRequest, tpAmb int) DPSXML {
	req.Sanitize()
	dpsID := req.GenerateDPSID()

	// Converte para o fuso horário de Brasília (-03:00) conforme exigido no TDateTimeUTC
	loc, err := time.LoadLocation("America/Sao_Paulo")
	dtEmissao := req.DtEmissao
	if err == nil {
		dtEmissao = dtEmissao.In(loc)
	} else {
		dtEmissao = dtEmissao.In(time.FixedZone("BRT", -3*3600))
	}

	opSimp := req.OpSimplesNacional
	if opSimp == 0 {
		opSimp = OpSimplesOptanteMEEPP
	}

	var regApTribSN *int
	if opSimp == OpSimplesOptanteMEEPP {
		val := 1 // 1 - Apuração pelo Simples Nacional (guia DAS)
		regApTribSN = &val
	}

	versao := config.Env.SEFIN_DPS_VERSAO
	if versao == "" {
		versao = DPSVersao
	}

	dps := DPSXML{
		Versao: versao,
		InfDPS: InfDPSXML{
			Id:       dpsID,
			TpAmb:    tpAmb,
			DhEmi:    dtEmissao.Format("2006-01-02T15:04:05-07:00"),
			VerAplic: "1.0.0",
			Serie:    req.SerieDPS,
			NDPS:     req.NumeroDPS,
			DCompet:  req.Competencia,
			TpEmit:   1, // 1 - Prestador
			CLocEmi:  req.CLocEmi,
			Prest: PrestadorXML{
				IM: req.IM,
				RegTrib: RegTribXML{
					OpSimpNac:   opSimp,
					RegApTribSN: regApTribSN,
					RegEspTrib:  req.RegEspTrib,
				},
			},
			Toma: TomadorXML{
				XNome: req.TomadorNome,
				Fone:  req.TomadorTelefone,
				Email: req.TomadorEmail,
			},
			Serv: ServicoXML{
				LocPrest: LocPrestXML{
					CLocPrestacao: req.MunicipioPrestacao,
				},
				CServ: CServXML{
					CTribNac:  req.CodServico,
					CTribMun:  req.CTribMun,
					XDescServ: req.Discriminacao,
				},
			},
			Valores: ValoresXML{
				VServPrest: VServPrestXML{
					VServ: fmt.Sprintf("%.2f", req.ValorServ),
				},
				Trib: TribXML{
					TribMun: TribMunXML{
						TribISSQN:  req.TribISSQN,
						TpRetISSQN: RetISSQNNaoRetido,
					},
				},
			},
		},
	}

	// Prestador CPF vs CNPJ
	if len(req.Remetente) == 11 {
		dps.InfDPS.Prest.CPF = req.Remetente
	} else {
		dps.InfDPS.Prest.CNPJ = req.Remetente
	}

	// Tomador CPF vs CNPJ
	if len(req.Tomador) == 11 {
		dps.InfDPS.Toma.CPF = req.Tomador
	} else if len(req.Tomador) == 14 {
		dps.InfDPS.Toma.CNPJ = req.Tomador
	}

	// Endereço do Tomador (conforme XSD TEndereco: endNac com cMun e CEP, e logradouro diretamente em end)
	if req.TomadorCidade != "" || req.TomadorCEP != "" || req.TomadorLogradouro != "" {
		dps.InfDPS.Toma.End = &EnderecoXML{
			EndNac: EndNacXML{
				CMun: req.TomadorCidade,
				CEP:  req.TomadorCEP,
			},
			XLgr:    req.TomadorLogradouro,
			Nro:     req.TomadorNumero,
			XCpl:    req.TomadorComplemento,
			XBairro: req.TomadorBairro,
		}
	}

	// Descontos e Deduções (conforme XSD TCInfoValores)
	if req.DescontoIncondicionado > 0 || req.DescontoCondicionado > 0 {
		desc := &VDescCondIncondXML{}
		if req.DescontoIncondicionado > 0 {
			desc.VDescIncond = fmt.Sprintf("%.2f", req.DescontoIncondicionado)
		}
		if req.DescontoCondicionado > 0 {
			desc.VDescCond = fmt.Sprintf("%.2f", req.DescontoCondicionado)
		}
		dps.InfDPS.Valores.VDescCondIncond = desc
	}
	if req.ValorDeducoes > 0 {
		dps.InfDPS.Valores.VDedRed = &VDedRedXML{
			VDed: fmt.Sprintf("%.2f", req.ValorDeducoes),
		}
	}

	// Alíquota e Retenção do ISS
	if req.ISS {
		dps.InfDPS.Valores.Trib.TribMun.TpRetISSQN = RetISSQNRetido // 2 = Retido pelo Tomador
		if req.Aliquota > 0 {
			dps.InfDPS.Valores.Trib.TribMun.PAliq = fmt.Sprintf("%.2f", req.Aliquota)
			baseCalculo := req.ValorServ - req.ValorDeducoes
			if baseCalculo > 0 {
				vISS := baseCalculo * (req.Aliquota / 100.0)
				dps.InfDPS.Valores.Trib.TribMun.VISSQN = fmt.Sprintf("%.2f", vISS)
			}
		}
	} else {
		dps.InfDPS.Valores.Trib.TribMun.TpRetISSQN = RetISSQNNaoRetido // 1 = Não Retido
	}

	// TotTrib (Obrigatório em trib conforme XSD TCTribTotal)
	if opSimp == OpSimplesOptanteMEEPP {
		pTot := req.PTotTribSN
		if pTot <= 0 {
			pTot = req.Aliquota
		}
		dps.InfDPS.Valores.Trib.TotTrib = TotTribXML{
			PTotTribSN: fmt.Sprintf("%.2f", pTot),
		}
	} else {
		valZero := 0
		dps.InfDPS.Valores.Trib.TotTrib = TotTribXML{
			IndTotTrib: &valZero,
		}
	}

	return dps
}

// EmissaoDPS_Sefin realiza o ciclo completo de emissão da DPS na API Sefin Nacional.
func EmissaoDPS_Sefin(req DPSRequest) (*SefinNFSeResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Determina o ambiente a partir da config
	tpAmb := TpAmbHomologacao
	if config.Env.SEFIN_AMBIENTE == "1" {
		tpAmb = TpAmbProducao
	}

	dpsObj := BuildDPSXML(req, tpAmb)
	dpsXMLBytes, err := xml.Marshal(dpsObj)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar XML da DPS: %v", err)
	}

	dpsXMLStr := `<?xml version="1.0" encoding="UTF-8"?>` + string(dpsXMLBytes)

	// Carrega certificado digital ICP-Brasil e chaves
	tlsCert, certPEM, keyPEM, err := security.TLSCert()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar certificado TLS: %v", err)
	}

	// Assina digitalmente o nó infDPS da DPS via XMLDSig SHA-256
	signedXML, err := security.SignDPS(dpsXMLStr, dpsObj.InfDPS.Id, keyPEM, certPEM)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar DPS: %v", err)
	}

	url := config.SefinURL()
	log.Printf("[SEFIN] Enviando DPS Id=%s para %s (Ambiente %d, Versao %s)", dpsObj.InfDPS.Id, url, tpAmb, dpsObj.Versao)
	log.Printf("[SEFIN] XML da DPS Assinada:\n%s", signedXML)

	// Compacta GZip e codifica Base64
	dpsXmlGZipB64, err := GzipCompressBase64([]byte(signedXML))
	if err != nil {
		return nil, fmt.Errorf("erro ao compactar DPS para GZip Base64: %v", err)
	}

	payload := SefinNFSeRequest{
		DPSXmlGZipB64: dpsXmlGZipB64,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload JSON para Sefin: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		Renegotiation:      tls.RenegotiateFreelyAsClient,
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP para Sefin: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("erro na comunicação com Sefin Nacional (%s): %v", url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da Sefin: %v", err)
	}

	log.Printf("[SEFIN] Resposta Sefin (HTTP %d):\n%s", resp.StatusCode, string(respBody))

	var sefinResp SefinNFSeResponse
	if err := json.Unmarshal(respBody, &sefinResp); err != nil {
		return nil, fmt.Errorf("resposta inesperada da Sefin (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		sefinResp.Sucesso = true
	} else {
		sefinResp.Sucesso = false
	}

	// Descompacta o XML da NFS-e se retornado pela Sefin
	if sefinResp.NFSeXmlGZipB64 != "" {
		if decompressed, err := GzipDecompressBase64(sefinResp.NFSeXmlGZipB64); err == nil {
			sefinResp.NFSeXML = string(decompressed)
		}
	}

	if !sefinResp.Sucesso && len(sefinResp.Erros) > 0 {
		return &sefinResp, fmt.Errorf("rejeição Sefin (%s): %s", sefinResp.Erros[0].Codigo, sefinResp.Erros[0].Descricao)
	}

	return &sefinResp, nil
}
