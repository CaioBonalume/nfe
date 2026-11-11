package nfe

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

func TestGzipCompressAndDecompress(t *testing.T) {
	original := "Texto de teste para compressão GZip e codificação Base64 no padrão nacional da NFS-e"

	compressed, err := GzipCompressBase64([]byte(original))
	if err != nil {
		t.Fatalf("GzipCompressBase64 falhou: %v", err)
	}
	if compressed == "" {
		t.Fatalf("esperado string comprimida não vazia")
	}

	decompressed, err := GzipDecompressBase64(compressed)
	if err != nil {
		t.Fatalf("GzipDecompressBase64 falhou: %v", err)
	}

	if string(decompressed) != original {
		t.Errorf("esperado '%s', obteve '%s'", original, string(decompressed))
	}
}

func TestBuildDPSXML(t *testing.T) {
	jsonContent, err := os.ReadFile("../../exemplo_JSON/emissao_dps_nacional.json")
	if err != nil {
		t.Fatalf("erro ao ler arquivo json: %v", err)
	}

	var req DPSRequest
	if err := json.Unmarshal(jsonContent, &req); err != nil {
		t.Fatalf("erro unmarshal json: %v", err)
	}

	dpsObj := BuildDPSXML(req, TpAmbHomologacao)

	b, _ := xml.MarshalIndent(dpsObj, "", "  ")
	t.Logf("XML GERADO COM EXEMPLO_JSON:\n%s", string(b))

	if dpsObj.InfDPS.TpAmb != TpAmbHomologacao {
		t.Errorf("esperado TpAmb=%d, obteve %d", TpAmbHomologacao, dpsObj.InfDPS.TpAmb)
	}
	if dpsObj.InfDPS.Prest.CNPJ != "43629976000130" {
		t.Errorf("esperado CNPJ do prestador correto, obteve: %s", dpsObj.InfDPS.Prest.CNPJ)
	}
	if dpsObj.InfDPS.Prest.RegTrib.OpSimpNac != OpSimplesOptanteMEEPP {
		t.Errorf("esperado OpSimpNac=3, obteve %d", dpsObj.InfDPS.Prest.RegTrib.OpSimpNac)
	}
	if dpsObj.InfDPS.Serv.CServ.CTribNac != "010701" {
		t.Errorf("esperado cTribNac '010701', obteve: %s", dpsObj.InfDPS.Serv.CServ.CTribNac)
	}
	if dpsObj.InfDPS.Valores.Trib.TribMun.TpRetISSQN != RetISSQNNaoRetido {
		t.Errorf("esperado TpRetISSQN=1 (não retido), obteve %d", dpsObj.InfDPS.Valores.Trib.TribMun.TpRetISSQN)
	}
	if !strings.HasPrefix(dpsObj.InfDPS.Id, "DPS3550308243629976000130") {
		t.Errorf("esperado prefixo de Id correto, obteve: %s", dpsObj.InfDPS.Id)
	}
}

