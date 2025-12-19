package main

import (
	"log"
	"nfe/config"
	"nfe/internal"
	"nfe/internal/models"
	"time"
)

func main() {
	config.LoadEnv()
	// Consulta CNPJ
	// status, body := internal.ConsultaCNPJ_V1("00000000000000", "00000000000000")

	// Consulta NFE por período
	// request := models.ConsultaNFERequest{
	// 	CNPJ_REMETENTE: "00000000000000",
	// 	CNPJ:           "00000000000000",
	// 	IE:             "00000000",
	// 	DTInicio:       time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
	// DTFim:          time.Date(2025, 12, 17, 23, 59, 59, 0, time.UTC),
	// 	Pagina:         1,
	// }
	// status, body := internal.ConsultaNFePeriodo(request)

	// Emissão de RPS
	request := models.RPSRequest{
		Remetente:     "00000000000000",
		Tomador:       "00000000000000",
		IE:            00000000,
		NumeroRPS:     21,
		DtEmissao:     time.Now(),
		ValorServ:     100,
		ValorDeducoes: 0,
		PIS:           0,
		COFINS:        0,
		INSS:          0,
		IR:            0,
		CSLL:          0,
		Aliquota:      0,
		CodServico:    "00000",
		ISS:           false,
		Discriminacao: "Teste",
	}

	status, body := internal.EmissaoRPS_V1(request)
	log.Println("Resposta da SEFAZ SP:")
	log.Println(status)
	log.Println(body)
}
