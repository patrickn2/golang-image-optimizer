package database

import "context"

type PkgDatabaseInterface interface {
	Set(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
}
