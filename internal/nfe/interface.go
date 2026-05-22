package nfe

type NFEInterface interface {
	ConsultaCNPJ(remetente, tomador string) (bool, error)
	ConsultaNFePeriodo(request ConsultaNFERequest) (bool, error)
	EmissaoRPS_V1(request RPSRequest) (bool, error)
	EmissaoLoteRPS_V1(requests []RPSRequest) (bool, error)
}
