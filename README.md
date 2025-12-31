# ClientHTTP

Uma biblioteca de cliente HTTP idiomática para Go que simplifica requisições HTTP com suporte para auditoria, rastreamento de IDs de correlação e manipulação flexível de requisições e respostas.

## Características

- API simples e idiomática
- Suporte completo para métodos HTTP (GET, POST, PUT, DELETE, PATCH)
- Adaptadores opcionais para auditoria de requisições
- Gerenciamento de IDs de correlação
- Configuração flexível via options (timeouts, connection pool, TLS/mTLS)
- Request options para headers, query params e autenticação por requisição
- Erros estruturados com suporte a `errors.Is` e `errors.As`

## Instalação

```bash
go get -u github.com/IsaacDSC/clienthttp
```

## Uso Básico

### Criando um Cliente

```go
package main

import (
    "context"
    "fmt"
    "clienthttp"
)

func main() {
    // Cria cliente com URL base
    client, err := clienthttp.New("https://api.example.com")
    if err != nil {
        panic(err)
    }
    
    // GET simples
    resp, err := client.Get(context.Background(), "/users")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.String())
}
```

### Enviando JSON

```go
// Marshal o body manualmente para ter controle total
data, _ := json.Marshal(User{Name: "John", Email: "john@example.com"})

resp, err := client.Post(ctx, "/users", data)
if err != nil {
    // tratar erro
}

// Unmarshal a resposta
var created User
if err := resp.JSON(&created); err != nil {
    // tratar erro de parsing
}
```

### Request Options

```go
// Query parameters
resp, err := client.Get(ctx, "/users",
    clienthttp.WithQuery("page", "1"),
    clienthttp.WithQuery("limit", "10"),
)

// Headers customizados
resp, err := client.Get(ctx, "/protected",
    clienthttp.WithHeader("X-Custom-Header", "value"),
    clienthttp.WithBearerToken("my-jwt-token"),
)

// Basic Auth
resp, err := client.Get(ctx, "/admin",
    clienthttp.WithBasicAuth("username", "password"),
)

// Múltiplos query params e headers de uma vez
resp, err := client.Get(ctx, "/search",
    clienthttp.WithQueries(map[string]string{
        "q":     "golang",
        "sort":  "stars",
        "order": "desc",
    }),
    clienthttp.WithHeaders(map[string]string{
        "Accept":          "application/json",
        "Accept-Language": "pt-BR",
    }),
)
```

### Formulários

```go
resp, err := client.PostForm(ctx, "/login", map[string]string{
    "username": "johndoe",
    "password": "secret",
})
```

### Trabalhando com Response

```go
resp, err := client.Get(ctx, "/users/123")
if err != nil {
    // tratar erro
}

// Verificar se foi sucesso (2xx)
if resp.OK() {
    // Unmarshal para struct
    var user User
    if err := resp.JSON(&user); err != nil {
        // tratar erro de parsing
    }
    
    // Ou como string
    body := resp.String()
    
    // Acessar headers
    contentType := resp.Headers.Get("Content-Type")
}
```

## Auditoria de Requisições

Implemente a interface `Auditor` para logging automático:

```go
type MyAuditor struct{}

func (a *MyAuditor) Log(ctx context.Context, req *clienthttp.AuditRequest, resp *clienthttp.AuditResponse) {
    log.Printf("%s %s -> %d", req.Method, req.URL, resp.StatusCode)
}

// Usar com o cliente
client, err := clienthttp.New("https://api.example.com",
    clienthttp.WithAuditor(&MyAuditor{}),
)
```

## Correlation ID

Configure uma função para extrair ou gerar correlation IDs:

```go
getCorrelationID := func(ctx context.Context) string {
    if id, ok := ctx.Value("correlation_id").(string); ok {
        return id
    }
    return uuid.New().String()
}

client, err := clienthttp.New("https://api.example.com",
    clienthttp.WithCorrelationID(getCorrelationID),
)
```

## Configuração Avançada

```go
import (
    "clienthttp"
    "crypto/tls"
    "time"
)

client, err := clienthttp.New("https://api.example.com",
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

    // Adapters
    clienthttp.WithAuditor(myAuditor),
    clienthttp.WithCorrelationID(correlationFunc),
    clienthttp.WithAuthCallback(addAuthHeader),
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
| `WithAuditor(a)` | Adaptador de auditoria | nil |
| `WithCorrelationID(fn)` | Função para correlation ID | nil |
| `WithAuthCallback(fn)` | Callback para autenticação | nil |
| `WithContentType(ct)` | Content-Type padrão | application/json |
| `WithCookies(cookies...)` | Cookies padrão | - |

### Request Options

| Option | Descrição |
|--------|-----------|
| `WithQuery(k, v)` | Adiciona query parameter |
| `WithQueries(map)` | Adiciona múltiplos query params |
| `WithHeader(k, v)` | Adiciona header |
| `WithHeaders(map)` | Adiciona múltiplos headers |
| `WithBearerToken(token)` | Adiciona Bearer token |
| `WithBasicAuth(user, pass)` | Adiciona Basic Auth |

## Tratamento de Erros

A biblioteca fornece erros estruturados:

```go
resp, err := client.Get(ctx, "/users/123")
if err != nil {
    // Verificar tipo de erro
    var httpErr *clienthttp.Error
    if errors.As(err, &httpErr) {
        fmt.Printf("Request to %s failed with status %d\n", 
            httpErr.URL, httpErr.StatusCode)
        fmt.Printf("Response body: %s\n", string(httpErr.Body))
    }
    
    // Verificar erros específicos
    if errors.Is(err, clienthttp.ErrInvalidURL) {
        fmt.Println("URL inválida")
    }
    if errors.Is(err, clienthttp.ErrRequestFailed) {
        fmt.Println("Requisição retornou status não-2xx")
    }
}
```

### Erros Disponíveis

| Erro | Descrição |
|------|-----------|
| `ErrInvalidURL` | URL base inválida |
| `ErrRequestFailed` | Requisição retornou status não-2xx |
| `ErrTimeout` | Requisição excedeu timeout |

## API Pública

### Tipos

- `Client` - Cliente HTTP
- `Response` - Resposta HTTP com métodos `OK()`, `JSON()`, `String()`
- `Auditor` - Interface para auditoria
- `AuditRequest` - Dados da requisição para auditoria
- `AuditResponse` - Dados da resposta para auditoria
- `Error` - Erro estruturado com contexto

### Métodos do Client

```go
// HTTP básico
Get(ctx, endpoint, opts...) (*Response, error)
Post(ctx, endpoint, body, opts...) (*Response, error)
Put(ctx, endpoint, body, opts...) (*Response, error)
Patch(ctx, endpoint, body, opts...) (*Response, error)
Delete(ctx, endpoint, opts...) (*Response, error)
Do(ctx, method, endpoint, body, opts...) (*Response, error)

// Form
PostForm(ctx, endpoint, data, opts...) (*Response, error)
```

## Exemplos

Veja a pasta `example/` para exemplos completos:

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
