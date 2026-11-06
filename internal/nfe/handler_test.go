package nfe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockNFEService struct {
	calledV1     bool
	calledV2     bool
	calledLoteV1 bool
	calledLoteV2 bool
}

func (m *mockNFEService) ConsultaCNPJ(remetente, tomador string) (string, error) {
	return "", nil
}

func (m *mockNFEService) ConsultaNFePeriodo(request ConsultaNFERequest) ([]ConsultaNFEResponse, error) {
	return nil, nil
}

func (m *mockNFEService) EmissaoRPS(request RPSRequest) (string, error) {
	m.calledV1 = true
	return "<Resultado>SucessoV1</Resultado>", nil
}

func (m *mockNFEService) EmissaoRPS_V2(request RPSRequestV2) (string, error) {
	m.calledV2 = true
	return "<Resultado>SucessoV2</Resultado>", nil
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

