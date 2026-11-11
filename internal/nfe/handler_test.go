package nfe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockNFEService struct {
	calledV1     bool
	calledV2     bool
	calledDPS    bool
	calledLoteV1 bool
	calledLoteV2 bool
	retV1        any
	errV1        error
	retV2        any
	errV2        error
	retDPS       *SefinNFSeResponse
	errDPS       error
}

func (m *mockNFEService) ConsultaCNPJ(remetente, tomador string) (string, error) {
	return "", nil
}

func (m *mockNFEService) ConsultaNFePeriodo(request ConsultaNFERequest) ([]ConsultaNFEResponse, error) {
	return nil, nil
}

func (m *mockNFEService) EmissaoRPS(request RPSRequest) (any, error) {
	m.calledV1 = true
	if m.errV1 != nil || m.retV1 != nil {
		return m.retV1, m.errV1
	}
	return map[string]any{"sucesso": true, "versao": 1}, nil
}

func (m *mockNFEService) EmissaoRPS_V2(request RPSRequestV2) (any, error) {
	m.calledV2 = true
	if m.errV2 != nil || m.retV2 != nil {
		return m.retV2, m.errV2
	}
	return map[string]any{"sucesso": true, "versao": 2}, nil
}

func (m *mockNFEService) EmissaoDPS(request DPSRequest) (*SefinNFSeResponse, error) {
	m.calledDPS = true
	if m.errDPS != nil || m.retDPS != nil {
		return m.retDPS, m.errDPS
	}
	return &SefinNFSeResponse{Sucesso: true, ChaveAcesso: "35503080000000000000000000000000000000000000000001"}, nil
}

func (m *mockNFEService) EmissaoLoteRPS(requests []RPSRequest) (any, error) {
	m.calledLoteV1 = true
	return map[string]any{"sucesso": true, "versao": 1}, nil
}

func (m *mockNFEService) EmissaoLoteRPS_V2(requests []RPSRequestV2) (any, error) {
	m.calledLoteV2 = true
	return map[string]any{"sucesso": true, "versao": 2}, nil
}

func (m *mockNFEService) CancelarNFe(request CancelarNFeRequest) (string, error) {
	return "", nil
}

func (m *mockNFEService) CancelarLoteNFe(requests []CancelarNFeRequest) (string, error) {
	return "", nil
}

func TestEmissaoRPS_MissingLeiaute(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := map[string]any{
		"remetente": "43629976000130",
		"tomador":   "54487608000111",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400 quando leiaute nao informado, obteve %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "leiaute não informado") {
		t.Errorf("esperado mensagem indicando ausencia de leiaute, obteve: %s", w.Body.String())
	}
}

func TestEmissaoRPS_Leiaute1(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := RPSRequest{
		Remetente:  "43629976000130",
		Tomador:    "54487608000111",
		ValorServ:  100.0,
		CodServico: "02800",
		DtEmissao:  time.Now(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps?leiaute=1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledV1 {
		t.Errorf("esperado chamada ao servico V1")
	}
}

func TestEmissaoRPS_Leiaute2(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := RPSRequestV2{
		Remetente:           "43629976000130",
		Tomador:             "54487608000111",
		ValorInicialCobrado: 150.0,
		NBS:                 "101010100",
		CIndOp:              "000001",
		CClassTrib:          "000001",
		CodServico:          "02800",
		DtEmissao:           time.Now(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps?leiaute=2", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledV2 {
		t.Errorf("esperado chamada ao servico V2")
	}
}

func TestEmissaoRPS_V2_DirectRoute_RejeicaoRetorna400ComJSON(t *testing.T) {
	mockSvc := &mockNFEService{
		retV2: &RetornoEnvioRPS{
			Cabecalho: CabecalhoRPS{Versao: "2", Sucesso: false},
			Erros: []ErroRetorno{
				{
					Codigo:    "1630",
					Descricao: "Código de serviço (7870) não permite indicação de Valor Total Recebido.",
					NumeroRPS: 741,
				},
			},
			Alertas: []Alerta{
				{
					Codigo:    1651,
					Descricao: "Aviso qualquer",
					NumeroRPS: 741,
				},
			},
		},
		errV2: fmt.Errorf("rps rejeitado"),
	}
	h := NewNFEHandler(mockSvc)

	payload := RPSRequestV2{
		Remetente:           "43629976000130",
		Tomador:             "54487608000111",
		ValorInicialCobrado: 150.0,
		NBS:                 "101010100",
		CIndOp:              "000001",
		CClassTrib:          "000001",
		CodServico:          "02800",
		DtEmissao:           time.Now(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps/v2", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS_V2(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400 quando rejeitado, obteve %d: %s", w.Code, w.Body.String())
	}

	respBody := w.Body.String()
	if !strings.Contains(respBody, `"sucesso":false`) {
		t.Errorf("esperado '\"sucesso\":false' no corpo JSON, obteve: %s", respBody)
	}
	if !strings.Contains(respBody, `"1630"`) {
		t.Errorf("esperado código de erro 1630 no corpo JSON, obteve: %s", respBody)
	}
}

func TestEmissaoRPS_Leiaute2_RejeicaoRetorna400(t *testing.T) {
	mockSvc := &mockNFEService{
		retV2: &RetornoEnvioRPS{
			Cabecalho: CabecalhoRPS{Versao: "2", Sucesso: false},
			Erros: []ErroRetorno{
				{
					Codigo:    "1630",
					Descricao: "Erro qualquer",
				},
			},
		},
		errV2: fmt.Errorf("rps rejeitado"),
	}
	h := NewNFEHandler(mockSvc)

	payload := RPSRequestV2{
		Remetente:           "43629976000130",
		Tomador:             "54487608000111",
		ValorInicialCobrado: 150.0,
		NBS:                 "101010100",
		CIndOp:              "000001",
		CClassTrib:          "000001",
		CodServico:          "02800",
		DtEmissao:           time.Now(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps?leiaute=2", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400 quando rejeitado via /emissao-rps?leiaute=2, obteve %d", w.Code)
	}
}

func TestEmissaoRPS_V2_DirectRoute(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := RPSRequestV2{
		Remetente:           "43629976000130",
		Tomador:             "54487608000111",
		ValorInicialCobrado: 150.0,
		NBS:                 "101010100",
		CIndOp:              "000001",
		CClassTrib:          "000001",
		CodServico:          "02800",
		DtEmissao:           time.Now(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps/v2", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS_V2(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledV2 {
		t.Errorf("esperado chamada ao servico V2")
	}
}

func TestEmissaoLoteRPS_MissingLeiaute(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := []map[string]any{
		{"remetente": "43629976000130", "tomador": "54487608000111"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-lote-rps", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoLoteRPS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400 quando lote nao especifica leiaute, obteve %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "leiaute do lote não informado") {
		t.Errorf("esperado mensagem indicando ausencia de leiaute do lote, obteve: %s", w.Body.String())
	}
}

func TestEmissaoLoteRPS_Leiaute1(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := []RPSRequest{
		{
			Remetente:  "43629976000130",
			Tomador:    "54487608000111",
			ValorServ:  100.0,
			CodServico: "02800",
			DtEmissao:  time.Now(),
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-lote-rps?leiaute=1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoLoteRPS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledLoteV1 {
		t.Errorf("esperado chamada a EmissaoLoteRPS (v1)")
	}
}

func TestEmissaoLoteRPS_Leiaute2(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	payload := []RPSRequestV2{
		{
			Remetente:           "43629976000130",
			Tomador:             "54487608000111",
			ValorInicialCobrado: 200.0,
			NBS:                 "101010100",
			CIndOp:              "000001",
			CClassTrib:          "000001",
			CodServico:          "02800",
			DtEmissao:           time.Now(),
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-lote-rps?leiaute=2", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoLoteRPS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledLoteV2 {
		t.Errorf("esperado chamada a EmissaoLoteRPS_V2")
	}
}

func TestEmissaoRPS_V1_DirectRoute(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	jsonPayload := `{
		"remetente": "59789981000140",
		"im": 17329248,
		"serie_rps": "NFBON",
		"numero_rps": 2067,
		"dt_emissao": "2026-11-06T10:00:00Z",
		"tributacao": "T",
		"valor_serv": 30545.00,
		"cod_servico": "07498",
		"tomador": "58801036000154"
	}`

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps/v1", strings.NewReader(jsonPayload))
	w := httptest.NewRecorder()

	h.EmissaoRPS_V1(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledV1 {
		t.Errorf("esperado chamada ao servico V1")
	}
}

func TestEmissaoLoteRPS_V1_DirectRoute(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	jsonPayload := `[
		{
			"remetente": "59789981000140",
			"im": 17329248,
			"serie_rps": "NFBON",
			"numero_rps": 2067,
			"dt_emissao": "2026-11-06T10:00:00Z",
			"tributacao": "T",
			"valor_serv": 30545.00,
			"cod_servico": "07498",
			"tomador": "58801036000154"
		}
	]`

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-lote-rps/v1", strings.NewReader(jsonPayload))
	w := httptest.NewRecorder()

	h.EmissaoLoteRPS_V1(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledLoteV1 {
		t.Errorf("esperado chamada a EmissaoLoteRPS (v1)")
	}
}

func TestEmissaoLoteRPS_V1_Wrapper(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	jsonPayload := `{
		"requests": [
			{
				"remetente": "59789981000140",
				"im": 17329248,
				"serie_rps": "NFBON",
				"numero_rps": 2067,
				"dt_emissao": "2026-11-06T10:00:00Z",
				"tributacao": "T",
				"valor_inicial_cobrado": 30545.00,
				"cod_servico": "07498",
				"tomador": "58801036000154"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-lote-rps/v1", strings.NewReader(jsonPayload))
	w := httptest.NewRecorder()

	h.EmissaoLoteRPS_V1(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledLoteV1 {
		t.Errorf("esperado chamada a EmissaoLoteRPS (v1)")
	}
}

func TestRPSRequest_ValidateAndAliases(t *testing.T) {
	// Alias valor_inicial_cobrado -> ValorServ
	payload := `{"remetente":"59789981000140","tomador":"58801036000154","valor_inicial_cobrado":1500.50}`
	var req RPSRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("falha no unmarshal: %v", err)
	}
	if req.ValorServ != 1500.50 {
		t.Errorf("esperado ValorServ=1500.50 via alias, obteve %.2f", req.ValorServ)
	}
	if err := req.Validate(); err != nil {
		t.Errorf("esperado valido, obteve erro: %v", err)
	}

	// Invalido: sem valor
	reqInvalida := RPSRequest{Remetente: "123", Tomador: "456"}
	if err := reqInvalida.Validate(); err == nil {
		t.Errorf("esperado erro por falta de valor_serv")
	}
}

func TestEmissaoDPS_DedicatedEndpoint(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	body := []byte(`{
		"remetente": "43629976000130",
		"tomador": "54487608000111",
		"numero_dps": 101,
		"cod_servico": "010701",
		"valor_serv": 1500.00,
		"discriminacao": "Consultoria"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-dps", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoDPS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledDPS {
		t.Errorf("esperado chamada a EmissaoDPS no mock")
	}
}

func TestEmissaoRPS_AutoDetectDPS(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	body := []byte(`{
		"remetente": "43629976000130",
		"tomador": "54487608000111",
		"numero_dps": 101,
		"op_simples_nacional": 2,
		"cod_servico": "010701",
		"valor_serv": 1500.00,
		"discriminacao": "Consultoria"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-rps", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoRPS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obteve %d: %s", w.Code, w.Body.String())
	}
	if !mockSvc.calledDPS {
		t.Errorf("esperado chamada a EmissaoDPS via auto-detect no /emissao-rps")
	}
}

func TestEmissaoDPS_MissingNumeroDPS(t *testing.T) {
	mockSvc := &mockNFEService{}
	h := NewNFEHandler(mockSvc)

	body := []byte(`{
		"remetente": "43629976000130",
		"tomador": "54487608000111",
		"cod_servico": "010701",
		"valor_serv": 1500.00,
		"discriminacao": "Consultoria"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/nfe/emissao-dps", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.EmissaoDPS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400 por falta de numero_dps, obteve %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "numero_dps") {
		t.Errorf("esperado menção a numero_dps no erro, obteve: %s", w.Body.String())
	}
}


