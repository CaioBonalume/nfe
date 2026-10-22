package security

import (
	"fmt"
	"strings"
)

// IsValidCPF valida os dígitos verificadores de um CPF (11 dígitos).
func IsValidCPF(cpf string) bool {
	if len(cpf) != 11 {
		return false
	}
	allSame := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	for i := 9; i < 11; i++ {
		sum := 0
		for j := 0; j < i; j++ {
			sum += int(cpf[j]-'0') * ((i + 1) - j)
		}
		digit := ((sum * 10) % 11) % 10
		if digit != int(cpf[i]-'0') {
			return false
		}
	}
	return true
}

// IsValidCNPJ valida os dígitos verificadores de um CNPJ (14 dígitos).
func IsValidCNPJ(cnpj string) bool {
	if len(cnpj) != 14 {
		return false
	}
	allSame := true
	for i := 1; i < 14; i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	sum1 := 0
	for i := 0; i < 12; i++ {
		sum1 += int(cnpj[i]-'0') * weights1[i]
	}
	d1 := sum1 % 11
	if d1 < 2 {
		d1 = 0
	} else {
		d1 = 11 - d1
	}
	if int(cnpj[12]-'0') != d1 {
		return false
	}

	sum2 := 0
	for i := 0; i < 13; i++ {
		sum2 += int(cnpj[i]-'0') * weights2[i]
	}
	d2 := sum2 % 11
	if d2 < 2 {
		d2 = 0
	} else {
		d2 = 11 - d2
	}
	return int(cnpj[13]-'0') == d2
}

// DetectTomadorDoc analisa o campo Tomador e determina se é CPF, CNPJ ou Não informado.
// Retorna:
//   - tipoDoc: "1" para CPF, "2" para CNPJ, "3" para Não informado
//   - docXML: valor numérico a ser injetado na tag <CPF> (11 dígitos) ou <CNPJ> (14 dígitos), vazio para tipo 3
//   - docAssinatura: 14 dígitos numéricos com zeros à esquerda para compor a string de assinatura do RPS
func DetectTomadorDoc(tomador string) (tipoDoc, docXML, docAssinatura string) {
	var clean strings.Builder
	for _, r := range tomador {
		if r >= '0' && r <= '9' {
			clean.WriteRune(r)
		}
	}
	digits := clean.String()

	// Verifica se é vazio ou composto apenas por zeros (ex: "00000000000000")
	isAllZeros := true
	for _, r := range digits {
		if r != '0' {
			isAllZeros = false
			break
		}
	}
	if len(digits) == 0 || isAllZeros {
		return "3", "", "00000000000000"
	}

	// 11 ou menos dígitos -> Trata como CPF (completa com zeros à esquerda até 11 posições)
	if len(digits) <= 11 {
		cpf := fmt.Sprintf("%011s", digits)
		return "1", cpf, fmt.Sprintf("%014s", cpf)
	}

	// Exatamente 14 dígitos
	if len(digits) == 14 {
		// Se valida como CNPJ com algoritmo de dígitos verificadores
		if IsValidCNPJ(digits) {
			return "2", digits, digits
		}

		// Se começa com "000" (preenchimento padrão de banco/ERP para CPF em coluna de 14 caracteres)
		// ou se os últimos 11 dígitos formam um CPF válido
		last11 := digits[len(digits)-11:]
		if strings.HasPrefix(digits, "000") || IsValidCPF(last11) {
			return "1", last11, fmt.Sprintf("%014s", last11)
		}

		// Fallback para CNPJ
		return "2", digits, digits
	}

	// Entre 12 e 13 dígitos
	if len(digits) > 11 && len(digits) < 14 {
		last11 := digits[len(digits)-11:]
		if IsValidCPF(last11) {
			return "1", last11, fmt.Sprintf("%014s", last11)
		}
		cnpj := fmt.Sprintf("%014s", digits)
		return "2", cnpj, cnpj
	}

	return "2", digits[:14], digits[:14]
}
