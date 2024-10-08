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
	"regexp"
	"strings"
	"time"

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
	ErrInvalidImageWidth   = errors.New("invalid image width")
	ErrInvalidImageUrl     = errors.New("invalid image url")
	ErrDomainNotAuthorized = errors.New("domain not authorized")
	ErrInvalidImageType    = errors.New("invalid image type")
	ErrInvalidImageSize    = errors.New("invalid image size")
	ErrInvalidQuality      = errors.New("invalid quality")
	ErrNotModified         = errors.New("not modified")
)

type OptimizeResponse struct {
	ImageData []byte
	Modified  *time.Time
	Cache     bool
}

type OptimizeRequest struct {
	Ctx               context.Context
	ImageUrl          string
	Height            int
	Width             int
	Quality           int
	IfModifiedSince   string
	CacheControl      string
	BrokenImage       bool
	MaxImageSize      int64
	AuthorizedDomains string
}

func (is *ImageService) Optimize(or *OptimizeRequest) (*OptimizeResponse, error) {
	u, err := url.ParseRequestURI(or.ImageUrl)
	if err != nil {
		return nil, ErrInvalidImageUrl
	}

	if or.AuthorizedDomains != "" {
		if !regexp.MustCompile(or.AuthorizedDomains).MatchString(u.Host) {
			return nil, ErrDomainNotAuthorized
		}
	}

	// Generate image name
	s := sha256.New()
	s.Write([]byte(or.ImageUrl))
	imageName := fmt.Sprintf("%x_%d_%d_%d.webp", s.Sum(nil), or.Quality, or.Width, or.Height)

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

	if response.ContentLength > or.MaxImageSize {
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
	compressRequest := &imagecompress.CompressImageRequest{
		ImageData: imageBuffer.Bytes(),
		Quality:   or.Quality,
		Width:     or.Width,
		Height:    or.Height,
	}
	compressedImage, err := is.ic.CompressImage(compressRequest)
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

type BrokenImageRequest struct {
	Ctx             context.Context
	Quality         int
	Width           int
	Height          int
	BrokenImageData []byte
}

func (is *ImageService) BrokenImage(bir *BrokenImageRequest) (*OptimizeResponse, error) {
	brokenImageName := fmt.Sprintf("broken_%d_%d_%d.webp", bir.Quality, bir.Width, bir.Height)
	compressedImage, modified, err := is.ir.GetImage(bir.Ctx, brokenImageName)
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

	compressRequest := &imagecompress.CompressImageRequest{
		ImageData: bir.BrokenImageData,
		Quality:   bir.Quality,
		Width:     bir.Width,
		Height:    bir.Height,
	}

	compressedImage, err = is.ic.CompressImage(compressRequest)
	if err != nil {
		return nil, err
	}
	err = is.ir.SaveImage(bir.Ctx, brokenImageName, compressedImage)
	if err != nil {
		log.Printf("Error saving Broken Image to cache: %v\n", err)
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
