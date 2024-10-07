package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type PkgDatabaseFile struct {
	path string
}

func NewDatabaseFile(path string) *PkgDatabaseFile {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Cache directory %s does not exist, creating it\n", path)
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			log.Fatalf("Error creating cache directory: %v\n", err)
		}
	}
	return &PkgDatabaseFile{
		path: path,
	}
}

func (db *PkgDatabaseFile) Set(ctx context.Context, key string, data []byte) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", db.path, key))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (db *PkgDatabaseFile) Get(ctx context.Context, key string) ([]byte, *time.Time, error) {
	filePath := fmt.Sprintf("%s/%s", db.path, key)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, nil, nil
		}
	}
	modTime := fileInfo.ModTime().UTC()
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer file.Close()

	imageData, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, &modTime, err
	}
	return imageData, &modTime, nil
}
