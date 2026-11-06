package nfe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// NFEHandler define os endpoints relacionados à NF-e.
type NFEHandler interface {
	ConsultaCNPJ(w http.ResponseWriter, r *http.Request)
	ConsultaNFePeriodo(w http.ResponseWriter, r *http.Request)
	EmissaoRPS(w http.ResponseWriter, r *http.Request)
	EmissaoRPS_V1(w http.ResponseWriter, r *http.Request)
	EmissaoRPS_V2(w http.ResponseWriter, r *http.Request)
	EmissaoLoteRPS(w http.ResponseWriter, r *http.Request)
	EmissaoLoteRPS_V1(w http.ResponseWriter, r *http.Request)
	EmissaoLoteRPS_V2(w http.ResponseWriter, r *http.Request)
	CancelarNFe(w http.ResponseWriter, r *http.Request)
	CancelarLoteNFe(w http.ResponseWriter, r *http.Request)
}

type nfeHandler struct {
	nfeService NFEServiceInterface
}

func NewNFEHandler(nfeService NFEServiceInterface) NFEHandler {
	return &nfeHandler{nfeService: nfeService}
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- ConsultaCNPJ ----
// POST /v1/nfe/consulta-cnpj/{remetente}/{tomador}
// multipart: cert (arquivo .pfx), password (string)

func (h *nfeHandler) ConsultaCNPJ(w http.ResponseWriter, r *http.Request) {
	// TODO: quando auth por token estiver pronto, remover params da URL
	remetente := chi.URLParam(r, "remetente")
	tomador := chi.URLParam(r, "tomador")

	if remetente == "" || tomador == "" {
		writeError(w, http.StatusBadRequest, "remetente e tomador obrigatórios")
		return
	}

	body, err := h.nfeService.ConsultaCNPJ(remetente, tomador)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
}

// ---- ConsultaNFePeriodo ----
// POST /v1/nfe/consulta-periodo
// JSON body: ConsultaNFERequest

func (h *nfeHandler) ConsultaNFePeriodo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CNPJRemetente string `json:"cnpj_remetente"`
		CNPJ          string `json:"cnpj"`
		IM            string `json:"im"`
		DTInicio      string `json:"dt_inicio"` // YYYY-MM-DD
		DTFim         string `json:"dt_fim"`    // YYYY-MM-DD
		Pagina        int    `json:"pagina"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido")
		return
	}

	layout := "2006-01-02"
	dtInicio, err := time.Parse(layout, req.DTInicio)
	if err != nil {
		writeError(w, http.StatusBadRequest, "dt_inicio inválida (use YYYY-MM-DD)")
		return
	}
	dtFim, err := time.Parse(layout, req.DTFim)
	if err != nil {
		writeError(w, http.StatusBadRequest, "dt_fim inválida (use YYYY-MM-DD)")
		return
	}

	pagina := req.Pagina
	if pagina < 1 {
		pagina = 1
	}

	nfeReq := ConsultaNFERequest{
		CNPJ_REMETENTE: req.CNPJRemetente,
		CNPJ:           req.CNPJ,
		IM:             req.IM,
		DTInicio:       dtInicio,
		DTFim:          dtFim,
		Pagina:         pagina,
	}

	body, err := h.nfeService.ConsultaNFePeriodo(nfeReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, body)
}

// detectLeiaute identifica se a requisição deve ser processada no Leiaute 1 ou Leiaute 2.
// Retorna 1, 2 ou 0 (se não especificado).
func detectLeiaute(r *http.Request, bodyBytes []byte) int {
	// 1. Query Param ?leiaute=1|2
	q := strings.TrimSpace(r.URL.Query().Get("leiaute"))
	if q == "1" {
		return 1
	}
	if q == "2" {
		return 2
	}

	// 2. Header HTTP X-Leiaute-Versao
	h := strings.TrimSpace(r.Header.Get("X-Leiaute-Versao"))
	if h == "1" {
		return 1
	}
	if h == "2" {
		return 2
	}

	// 3. Inspeção do payload JSON
	if len(bodyBytes) > 0 {
		var peek struct {
			Leiaute int `json:"leiaute"`
		}
		if err := json.Unmarshal(bodyBytes, &peek); err == nil && (peek.Leiaute == 1 || peek.Leiaute == 2) {
			return peek.Leiaute
		}
		// Se for array: analisa se o primeiro item especifica leiaute
		var peekArr []struct {
			Leiaute int `json:"leiaute"`
		}
		if err := json.Unmarshal(bodyBytes, &peekArr); err == nil && len(peekArr) > 0 && (peekArr[0].Leiaute == 1 || peekArr[0].Leiaute == 2) {
			return peekArr[0].Leiaute
		}
	}

	return 0
}

// ---- EmissaoRPS ----
// POST /v1/nfe/emissao-rps
// Suporta Leiaute 1 e 2. Se o leiaute não for especificado, retorna HTTP 400 perguntando ao cliente.

func (h *nfeHandler) EmissaoRPS(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "erro ao ler corpo da requisição")
		return
	}

	leiaute := detectLeiaute(r, bodyBytes)
	if leiaute == 0 {
		writeError(w, http.StatusBadRequest, "leiaute não informado. Para emissões individuais, especifique o leiaute desejado (1 ou 2) via query param ?leiaute=1|2, header X-Leiaute-Versao ou campo 'leiaute' no JSON")
		return
	}

	if leiaute == 1 {
		var req RPSRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			writeError(w, http.StatusBadRequest, "body JSON inválido para Leiaute 1: "+err.Error())
			return
		}
		if req.Remetente == "" || req.Tomador == "" {
			writeError(w, http.StatusBadRequest, "remetente e tomador obrigatórios")
			return
		}
		body, err := h.nfeService.EmissaoRPS(req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
		return
	}

	// Leiaute 2
	var reqV2 RPSRequestV2
	if err := json.Unmarshal(bodyBytes, &reqV2); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido para Leiaute 2: "+err.Error())
		return
	}
	if err := reqV2.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := h.nfeService.EmissaoRPS_V2(reqV2)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
}

// ---- EmissaoRPS_V1 ----
// POST /v1/nfe/emissao-rps/v1 (ou endpoint legado)
// JSON body: RPSRequest

func (h *nfeHandler) EmissaoRPS_V1(w http.ResponseWriter, r *http.Request) {
	var req RPSRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido: "+err.Error())
		return
	}

	if req.Remetente == "" || req.Tomador == "" {
		writeError(w, http.StatusBadRequest, "remetente e tomador obrigatórios")
		return
	}

	body, err := h.nfeService.EmissaoRPS(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
}

// ---- EmissaoRPS_V2 ----
// POST /v1/nfe/emissao-rps/v2
// JSON body: RPSRequestV2

func (h *nfeHandler) EmissaoRPS_V2(w http.ResponseWriter, r *http.Request) {
	var req RPSRequestV2

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido para Leiaute 2: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	body, err := h.nfeService.EmissaoRPS_V2(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
}

// ---- EmissaoLoteRPS ----
// POST /v1/nfe/emissao-lote-rps
// O leiaute escolhido serve uniformemente para todos os RPS do lote.

func (h *nfeHandler) EmissaoLoteRPS(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "erro ao ler corpo da requisição")
		return
	}

	leiaute := detectLeiaute(r, bodyBytes)
	if leiaute == 0 {
		writeError(w, http.StatusBadRequest, "leiaute do lote não informado. Especifique o leiaute para todo o lote (1 ou 2) via query param ?leiaute=1|2, header X-Leiaute-Versao ou campo 'leiaute' no JSON")
		return
	}

	if leiaute == 1 {
		var requests []RPSRequest
		if err := json.Unmarshal(bodyBytes, &requests); err != nil || len(requests) == 0 {
			var wrapper struct {
				Requests []RPSRequest `json:"requests"`
			}
			if err2 := json.Unmarshal(bodyBytes, &wrapper); err2 != nil || len(wrapper.Requests) == 0 {
				writeError(w, http.StatusBadRequest, "lista de RPS vazia ou formato inválido para Leiaute 1")
				return
			}
			requests = wrapper.Requests
		}

		body, err := h.nfeService.EmissaoLoteRPS(requests)
		if err != nil {
			if body != nil {
				writeJSON(w, http.StatusBadRequest, body)
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, body)
		return
	}

	// Leiaute 2
	var requestsV2 []RPSRequestV2
	if err := json.Unmarshal(bodyBytes, &requestsV2); err != nil || len(requestsV2) == 0 {
		var wrapper LoteRPSRequestV2
		if err2 := json.Unmarshal(bodyBytes, &wrapper); err2 != nil || len(wrapper.Requests) == 0 {
			writeError(w, http.StatusBadRequest, "lista de RPS vazia ou formato inválido para Leiaute 2")
			return
		}
		requestsV2 = wrapper.Requests
	}

	for i, rps := range requestsV2 {
		if err := rps.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("rps [%d] invalido: %v", i+1, err))
			return
		}
	}

	body, err := h.nfeService.EmissaoLoteRPS_V2(requestsV2)
	if err != nil {
		if body != nil {
			writeJSON(w, http.StatusBadRequest, body)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, body)
}

// ---- EmissaoLoteRPS_V1 ----
// POST /v1/nfe/emissao-lote-rps/v1 (ou endpoint legado)
// JSON body: []RPSRequest

func (h *nfeHandler) EmissaoLoteRPS_V1(w http.ResponseWriter, r *http.Request) {
	var requests []RPSRequest

	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido")
		return
	}

	if len(requests) == 0 {
		writeError(w, http.StatusBadRequest, "lista de RPS vazia")
		return
	}

	body, err := h.nfeService.EmissaoLoteRPS(requests)
	if err != nil {
		if body != nil {
			writeJSON(w, http.StatusBadRequest, body)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, body)
}

// ---- EmissaoLoteRPS_V2 ----
// POST /v1/nfe/emissao-lote-rps/v2
// JSON body: []RPSRequestV2 ou LoteRPSRequestV2

func (h *nfeHandler) EmissaoLoteRPS_V2(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "erro ao ler corpo da requisição")
		return
	}

	var requestsV2 []RPSRequestV2
	if err := json.Unmarshal(bodyBytes, &requestsV2); err != nil || len(requestsV2) == 0 {
		var wrapper LoteRPSRequestV2
		if err2 := json.Unmarshal(bodyBytes, &wrapper); err2 != nil || len(wrapper.Requests) == 0 {
			writeError(w, http.StatusBadRequest, "lista de RPS vazia ou formato inválido para Leiaute 2")
			return
		}
		requestsV2 = wrapper.Requests
	}

	for i, rps := range requestsV2 {
		if err := rps.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("rps [%d] invalido: %v", i+1, err))
			return
		}
	}

	body, err := h.nfeService.EmissaoLoteRPS_V2(requestsV2)
	if err != nil {
		if body != nil {
			writeJSON(w, http.StatusBadRequest, body)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, body)
}

// ---- CancelarNFe ----
// DELETE /v1/nfe/cancelar-nfe
// JSON body: CancelarNFeRequest

func (h *nfeHandler) CancelarNFe(w http.ResponseWriter, r *http.Request) {
	var req CancelarNFeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido")
		return
	}

	if req.Remetente == "" || req.Inscricao == "" || req.NumeroNFe == 0 || req.CodigoVerificacao == "" {
		writeError(w, http.StatusBadRequest, "remetente, inscricao_prestador, numero_nfe e codigo_verificacao são obrigatórios")
		return
	}

	body, err := h.nfeService.CancelarNFe(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
}

// ---- CancelarLoteNFe ----
// DELETE /v1/nfe/cancelar-lote-nfe
// JSON body: []CancelarNFeRequest

func (h *nfeHandler) CancelarLoteNFe(w http.ResponseWriter, r *http.Request) {
	var requests []CancelarNFeRequest

	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido")
		return
	}

	if len(requests) == 0 {
		writeError(w, http.StatusBadRequest, "lista de cancelamento vazia")
		return
	}

	if len(requests) > 50 {
		writeError(w, http.StatusBadRequest, "limite de notas para cancelamento em lote excedido (max 50)")
		return
	}

	body, err := h.nfeService.CancelarLoteNFe(requests)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"resultado": body})
}
