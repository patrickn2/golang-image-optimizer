package database

import (
	"context"
	"log"
	"time"
)

type InMemObject struct {
	Data     []byte
	ExpireAt time.Time
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
		go dbInMem.checkExpiredImages()
	}

	return dbInMem
}

func (db *PkgDatabaseInMemory) Set(ctx context.Context, key string, data []byte) error {
	db.data[key] = InMemObject{
		Data:     data,
		ExpireAt: time.Now().Add(time.Minute * time.Duration(db.cacheExpiration)),
	}
	return nil
}

func (db *PkgDatabaseInMemory) Get(ctx context.Context, key string) ([]byte, error) {
	data, ok := db.data[key]
	if !ok {
		return nil, nil
	}
	return data.Data, nil
}

func (db *PkgDatabaseInMemory) checkExpiredImages() {
	for {
		time.Sleep(time.Minute)
		log.Println("Checking for expired images")
		for imageName, image := range db.data {
			if time.Now().After(image.ExpireAt) {
				log.Printf("Image %s expired\n", imageName)
				delete(db.data, imageName)
			}
		}
		log.Println("Done checking for expired images")
	}
}
