package nfe

const (
	// Sefin Nacional NFS-e Endpoints (Emissão síncrona POST /nfse)
	SefinHomologacaoURL = "https://sefin.producaorestrita.nfse.gov.br/SefinNacional/nfse"
	SefinProducaoURL    = "https://sefin.nfse.gov.br/SefinNacional/nfse"

	// ADN (Ambiente de Dados Nacional) Endpoints (Distribuição e consultas)
	ADNHomologacaoURL = "https://adn.producaorestrita.nfse.gov.br"
	ADNProducaoURL    = "https://adn.nfse.gov.br"

	// Namespace oficial do padrão nacional da NFS-e
	DPSNamespace = "http://www.sped.fazenda.gov.br/nfse"

	// Versão do Schema da DPS (v1.01 é o leiaute padrão aceito pela Sefin Nacional)
	DPSVersao = "1.01"

	// Identificação do Ambiente (tpAmb)
	TpAmbProducao    = 1
	TpAmbHomologacao = 2

	// Opção pelo Simples Nacional (opSimpNac)
	OpSimplesNaoOptante   = 1 // Não optante
	OpSimplesOptanteMEI   = 2 // Optante - Microempreendedor Individual (MEI)
	OpSimplesOptanteMEEPP = 3 // Optante - Microempresa ou Empresa de Pequeno Porte (ME/EPP)

	// Regime Especial de Tributação (regEspTrib)
	RegEspTribNenhum = 0

	// Tributação do ISSQN (tribISSQN)
	TribISSQNOperacaoTributavel = 1
	TribISSQNImunidade          = 2
	TribISSQNExportacao         = 3
	TribISSQNNaoIncidencia      = 4

	// Retenção do ISSQN (tpRetISSQN)
	RetISSQNNaoRetido = 1 // 1 - ISS Não Retido
	RetISSQNRetido    = 2 // 2 - ISS Retido pelo Tomador
)

