# TLS/mTLS Example

Este exemplo demonstra como usar o `clienthttp` com configurações TLS/mTLS (mutual TLS).

## Estrutura

```
example/tls/
├── README.md           # Este arquivo
├── certs/
│   ├── generate.sh     # Script para gerar certificados
│   ├── ca.pem          # CA certificate (gerado)
│   ├── ca-key.pem      # CA private key (gerado)
│   ├── server.pem      # Server certificate (gerado)
│   ├── server-key.pem  # Server private key (gerado)
│   ├── client.pem      # Client certificate (gerado)
│   └── client-key.pem  # Client private key (gerado)
├── server/
│   └── main.go         # Servidor HTTPS com suporte mTLS
└── client/
    └── main.go         # Cliente usando clienthttp com TLS
```

## Pré-requisitos

- Go 1.21+
- OpenSSL (para gerar certificados)

## Como Usar

### 1. Gerar Certificados

Primeiro, gere os certificados de teste:

```bash
cd example/tls/certs
chmod +x generate.sh
./generate.sh
```

Isso criará:
- `ca.pem` / `ca-key.pem` - Certificado e chave da CA
- `server.pem` / `server-key.pem` - Certificado e chave do servidor
- `client.pem` / `client-key.pem` - Certificado e chave do cliente (para mTLS)

### 2. Iniciar o Servidor

**Modo TLS simples:**
```bash
cd example/tls/server
go run main.go
```

**Modo mTLS (requer certificado do cliente):**
```bash
cd example/tls/server
go run main.go -mtls
```

O servidor escuta em `https://localhost:8443` por padrão.

### 3. Executar o Cliente

**Conexão TLS simples:**
```bash
cd example/tls/client
go run main.go
```

**Conexão mTLS (com certificado do cliente):**
```bash
cd example/tls/client
go run main.go -mtls
```

**Modo inseguro (pular verificação de certificado):**
```bash
cd example/tls/client
go run main.go -insecure
```

## Cenários Demonstrados

### 1. TLS com CA Customizada

O cliente usa `WithRootCA()` para confiar no certificado self-signed do servidor:

```go
client, _ := clienthttp.NewClientHttp(serverURL, nil, nil,
    clienthttp.WithRootCA("certs/ca.pem"),
    clienthttp.WithTLSMinVersion(tls.VersionTLS12),
)
```

### 2. mTLS (Mutual TLS)

Autenticação bidirecional onde servidor e cliente verificam certificados mutuamente:

```go
client, _ := clienthttp.NewClientHttp(serverURL, nil, nil,
    clienthttp.WithRootCA("certs/ca.pem"),
    clienthttp.WithClientCertificate("certs/client.pem", "certs/client-key.pem"),
)
```

### 3. Skip Verification (Desenvolvimento)

Para desenvolvimento local sem certificados válidos:

```go
client, _ := clienthttp.NewClientHttp(serverURL, nil, nil,
    clienthttp.WithInsecureSkipVerify(),
)
```

⚠️ **AVISO:** Nunca use `WithInsecureSkipVerify()` em produção!

## Options Disponíveis

| Option | Descrição |
|--------|-----------|
| `WithRootCA(file)` | Adiciona CA customizada via arquivo PEM |
| `WithRootCAFromPEM(bytes)` | Adiciona CA customizada via bytes PEM |
| `WithClientCertificate(cert, key)` | Configura certificado client para mTLS |
| `WithClientCertificateFromPEM(cert, key)` | Configura certificado client via bytes |
| `WithTLSConfig(config)` | Usa configuração TLS customizada completa |
| `WithTLSMinVersion(version)` | Define versão TLS mínima (padrão: TLS 1.2) |
| `WithTLSMaxVersion(version)` | Define versão TLS máxima |
| `WithInsecureSkipVerify()` | Desabilita verificação de certificado |

## Testando com cURL

Você também pode testar o servidor com cURL:

**TLS simples:**
```bash
curl --cacert certs/ca.pem https://localhost:8443/
```

**mTLS:**
```bash
curl --cacert certs/ca.pem \
     --cert certs/client.pem \
     --key certs/client-key.pem \
     https://localhost:8443/
```

## Troubleshooting

### "certificate signed by unknown authority"
- Verifique se o CA certificate foi gerado corretamente
- Use `WithRootCA()` para especificar o CA

### "remote error: tls: bad certificate"
- O servidor está em modo mTLS mas o cliente não enviou certificado
- Use `WithClientCertificate()` para especificar o certificado do cliente

### "tls: client didn't provide a certificate"
- O servidor requer certificado do cliente (mTLS)
- Execute o cliente com `-mtls` flag

