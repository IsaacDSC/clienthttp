package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IsaacDSC/clienthttp"

	"github.com/google/uuid"
)

func GetCorrelation(ctx context.Context) string {
	return uuid.New().String()
}

func main() {
	// Configurar log para mostrar o arquivo e linha nos logs de erro
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	// Usando httpbin.org como endpoint de teste que sabemos que está disponível
	baseURL := "https://httpbin.org"

	fmt.Println("Inicializando o cliente HTTP...")
	auditoryService := NewAuditoryService()

	// Verificar se auditoryService está correto
	if auditoryService == nil {
		log.Fatal("AuditoryService não pôde ser criado")
	}

	client, err := clienthttp.New(baseURL,
		clienthttp.WithAuditor(auditoryService),
		clienthttp.WithCorrelationID(GetCorrelation),
		clienthttp.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Erro ao criar cliente HTTP: %v", err)
	}

	fmt.Println("Cliente HTTP inicializado com sucesso!")

	// Criando o payload para teste
	payload := map[string]interface{}{
		"username":  "isaacdsc",
		"timestamp": time.Now().Unix(),
	}

	fmt.Println("Preparando payload JSON...")
	b, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Erro ao serializar JSON: %v", err)
	}

	fmt.Println("JSON gerado com sucesso:", string(b))

	fmt.Println("Enviando requisição POST...")
	// Usando o endpoint /post do httpbin.org para testar
	res, err := client.Post(ctx, "/post", b,
		clienthttp.WithQuery("test", "true"),
		clienthttp.WithHeader("Content-Type", "application/json"),
	)

	// Tratar erro na requisição detalhadamente
	if err != nil {
		log.Printf("Erro detalhado na requisição POST: %+v", err)
		os.Exit(1)
	}

	// Verificar resposta
	if res == nil {
		log.Fatal("A resposta recebida é nula")
	}

	fmt.Println("Resposta recebida com sucesso!")
	fmt.Println("Status code:", res.StatusCode)
	fmt.Println("Corpo da resposta:")
	fmt.Println(res.String())

	// Exemplo com GET e parsing JSON da resposta
	fmt.Println("\n--- Exemplo com GET e JSON ---")
	getRes, err := client.Get(ctx, "/get",
		clienthttp.WithQuery("example", "json-parsing"),
	)
	if err != nil {
		log.Printf("Erro no GET: %v", err)
	} else {
		fmt.Println("GET funcionou!")
		fmt.Printf("Status: %d\n", getRes.StatusCode)

		// Parse JSON manualmente - mais controle sobre erros
		var result map[string]interface{}
		if err := getRes.JSON(&result); err != nil {
			log.Printf("Erro ao parsear JSON: %v", err)
		} else {
			fmt.Printf("Args recebidos: %v\n", result["args"])
		}
	}

	fmt.Println("\nLog de auditoria salvo em 'auditory_payload.json'")
}
