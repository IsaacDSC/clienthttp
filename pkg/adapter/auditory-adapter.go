package adapter

import (
	"clienthttp/pkg/structs"
	"context"
)

type AuditoryAdapter interface {
	Save(ctx context.Context, request *structs.Request, response *structs.Response)
}
