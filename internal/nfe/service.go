package nfe

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// NFEServiceInterface define os métodos disponíveis para o serviço de NFE.
type NFEServiceInterface interface {
	ConsultaCNPJ(remetente, tomador string) (string, error)
	ConsultaNFePeriodo(request ConsultaNFERequest) ([]ConsultaNFEResponse, error)
	EmissaoRPS(request RPSRequest) (string, error)
	EmissaoLoteRPS(requests []RPSRequest) (any, error)
	CancelarNFe(request CancelarNFeRequest) (string, error)
	CancelarLoteNFe(requests []CancelarNFeRequest) (string, error)
}

type nfeService struct{}

func NewNFEService() NFEServiceInterface {
	return &nfeService{}
}

func (s *nfeService) ConsultaCNPJ(remetente, tomador string) (string, error) {
	_, body := ConsultaCNPJ_V1(remetente, tomador)
	return body, nil
}

func (s *nfeService) ConsultaNFePeriodo(request ConsultaNFERequest) ([]ConsultaNFEResponse, error) {
	_, body := ConsultaNFePeriodo(request)

	startIdx := strings.Index(body, "<RetornoConsulta")
	endIdx := strings.Index(body, "</RetornoConsulta>")
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("tag RetornoConsulta não encontrada na resposta")
	}

	xmlData := body[startIdx : endIdx+len("</RetornoConsulta>")]
	var retorno RetornoConsulta
	if err := xml.Unmarshal([]byte(xmlData), &retorno); err != nil {
		return nil, fmt.Errorf("erro ao decodificar XML da resposta: %v", err)
	}

	return retorno.NFes, nil
}

func (s *nfeService) EmissaoRPS(request RPSRequest) (string, error) {
	_, body := EmissaoRPS_V1(request)
	return body, nil
}

func (s *nfeService) EmissaoLoteRPS(requests []RPSRequest) (any, error) {
	_, body, err := EmissaoLoteRPS_V1(requests)
	if err != nil {
		return nil, err
	}

	startIdx := strings.Index(body, "<RetornoEnvioLoteRPS")
	endIdx := strings.Index(body, "</RetornoEnvioLoteRPS>")
	if startIdx == -1 || endIdx == -1 {
		return map[string]string{"resposta": body}, fmt.Errorf("tag RetornoEnvioLoteRPS não encontrada na resposta")
	}

	xmlData := body[startIdx : endIdx+len("</RetornoEnvioLoteRPS>")]
	var retorno RetornoEnvioLoteRPS
	println("XML DATA: ", xmlData)
	if err := xml.Unmarshal([]byte(xmlData), &retorno); err != nil {
		return map[string]string{"resposta": body}, fmt.Errorf("erro ao decodificar XML da resposta: %v", err)
	}

	// Popula o número do RPS em cada erro a partir de ChaveRPS ou tag direta
	for i := range retorno.Erros {
		if retorno.Erros[i].NumeroRPS == 0 {
			if retorno.Erros[i].ChaveRPS.NumeroRPS != 0 {
				retorno.Erros[i].NumeroRPS = retorno.Erros[i].ChaveRPS.NumeroRPS
			} else if retorno.Erros[i].DirectNumeroRPS != 0 {
				retorno.Erros[i].NumeroRPS = retorno.Erros[i].DirectNumeroRPS
			}
		}
	}

	// Popula o número do RPS em cada alerta a partir de ChaveRPS ou tag direta
	for i := range retorno.Alertas {
		if retorno.Alertas[i].NumeroRPS == 0 {
			if retorno.Alertas[i].ChaveRPS.NumeroRPS != 0 {
				retorno.Alertas[i].NumeroRPS = retorno.Alertas[i].ChaveRPS.NumeroRPS
			} else if retorno.Alertas[i].DirectNumeroRPS != 0 {
				retorno.Alertas[i].NumeroRPS = retorno.Alertas[i].DirectNumeroRPS
			}
		}
	}

	if !retorno.Cabecalho.Sucesso {
		errosAgrupados := agruparErros(retorno.Erros)
		alertasAgrupados := agruparAlertas(retorno.Alertas)

		if len(errosAgrupados) > 0 || len(alertasAgrupados) > 0 {
			res := map[string]any{
				"status_operacao": false,
			}
			if len(errosAgrupados) > 0 {
				res["erros"] = errosAgrupados
			} else {
				res["erros"] = alertasAgrupados
			}
			if len(alertasAgrupados) > 0 {
				res["alertas"] = alertasAgrupados
			}
			return res, fmt.Errorf("lote rps rejeitado")
		}
		return map[string]string{"resposta": body}, fmt.Errorf("lote rps rejeitado")
	}

	return &retorno, nil
}

func (s *nfeService) CancelarNFe(request CancelarNFeRequest) (string, error) {
	_, body := CancelarNFe(request)
	return body, nil
}

func (s *nfeService) CancelarLoteNFe(requests []CancelarNFeRequest) (string, error) {
	_, body := CancelarLoteNFe(requests)
	return body, nil
}

func agruparErros(erros []ErroRetorno) []ErroAgrupado {
	var resultado []ErroAgrupado
	indexMap := make(map[string]int)

	for _, e := range erros {
		desc := strings.TrimSpace(e.Descricao)
		key := fmt.Sprintf("%s|%s", e.Codigo, desc)
		idx, exists := indexMap[key]
		if !exists {
			item := ErroAgrupado{
				Codigo:    e.Codigo,
				Descricao: desc,
				RPS:       []int{},
			}
			if e.NumeroRPS != 0 {
				item.RPS = append(item.RPS, e.NumeroRPS)
			}
			indexMap[key] = len(resultado)
			resultado = append(resultado, item)
		} else if e.NumeroRPS != 0 {
			duplicado := false
			for _, rps := range resultado[idx].RPS {
				if rps == e.NumeroRPS {
					duplicado = true
					break
				}
			}
			if !duplicado {
				resultado[idx].RPS = append(resultado[idx].RPS, e.NumeroRPS)
			}
		}
	}

	for i := range resultado {
		sort.Ints(resultado[i].RPS)
	}

	return resultado
}

func agruparAlertas(alertas []Alerta) []ErroAgrupado {
	var resultado []ErroAgrupado
	indexMap := make(map[string]int)

	for _, a := range alertas {
		codStr := fmt.Sprintf("%d", a.Codigo)
		desc := strings.TrimSpace(a.Descricao)
		key := fmt.Sprintf("%s|%s", codStr, desc)
		idx, exists := indexMap[key]
		if !exists {
			item := ErroAgrupado{
				Codigo:    codStr,
				Descricao: desc,
				RPS:       []int{},
			}
			if a.NumeroRPS != 0 {
				item.RPS = append(item.RPS, a.NumeroRPS)
			}
			indexMap[key] = len(resultado)
			resultado = append(resultado, item)
		} else if a.NumeroRPS != 0 {
			duplicado := false
			for _, rps := range resultado[idx].RPS {
				if rps == a.NumeroRPS {
					duplicado = true
					break
				}
			}
			if !duplicado {
				resultado[idx].RPS = append(resultado[idx].RPS, a.NumeroRPS)
			}
		}
	}

	for i := range resultado {
		sort.Ints(resultado[i].RPS)
	}

	return resultado
}

