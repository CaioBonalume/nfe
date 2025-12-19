package models

import "time"

type ConsultaNFERequest struct {
	CNPJ_REMETENTE string    // 14 dígitos
	IE             string    // V.1 8 dígitos / V.2 12 dígitos
	CNPJ           string    // 14 dígitos
	DTInicio       time.Time // Formato YYYY-MM-DD
	DTFim          time.Time // Formato YYYY-MM-DD
	Pagina         int
}
