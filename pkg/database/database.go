package database

import (
	"context"
	"time"
)

type PkgDatabaseInterface interface {
	Set(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, *time.Time, error)
}
