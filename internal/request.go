package internal

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func xmlCleanup(xmlPath string) string {
	xmlContent, err := os.ReadFile(xmlPath)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo: %v", err)
	}
	xmlFinal := string(xmlContent)
	xmlFinal = strings.ReplaceAll(xmlFinal, "\n", "")
	xmlFinal = strings.ReplaceAll(xmlFinal, "\t", "")
	xmlFinal = strings.ReplaceAll(xmlFinal, "\r", "")
	// regex que remove espaço entre tags
	re := regexp.MustCompile(`>\s+<`)
	xmlFinal = re.ReplaceAllString(xmlFinal, "><")
	xmlFinal = strings.TrimSpace(xmlFinal)
	return xmlFinal
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

func ConsultaCNPJ_V1(cnpj string) (status string, body string) {
	var tlsCert, certPEM, keyPEM, _ = TLSCert()
	// --- MONTAGEM DO XML V1 ---
	msg := `<p1:PedidoConsultaCNPJ xmlns:p1="http://www.prefeitura.sp.gov.br/nfe"><Cabecalho Versao="1"><CPFCNPJRemetente><CNPJ>{CNPJ}</CNPJ></CPFCNPJRemetente></Cabecalho><CNPJContribuinte><CNPJ>{CNPJ}</CNPJ></CNPJContribuinte></p1:PedidoConsultaCNPJ>`
	msg = strings.ReplaceAll(msg, "{CNPJ}", cnpj)

	// GARANTIA EXTRA: Remove qualquer espaço em branco ou quebra de linha acidental
	msg = strings.ReplaceAll(msg, "\n", "")
	msg = strings.ReplaceAll(msg, "\r", "")
	msg = strings.ReplaceAll(msg, "\t", "")
	msg = strings.TrimSpace(msg)

	// 2. Gera APENAS o bloco da assinatura com base nessa string exata
	signatureBlock, err := SignBlock(msg, keyPEM, certPEM)
	if err != nil {
		log.Fatalf("Erro ao assinar: %v", err)
	}

	// 3. Injeta a assinatura manualmente (Concatenando strings)
	// Isso evita que o parser XML altere a estrutura do 'msg' original
	msgSigned := strings.Replace(msg, "</p1:PedidoConsultaCNPJ>", signatureBlock+"</p1:PedidoConsultaCNPJ>", 1)

	soapEnvelope := `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	  <soap:Body>
	    <ConsultaCNPJRequest xmlns="http://www.prefeitura.sp.gov.br/nfe">
	      <VersaoSchema>1</VersaoSchema>
	      <MensagemXML><![CDATA[` + msgSigned + `]]></MensagemXML>
	    </ConsultaCNPJRequest>
	  </soap:Body>
	</soap:Envelope>`

	// --- ENVIO DA REQUISIÇÃO ---
	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL"
	soapAction := "http://www.prefeitura.sp.gov.br/nfe/ws/ConsultaCNPJ"

	resp, err := newRequest(url, soapEnvelope, tlsCert, 1, soapAction)
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
	return resp.Status, respBody.String()
}
