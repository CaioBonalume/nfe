package nfe
import (
	"encoding/xml"
	"fmt"
	"strings"
)

// NFEServiceInterface define os métodos disponíveis para o serviço de NFE.
type NFEServiceInterface interface {
	ConsultaCNPJ(remetente, tomador string) (string, error)
	ConsultaNFePeriodo(request ConsultaNFERequest) ([]ConsultaNFEResponse, error)
	EmissaoRPS(request RPSRequest) (string, error)
	EmissaoLoteRPS(requests []RPSRequest) (*RetornoEnvioLoteRPS, error)
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

func (s *nfeService) EmissaoLoteRPS(requests []RPSRequest) (*RetornoEnvioLoteRPS, error) {
	_, body, err := EmissaoLoteRPS_V1(requests)
	if err != nil {
		return nil, err
	}

	startIdx := strings.Index(body, "<RetornoEnvioLoteRPS")
	endIdx := strings.Index(body, "</RetornoEnvioLoteRPS>")
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("tag RetornoEnvioLoteRPS não encontrada na resposta")
	}

	xmlData := body[startIdx : endIdx+len("</RetornoEnvioLoteRPS>")]
	var retorno RetornoEnvioLoteRPS
	if err := xml.Unmarshal([]byte(xmlData), &retorno); err != nil {
		return nil, fmt.Errorf("erro ao decodificar XML da resposta: %v", err)
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
