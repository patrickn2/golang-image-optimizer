package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/memcachier/mc/v3"
	"github.com/redis/go-redis/v9"
)

type PkgDatabaseMemcache struct {
	conn            *mc.Client
	cacheExpiration uint
}

func NewDatabaseMemcache(host string, port int, username string, password string, ce uint) *PkgDatabaseMemcache {
	server := fmt.Sprintf("%s:%d", host, port)
	c := mc.NewMC(server, username, password)
	if _, err := c.Stats(); err != nil {
		log.Fatalf("Error connecting to Memcache Cache server: %v\n", err)
	}
	return &PkgDatabaseMemcache{
		conn:            c,
		cacheExpiration: ce,
	}
}

func (db *PkgDatabaseMemcache) Set(ctx context.Context, key string, data []byte) error {
	_, err := db.conn.Set(key+":created_at", time.Now().UTC().String(), 0, uint32(db.cacheExpiration*60), 0)
	if err != nil {
		return err
	}
	_, err = db.conn.Set(key, string(data), 0, uint32(db.cacheExpiration*60), 0)
	return err
}

func (db *PkgDatabaseMemcache) Get(ctx context.Context, key string) ([]byte, *time.Time, error) {
	modified, _, _, err := db.conn.Get(key + ":created_at")
	if err != nil {
		if err != mc.ErrNotFound {
			return nil, nil, err
		}
		return nil, nil, nil
	}
	modifiedTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", modified)
	data, _, _, err := db.conn.Get(key)
	if err != nil && err != redis.Nil {
		return nil, nil, err
	}
	return []byte(data), &modifiedTime, nil
}
