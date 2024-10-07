package config

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sethvargo/go-envconfig"
)

type envs struct {
	ApiPort         string `env:"API_PORT, required"`
	ImageApiPath    string `env:"IMAGE_API_PATH"`
	BrokenImagePath string `env:"BROKEN_IMAGE_PATH"`
	MIS             string `env:"MAX_IMAGE_SIZE, required"`
	MaxImageSize    int64
	CacheType       string `env:"CACHE_TYPE, required"`
	CachePath       string `env:"CACHE_PATH"`
	CacheExpiration uint   `env:"CACHE_EXPIRATION"`
	RedisHost       string `env:"REDIS_HOST"`
	RedisPort       int    `env:"REDIS_PORT"`
	RedisPassword   string `env:"REDIS_PASSWORD"`
	RedisDB         int    `env:"REDIS_DB"`
	BrokenImageData []byte
}

var envList envs

func Init() *envs {
	log.Println("Initializing Image optimizer")
	if err := envconfig.Process(context.Background(), &envList); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}
	MaxImageSize, err := convertToBytes(envList.MIS)
	if err != nil {
		log.Fatalf("Invalid MAX_IMAGE_SIZE env value: %v\n", err)
	}
	envList.MaxImageSize = MaxImageSize

	if envList.CacheType != "file" && envList.CacheType != "redis" && envList.CacheType != "in-memory" {
		log.Fatalf("CACHE_TYPE env value is invalid\n")
	}
	if envList.CacheType == "file" && envList.CachePath == "" {
		log.Fatalf("CACHE_PATH env value is required when CACHE_TYPE is file\n")
	}
	if envList.CacheType == "redis" && (envList.RedisPort == 0) {
		log.Fatalf("REDIS_PORT env value is required when CACHE_TYPE is redis\n")
	}
	// Initialization Messages
	if envList.ImageApiPath == "" {
		envList.ImageApiPath = "/image"
	}
	log.Printf("Max image size: %s\n", envList.MIS)
	if envList.BrokenImagePath != "" {
		log.Printf("Default image: %s\n", envList.BrokenImagePath)
		// Load broken image
		_, err := url.ParseRequestURI(envList.BrokenImagePath)
		if err != nil {
			// Image is a local file
			envList.BrokenImageData, err = os.ReadFile(envList.BrokenImagePath)
			if err != nil {
				log.Fatalf("Error loading broken image: %v\n", err)
			}
		} else {
			// Image is a URL
			response, err := http.Get(envList.BrokenImagePath)
			if err != nil {
				log.Fatalf("Error loading broken image: %v\n", err)
			}
			if response.StatusCode != 200 {
				log.Fatalf("Error loading broken image: %d\n", response.StatusCode)
			}
			defer response.Body.Close()

			var imageBuffer bytes.Buffer
			_, err = imageBuffer.ReadFrom(response.Body)
			if err != nil {
				log.Fatalf("Error loading broken image: %v\n", err)
			}
			envList.BrokenImageData = imageBuffer.Bytes()

		}

	}

	log.Printf("Cache type: %s\n", envList.CacheType)
	if envList.CacheType == "file" {
		log.Printf("Your Images will be saved locally in the Hard Drive path: %s\n", envList.CachePath)
	}
	if envList.CacheType == "in-memory" {
		log.Printf("Your Images will be saved in the RAM Memory\n")
	}
	if envList.CacheType == "redis" {
		log.Printf("Your Images will be saved in the Redis cache\n")
	}
	log.Printf("API Image Path: %s\n", envList.ImageApiPath)

	return &envList
}

func GetEnvs() *envs {
	return &envList
}

func convertToBytes(size string) (int64, error) {
	size = strings.ToUpper(size)
	switch {
	case strings.Contains(size, "KB"):
		s, err := strconv.Atoi(size[:len(size)-2])
		if err != nil {
			return 0, err
		}
		return int64(s) * 1024, nil
	case strings.Contains(size, "MB"):
		s, err := strconv.Atoi(size[:len(size)-2])
		if err != nil {
			return 0, err
		}
		return int64(s) * 1024 * 1024, nil
	default:
		s, err := strconv.Atoi(size)
		if err != nil {
			return 0, err
		}
		return int64(s), nil
	}
}
