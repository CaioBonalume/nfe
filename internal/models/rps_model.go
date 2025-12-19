package models

import "time"

type RPSRequest struct {
	Remetente     string
	Tomador       string
	IE            int
	SerieRPS      string // opcional, use "" para padrão
	NumeroRPS     int
	DtEmissao     time.Time
	Tributacao    string // T - Tributado em SP | F - Fora de SP | A - Isento | B - Tributado fora de SP e Isento | M - Imune | N - Serviço não listado no ISS | X - Tributado em SP com Exigibilidade Suspensa | P - Exportação de serviços
	ValorServ     float64
	ValorDeducoes float64 // opcional, use 0 para padrão
	PIS           float64 // opcional, use 0 para padrão
	COFINS        float64 // opcional, use 0 para padrão
	INSS          float64 // opcional, use 0 para padrão
	IR            float64 // opcional, use 0 para padrão
	CSLL          float64 // opcional, use 0 para padrão
	CodServico    string  // 5 dígitos
	Aliquota      float64
	ISS           bool
	Discriminacao string
}
