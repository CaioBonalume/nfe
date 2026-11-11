package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	// Sefin Nacional NFS-e Endpoints (Emissão síncrona POST /nfse)
	SefinHomologacaoURL = "https://sefin.producaorestrita.nfse.gov.br/SefinNacional/nfse"
	SefinProducaoURL    = "https://sefin.nfse.gov.br/SefinNacional/nfse"

	// ADN (Ambiente de Dados Nacional) Endpoints (Distribuição e consultas)
	ADNHomologacaoURL = "https://adn.producaorestrita.nfse.gov.br"
	ADNProducaoURL    = "https://adn.nfse.gov.br"
)

type env struct {
	CERT_PATH        string `env:"PFXPATH"`
	CERT_PASS        string `env:"PFXPASSWORD"`
	SEFIN_AMBIENTE   string `env:"SEFIN_AMBIENTE"`   // "1" para Produção, "2" para Homologação (default "2")
	SEFIN_DPS_VERSAO string `env:"SEFIN_DPS_VERSAO"` // "1.01" (default)
}

var Env env

func LoadEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	sefinAmb := os.Getenv("SEFIN_AMBIENTE")
	if sefinAmb == "" {
		sefinAmb = "2" // default para Homologação (Produção Restrita)
	}

	sefinVersao := os.Getenv("SEFIN_DPS_VERSAO")
	if sefinVersao == "" {
		sefinVersao = "1.01"
	}

	Env = env{
		CERT_PATH:        os.Getenv("PFXPATH"),
		CERT_PASS:        os.Getenv("PFXPASSWORD"),
		SEFIN_AMBIENTE:   sefinAmb,
		SEFIN_DPS_VERSAO: sefinVersao,
	}
}

// SefinURL retorna a URL de emissão conforme o ambiente configurado (1 = Produção, outro = Homologação).
func SefinURL() string {
	if Env.SEFIN_AMBIENTE == "1" {
		return SefinProducaoURL
	}
	return SefinHomologacaoURL
}

