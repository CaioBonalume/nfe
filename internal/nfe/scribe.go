package nfe

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"nfe/internal/security"
	"os"
	"strconv"
	"strings"
	"time"
)

func xmlCleanup(content string) string {
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	content = strings.ReplaceAll(content, "\t", " ")
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	content = strings.ReplaceAll(content, "> <", "><")
	return strings.TrimSpace(content)
}

func fixResponseXML(content string) string {
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	return content
}

func newRequest(url, xmlFinal string, cert tls.Certificate, soap int, soapAction ...string) (*http.Response, error) {
	// 1. Configurar o Cliente HTTP Seguro
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Baixe a cadeia de certificados da ICP-Brasil (Raiz V5 e V10) no site do ITI.
		InsecureSkipVerify: true, // Apenas descomente se tiver erro de verificação da cadeia da SEFAZ (não recomendado)
		Renegotiation:      tls.RenegotiateOnceAsClient,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 240 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(xmlFinal))
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	// Headers obrigatórios para SOAP
	if soap == 1 {
		// SOAP 1.1
		fmt.Printf("SOAP 1.1\n")

		content := fmt.Sprintf("text/xml; charset=utf-8; action=%s", soapAction[0])
		req.Header.Set("Content-Type", content)
	} else {
		// SOAP 1.2
		fmt.Printf("SOAP 1.2\n")
		content := fmt.Sprintf("application/soap+xml; charset=utf-8; action=%s", soapAction[0])
		req.Header.Set("Content-Type", content)
	}

	// Enviar e Receber
	fmt.Println("Enviando requisição...")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Erro na conexão (verifique certificado/internet): %v", err)
	}
	return resp, nil
}

func ConsultaCNPJ_V1(remetente, tomador string) (status, body string) {
	xmlEnvelope, err := os.ReadFile("../assets/schemas/xml/requestModel_1.1.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlEnvelopeStr := string(xmlEnvelope)
	xmlEnvelopeStr = strings.ReplaceAll(xmlEnvelopeStr, "WRAPPER", "ConsultaCNPJRequest")

	// CARTA ESCRITA
	xmlMSG, err := os.ReadFile("../assets/schemas/xml/consulta_CNPJ_request.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlMSGStr := string(xmlMSG)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{REMETENTE}", remetente, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{TOMADOR}", tomador, 1)
	xmlMSGStr = xmlCleanup(xmlMSGStr)

	// CARTA ASSINADA
	var tlsCert, certPEM, keyPEM, _ = security.TLSCert()
	signatureBlock, _ := security.SignBlock(xmlMSGStr, keyPEM, certPEM)
	xmlMSGStr = strings.Replace(xmlMSGStr, "</p1:PedidoConsultaCNPJ>", signatureBlock+"</p1:PedidoConsultaCNPJ>", 1)

	// CARTA ENVELOPADA
	mail := strings.Replace(xmlEnvelopeStr, "{MSG}", xmlMSGStr, 1)
	mail = xmlCleanup(mail)

	// fmt.Println(xml)
	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL"
	soapAction := "http://www.prefeitura.sp.gov.br/nfe/ws/ConsultaCNPJ"

	resp, err := newRequest(url, mail, tlsCert, 1, soapAction)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	// 4. Ler o corpo da resposta
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}
	defer resp.Body.Close()
	respStr := fixResponseXML(respBody.String())

	return resp.Status, respStr
}

func ConsultaNFePeriodo(request ConsultaNFERequest) (status, body string) {
	xmlEnvelope, err := os.ReadFile("../assets/schemas/xml/requestModel_1.1.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlEnvelopeStr := string(xmlEnvelope)

	wrapper := "ConsultaNFeEmitidasRequest" // futuramente fazer um if para ConsultaNFeRecebidasRequest
	xmlEnvelopeStr = strings.ReplaceAll(xmlEnvelopeStr, "WRAPPER", wrapper)

	// CARTA ESCRITA
	xmlMSG, err := os.ReadFile("../assets/schemas/xml/consulta_nfe_request.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlMSGStr := string(xmlMSG)

	xmlMSGStr = strings.Replace(xmlMSGStr, "{CNPJ_REMETENTE}", request.CNPJ_REMETENTE, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{CNPJ}", request.CNPJ, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{IE}", request.IE, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DT_INICIO}", request.DTInicio.Format("2006-01-02"), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DT_FIM}", request.DTFim.Format("2006-01-02"), 1)
	if request.Pagina < 1 {
		xmlMSGStr = strings.Replace(xmlMSGStr, "{NUMERO_PAGINA}", "1", 1)
	} else {
		xmlMSGStr = strings.Replace(xmlMSGStr, "{NUMERO_PAGINA}", fmt.Sprintf("%d", request.Pagina), 1)
	}
	xmlMSGStr = xmlCleanup(xmlMSGStr)

	// CARTA ASSINADA
	var tlsCert, certPEM, keyPEM, _ = security.TLSCert()
	signatureBlock, _ := security.SignBlock(xmlMSGStr, keyPEM, certPEM)
	xmlMSGStr = strings.Replace(xmlMSGStr, "</p1:PedidoConsultaNFePeriodo>", signatureBlock+"</p1:PedidoConsultaNFePeriodo>", 1)

	// CARTA ENVELOPADA
	mail := strings.Replace(xmlEnvelopeStr, "{MSG}", xmlMSGStr, 1)
	mail = xmlCleanup(mail)

	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL"
	// EMITIDAS
	soapAction := "http://www.prefeitura.sp.gov.br/nfe/ws/ConsultaNFeEmitidas"
	// RECEBIDAS
	// soapAction := "http://www.prefeitura.sp.gov.br/nfe/ConsultaNFeRecebidas"

	resp, err := newRequest(url, mail, tlsCert, 1, soapAction)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	// 4. Ler o corpo da resposta
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}
	defer resp.Body.Close()
	respStr := fixResponseXML(respBody.String())

	return resp.Status, respStr
}

func EmissaoRPS_V1(request RPSRequest) (status, body string) {
	xmlEnvelope, err := os.ReadFile("../assets/schemas/xml/requestModel_1.1.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlEnvelopeStr := string(xmlEnvelope)

	wrapper := "EnvioRPSRequest"
	xmlEnvelopeStr = strings.ReplaceAll(xmlEnvelopeStr, "WRAPPER", wrapper)

	var tlsCert, certPEM, keyPEM, _ = security.TLSCert()

	// CARTA ESCRITA
	xmlMSGStr := writeRPS(request, keyPEM)

	// CARTA ASSINADA
	signatureBlock, _ := security.SignBlock(xmlMSGStr, keyPEM, certPEM)
	xmlMSGStr = strings.Replace(xmlMSGStr, "</p1:PedidoEnvioRPS>", signatureBlock+"</p1:PedidoEnvioRPS>", 1)

	// CARTA ENVELOPADA
	mail := strings.Replace(xmlEnvelopeStr, "{MSG}", xmlMSGStr, 1)
	mail = xmlCleanup(mail)

	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL"
	soapAction := "http://www.prefeitura.sp.gov.br/nfe/ws/EnvioRPS"

	resp, err := newRequest(url, mail, tlsCert, 1, soapAction)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	// 4. Ler o corpo da resposta
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}
	defer resp.Body.Close()
	respStr := fixResponseXML(respBody.String())

	return resp.Status, respStr
}

func EmissaoLoteRPS_V1(requests []RPSRequest) (status, body string, err error) {
	xmlEnvelope, err := os.ReadFile("../assets/schemas/xml/requestModel_1.1.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlEnvelopeStr := string(xmlEnvelope)
	wrapper := "EnvioLoteRPSRequest"
	xmlEnvelopeStr = strings.ReplaceAll(xmlEnvelopeStr, "WRAPPER", wrapper)

	var tlsCert, certPEM, keyPEM, _ = security.TLSCert()

	// CARTA ESCRITA
	xmlMSG, err := os.ReadFile("../assets/schemas/xml/envio_lote_rps_request.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlMSGStr := string(xmlMSG)

	xmlMSGStr = strings.Replace(xmlMSGStr, "{REMETENTE}", requests[0].Remetente, 1)
	min := requests[0].DtEmissao
	max := requests[0].DtEmissao
	for _, r := range requests[1:] {
		if r.DtEmissao.Before(min) {
			min = r.DtEmissao
		}
		if r.DtEmissao.After(max) {
			max = r.DtEmissao
		}
	}
	dtIniStr := min.Format("2006-01-02")
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DT_INICIO}", dtIniStr, 1)
	dtFimStr := max.Format("2006-01-02")
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DT_FIM}", dtFimStr, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{QTD_RPS}", strconv.Itoa(len(requests)), 1)
	valorTotal := 0.0
	valorDeducoes := 0.0
	for _, r := range requests {
		valorTotal += r.ValorServ
		valorDeducoes += r.ValorDeducoes
	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_TOTAL}", fmt.Sprintf("%.2f", valorTotal), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_DEDUCOES}", fmt.Sprintf("%.2f", valorDeducoes), 1)

	countComNumero := 0
	for _, r := range requests {
		if r.NumeroRPS != 0 {
			countComNumero++
		}
	}

	if countComNumero > 0 && countComNumero < len(requests) {
		return "", "", fmt.Errorf("lote invalido: mistura de RPS com e sem NumeroRPS. Forneca todos ou nenhum para auto-preenchimento")
	}

	if countComNumero == 0 {
		lastNum, err := lastRPSNumber(requests[0])
		if err != nil {
			return "", "", fmt.Errorf("erro ao consultar ultimo RPS: %v", err)
		}
		for i := range requests {
			requests[i].NumeroRPS = lastNum + i + 1
		}
	}

	rpsBlocks := ""
	for _, r := range requests {
		rpsBlock := writeRPS(r, keyPEM, true)
		rpsBlocks += rpsBlock
	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{RPS_LIST}", rpsBlocks, 1)
	xmlMSGStr = xmlCleanup(xmlMSGStr)

	// CARTA ASSINADA
	signatureBlock, _ := security.SignBlock(xmlMSGStr, keyPEM, certPEM)
	xmlMSGStr = strings.Replace(xmlMSGStr, "</p1:PedidoEnvioLoteRPS>", signatureBlock+"</p1:PedidoEnvioLoteRPS>", 1)

	// CARTA ENVELOPADA
	mail := strings.Replace(xmlEnvelopeStr, "{MSG}", xmlMSGStr, 1)
	mail = xmlCleanup(mail)

	// VERIFICAR DAQUI PRA FRENTE <-------------------------------------------
	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL"
	soapAction := "http://www.prefeitura.sp.gov.br/nfe/ws/EnvioLoteRPS"

	resp, err := newRequest(url, mail, tlsCert, 1, soapAction)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	// 4. Ler o corpo da resposta
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}
	defer resp.Body.Close()
	respStr := fixResponseXML(respBody.String())

	return resp.Status, respStr, nil
}

func lastRPSNumber(request RPSRequest) (int, error) {
	valStr := ""
	consultRequest := ConsultaNFERequest{
		CNPJ_REMETENTE: request.Remetente,
		CNPJ:           request.Remetente,
		IE:             strconv.Itoa(request.IE),
		DTInicio:       time.Now().AddDate(0, 0, -30),
		DTFim:          time.Now(),
		Pagina:         1,
	}

	_, consultBody := ConsultaNFePeriodo(consultRequest)
	i := strings.LastIndex(consultBody, "<NumeroRPS>")
	if i != -1 {
		j := strings.Index(consultBody[i+len("<NumeroRPS>"):], "</NumeroRPS>")
		if j != -1 {
			valStr = consultBody[i+len("<NumeroRPS>") : i+len("<NumeroRPS>")+j]
		}
	} else {
		return 0, fmt.Errorf("Não foi encontrado ultimo RPS emitido")
	}

	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("Não foi encontrado ultimo RPS emitido")
	}
	return valInt, nil
}

func writeRPS(request RPSRequest, keyPEM []byte, onlyRPSOpt ...bool) string {
	// CARTA ESCRITA
	xmlMSG, err := os.ReadFile("../assets/schemas/xml/envio_rps_request.xml")
	if err != nil {
		log.Fatalf("Erro ao ler envio_rps_request.xml: %v", err)
	}
	xmlMSGStr := string(xmlMSG)
	onlyRPS := false
	if len(onlyRPSOpt) > 0 {
		onlyRPS = onlyRPSOpt[0]
	}

	if !onlyRPS {
		xmlMSGStr = strings.Replace(xmlMSGStr, "{REMETENTE}", request.Remetente, 1)
	} else {
		// mantém apenas o bloco <RPS>...</RPS> (fallback: mantém original se não encontrar)
		start := strings.Index(xmlMSGStr, "<RPS")
		if start != -1 {
			endRel := strings.Index(xmlMSGStr[start:], "</RPS>")
			if endRel != -1 {
				xmlMSGStr = xmlMSGStr[start : start+endRel+len("</RPS>")]
			}
		}
	}

	// NUMERO RPS
	if request.NumeroRPS < 1 {
		numRps, err := lastRPSNumber(request)
		if err != nil {
			log.Fatalf("%v", err)
		}
		request.NumeroRPS = numRps + 1

	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{NUMERO_RPS}", strconv.Itoa(request.NumeroRPS), 1)

	xmlMSGStr = strings.Replace(xmlMSGStr, "{DATA_EMISSAO}", request.DtEmissao.Format("2006-01-02"), 1)

	if request.Tributacao != "" {
		xmlMSGStr = strings.Replace(xmlMSGStr, "{TRIBUTACAO_RPS}", request.Tributacao, 1)
	} else {
		request.Tributacao = "T"
	}

	IE := strconv.Itoa(request.IE)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{IE}", IE, 1)

	if request.SerieRPS == "" {
		request.SerieRPS = "NFBON"
	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{SERIE_RPS}", request.SerieRPS, 1)

	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_SERVICO}", fmt.Sprintf("%.2f", request.ValorServ), 1)

	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_DEDUCOES}", fmt.Sprintf("%.2f", request.ValorDeducoes), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_PIS}", fmt.Sprintf("%.2f", request.PIS), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_COFINS}", fmt.Sprintf("%.2f", request.COFINS), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_INSS}", fmt.Sprintf("%.2f", request.INSS), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_IR}", fmt.Sprintf("%.2f", request.IR), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{VALOR_CSLL}", fmt.Sprintf("%.2f", request.CSLL), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{ALIQUOTA}", fmt.Sprintf("%.2f", request.Aliquota), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{COD_SERVICO}", request.CodServico, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{ISS}", fmt.Sprintf("%t", request.ISS), 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{TOMADOR}", request.Tomador, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DISCRIMINACAO}", request.Discriminacao, 1)
	// ASSINATURA RPS
	hashRPS, _ := security.SignRPS(
		strconv.Itoa(request.IE),
		request.SerieRPS,
		strconv.Itoa(request.NumeroRPS),
		request.DtEmissao.Format("2006-01-02"),
		request.Tributacao,
		"N",
		fmt.Sprintf("%d", int(request.ValorServ*100)),
		fmt.Sprintf("%d", int(request.ValorDeducoes*100)),
		request.CodServico,
		request.Tomador,
		request.ISS,
		keyPEM,
	)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{HASH_RPS}", hashRPS, 1)
	xmlMSGStr = xmlCleanup(xmlMSGStr)
	return xmlMSGStr
}

func CancelarNFe(request CancelarNFeRequest) (status, body string) {
	return CancelarLoteNFe([]CancelarNFeRequest{request})
}

func CancelarLoteNFe(requests []CancelarNFeRequest) (status, body string) {
	xmlEnvelope, err := os.ReadFile("../assets/schemas/xml/requestModel_1.1.xml")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo XML: %v", err)
	}
	xmlEnvelopeStr := string(xmlEnvelope)
	wrapper := "CancelamentoNFeRequest"
	xmlEnvelopeStr = strings.ReplaceAll(xmlEnvelopeStr, "WRAPPER", wrapper)

	var tlsCert, certPEM, keyPEM, _ = security.TLSCert()

	// CARTA ESCRITA
	xmlMSG, err := os.ReadFile("../assets/schemas/xml/cancelamento_nfe_request.xml")
	if err != nil {
		log.Fatalf("Erro ao ler cancelamento_nfe_request.xml: %v", err)
	}
	xmlMSGStr := string(xmlMSG)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{REMETENTE}", requests[0].Remetente, 1)

	detalhes := ""
	for _, r := range requests {
		detalhes += writeCancelamentoDetalhe(r, keyPEM)
	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DETALHE_LIST}", detalhes, 1)
	xmlMSGStr = xmlCleanup(xmlMSGStr)

	// CARTA ASSINADA
	signatureBlock, _ := security.SignBlock(xmlMSGStr, keyPEM, certPEM)
	xmlMSGStr = strings.Replace(xmlMSGStr, "</p1:PedidoCancelamentoNFe>", signatureBlock+"</p1:PedidoCancelamentoNFe>", 1)

	// CARTA ENVELOPADA
	mail := strings.Replace(xmlEnvelopeStr, "{MSG}", xmlMSGStr, 1)
	mail = xmlCleanup(mail)

	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL"
	soapAction := "http://www.prefeitura.sp.gov.br/nfe/ws/CancelamentoNFe"

	resp, err := newRequest(url, mail, tlsCert, 1, soapAction)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	// 4. Ler o corpo da resposta
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler corpo da resposta: %v", err)
	}
	defer resp.Body.Close()
	respStr := fixResponseXML(respBody.String())

	return resp.Status, respStr
}

func writeCancelamentoDetalhe(request CancelarNFeRequest, keyPEM []byte) string {
	xmlMSG, err := os.ReadFile("../assets/schemas/xml/cancelamento_nfe_detalhe.xml")
	if err != nil {
		log.Fatalf("Erro ao ler cancelamento_nfe_detalhe.xml: %v", err)
	}
	xmlMSGStr := string(xmlMSG)

	inscricao := request.Inscricao
	numeroNFe := strconv.Itoa(request.NumeroNFe)

	xmlMSGStr = strings.Replace(xmlMSGStr, "{INSCRICAO_PRESTADOR}", inscricao, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{NUMERO_NFE}", numeroNFe, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{CODIGO_VERIFICACAO}", request.CodigoVerificacao, 1)

	// ASSINATURA CANCELAMENTO
	hashCancelamento, err := security.SignCancelamento(inscricao, numeroNFe, keyPEM)
	if err != nil {
		log.Fatalf("Erro ao assinar cancelamento: %v", err)
	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{ASSINATURA_CANCELAMENTO}", hashCancelamento, 1)

	return xmlMSGStr
}
