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
	"slices"
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
	ErrTimeout             = errors.New("image download timeout")
)

type OptimizeResponse struct {
	ImageData   []byte
	ImageFormat string
	Modified    *time.Time
	Cache       bool
}

type OptimizeRequest struct {
	Ctx                  context.Context
	ImageUrl             string
	Height               int
	Width                int
	Quality              int
	IfModifiedSince      string
	CacheControl         string
	BrokenImage          bool
	MaxImageSize         int64
	AuthorizedDomains    string
	ImageDownloadTimeout int
	AcceptedFormats      []string
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

	acceptWebp := slices.Contains[[]string, string](or.AcceptedFormats, "image/webp")

	// Generate image name
	s := sha256.New()
	s.Write([]byte(or.ImageUrl))
	imageName := fmt.Sprintf("%x_%d_%d_%d_%v", s.Sum(nil), or.Quality, or.Width, or.Height, acceptWebp)

	// Check if image is in the cache
	optimizedImage, modified, err := is.ir.GetImage(or.Ctx, imageName)
	if err != nil {
		return nil, err
	}
	if optimizedImage != nil {
		log.Println("Image found in cache", imageName)
		// Convert If-Modified-Since header to time.Time
		var ifModified time.Time
		if or.IfModifiedSince != "" {
			ifModified, err = time.Parse(time.RFC1123, or.IfModifiedSince)
			if err != nil {
				ifModified = time.Now().UTC()
			}
		}
		log.Println("Modified", modified)
		// Check if image is modified and if cache is enabled and will return 304 Not Modified
		if modified != nil && modified.After(ifModified) && or.CacheControl != "no-cache" && or.CacheControl != "no-store" && or.CacheControl != "" {
			log.Println("Returning 304")
			return &OptimizeResponse{
				Modified: modified,
				Cache:    true,
			}, nil
		}
		log.Println("Returning Cache")
		return &OptimizeResponse{
			ImageData:   optimizedImage,
			ImageFormat: http.DetectContentType(optimizedImage),
			Modified:    modified,
			Cache:       true,
		}, nil
	}

	// Check image size
	httpClient := &http.Client{
		Timeout: time.Duration(or.ImageDownloadTimeout) * time.Second,
	}
	// Download Image Header
	head, err := httpClient.Head(or.ImageUrl)
	if err != nil || head.StatusCode != 200 {
		if err == http.ErrHandlerTimeout {
			return nil, ErrTimeout
		}
		return nil, ErrInvalidImageUrl
	}
	defer head.Body.Close()

	// Check Image Size
	if head.ContentLength > or.MaxImageSize {
		return nil, ErrInvalidImageSize
	}
	// Check if it is an image
	if !strings.HasPrefix(head.Header.Get("Content-Type"), "image/") {
		return nil, ErrInvalidImageType
	}

	// Download image
	res, err := httpClient.Get(or.ImageUrl)
	if err != nil || res.StatusCode != 200 {
		if err == http.ErrHandlerTimeout {
			return nil, ErrTimeout
		}
		return nil, ErrInvalidImageUrl
	}
	defer res.Body.Close()

	var imageBuffer bytes.Buffer
	_, err = imageBuffer.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	// Check Image Type Again (Protection against type manipulation)
	downloadedImageRealType := http.DetectContentType(imageBuffer.Bytes())
	if !strings.HasPrefix(downloadedImageRealType, "image/") {
		return nil, ErrInvalidImageType
	}

	newImageType := chooseImageFormat(downloadedImageRealType, or.AcceptedFormats)
	log.Println("Downloaded Image Type", downloadedImageRealType, "New Image Type", newImageType)

	// Resizing and compressing image
	compressRequest := &imagecompress.CompressImageRequest{
		ImageData: imageBuffer.Bytes(),
		Quality:   or.Quality,
		Width:     or.Width,
		Height:    or.Height,
		NewType:   newImageType,
	}
	compressedImage, err := is.ic.CompressImage(compressRequest)
	if err != nil {
		return nil, err
	}

	// If the mew image is bigger than the original image, save the old image instead of the new one
	if len(compressedImage) > imageBuffer.Len() && newImageType == downloadedImageRealType {
		compressedImage = imageBuffer.Bytes()
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
		ImageData:   compressedImage,
		ImageFormat: newImageType,
		Modified:    modified,
		Cache:       false,
	}, nil
}

type BrokenImageRequest struct {
	Ctx             context.Context
	Quality         int
	Width           int
	Height          int
	BrokenImageData []byte
	AcceptedFormats []string
}

func (is *ImageService) BrokenImage(bir *BrokenImageRequest) (*OptimizeResponse, error) {
	acceptWebp := slices.Contains[[]string, string](bir.AcceptedFormats, "image/webp")
	brokenImageName := fmt.Sprintf("broken_%d_%d_%d_%v", bir.Quality, bir.Width, bir.Height, acceptWebp)
	compressedImage, modified, err := is.ir.GetImage(bir.Ctx, brokenImageName)
	if err != nil {
		return nil, err
	}
	if compressedImage != nil {
		return &OptimizeResponse{
			ImageData:   compressedImage,
			ImageFormat: http.DetectContentType(compressedImage),
			Modified:    modified,
			Cache:       true,
		}, nil
	}

	newImageType := chooseImageFormat(http.DetectContentType(bir.BrokenImageData), bir.AcceptedFormats)

	compressRequest := &imagecompress.CompressImageRequest{
		ImageData: bir.BrokenImageData,
		Quality:   bir.Quality,
		Width:     bir.Width,
		Height:    bir.Height,
		NewType:   newImageType,
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
		ImageData:   compressedImage,
		ImageFormat: newImageType,
		Modified:    modified,
		Cache:       false,
	}, nil
}

func chooseImageFormat(imageFormat string, acceptFormats []string) string {
	var acceptAvif bool
	if imageFormat == "image/svg+xml" {
		return "image/svg+xml"
	}
	for _, format := range acceptFormats {
		switch format {
		case "image/webp":
			return "image/webp"
		case "image/avif":
			acceptAvif = true
		}
	}
	switch imageFormat {
	case "image/webp":
		return "image/png"
	case "image/avif":
		if acceptAvif {
			return "image/avif"
		}
		return "image/png"
	default:
		return imageFormat
	}
}
