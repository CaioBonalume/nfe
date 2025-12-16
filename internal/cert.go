package internal

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"nfe/config"
	"os"
	"strings"

	"golang.org/x/crypto/pkcs12"
)

func ExportPEM() ([]byte, []byte, error) {
	fmt.Println("Exportando certificado PFX para PEM...")
	// 1. Ler o arquivo PFX do disco
	// log.Printf("path: %s", config.Env.CERT_PATH)
	pfxData, err := os.ReadFile(config.Env.CERT_PATH)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao ler arquivo PFX: %v", err)
	}

	// 2. Converter TODO o conteúdo do PFX para blocos PEM
	// A função ToPEM não reclama se tiverem 3, 4 ou 10 itens dentro do PFX (Cadeia completa)
	blocks, err := pkcs12.ToPEM(pfxData, config.Env.CERT_PASS)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao converter PFX para PEM (senha incorreta?): %v", err)
	}

	var pemData []byte // Vai guardar todos os certificados (O seu + Cadeia)
	var keyData []byte // Vai guardar a chave privada

	// 3. Iterar sobre os blocos encontrados e separar
	for _, b := range blocks {
		if b.Type == "CERTIFICATE" {
			// Concatena o certificado na lista (preserva a cadeia)
			pemData = append(pemData, pem.EncodeToMemory(b)...)
		} else if b.Type == "PRIVATE KEY" || strings.Contains(b.Type, "PRIVATE KEY") {
			// Encontrou a chave privada
			keyData = append(keyData, pem.EncodeToMemory(b)...)
		}
	}

	if len(pemData) == 0 {
		return nil, nil, fmt.Errorf("nenhum certificado encontrado no PFX")
	}
	if len(keyData) == 0 {
		return nil, nil, fmt.Errorf("nenhuma chave privada encontrada no PFX")
	}

	return pemData, keyData, nil
}

func TLSCert() (tlsCert tls.Certificate, certPEM []byte, keyPEM []byte, err error) {
	cert, key, err := ExportPEM()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Chave PEM gerada com sucesso!")
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return tls.Certificate{}, nil, nil, fmt.Errorf("erro ao carregar par chave/certificado: %v", err)
	}
	return certificate, cert, key, nil
}

// SignBlock assina um arquivo XML (ou string XML) injetando a assinatura DSIG
// Este exemplo assume que você quer assinar o conteúdo inteiro ou uma tag específica.
// Para SP (ConsultaCNPJ), a assinatura é "Enveloped", ou seja, ela vai DENTRO do XML,
// mas assinando o elemento pai.
func SignBlock(xmlContent string, keyPEM []byte, certPEM []byte) (string, error) {
	// 1. Calcular o Digest do Payload (SHA1)
	// Como estamos usando Enveloped Signature, o digest é calculado sobre o conteúdo ORIGINAL (sem a assinatura).
	hasher := sha1.New()
	hasher.Write([]byte(xmlContent))
	digest := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	// 2. Definir os Templates do SignedInfo

	// A) Template para CÁLCULO DO HASH (Matemática)
	// CORREÇÃO CRÍTICA: Adicionamos 'xmlns:p1' aqui.
	// O C14N Inclusivo exige que o SignedInfo herde os namespaces do pai (p1) e declare o seu próprio (ds).
	// A ordem alfabética dos atributos xmlns é obrigatória: 'ds' vem antes de 'p1'.
	signedInfoTemplateHash := `<ds:SignedInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:p1="http://www.prefeitura.sp.gov.br/nfe"><ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></ds:CanonicalizationMethod><ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"></ds:SignatureMethod><ds:Reference URI=""><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"></ds:Transform><ds:Transform Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></ds:Transform></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></ds:Reference></ds:SignedInfo>`

	// B) Template para TRANSPORTE (O que vai no XML)
	// Aqui NÃO colocamos os namespaces, pois eles já existem no XML onde esse bloco será colado.
	signedInfoTemplateTransport := `<ds:SignedInfo><ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></ds:CanonicalizationMethod><ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"></ds:SignatureMethod><ds:Reference URI=""><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"></ds:Transform><ds:Transform Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></ds:Transform></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></ds:Reference></ds:SignedInfo>`

	// Preenche o Digest nos dois templates
	stringParaHash := fmt.Sprintf(signedInfoTemplateHash, digest)
	stringParaTransporte := fmt.Sprintf(signedInfoTemplateTransport, digest)

	// 3. Assinar a String de Hash (Essa contém os namespaces herdados)
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return "", fmt.Errorf("chave privada invalida")
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("erro parse rsa: %v", err)
	}

	siHasher := sha1.New()
	siHasher.Write([]byte(stringParaHash)) // Assina a versão "cheia" de namespaces
	siHash := siHasher.Sum(nil)

	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA1, siHash)
	if err != nil {
		return "", fmt.Errorf("erro assinatura rsa: %v", err)
	}
	signatureValue := base64.StdEncoding.EncodeToString(sigBytes)

	// 4. Tratar o Certificado
	blockCert, _ := pem.Decode(certPEM)
	var certBase64 string
	if blockCert != nil {
		certBase64 = base64.StdEncoding.EncodeToString(blockCert.Bytes)
	} else {
		// Fallback para limpeza manual se necessário
		certBase64 = string(certPEM)
		certBase64 = strings.ReplaceAll(certBase64, "-----BEGIN CERTIFICATE-----", "")
		certBase64 = strings.ReplaceAll(certBase64, "-----END CERTIFICATE-----", "")
		certBase64 = strings.ReplaceAll(certBase64, "\n", "")
		certBase64 = strings.ReplaceAll(certBase64, "\r", "")
	}

	// 5. Montar o Bloco Final usando a String de Transporte
	// O 'xmlns:ds' aqui vai no pai <Signature>, então o filho <SignedInfo> herda ele automaticamente no transporte.
	finalSignature := fmt.Sprintf(`<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">%s<ds:SignatureValue>%s</ds:SignatureValue><ds:KeyInfo><ds:X509Data><ds:X509Certificate>%s</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature>`,
		stringParaTransporte, signatureValue, certBase64)

	return finalSignature, nil
}
