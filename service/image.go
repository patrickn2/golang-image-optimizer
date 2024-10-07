package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	Error     error
}

type OptimizeRequest struct {
	Ctx             context.Context
	ImageUrl        string
	Width           string
	Quality         string
	IfModifiedSince string
	CacheControl    string
}

func (is *ImageService) Optimize(or *OptimizeRequest) *OptimizeResponse {
	envs := config.GetEnvs()
	if or.Quality == "" {
		or.Quality = "75"
	}
	if or.ImageUrl == "" {
		return &OptimizeResponse{
			Error: ErrInvalidImageUrl,
		}
	}
	if or.Width == "" || or.Width == "0" {
		return &OptimizeResponse{
			Error: ErrInvalidImageWidth,
		}
	}
	w, err := strconv.Atoi(or.Width)
	if err != nil {
		return &OptimizeResponse{
			Error: ErrInvalidImageWidth,
		}
	}
	q, err := strconv.Atoi(or.Quality)
	if err != nil {
		return &OptimizeResponse{
			Error: ErrInvalidQuality,
		}
	}
	// Check if the quality is between 0 and 100
	if q < 0 || q > 100 {
		return &OptimizeResponse{
			Error: ErrInvalidQuality,
		}
	}
	// Check if url is valid
	u, err := url.ParseRequestURI(or.ImageUrl)
	if err != nil {
		return &OptimizeResponse{
			Error: ErrInvalidImageUrl,
		}
	}

	// Generate image name
	s := sha256.New()
	s.Write([]byte(or.ImageUrl))
	imageName := fmt.Sprintf("%x_%d_%d.webp", s.Sum(nil), w, q)

	// Convert If-Modified-Since header to time.Time
	var ifModified time.Time
	if or.IfModifiedSince != "" {
		ifModified, err = time.Parse(time.RFC1123, or.IfModifiedSince)
		if err != nil {
			ifModified = time.Now()
		}
	}

	// Check if image is in the cache
	optimizedImage, modified, err := is.ir.GetImage(or.Ctx, imageName)

	if err != nil {
		return &OptimizeResponse{
			Error: err,
		}
	}
	if optimizedImage != nil {
		// Check if image is modified and if cache is enabled and will return 304 Not Modified
		if modified != nil && modified.After(ifModified) && or.CacheControl != "no-cache" && or.CacheControl != "no-store" && or.CacheControl != "" {
			return &OptimizeResponse{
				Modified: modified,
				Cache:    true,
			}
		}
		return &OptimizeResponse{
			ImageData: optimizedImage,
			Modified:  modified,
			Cache:     true,
		}
	}

	// Check image size
	response, err := http.Head(u.String())
	if err != nil {
		return &OptimizeResponse{
			Error: ErrInvalidImageUrl,
		}
	}
	defer response.Body.Close()

	if response.ContentLength > int64(envs.MaxImageSize) {
		return &OptimizeResponse{
			Error: ErrInvalidImageSize,
		}
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "image/") {
		return &OptimizeResponse{
			Error: ErrInvalidImageType,
		}
	}

	// Download image
	response, err = http.Get(u.String())
	if err != nil {
		return &OptimizeResponse{
			Error: ErrInvalidImageUrl,
		}
	}
	defer response.Body.Close()

	var imageBuffer bytes.Buffer
	_, err = imageBuffer.ReadFrom(response.Body)
	if err != nil {
		return &OptimizeResponse{
			Error: err,
		}
	}
	// Resizing and compressing image to webp
	compressedImage, err := is.ic.CompressImage(imageBuffer.Bytes(), q, w)
	err = is.ir.SaveImage(or.Ctx, imageName, compressedImage)
	if err != nil {
		return &OptimizeResponse{
			Error: err,
		}
	}

	if modified == nil {
		m := time.Now().UTC()
		modified = &m
	}
	return &OptimizeResponse{
		ImageData: compressedImage,
		Modified:  modified,
		Cache:     false,
	}
}
