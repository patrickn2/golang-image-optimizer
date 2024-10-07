package database

import (
	"context"
	"log"
	"time"
)

type InMemObject struct {
	Data      []byte
	CreatedAt time.Time
	ExpireAt  time.Time
}

type PkgDatabaseInMemory struct {
	data            map[string]InMemObject
	cacheExpiration uint
}

func NewDatabaseInMemory(cacheExpiration uint) *PkgDatabaseInMemory {
	dbInMem := &PkgDatabaseInMemory{
		data:            make(map[string]InMemObject),
		cacheExpiration: cacheExpiration,
	}
	if cacheExpiration != 0 {
		go dbInMem.checkExpiredData()
	}

	return dbInMem
}

func (db *PkgDatabaseInMemory) Set(ctx context.Context, key string, data []byte) error {
	db.data[key] = InMemObject{
		Data:      data,
		CreatedAt: time.Now().UTC(),
		ExpireAt:  time.Now().UTC().Add(time.Minute * time.Duration(db.cacheExpiration)),
	}
	return nil
}

func (db *PkgDatabaseInMemory) Get(ctx context.Context, key string) ([]byte, *time.Time, error) {
	data, ok := db.data[key]
	if !ok {
		return nil, nil, nil
	}
	return data.Data, &data.CreatedAt, nil
}

func (db *PkgDatabaseInMemory) checkExpiredData() {
	for {
		time.Sleep(time.Minute)
		log.Println("Checking for expired data")
		for imageName, image := range db.data {
			if time.Now().After(image.ExpireAt) {
				log.Printf("Data %s expired\n", imageName)
				delete(db.data, imageName)
			}
		}
		log.Println("Done checking for expired data")
	}
}
