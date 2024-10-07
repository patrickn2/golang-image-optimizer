package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type PkgDatabaseRedis struct {
	conn            *redis.Client
	cacheExpiration uint
}

func NewDatabaseRedis(host string, port int, password string, db int, ce uint) *PkgDatabaseRedis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Error connecting to Redis Cache server: %v\n", err)
	}
	return &PkgDatabaseRedis{
		conn:            rdb,
		cacheExpiration: ce,
	}
}

func (db *PkgDatabaseRedis) Set(ctx context.Context, key string, data []byte) error {
	err := db.conn.Set(ctx, key+":created_at", time.Now().UTC(), time.Minute*time.Duration(db.cacheExpiration)).Err()
	if err != nil {
		return err
	}
	return db.conn.Set(ctx, key, data, time.Minute*time.Duration(db.cacheExpiration)).Err()
}

func (db *PkgDatabaseRedis) Get(ctx context.Context, key string) ([]byte, *time.Time, error) {
	modified, err := db.conn.Get(ctx, key+":created_at").Time()
	if err != nil {
		if err != redis.Nil {
			return nil, nil, err
		}
		return nil, nil, nil
	}
	data, err := db.conn.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, nil, err
	}
	return data, &modified, nil
}
