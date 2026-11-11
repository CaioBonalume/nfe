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
		// RPS unitário (detecta leiaute via param/body ou exige especificação)
		r.Post("/emissao-rps", h.EmissaoRPS)
		r.Post("/emissao-rps/v1", h.EmissaoRPS_V1)
		r.Post("/emissao-rps/v2", h.EmissaoRPS_V2)
		// DPS unitário (Emissor Nacional Sefin)
		r.Post("/emissao-dps", h.EmissaoDPS)

		// lote de RPS (aplica leiaute uniforme para todos os RPS do lote)
		r.Post("/emissao-lote-rps", h.EmissaoLoteRPS)
		r.Post("/emissao-lote-rps/v1", h.EmissaoLoteRPS_V1)
		r.Post("/emissao-lote-rps/v2", h.EmissaoLoteRPS_V2)

		// CANCELAR
		// NFe unitário
		// DELETE /v1/nfe/cancelar-nfe  body: CancelarNFeRequest JSON
		r.Delete("/cancelar-nfe", h.CancelarNFe)

		// lote de cancelamento
		// DELETE /v1/nfe/cancelar-lote-nfe  body: []CancelarNFeRequest JSON
		r.Delete("/cancelar-lote-nfe", h.CancelarLoteNFe)
	})
}
