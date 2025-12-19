# NFE

API para gerar NFE.

<!-- ### Rotas
Rota|Função
|-----|-----|
/consulta|Consulta o CNPJ -->

## Estruturação de requisições
Daqui em diante irei tentar desmistificar como montar suas proprias requisições para utilizar em qualquer linguagem de programação e colocar direto e reto onde encontrar a informação.
### ROTA da API nota fiscal paulistana
Esta é a rota para todas as requisições

`https://nfews.prefeitura.sp.gov.br/lotenfe.asmx?WSDL` - *Pág. 17 do manual*

O que vai mudar serão 2 coisas. 
1. A `action` dentro do `Content-Type` que irá dizer o que você esta requisitando.
2. No corpo do envelope o nome da tag. Ex. `ConsultaCNPJRequest`.

#### Exemplo Corpo de envelope para uma requisição de consulta de CNPJ
```xml
<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	  <soap:Body>
	    <ConsultaCNPJRequest xmlns="http://www.prefeitura.sp.gov.br/nfe">
	      <VersaoSchema>1</VersaoSchema>
	      <MensagemXML><![CDATA[` + {XMLRequest} + `]]></MensagemXML>
	    </ConsultaCNPJRequest>
	  </soap:Body>
	</soap:Envelope>
```
> [!NOTE]
> Perceba que iremos criar 2 XML, e colocar um dentro do outro, porém o interno será acrescentado como `string`

#### Exemplo de MensagemXML para consulta de CNPJ
```xml
<p1:PedidoConsultaCNPJ xmlns:p1="http://www.prefeitura.sp.gov.br/nfe"><Cabecalho Versao="1"><CPFCNPJRemetente><CNPJ>{CNPJ}</CNPJ></CPFCNPJRemetente></Cabecalho><CNPJContribuinte><CNPJ>{CNPJ}</CNPJ></CNPJContribuinte></p1:PedidoConsultaCNPJ>
```
> [!IMPORTANT]
> O Prefixo `"p1"` esta sendo utilizado para que o mesmo não herde `tags` do seu pai e seus filhos não herdem suas `tags` visto que isso poderá causar um erro.

:shipit: - Vamos voltar aqui depois, mas coloquei o exemplo porque eu sei que o que a gente quer é ver logo o código e deduzir a solução para partirmos logo para o próximo bug. Go Horses!

### Protocolo SOAP
Vamos começar pela formatação da requisição,a nota fiscal paulistana aceita apenas o formato SOAP. O protocolo SOAP só funciona através do envio de envelopes o que significa que você sempre utilizará o método `POST`. Existem 2 versões de SOAP, 1.1 e 1.2, ambas são aceitas pelo sistema da prefeitura. Portanto vamos preparar nossas requisições para ambos os casos. Se lembra do exemplo de envelope acima, vamo analisa-lo.
#### SOAP 1.1
1. Declare a `tag` `xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"` no elemento raiz `<Envelope>` para especificar o namespace do protocolo SOAP 1.1 e garantir que o servidor interprete corretamente a estrutura da mensagem XML. O namespace é essencial para padronizar a comunicação e evitar conflitos de nomes de elementos no documento XML.
2. Na requisição você precisará declarar também no header o tipo que esta enviando, e a ação que espera que o servidor realize com os dados que enviou para ele. Portanto header fica assim: `Content-Type:text/xml; charset=utf-8; action={CONSULTA/EMISSÃO/CANCELAMENTO}`.

#### SOAP 1.2
1. Declare a `xmlns:soap="http://www.w3.org/2003/05/soap-envelope"` no elemento raiz `<Envelope>` para especificar o namespace do protocolo SOAP 1.2 e garantir que o servidor interprete corretamente a estrutura da mensagem XML. O namespace é essencial para padronizar a comunicação e evitar conflitos de nomes de elementos no documento XML.
2. Na requisição você precisará declarar também no header o tipo que esta enviando, e a ação que espera que o servidor realize com os dados que enviou para ele. Portanto header fica assim: `Content-Type:application/soap+xml; charset=utf-8; action={CONSULTA/EMISSÃO/CANCELAMENTO}`.

#### Exemplo
`Content-Type:application/soap+xml; charset=utf-8; action=http://www.prefeitura.sp.gov.br/nfe/ws/ConsultaCNPJ`

#### Como saber qual url da action que desejo?
Eu também não sei, pedi para a IA me responder isso e deu certo.

### Como criar um corpo de mensagem para solicitar algo ao servidor
Vamos começar entendendo o que precisará ser enviado, será um envelope XML com outro XML encapsulado dentro dele. Por mais estranho que pareça, esta é uma prática comum de arquitetura, isso ocorre porque vamos transportar uma mensagem que é "original" que não pode ser alterada, por isso ela é encapsulada dentro de uma estrutura maior, o envelope transporta metadados como, remetende, data, tipo de conteúdo, enquanto a mensagem possui apenas 2 informações, a requisição e a assinatura de quem esta pedindo, é uma autentificação que garante que quem enviou foi o próprio solicitante.
Portanto vamos nomear esses 2 XML como:
- XML Envelope
- XML Requisição

### Como criar um envelope para uma requisição
Para saber como é a formatação de um envelope da requisição que você deseja realizar, no manual há os exemplos de envelope para cada tipo de requisição.

*Ex. Requisição de Consulta de CNPJ pág. 67 do manual*

### Como criar uma MensagemXML
Utilizando modelos, denominados **schemas**, que especificam os campos obrigatórios e os tipos de informação esperados para um determinado documento. Eles atuam como um gabarito estrutural, comumente implementado em arquivos com a extensão .XSD.
#### O que são .XSD
XSD (XML Schema Definition) é uma linguagem baseada em XML que define a estrutura, o conteúdo e as regras de um documento XML.
#### Onde encontro estes esquemas?
Desça ao fim da página e abra a aba **Baixar Schemas XML**

[Schemas](https://notadomilhao.sf.prefeitura.sp.gov.br/desenvolvedor/)
#### Como fazer um XML a partir do XSD?
Com certeza existe uma maneira de ler esses schemas, as quais eu admito que não sei, o que fiz foi ler o nome do arquivo ver que era o que eu precisava e pedir para uma IA criar um arquivo XML baseado nessas regras do .XSD, dessa forma tendo o arquivo que eu precisava colocar dentro do envelope.
#### Com o XML criado onde devo usa-lo?
Dentro do envelope haverá uma TAG `<MensagemXML></MensagemXML>` dentro dela você precisará colocar esse xml que você criou, mas como `string`, você pode fazer isso utilizando `<![CDATA[` + {XMLRequest} + `]]>.`

### Como assinar o XML Requisição
#### Objetos da chave PEM que devem estar na requisição
- `<KeyValue>`
- `<RSAKeyValue>`
- `<Modulus>`
- `<Exponent>`

#### Objetos da chave PEM que NÃO devem estar na requisição
- `<X509SubjectName>`
- `<X509IssuerSerial>`
- `<X509IssuerName>`
- `<X509SerialNumber>`
- `<X509SKI>`
#### Assinando certificado TLS 
#### Assinando MensagemXML
## Documentation

Neste link você encontra: 
- Manual de Utilização do Webservice de NFS-e
- NFS-e – Reforma tributária 2026 (serviços síncronos e assíncronos) – Atualizado em 04/11/2025!
[Documentação (Manual) e Models (Structs) WebService Nota Fiscal Paulistana (Antiga Nota do Milhão)](https://notadomilhao.sf.prefeitura.sp.gov.br/desenvolvedor/)
## License

[CC BY-NC](https://creativecommons.org/licenses/by-nc/4.0/legalcode)

