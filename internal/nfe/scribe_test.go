package nfe

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("falha ao gerar chave rsa de teste: %v", err)
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func TestWriteRPS_V1(t *testing.T) {
	keyPEM := generateTestKey(t)

	req := RPSRequest{
		Remetente:     "43629976000130",
		IM:            70718911,
		SerieRPS:      "NFBON",
		NumeroRPS:     101,
		DtEmissao:     time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC),
		Tributacao:    "T",
		ValorServ:     1250.50,
		ValorDeducoes: 50.00,
		PIS:           10.00,
		COFINS:        20.00,
		INSS:          0.00,
		IR:            0.00,
		CSLL:          0.00,
		CodServico:    "02800",
		Aliquota:      0.02,
		ISS:           false,
		Tomador:       "54487608000111",
		Discriminacao: "Servico teste V1",
	}

	xml := writeRPS(req, keyPEM)

	if !strings.Contains(xml, `Cabecalho Versao="1"`) {
		t.Errorf("esperado Cabecalho Versao=\"1\", obteve:\n%s", xml)
	}
	if !strings.Contains(xml, "<ValorServicos>1250.50</ValorServicos>") {
		t.Errorf("esperado <ValorServicos>1250.50</ValorServicos>, obteve:\n%s", xml)
	}
	if !strings.Contains(xml, "<Assinatura>") {
		t.Errorf("esperado tag <Assinatura> presente")
	}
	if strings.Contains(xml, "<IBSCBS>") {
		t.Errorf("nao esperado <IBSCBS> no Leiaute 1")
	}
	if strings.Contains(xml, "<ValorInicialCobrado>") {
		t.Errorf("nao esperado <ValorInicialCobrado> no Leiaute 1")
	}
}

func TestWriteRPSV2(t *testing.T) {
	keyPEM := generateTestKey(t)

	req := RPSRequestV2{
		Remetente:             "43629976000130",
		IM:                    70718911,
		SerieRPS:              "NFBON",
		NumeroRPS:             102,
		DtEmissao:             time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC),
		Tributacao:            "T",
		ValorInicialCobrado:   1500.00,
		ValorDeducoes:         0.00,
		ValorTotalRecebido:    1500.00,
		ValorIPI:              0.00,
		ExigibilidadeSuspensa: 0,
		PagamentoParcelado:    0,
		NBS:                   "101010100",
		CLocPrestacao:         "3550308",
		FinNFSe:               0,
		IndFinal:              0,
		CIndOp:                "000001",
		IndDest:               0,
		CClassTrib:            "000001",
		CodServico:            "02800",
		Aliquota:              0.02,
		ISS:                   false,
		Tomador:               "54487608000111",
		Discriminacao:         "Servico teste V2",
	}

	xml := writeRPSV2(req, keyPEM)

	if !strings.Contains(xml, `Cabecalho Versao="2"`) {
		t.Errorf("esperado Cabecalho Versao=\"2\", obteve:\n%s", xml)
	}
	if !strings.Contains(xml, "<ValorInicialCobrado>1500.00</ValorInicialCobrado>") {
		t.Errorf("esperado <ValorInicialCobrado>1500.00</ValorInicialCobrado>, obteve:\n%s", xml)
	}
	if strings.Contains(xml, "<ValorServicos>") {
		t.Errorf("nao esperado <ValorServicos> no Leiaute 2")
	}
	if !strings.Contains(xml, "<ValorIPI>0.00</ValorIPI>") {
		t.Errorf("esperado <ValorIPI>0.00</ValorIPI>")
	}
	if !strings.Contains(xml, "<ExigibilidadeSuspensa>0</ExigibilidadeSuspensa>") {
		t.Errorf("esperado <ExigibilidadeSuspensa>0</ExigibilidadeSuspensa>")
	}
	if !strings.Contains(xml, "<NBS>101010100</NBS>") {
		t.Errorf("esperado <NBS>101010100</NBS>")
	}
	if !strings.Contains(xml, "<cLocPrestacao>3550308</cLocPrestacao>") {
		t.Errorf("esperado <cLocPrestacao>3550308</cLocPrestacao>")
	}
	if !strings.Contains(xml, "<IBSCBS>") || !strings.Contains(xml, "</IBSCBS>") {
		t.Errorf("esperado bloco <IBSCBS>")
	}
	if !strings.Contains(xml, "<cIndOp>000001</cIndOp>") {
		t.Errorf("esperado <cIndOp>000001</cIndOp>")
	}
	if !strings.Contains(xml, "<cClassTrib>000001</cClassTrib>") {
		t.Errorf("esperado <cClassTrib>000001</cClassTrib>")
	}
	if !strings.Contains(xml, "<Assinatura>") {
		t.Errorf("esperado <Assinatura> no RPS V2")
	}
}

func TestWriteRPSV2_OnlyRPS(t *testing.T) {
	keyPEM := generateTestKey(t)

	req := RPSRequestV2{
		Remetente:           "43629976000130",
		IM:                  70718911,
		SerieRPS:            "NFBON",
		NumeroRPS:           103,
		DtEmissao:           time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC),
		Tributacao:          "T",
		ValorInicialCobrado: 2000.00,
		NBS:                 "101010100",
		CIndOp:              "000001",
		CClassTrib:          "000001",
		CodServico:          "02800",
		Tomador:             "07766570890",
		Discriminacao:       "RPS em bloco de lote",
	}

	xml := writeRPSV2(req, keyPEM, true)

	if strings.Contains(xml, "<PedidoEnvioRPS") {
		t.Errorf("nao esperado PedidoEnvioRPS quando onlyRPS=true")
	}
	if !strings.HasPrefix(strings.TrimSpace(xml), "<RPS") || !strings.HasSuffix(strings.TrimSpace(xml), "</RPS>") {
		t.Errorf("esperado bloco iniciando com <RPS e finalizando com </RPS>, obteve:\n%s", xml)
	}
}

func TestLoteRPS_V2_HeaderXML(t *testing.T) {
	keyPEM := generateTestKey(t)

	requests := []RPSRequestV2{
		{
			Remetente:           "43629976000130",
			IM:                  70718911,
			SerieRPS:            "NFBON",
			NumeroRPS:           201,
			DtEmissao:           time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC),
			Tributacao:          "T",
			ValorInicialCobrado: 1000.00,
			NBS:                 "101010100",
			CIndOp:              "000001",
			CClassTrib:          "000001",
			CodServico:          "02800",
			Tomador:             "54487608000111",
			Discriminacao:       "Lote item 1",
		},
		{
			Remetente:           "43629976000130",
			IM:                  70718911,
			SerieRPS:            "NFBON",
			NumeroRPS:           202,
			DtEmissao:           time.Date(2026, 11, 6, 11, 0, 0, 0, time.UTC),
			Tributacao:          "T",
			ValorInicialCobrado: 1500.00,
			NBS:                 "101010100",
			CIndOp:              "000001",
			CClassTrib:          "000001",
			CodServico:          "02800",
			Tomador:             "07766570890",
			Discriminacao:       "Lote item 2",
		},
	}

	xmlMSG, err := readSchemaFile("../assets/schemas/xml/envio_lote_rps_request_v2.xml")
	if err != nil {
		t.Fatalf("falha ao ler envio_lote_rps_request_v2.xml: %v", err)
	}
	xmlMSGStr := string(xmlMSG)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{REMETENTE}", requests[0].Remetente, 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DT_INICIO}", "2026-11-06", 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{DT_FIM}", "2026-11-06", 1)
	xmlMSGStr = strings.Replace(xmlMSGStr, "{QTD_RPS}", "2", 1)

	rpsBlocks := ""
	for _, r := range requests {
		rpsBlocks += writeRPSV2(r, keyPEM, true)
	}
	xmlMSGStr = strings.Replace(xmlMSGStr, "{RPS_LIST}", rpsBlocks, 1)

	if !strings.Contains(xmlMSGStr, `Cabecalho Versao="2"`) {
		t.Errorf("esperado Cabecalho Versao=\"2\"")
	}
	if strings.Contains(xmlMSGStr, "ValorTotalServicos") {
		t.Errorf("nao esperado ValorTotalServicos no cabecalho do lote V2")
	}
	if strings.Contains(xmlMSGStr, "ValorTotalDeducoes") {
		t.Errorf("nao esperado ValorTotalDeducoes no cabecalho do lote V2")
	}
	if strings.Count(xmlMSGStr, "<RPS>") != 2 {
		t.Errorf("esperado exatamente 2 blocos <RPS>, obteve %d", strings.Count(xmlMSGStr, "<RPS>"))
	}
}

func TestRPSRequestV2_Validate(t *testing.T) {
	req := RPSRequestV2{
		Remetente:           "43629976000130",
		Tomador:             "54487608000111",
		ValorInicialCobrado: 100.0,
		NBS:                 "101010100",
		CIndOp:              "000001",
		CClassTrib:          "000001",
	}

	if err := req.Validate(); err != nil {
		t.Errorf("esperado valido, erro: %v", err)
	}

	// Sem NBS
	invalidReq := req
	invalidReq.NBS = ""
	if err := invalidReq.Validate(); err == nil {
		t.Errorf("esperado erro quando NBS vazio")
	}

	// NBS tamanho incorreto
	invalidReq.NBS = "123"
	if err := invalidReq.Validate(); err == nil {
		t.Errorf("esperado erro quando NBS diferente de 9 digitos")
	}

	// Sem c_ind_op
	invalidReq = req
	invalidReq.CIndOp = ""
	if err := invalidReq.Validate(); err == nil {
		t.Errorf("esperado erro quando CIndOp vazio")
	}

	// Sem c_class_trib
	invalidReq = req
	invalidReq.CClassTrib = ""
	if err := invalidReq.Validate(); err == nil {
		t.Errorf("esperado erro quando CClassTrib vazio")
	}

	// Sem valor
	invalidReq = req
	invalidReq.ValorInicialCobrado = 0
	invalidReq.ValorServ = 0
	if err := invalidReq.Validate(); err == nil {
		t.Errorf("esperado erro quando valor cobrado <= 0")
	}
}
