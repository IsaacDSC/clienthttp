# Retry Feature - Especificação de Abordagens

## Objetivo

Implementar uma feature de retry para o cliente HTTP que seja:
- Thread-safe (sem race conditions)
- Idiomática em Go
- Configurável e extensível

---

## Abordagem 1: Retry via Options (Functional Options Pattern)

Utiliza o padrão de options já existente no cliente para configurar retry.

### Estrutura

```go
// retryConfig holds retry configuration
type retryConfig struct {
    maxAttempts   int
    initialDelay  time.Duration
    maxDelay      time.Duration
    multiplier    float64
    retryableFunc func(resp *Response, err error) bool
}
```

### Options

```go
// WithRetry configures the retry behavior
func WithRetry(maxAttempts int) Option {
    return func(c *config) {
        c.retry.maxAttempts = maxAttempts
    }
}

// WithRetryBackoff configures exponential backoff
func WithRetryBackoff(initial, max time.Duration, multiplier float64) Option {
    return func(c *config) {
        c.retry.initialDelay = initial
        c.retry.maxDelay = max
        c.retry.multiplier = multiplier
    }
}

// WithRetryCondition allows custom retry logic
func WithRetryCondition(fn func(resp *Response, err error) bool) Option {
    return func(c *config) {
        c.retry.retryableFunc = fn
    }
}
```

### Implementação no Do()

```go
func (c *Client) Do(ctx context.Context, method, endpoint string, body []byte, opts ...RequestOption) (*Response, error) {
    var lastErr error
    var lastResp *Response

    for attempt := 0; attempt <= c.config.retry.maxAttempts; attempt++ {
        if attempt > 0 {
            delay := c.calculateBackoff(attempt)
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(delay):
            }
        }

        resp, err := c.doRequest(ctx, method, endpoint, body, opts...)
        if err == nil && !c.shouldRetry(resp, nil) {
            return resp, nil
        }

        lastResp = resp
        lastErr = err

        if !c.shouldRetry(resp, err) {
            break
        }
    }

    return lastResp, lastErr
}
```

### Prós
- Integra naturalmente com o padrão existente
- Configuração imutável após criação do cliente
- Sem estado compartilhado entre goroutines

### Contras
- Configuração global para todas as requests
- Menos flexível para casos específicos

---

## Abordagem 2: Retry via RequestOption (Per-Request)

Permite configurar retry individualmente por request.

### Estrutura

```go
type requestConfig struct {
    headers     http.Header
    queryParams map[string]string
    retry       *retryConfig  // novo campo
}
```

### RequestOption

```go
// WithRequestRetry configures retry for a specific request
func WithRequestRetry(maxAttempts int) RequestOption {
    return func(rc *requestConfig) {
        if rc.retry == nil {
            rc.retry = &retryConfig{}
        }
        rc.retry.maxAttempts = maxAttempts
    }
}
```

### Uso

```go
resp, err := client.Get(ctx, "/api/data", 
    WithRequestRetry(3),
    WithRequestBackoff(time.Second, 30*time.Second, 2.0),
)
```

### Prós
- Controle granular por request
- Composição flexível

### Contras
- Pode aumentar a complexidade do código

---

## Abordagem 3: Retry com Interface Strategy

Utiliza uma interface para permitir estratégias de retry customizadas.

### Interface

```go
// RetryStrategy defines how retries should be performed
type RetryStrategy interface {
    // ShouldRetry determines if the request should be retried
    ShouldRetry(attempt int, resp *Response, err error) bool
    
    // NextDelay returns the delay before the next attempt
    NextDelay(attempt int) time.Duration
}
```

### Implementações Built-in

```go
// ExponentialBackoff implements RetryStrategy with exponential backoff
type ExponentialBackoff struct {
    MaxAttempts  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func (e *ExponentialBackoff) ShouldRetry(attempt int, resp *Response, err error) bool {
    if attempt >= e.MaxAttempts {
        return false
    }
    return isRetryableError(err) || isRetryableStatus(resp)
}

func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
    delay := time.Duration(float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt)))
    if delay > e.MaxDelay {
        return e.MaxDelay
    }
    return delay
}

// ConstantBackoff implements RetryStrategy with constant delay
type ConstantBackoff struct {
    MaxAttempts int
    Delay       time.Duration
}

func (c *ConstantBackoff) ShouldRetry(attempt int, resp *Response, err error) bool {
    return attempt < c.MaxAttempts && (isRetryableError(err) || isRetryableStatus(resp))
}

func (c *ConstantBackoff) NextDelay(attempt int) time.Duration {
    return c.Delay
}
```

### Configuração

```go
func WithRetryStrategy(strategy RetryStrategy) Option {
    return func(c *config) {
        c.retryStrategy = strategy
    }
}
```

### Prós
- Altamente extensível
- Permite estratégias customizadas (circuit breaker, jitter, etc.)
- Testável via mock

### Contras
- Mais complexo para casos simples

---

## Funções Auxiliares (Comum a todas abordagens)

```go
// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
    if err == nil {
        return false
    }
    
    // Network errors são geralmente retryable
    var netErr net.Error
    if errors.As(err, &netErr) {
        return netErr.Temporary() || netErr.Timeout()
    }
    
    // Context cancelado não deve retry
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        return false
    }
    
    return false
}

// isRetryableStatus checks if a status code is retryable
func isRetryableStatus(resp *Response) bool {
    if resp == nil {
        return false
    }
    
    switch resp.StatusCode {
    case http.StatusTooManyRequests,        // 429
         http.StatusServiceUnavailable,      // 503
         http.StatusGatewayTimeout,          // 504
         http.StatusBadGateway:              // 502
        return true
    }
    
    return false
}

// calculateDelayWithJitter adds jitter to prevent thundering herd
func calculateDelayWithJitter(baseDelay time.Duration, jitterFactor float64) time.Duration {
    jitter := time.Duration(rand.Float64() * jitterFactor * float64(baseDelay))
    return baseDelay + jitter
}
```

---

## Garantias de Thread Safety

### 1. Sem Estado Compartilhado Mutável

```go
// A configuração é imutável após criação
type config struct {
    retry retryConfig  // valor, não ponteiro
}
```

### 2. Contexto para Cancelamento

```go
func (c *Client) doWithRetry(ctx context.Context, ...) (*Response, error) {
    for attempt := range c.config.retry.maxAttempts {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            // proceed with request
        }
        
        // ...
    }
}
```

### 3. Deep Copy de Body
```go
// Body é passado por valor ([]byte copiado) em cada tentativa
func (c *Client) Do(ctx context.Context, method, endpoint string, body []byte, opts ...RequestOption) (*Response, error) {
    // bytes.NewReader cria um novo reader para cada tentativa
    // sem compartilhar estado
}
```

---

## Recomendação

**Para este projeto, recomendo a Abordagem 2 (Retry via RequestOption)** combinada com valores default da Abordagem 1:

1. Define defaults sensatos no cliente via `WithRetry()`
2. Permite override por request via `WithRequestRetry()`
3. Mantém compatibilidade com o código existente
4. Segue o padrão idiomático já estabelecido no projeto

### Exemplo de Uso Final

```go
// Cliente com retry padrão
client, _ := clienthttp.New("https://api.example.com",
    clienthttp.WithRetry(3),
    clienthttp.WithRetryBackoff(time.Second, 30*time.Second, 2.0),
)

// Request com retry customizado
resp, err := client.Get(ctx, "/important-data",
    clienthttp.WithRequestRetry(5),  // override para 5 tentativas
)

// Request sem retry
resp, err := client.Post(ctx, "/idempotent-action", body,
    clienthttp.WithRequestRetry(0),  // desabilita retry
)
```

---

## Referências

- [Exponential Backoff And Jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
- [Go HTTP Client Best Practices](https://blog.golang.org/http-tracing)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
