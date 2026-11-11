package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("falha ao gerar chave RSA de teste: %v", err)
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func TestSignRPS_V1_vs_V2(t *testing.T) {
	keyPEM := generateTestKey(t)

	// Dados reais fornecidos pelo usuário
	ccm := "17329248"
	serie := "NFBON"
	numeroRPS := "2067"
	dataEmissao := "2026-11-06"
	tributacao := "T"
	status := "N"
	valorStr := "3054500" // R$ 30.545,00 * 100
	deducoesStr := "0"
	codServico := "07498"
	cnpjTomador := "58801036000154"
	issRetido := false

	sigV1, err := SignRPS(ccm, serie, numeroRPS, dataEmissao, tributacao, status, valorStr, deducoesStr, codServico, cnpjTomador, issRetido, keyPEM)
	if err != nil {
		t.Fatalf("SignRPS V1 falhou: %v", err)
	}
	if sigV1 == "" {
		t.Errorf("esperado assinatura V1 nao vazia")
	}

	sigV2, err := SignRPS_V2(ccm, serie, numeroRPS, dataEmissao, tributacao, status, valorStr, deducoesStr, codServico, cnpjTomador, issRetido, keyPEM)
	if err != nil {
		t.Fatalf("SignRPS V2 falhou: %v", err)
	}
	if sigV2 == "" {
		t.Errorf("esperado assinatura V2 nao vazia")
	}

	// As assinaturas V1 e V2 devem ser diferentes porque V1 usa CCM de 8 digitos (string de 86 chars)
	// e V2 usa CCM de 12 digitos (string de 90 chars, conforme exigido pela Prefeitura de SP no Leiaute 2).
	if sigV1 == sigV2 {
		t.Errorf("esperado que assinaturas V1 e V2 sejam diferentes devido ao tamanho do CCM")
	}
}

func TestSignDPS(t *testing.T) {
	keyPEM := generateTestKey(t)
	dummyCert := []byte("-----BEGIN CERTIFICATE-----\nMIIBojCCAUqgAwIBAgII...\n-----END CERTIFICATE-----")

	dpsXML := `<DPS xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.01"><infDPS Id="DPS355030824362997600013000001000000000000105"><tpAmb>2</tpAmb></infDPS></DPS>`
	dpsID := "DPS355030824362997600013000001000000000000105"

	signedXML, err := SignDPS(dpsXML, dpsID, keyPEM, dummyCert)
	if err != nil {
		t.Fatalf("SignDPS falhou: %v", err)
	}

	if !strings.HasPrefix(signedXML, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("esperado prologo XML UTF-8 no inicio do XML assinado")
	}
	if !strings.Contains(signedXML, "<Signature xmlns=\"http://www.w3.org/2000/09/xmldsig#\">") {
		t.Errorf("esperado Signature presente no XML assinado")
	}
	if !strings.Contains(signedXML, "URI=\"#DPS355030824362997600013000001000000000000105\"") {
		t.Errorf("esperado URI com Id da DPS no XML assinado")
	}
	if !strings.Contains(signedXML, "Algorithm=\"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256\"") {
		t.Errorf("esperado rsa-sha256 no XML assinado")
	}

	// Teste com XML ja contendo o prologo
	signedXML2, err := SignDPS(signedXML, dpsID, keyPEM, dummyCert)
	if err != nil {
		t.Fatalf("SignDPS com prologo falhou: %v", err)
	}
	if strings.Count(signedXML2, "<?xml") != 1 {
		t.Errorf("prologo XML duplicado detectado")
	}
}


