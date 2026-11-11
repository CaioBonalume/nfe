package nfe

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDPSRequest_ValidateAndSanitize(t *testing.T) {
	req := DPSRequest{
		Remetente:     "43.629.976/0001-30",
		Tomador:       "54.487.608/0001-11",
		NumeroDPS:     101,
		CodServico:    "01.07.01",
		ValorServ:     1500.50,
		Aliquota:      2.0,
		ISS:           false,
		Discriminacao: "Serviço de desenvolvimento de software",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("esperado validação com sucesso, obteve: %v", err)
	}

	if req.Remetente != "43629976000130" {
		t.Errorf("esperado remetente sanitizado, obteve: %s", req.Remetente)
	}
	if req.Tomador != "54487608000111" {
		t.Errorf("esperado tomador sanitizado, obteve: %s", req.Tomador)
	}
	if req.CodServico != "010701" {
		t.Errorf("esperado cod_servico sanitizado, obteve: %s", req.CodServico)
	}
	if req.OpSimplesNacional != OpSimplesOptanteMEEPP {
		t.Errorf("esperado default OpSimplesOptanteMEEPP (2), obteve: %d", req.OpSimplesNacional)
	}
	if req.CLocEmi != "3550308" {
		t.Errorf("esperado default CLocEmi '3550308', obteve: %s", req.CLocEmi)
	}
}

func TestDPSRequest_GenerateDPSID(t *testing.T) {
	req := DPSRequest{
		CLocEmi:   "3550308",
		Remetente: "43629976000130",
		SerieDPS:  "1",
		NumeroDPS: 105,
	}

	id := req.GenerateDPSID()
	// Id deve ter 45 caracteres:
	// "DPS" (3) + "3550308" (7) + "2" (1) + "43629976000130" (14) + "00001" (5) + "000000000000105" (15) = 45
	if len(id) != 45 {
		t.Fatalf("esperado Id com 45 caracteres, obteve %d (%s)", len(id), id)
	}

	expectedPrefix := "DPS355030824362997600013000001000000000000105"
	if id != expectedPrefix {
		t.Errorf("esperado %s, obteve %s", expectedPrefix, id)
	}
}

func TestDPSRequest_UnmarshalJSON_Aliases(t *testing.T) {
	jsonInput := `{
		"prestador_cnpj": "43629976000130",
		"tomador": "54487608000111",
		"razao_social_tomador": "Empresa Tomadora LTDA",
		"numero_rps": 55,
		"serie_rps": "NFBON",
		"valor_servicos": 2500.00,
		"c_trib_nac": "010701",
		"c_loc_prestacao": "3550308",
		"aliquota": 2.0,
		"iss_retido": true,
		"discriminacao": "Consultoria técnica"
	}`

	var req DPSRequest
	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		t.Fatalf("erro ao deserializar JSON com aliases: %v", err)
	}

	if req.Remetente != "43629976000130" {
		t.Errorf("esperado Remetente preenchido via prestador_cnpj, obteve: %s", req.Remetente)
	}
	if req.TomadorNome != "Empresa Tomadora LTDA" {
		t.Errorf("esperado TomadorNome preenchido via razao_social_tomador, obteve: %s", req.TomadorNome)
	}
	if req.NumeroDPS != 55 {
		t.Errorf("esperado NumeroDPS=55 via numero_rps, obteve: %d", req.NumeroDPS)
	}
	if req.SerieDPS != "NFBON" {
		t.Errorf("esperado SerieDPS=NFBON via serie_rps, obteve: %s", req.SerieDPS)
	}
	if req.ValorServ != 2500.00 {
		t.Errorf("esperado ValorServ=2500.00 via valor_servicos, obteve: %.2f", req.ValorServ)
	}
	if req.CodServico != "010701" {
		t.Errorf("esperado CodServico=010701 via c_trib_nac, obteve: %s", req.CodServico)
	}
	if !req.ISS {
		t.Errorf("esperado ISS=true via iss_retido, obteve false")
	}
}

func TestDPSXML_Marshal(t *testing.T) {
	dps := DPSXML{
		Versao: "1.01",
		InfDPS: InfDPSXML{
			Id:       "DPS355030824362997600013000001000000000000105",
			TpAmb:    2,
			DhEmi:    time.Now().Format("2006-01-02T15:04:05-07:00"),
			VerAplic: "1.0.0",
			Serie:    "1",
			NDPS:     105,
			DCompet:  "2026-11-10",
			TpEmit:   1,
			CLocEmi:  "3550308",
			Prest: PrestadorXML{
				CNPJ: "43629976000130",
				RegTrib: RegTribXML{
					OpSimpNac:  3,
					RegEspTrib: 0,
				},
			},
			Toma: TomadorXML{
				CNPJ:  "54487608000111",
				XNome: "Tomador Teste LTDA",
			},
			Serv: ServicoXML{
				LocPrest: LocPrestXML{CLocPrestacao: "3550308"},
				CServ: CServXML{
					CTribNac:  "010701",
					XDescServ: "Desenvolvimento de software",
				},
			},
			Valores: ValoresXML{
				VServPrest: VServPrestXML{
					VServ: "1500.00",
				},
				Trib: TribXML{
					TribMun: TribMunXML{
						TribISSQN:  1,
						TpRetISSQN: 2,
						PAliq:      "2.00",
					},
					TotTrib: TotTribXML{
						PTotTribSN: "2.00",
					},
				},
			},
		},
	}

	data, err := xml.MarshalIndent(dps, "", "  ")
	if err != nil {
		t.Fatalf("erro ao serializar XML da DPS: %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, `<DPS xmlns="http://www.sped.fazenda.gov.br/nfse"`) {
		t.Errorf("esperado namespace da DPS no XML, obteve:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, `<opSimpNac>3</opSimpNac>`) {
		t.Errorf("esperado <opSimpNac>3</opSimpNac> no XML, obteve:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, `<cTribNac>010701</cTribNac>`) {
		t.Errorf("esperado <cTribNac>010701</cTribNac> no XML, obteve:\n%s", xmlStr)
	}
}

func TestDPSXML_Exemplo(t *testing.T) {
	data, err := os.ReadFile("../../exemplo_JSON/emissao_dps_nacional.json")
	if err != nil {
		t.Fatalf("erro ao ler exemplo: %v", err)
	}

	var req DPSRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("erro unmarshal: %v", err)
	}

	dps := BuildDPSXML(req, 2)
	xmlBytes, err := xml.MarshalIndent(dps, "", "  ")
	if err != nil {
		t.Fatalf("erro marshal: %v", err)
	}

	t.Logf("XML GERADO:\n%s", string(xmlBytes))
}

func TestDPSRequest_SerieComLetraEMunicipalCodServico(t *testing.T) {
	req := DPSRequest{
		Remetente:     "59789981000140",
		Tomador:       "54325709000196",
		SerieDPS:      "A", // Letra que causava o erro TSIdDPS
		NumeroDPS:     2196,
		ValorServ:     7779.71,
		CodServico:    "7498", // Código municipal SP 4 dígitos
		Aliquota:      0.02,   // 0.02 enviado como fração decimal
		ISS:           true,   // Retido pelo tomador
		Discriminacao: "Instalação e manutenção",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("erro inesperado na validação: %v", err)
	}

	if req.SerieDPS != "1" {
		t.Errorf("esperado serie normalizada para '1', obteve: %s", req.SerieDPS)
	}
	if req.CodServico != "140101" {
		t.Errorf("esperado cTribNac mapeado para '140101', obteve: %s", req.CodServico)
	}
	if req.CTribMun != "07498" {
		t.Errorf("esperado cTribMun '07498', obteve: %s", req.CTribMun)
	}
	if req.Aliquota != 2.0 {
		t.Errorf("esperado aliquota convertida para 2.0, obteve: %.2f", req.Aliquota)
	}

	id := req.GenerateDPSID()
	expectedID := "DPS355030825978998100014000001000000000002196"
	if id != expectedID {
		t.Errorf("esperado Id '%s', obteve: '%s'", expectedID, id)
	}

	dps := BuildDPSXML(req, 2)
	xmlBytes, err := xml.MarshalIndent(dps, "", "  ")
	if err != nil {
		t.Fatalf("erro ao serializar XML: %v", err)
	}
	xmlStr := string(xmlBytes)

	// Id sem letras
	if strings.Contains(xmlStr, "0000A") {
		t.Errorf("Id contem letra 'A' indevida no XML:\n%s", xmlStr)
	}
	// serie 1
	if !strings.Contains(xmlStr, "<serie>1</serie>") {
		t.Errorf("esperado <serie>1</serie> no XML, obteve:\n%s", xmlStr)
	}
	// cTribNac 140101
	if !strings.Contains(xmlStr, "<cTribNac>140101</cTribNac>") {
		t.Errorf("esperado <cTribNac>140101</cTribNac> no XML, obteve:\n%s", xmlStr)
	}
	// tpRetISSQN 2 e pAliq 2.00
	if !strings.Contains(xmlStr, "<tpRetISSQN>2</tpRetISSQN>") || !strings.Contains(xmlStr, "<pAliq>2.00</pAliq>") {
		t.Errorf("esperado tpRetISSQN 2 e pAliq 2.00 no XML, obteve:\n%s", xmlStr)
	}
	// vISSQN 155.59 (7779.71 * 0.02)
	if !strings.Contains(xmlStr, "<vISSQN>155.59</vISSQN>") {
		t.Errorf("esperado vISSQN 155.59 no XML, obteve:\n%s", xmlStr)
	}
}



