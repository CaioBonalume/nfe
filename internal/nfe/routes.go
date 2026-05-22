package nfe

import "github.com/go-chi/chi/v5"

func SetupRoutes(r chi.Router, h NFEHandler) {
	r.Route("/nfe", func(r chi.Router) {
		// CONSULTA
		// CNPJ do tomador
		// GET /v1/nfe/consulta-cnpj/{remetente}/{tomador}
		r.Get("/consulta-cnpj/{remetente}/{tomador}", h.ConsultaCNPJ)
		// NF-e emitidas em período
		// GET /v1/nfe/consulta-periodo  body: ConsultaNFERequest JSON
		r.Get("/consulta-periodo", h.ConsultaNFePeriodo)

		// EMISSÃO
		// RPS unitário
		// POST /v1/nfe/emissao-rps  body: RPSRequest JSON
		r.Post("/emissao-rps", h.EmissaoRPS_V1)
		// lote de RPS
		// POST /v1/nfe/emissao-lote-rps  body: []RPSRequest JSON
		r.Post("/emissao-lote-rps", h.EmissaoLoteRPS_V1)

		// CANCELAR
		// NFe unitário
		// DELETE /v1/nfe/cancelar-nfe  body: CancelarNFeRequest JSON
		r.Delete("/cancelar-nfe", h.CancelarNFe)

		// lote de cancelamento
		// DELETE /v1/nfe/cancelar-lote-nfe  body: []CancelarNFeRequest JSON
		r.Delete("/cancelar-lote-nfe", h.CancelarLoteNFe)
	})
}
