package main

import (
	"github.com/patrickn2/go-image-optimizator/config"
	"github.com/patrickn2/go-image-optimizator/handler"
	"github.com/patrickn2/go-image-optimizator/httpserver"
	"github.com/patrickn2/go-image-optimizator/pkg/database"
	"github.com/patrickn2/go-image-optimizator/pkg/imagecompress"
	"github.com/patrickn2/go-image-optimizator/repository"
	"github.com/patrickn2/go-image-optimizator/service"
)

func main() {
	envs := config.Init()
	ic := imagecompress.NewImageCompress()
	var db database.PkgDatabaseInterface

	switch envs.CacheType {
	case "file":
		db = database.NewDatabaseFile(envs.CachePath)
	case "redis":
		db = database.NewDatabaseRedis(envs.RedisHost, envs.RedisPort, envs.RedisPassword, envs.RedisDB, envs.CacheExpiration)
	case "in-memory":
		db = database.NewDatabaseInMemory(envs.CacheExpiration)
	}

	imageRepository := repository.NewImageRepository(db)
	imageService := service.NewImageService(ic, imageRepository)
	h := handler.New(imageService)
	httpserver.Start(h, envs.ApiPort, envs.ImageApiPath)
}
