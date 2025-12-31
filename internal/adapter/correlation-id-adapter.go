package adapter

import "context"

type CorrelationIDAdapter func(ctx context.Context) string
