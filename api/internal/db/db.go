package db

import "context"

type CloseDbFunc func(ctx context.Context) error
