package nfe

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseRetornoEnvioRPS_Sucesso(t *testing.T) {
	xmlPositivo := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <EnvioRPSResponse xmlns="http://www.prefeitura.sp.gov.br/nfe">
      <RetornoXML><?xml version="1.0" encoding="UTF-8"?>
        <RetornoEnvioRPS xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.prefeitura.sp.gov.br/nfe">
          <Cabecalho Versao="2" xmlns="">
            <Sucesso>true</Sucesso>
          </Cabecalho>
          <Alerta xmlns="">
            <Codigo>1651</Codigo>
            <Descricao>Atenção! O Sistema da NFS-e não faz a restrição de leiaute de emissão. Ambos leiautes permanecem disponíveis, e o emissor deve estar ciente da data de obrigatoriedade das informações para destaque do IBS e CBS, segundo o Ato Conjunto RFB/CGIBS nº 4/2026 ou posterior, e sua desconformidade estará sujeita às devidas sanções.</Descricao>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </Alerta>
          <Alerta xmlns="">
            <Codigo>648</Codigo>
            <Descricao>O campo PagamentoParceladoAntecipado não será mais utilizado e os valores enviados serão desconsiderados.</Descricao>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </Alerta>
          <ChaveNFeRPS xmlns="">
            <ChaveNFe>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <NumeroNFe>811</NumeroNFe>
              <CodigoVerificacao>RISMJCZB</CodigoVerificacao>
              <ChaveNotaNacional>35503081246408769000144000000000081126110477879915</ChaveNotaNacional>
            </ChaveNFe>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </ChaveNFeRPS>
        </RetornoEnvioRPS>
      </RetornoXML>
    </EnvioRPSResponse>
  </soap:Body>
</soap:Envelope>`

	res, err := parseRetornoEnvioRPS(xmlPositivo)
	if err != nil {
		t.Fatalf("esperado sucesso, obteve erro: %v", err)
	}

	ret, ok := res.(*RetornoEnvioRPS)
	if !ok {
		t.Fatalf("esperado tipo *RetornoEnvioRPS, obteve %T", res)
	}

	if !ret.Cabecalho.Sucesso {
		t.Errorf("esperado Cabecalho.Sucesso=true")
	}
	if ret.Cabecalho.Versao != "2" {
		t.Errorf("esperado Cabecalho.Versao=2, obteve %s", ret.Cabecalho.Versao)
	}
	if len(ret.Alertas) != 2 {
		t.Errorf("esperado 2 alertas, obteve %d", len(ret.Alertas))
	}
	if ret.Alertas[0].NumeroRPS != 741 {
		t.Errorf("esperado NumeroRPS 741 no alerta, obteve %d", ret.Alertas[0].NumeroRPS)
	}
	if ret.ChaveNFeRPS == nil {
		t.Fatalf("esperado ChaveNFeRPS preenchida")
	}
	if ret.ChaveNFeRPS.ChaveNFe.NumeroNFe != 811 {
		t.Errorf("esperado NumeroNFe=811, obteve %d", ret.ChaveNFeRPS.ChaveNFe.NumeroNFe)
	}
	if ret.ChaveNFeRPS.ChaveNFe.CodigoVerificacao != "RISMJCZB" {
		t.Errorf("esperado CodigoVerificacao=RISMJCZB, obteve %s", ret.ChaveNFeRPS.ChaveNFe.CodigoVerificacao)
	}
	if ret.ChaveNFeRPS.ChaveNFe.ChaveNotaNacional != "35503081246408769000144000000000081126110477879915" {
		t.Errorf("esperado ChaveNotaNacional correta, obteve %s", ret.ChaveNFeRPS.ChaveNFe.ChaveNotaNacional)
	}
	if ret.ChaveNFeRPS.ChaveRPS.NumeroRPS != 741 {
		t.Errorf("esperado ChaveRPS.NumeroRPS=741, obteve %d", ret.ChaveNFeRPS.ChaveRPS.NumeroRPS)
	}

	// Serialização JSON
	jsonBytes, err := json.Marshal(ret)
	if err != nil {
		t.Fatalf("falha ao serializar para JSON: %v", err)
	}
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"sucesso":true`) {
		t.Errorf("JSON esperado conter '\"sucesso\":true', obteve: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"numero_nfe":811`) {
		t.Errorf("JSON esperado conter '\"numero_nfe\":811', obteve: %s", jsonStr)
	}
}

func TestParseRetornoEnvioRPS_Rejeicao(t *testing.T) {
	xmlNegativo := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <EnvioRPSResponse xmlns="http://www.prefeitura.sp.gov.br/nfe">
      <RetornoXML><?xml version="1.0" encoding="UTF-8"?>
        <RetornoEnvioRPS xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.prefeitura.sp.gov.br/nfe">
          <Cabecalho Versao="2" xmlns="">
            <Sucesso>false</Sucesso>
          </Cabecalho>
          <Alerta xmlns="">
            <Codigo>1651</Codigo>
            <Descricao>Atenção! O Sistema da NFS-e não faz a restrição de leiaute de emissão. Ambos leiautes permanecem disponíveis, e o emissor deve estar ciente da data de obrigatoriedade das informações para destaque do IBS e CBS, segundo o Ato Conjunto RFB/CGIBS nº 4/2026 ou posterior, e sua desconformidade estará sujeita às devidas sanções.</Descricao>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </Alerta>
          <Alerta xmlns="">
            <Codigo>648</Codigo>
            <Descricao>O campo PagamentoParceladoAntecipado não será mais utilizado e os valores enviados serão desconsiderados.</Descricao>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </Alerta>
          <Alerta xmlns="">
            <Codigo>208</Codigo>
            <Descricao>Alíquota informada (0,05) difere da alíquota vigente (0,02) para o código de serviço informado. O sistema irá adotar a alíquota vigente.</Descricao>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </Alerta>
          <Erro xmlns="">
            <Codigo>1630</Codigo>
            <Descricao>Código de serviço (7870) não permite indicação de Valor Total Recebido.</Descricao>
            <ChaveRPS>
              <InscricaoPrestador>73149870</InscricaoPrestador>
              <SerieRPS>NFBON</SerieRPS>
              <NumeroRPS>741</NumeroRPS>
            </ChaveRPS>
          </Erro>
        </RetornoEnvioRPS>
      </RetornoXML>
    </EnvioRPSResponse>
  </soap:Body>
</soap:Envelope>`

	res, err := parseRetornoEnvioRPS(xmlNegativo)
	if err == nil {
		t.Fatalf("esperado erro por Sucesso=false, obteve nil")
	}

	ret, ok := res.(*RetornoEnvioRPS)
	if !ok {
		t.Fatalf("esperado tipo *RetornoEnvioRPS mesmo com erro, obteve %T", res)
	}

	if ret.Cabecalho.Sucesso {
		t.Errorf("esperado Cabecalho.Sucesso=false")
	}
	if len(ret.Erros) != 1 {
		t.Fatalf("esperado 1 erro, obteve %d", len(ret.Erros))
	}
	if ret.Erros[0].Codigo != "1630" {
		t.Errorf("esperado erro 1630, obteve %s", ret.Erros[0].Codigo)
	}
	if ret.Erros[0].NumeroRPS != 741 {
		t.Errorf("esperado NumeroRPS=741 no erro, obteve %d", ret.Erros[0].NumeroRPS)
	}
	if len(ret.Alertas) != 3 {
		t.Errorf("esperado 3 alertas, obteve %d", len(ret.Alertas))
	}

	// Serialização JSON
	jsonBytes, err := json.Marshal(ret)
	if err != nil {
		t.Fatalf("falha ao serializar para JSON: %v", err)
	}
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"sucesso":false`) {
		t.Errorf("JSON esperado conter '\"sucesso\":false', obteve: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"1630"`) {
		t.Errorf("JSON esperado conter '1630', obteve: %s", jsonStr)
	}
}

func TestParseRetornoEnvioRPS_SoapFault(t *testing.T) {
	soapFault := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Client</faultcode>
      <faultstring>Assinatura inválida do RPS</faultstring>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`

	res, err := parseRetornoEnvioRPS(soapFault)
	if err == nil {
		t.Fatalf("esperado erro para SOAP Fault")
	}
	m, ok := res.(map[string]string)
	if !ok {
		t.Fatalf("esperado map[string]string para SOAP Fault")
	}
	if m["error"] != "Assinatura inválida do RPS" {
		t.Errorf("esperado mensagem de fault, obteve: %v", m["error"])
	}
}

