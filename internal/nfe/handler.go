package nfe

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// NFEHandler define os endpoints relacionados à NF-e.
type NFEHandler interface {
	ConsultaCNPJ(w http.ResponseWriter, r *http.Request)
	ConsultaNFePeriodo(w http.ResponseWriter, r *http.Request)
	EmissaoRPS_V1(w http.ResponseWriter, r *http.Request)
	EmissaoLoteRPS_V1(w http.ResponseWriter, r *http.Request)
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

// ---- EmissaoRPS_V1 ----
// POST /v1/nfe/emissao-rps
// JSON body: RPSRequest

func (h *nfeHandler) EmissaoRPS_V1(w http.ResponseWriter, r *http.Request) {
	var req RPSRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON inválido")
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

// ---- EmissaoLoteRPS_V1 ----
// POST /v1/nfe/emissao-lote-rps
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
