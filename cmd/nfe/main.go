package main

import (
	"log"
	"nfe/config"
	"nfe/internal"
)

func main() {
	config.LoadEnv()
	// Consulta CNPJ
	status, body := internal.ConsultaCNPJ_V1("59530271000100")
	log.Println("Resposta da SEFAZ SP:")
	log.Println(status)
	log.Println(body)
}
