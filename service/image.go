package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patrickn2/go-image-optimizer/config"
	"github.com/patrickn2/go-image-optimizer/pkg/imagecompress"
	"github.com/patrickn2/go-image-optimizer/repository"
)

type ImageService struct {
	ir *repository.ImageRepository
	ic imagecompress.PkgImgCompressInterface
}

func NewImageService(ic imagecompress.PkgImgCompressInterface, ir *repository.ImageRepository) *ImageService {
	return &ImageService{
		ir: ir,
		ic: ic,
	}
}

var (
	ErrInvalidImageWidth = errors.New("invalid image width")
	ErrInvalidImageUrl   = errors.New("invalid image url")
	ErrInvalidImageType  = errors.New("invalid image type")
	ErrInvalidImageSize  = errors.New("invalid image size")
	ErrInvalidQuality    = errors.New("invalid quality")
	ErrNotModified       = errors.New("not modified")
)

type OptimizeResponse struct {
	ImageData []byte
	Modified  *time.Time
	Cache     bool
}

type OptimizeRequest struct {
	Ctx             context.Context
	ImageUrl        string
	Width           int
	Quality         int
	IfModifiedSince string
	CacheControl    string
	BrokenImage     bool
}

func (is *ImageService) Optimize(or *OptimizeRequest) (*OptimizeResponse, error) {
	envs := config.GetEnvs()

	_, err := url.ParseRequestURI(or.ImageUrl)
	if err != nil {
		return nil, ErrInvalidImageUrl
	}

	// Generate image name
	s := sha256.New()
	s.Write([]byte(or.ImageUrl))
	imageName := fmt.Sprintf("%x_%d_%d.webp", s.Sum(nil), or.Width, or.Quality)

	// Check if image is in the cache
	optimizedImage, modified, err := is.ir.GetImage(or.Ctx, imageName)
	if err != nil {
		return nil, err
	}
	if optimizedImage != nil {
		// Convert If-Modified-Since header to time.Time
		var ifModified time.Time
		if or.IfModifiedSince != "" {
			ifModified, err = time.Parse(time.RFC1123, or.IfModifiedSince)
			if err != nil {
				ifModified = time.Now().UTC()
			}
		}
		// Check if image is modified and if cache is enabled and will return 304 Not Modified
		if modified != nil && modified.After(ifModified) && or.CacheControl != "no-cache" && or.CacheControl != "no-store" && or.CacheControl != "" {
			return &OptimizeResponse{
				Modified: modified,
				Cache:    true,
			}, nil
		}
		return &OptimizeResponse{
			ImageData: optimizedImage,
			Modified:  modified,
			Cache:     true,
		}, nil
	}

	// Check image size
	response, err := http.Head(or.ImageUrl)
	if err != nil || response.StatusCode != 200 {
		return nil, ErrInvalidImageUrl

	}
	defer response.Body.Close()

	if response.ContentLength > int64(envs.MaxImageSize) {
		return nil, ErrInvalidImageSize

	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "image/") {
		return nil, ErrInvalidImageType
	}

	// Download image
	response, err = http.Get(or.ImageUrl)
	if err != nil || response.StatusCode != 200 {
		return nil, ErrInvalidImageUrl
	}
	defer response.Body.Close()

	var imageBuffer bytes.Buffer
	_, err = imageBuffer.ReadFrom(response.Body)
	if err != nil {
		return nil, err
	}
	// Resizing and compressing image to webp
	compressedImage, err := is.ic.CompressImage(imageBuffer.Bytes(), or.Quality, or.Width)
	if err != nil {
		return nil, err
	}
	err = is.ir.SaveImage(or.Ctx, imageName, compressedImage)
	if err != nil {
		log.Printf("Error saving Image to cache: %v\n", err)
	}

	if modified == nil {
		m := time.Now().UTC()
		modified = &m
	}
	return &OptimizeResponse{
		ImageData: compressedImage,
		Modified:  modified,
		Cache:     false,
	}, nil
}

func (is *ImageService) BrokenImage(ctx context.Context, width int) (*OptimizeResponse, error) {
	envs := config.GetEnvs()
	brokenImageName := fmt.Sprintf("broken_%d_%d.webp", width, 75)
	compressedImage, modified, err := is.ir.GetImage(ctx, brokenImageName)
	if err != nil {
		return nil, err
	}
	if compressedImage != nil {
		return &OptimizeResponse{
			ImageData: compressedImage,
			Modified:  modified,
			Cache:     true,
		}, nil
	}

	compressedImage, err = is.ic.CompressImage(envs.BrokenImageData, 75, width)
	if err != nil {
		return nil, err
	}
	err = is.ir.SaveImage(ctx, brokenImageName, compressedImage)
	if err != nil {
		log.Printf("Error saving Image to cache: %v\n", err)
	}
	if modified == nil {
		m := time.Now().UTC()
		modified = &m
	}
	return &OptimizeResponse{
		ImageData: compressedImage,
		Modified:  modified,
		Cache:     false,
	}, nil
}
