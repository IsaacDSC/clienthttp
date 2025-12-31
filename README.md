# ClientHTTP

Uma biblioteca de cliente HTTP para Go que simplifica requisições HTTP com suporte para auditoria, rastreamento de IDs de correlação e manipulação flexível de requisições e respostas.

## Características

- Suporte completo para métodos HTTP (GET, POST, PUT, DELETE, PATCH)
- Adaptadores personalizáveis para auditoria de requisições
- Gerenciamento de IDs de correlação
- Configuração flexível via opções (timeouts, connection pool, TLS/mTLS)
- Manipulação simplificada de headers, query parameters e cookies
- Suporte para submissão de formulários
- Interface pública limpa com implementação interna encapsulada
- Erros sentinel para tratamento de erros consistente

## Instalação

```bash
go get -u github.com/IsaacDSC/clienthttp
```

## Uso Básico

### Inicialização do Cliente

```go
package main

import (
    "context"
    "fmt"
    "clienthttp"
)

func main() {
    // Função para obter correlation ID do contexto
    getCorrelation := func(ctx context.Context) string {
        return "my-correlation-id"
    }
    
    // Inicializa o cliente com a URL base
    client, err := clienthttp.New(
        "https://api.example.com",
        nil,              // adapter de auditoria (opcional)
        getCorrelation,   // função de correlation ID (opcional)
    )
    if err != nil {
        panic(err)
    }
    
    // Agora você pode usar o cliente para fazer requisições
}
```

### Fazendo uma requisição GET

```go
func makeGetRequest(client clienthttp.Client) {
    ctx := context.Background()
    
    // Cria a requisição GET
    request := clienthttp.GetRequest{
        BaseInput: clienthttp.BaseInput{
            Endpoint: "users",
            QueryParams: map[string]string{
                "page":  "1",
                "limit": "10",
            },
            Headers: map[string]string{
                "Accept": "application/json",
            },
        },
    }
    
    // Executa a requisição
    response, err := client.Get(ctx, request)
    if err != nil {
        fmt.Printf("Erro na requisição: %v\n", err)
        return
    }
    
    fmt.Printf("Status Code: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", string(response.Body))
}
```

### Fazendo uma requisição POST

```go
func makePostRequest(client clienthttp.Client) {
    ctx := context.Background()
    
    // Dados para enviar no corpo da requisição
    body := []byte(`{"name": "John", "email": "john@example.com"}`)
    
    // Cria a requisição POST
    request := clienthttp.PostRequest{
        BaseInput: clienthttp.BaseInput{
            Endpoint: "users",
            Headers: map[string]string{
                "Content-Type": "application/json",
            },
        },
        Body: body,
    }
    
    // Executa a requisição
    response, err := client.Post(ctx, request)
    if err != nil {
        fmt.Printf("Erro na requisição: %v\n", err)
        return
    }
    
    fmt.Printf("Status Code: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", string(response.Body))
}
```

### Submissão de Formulário

```go
func submitForm(client clienthttp.Client) {
    ctx := context.Background()
    
    // Dados do formulário
    formData := map[string]string{
        "username": "johndoe",
        "password": "secretpassword",
    }
    
    // Executa a requisição de formulário
    response, err := client.DoFormRequest(ctx, "login", formData)
    if err != nil {
        fmt.Printf("Erro na requisição: %v\n", err)
        return
    }
    
    fmt.Printf("Status Code: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", string(response.Body))
}
```

## Auditoria de Requisições

A biblioteca suporta auditoria automática de requisições e respostas através do adaptador de auditoria.

```go
// Implementação de um adaptador de auditoria personalizado
type MyAuditAdapter struct{}

func (a *MyAuditAdapter) Save(ctx context.Context, request *clienthttp.Request, response *clienthttp.Response) {
    // Lógica para salvar a auditoria (ex: log, banco de dados, etc)
    fmt.Printf("Audit: %s %s - Status: %d\n", request.Method, request.Url, response.StatusCode)
}

// Uso:
auditAdapter := &MyAuditAdapter{}
client, _ := clienthttp.New("https://api.example.com", auditAdapter, nil)
```

## Configuração Avançada

A biblioteca suporta configuração através de opções:

```go
import (
    "clienthttp"
    "crypto/tls"
    "time"
)

client, err := clienthttp.New(
    "https://api.example.com",
    auditAdapter,
    correlationFunc,
    // Timeouts
    clienthttp.WithTimeout(30*time.Second),
    clienthttp.WithDialTimeout(10*time.Second),
    clienthttp.WithTLSHandshakeTimeout(10*time.Second),
    clienthttp.WithResponseHeaderTimeout(10*time.Second),
    // Connection Pool
    clienthttp.WithMaxIdleConns(100),
    clienthttp.WithMaxIdleConnsPerHost(10),
    clienthttp.WithMaxConnsPerHost(100),
    clienthttp.WithIdleConnTimeout(90*time.Second),
    // TLS
    clienthttp.WithTLSMinVersion(tls.VersionTLS12),
    clienthttp.WithRootCA("/path/to/ca.pem"),
    clienthttp.WithClientCertificate("/path/to/cert.pem", "/path/to/key.pem"),
)
```

### Options Disponíveis

| Option | Descrição | Default |
|--------|-----------|---------|
| `WithTimeout(d)` | Timeout total da requisição | 30s |
| `WithDialTimeout(d)` | Timeout para estabelecer conexão | 10s |
| `WithTLSHandshakeTimeout(d)` | Timeout para handshake TLS | 10s |
| `WithResponseHeaderTimeout(d)` | Timeout para receber headers | 10s |
| `WithMaxIdleConns(n)` | Máximo de conexões idle | 100 |
| `WithMaxIdleConnsPerHost(n)` | Máximo de conexões idle por host | 10 |
| `WithMaxConnsPerHost(n)` | Máximo de conexões por host | 100 |
| `WithIdleConnTimeout(d)` | Timeout de conexão idle | 90s |
| `WithTLSConfig(config)` | Configuração TLS customizada | - |
| `WithTLSMinVersion(v)` | Versão TLS mínima | TLS 1.2 |
| `WithTLSMaxVersion(v)` | Versão TLS máxima | - |
| `WithRootCA(file)` | CA customizada via arquivo | - |
| `WithRootCAFromPEM(bytes)` | CA customizada via bytes | - |
| `WithClientCertificate(cert, key)` | Certificado client para mTLS | - |
| `WithClientCertificateFromPEM(cert, key)` | Certificado client via bytes | - |
| `WithInsecureSkipVerify()` | Desabilita verificação TLS | false |

## Tratamento de Erros

A biblioteca fornece erros sentinel para tratamento consistente:

```go
import "errors"

client, err := clienthttp.New("invalid-url", nil, nil)
if errors.Is(err, clienthttp.ErrInvalidBaseURL) {
    fmt.Println("URL inválida fornecida")
}

// Erros disponíveis:
// - clienthttp.ErrInvalidBaseURL    - URL base inválida
// - clienthttp.ErrRequestFailed     - Requisição falhou (status não 2xx)
// - clienthttp.ErrReadResponseBody  - Erro ao ler corpo da resposta
```

## API Pública

A biblioteca expõe apenas:

- **Interface `Client`** - para fazer requisições HTTP
- **Tipos de dados** - `Request`, `Response`, `GetRequest`, `PostRequest`, etc.
- **Interfaces** - `AuditoryAdapter`, `CorrelationIDAdapter`
- **Options** - funções `With*` para configuração
- **Erros sentinel** - para tratamento de erros
- **Constantes** - valores default de configuração

A implementação interna está protegida em `internal/` e não pode ser acessada.

## Exemplos Completos

Veja a pasta `example/` para exemplos completos de uso da biblioteca:

```bash
# Executar exemplo básico
go run ./example/

# Executar exemplo TLS/mTLS
cd example/tls
./certs/generate.sh  # Gerar certificados primeiro
go run ./server/     # Em um terminal
go run ./client/     # Em outro terminal
```

## Contribuição

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou PRs.

## Licença

MIT License
