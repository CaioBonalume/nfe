package security

import (
	"testing"
)

func TestDetectTomadorDoc(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTipo  string
		expectedXML   string
		expectedSig   string
	}{
		{
			name:         "CPF com zeros a esquerda (14 digitos)",
			input:        "00007766570890",
			expectedTipo: "1",
			expectedXML:  "07766570890",
			expectedSig:  "00007766570890",
		},
		{
			name:         "CPF com 10 digitos (sem zero a esquerda)",
			input:        "7766570890",
			expectedTipo: "1",
			expectedXML:  "07766570890",
			expectedSig:  "00007766570890",
		},
		{
			name:         "CPF padrao com 11 digitos",
			input:        "07766570890",
			expectedTipo: "1",
			expectedXML:  "07766570890",
			expectedSig:  "00007766570890",
		},
		{
			name:         "CNPJ padrao com 14 digitos",
			input:        "54487608000111",
			expectedTipo: "2",
			expectedXML:  "54487608000111",
			expectedSig:  "54487608000111",
		},
		{
			name:         "CNPJ com 3 zeros a esquerda",
			input:        "00089289000167",
			expectedTipo: "2",
			expectedXML:  "00089289000167",
			expectedSig:  "00089289000167",
		},
		{
			name:         "Tomador nao informado (14 zeros)",
			input:        "00000000000000",
			expectedTipo: "3",
			expectedXML:  "",
			expectedSig:  "00000000000000",
		},
		{
			name:         "Tomador vazio",
			input:        "",
			expectedTipo: "3",
			expectedXML:  "",
			expectedSig:  "00000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tipo, xmlVal, sigVal := DetectTomadorDoc(tt.input)
			if tipo != tt.expectedTipo {
				t.Errorf("tipoDoc = %v, esperado %v", tipo, tt.expectedTipo)
			}
			if xmlVal != tt.expectedXML {
				t.Errorf("docXML = %v, esperado %v", xmlVal, tt.expectedXML)
			}
			if sigVal != tt.expectedSig {
				t.Errorf("docAssinatura = %v, esperado %v", sigVal, tt.expectedSig)
			}
		})
	}
}
