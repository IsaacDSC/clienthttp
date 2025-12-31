package adapter

import (
	"clienthttp/internal/structs"
	"context"
)

type AuditoryAdapter interface {
	Save(ctx context.Context, request *structs.Request, response *structs.Response)
}
