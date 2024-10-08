package main

import (
	"github.com/patrickn2/go-image-optimizer/config"
	"github.com/patrickn2/go-image-optimizer/handler"
	"github.com/patrickn2/go-image-optimizer/httpserver"
	"github.com/patrickn2/go-image-optimizer/pkg/database"
	"github.com/patrickn2/go-image-optimizer/pkg/imagecompress"
	"github.com/patrickn2/go-image-optimizer/repository"
	"github.com/patrickn2/go-image-optimizer/service"
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
	case "memcache":
		db = database.NewDatabaseMemcache(envs.MemcacheHost, envs.MemcachePort, envs.MemcacheUser, envs.MemcachePassword, envs.CacheExpiration)
	case "in-memory":
		db = database.NewDatabaseInMemory(envs.CacheExpiration)
	}

	imageRepository := repository.NewImageRepository(db)
	imageService := service.NewImageService(ic, imageRepository)
	h := handler.New(imageService, envs)
	httpserver.Start(h, envs.ApiPort, envs.ImageApiPath)
}
