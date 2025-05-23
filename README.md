# ClientHTTP

Uma biblioteca de cliente HTTP para Go que simplifica requisições HTTP com suporte para auditoria, rastreamento de IDs de correlação e manipulação flexível de requisições e respostas.

## Características

- Suporte completo para métodos HTTP (GET, POST, PUT, DELETE, PATCH)
- Adaptadores personalizáveis para auditoria de requisições
- Gerenciamento de IDs de correlação
- Configuração flexível via opções
- Manipulação simplificada de headers, query parameters e cookies
- Suporte para submissão de formulários

## Instalação

```bash
go get github.com/yourusername/clienthttp
```

## Uso Básico

### Inicialização do Cliente

```go
package main

import (
    "context"
    "fmt"
    "clienthttp/pkg/adapter"
    "clienthttp/pkg/clienthttp"
    "clienthttp/pkg/structs"
)

func main() {
    // Cria adaptadores para auditoria e correlação de ID
    auditAdapter := adapter.NewAuditoryAdapter()
    correlationAdapter := adapter.NewCorrelationIDAdapter()
    
    // Inicializa o cliente com a URL base
    client, err := clienthttp.NewClientHttp(
        "https://api.example.com",
        auditAdapter,
        correlationAdapter,
    )
    if err != nil {
        panic(err)
    }
    
    // Agora você pode usar o cliente para fazer requisições
}
```

### Fazendo uma requisição GET

```go
func makeGetRequest(client *clienthttp.ClientHttp) {
    ctx := context.Background()
    
    // Cria a requisição GET
    request := structs.GetRequest{
        Endpoint: "users",
        QueryParams: map[string]string{
            "page": "1",
            "limit": "10",
        },
        Headers: map[string]string{
            "Accept": "application/json",
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
func makePostRequest(client *clienthttp.ClientHttp) {
    ctx := context.Background()
    
    // Dados para enviar no corpo da requisição
    body := []byte(`{"name": "John", "email": "john@example.com"}`)
    
    // Cria a requisição POST
    request := structs.PostRequest{
        Endpoint: "users",
        Body: body,
        Headers: map[string]string{
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
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
func submitForm(client *clienthttp.ClientHttp) {
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
type MyAuditAdapter struct {}

func (a *MyAuditAdapter) Save(ctx context.Context, request *structs.Request, response *structs.Response) {
    // Lógica para salvar a auditoria (ex: log, banco de dados, etc)
    fmt.Printf("Audit: %s %s - Status: %d\n", request.Method, request.Url, response.StatusCode)
}

// Uso:
auditAdapter := &MyAuditAdapter{}
client, _ := clienthttp.NewClientHttp("https://api.example.com", auditAdapter, correlationAdapter)
```

## Configuração Avançada

A biblioteca suporta configuração através de opções:

```go
// Exemplo de configuração com opções personalizadas
client, err := clienthttp.NewClientHttp(
    "https://api.example.com",
    auditAdapter,
    correlationAdapter,
    clienthttp.WithTimeout(30), // timeout em segundos
    // outras opções aqui
)
```

## Exemplos Completos

Veja a pasta `example/` para exemplos completos de uso da biblioteca.

## Contribuição

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou PRs.

## Licença

[Especificar a licença aqui]

