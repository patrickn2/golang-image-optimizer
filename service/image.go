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

	"github.com/patrickn2/go-image-optimizator/config"
	"github.com/patrickn2/go-image-optimizator/pkg/imagecompress"
	"github.com/patrickn2/go-image-optimizator/repository"
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
)

func (is *ImageService) Optimize(ctx context.Context, imageUrl string, width string, quality string) ([]byte, bool, error) {
	envs := config.GetEnvs()
	if quality == "" {
		quality = "75"
	}
	if imageUrl == "" {
		return nil, false, ErrInvalidImageUrl
	}
	if width == "" || width == "0" {
		return nil, false, ErrInvalidImageWidth
	}
	w, err := strconv.Atoi(width)
	if err != nil {
		return nil, false, ErrInvalidImageWidth
	}
	q, err := strconv.Atoi(quality)
	if err != nil {
		return nil, false, ErrInvalidQuality
	}
	// Check if the quality is between 0 and 100
	if q < 0 || q > 100 {
		return nil, false, ErrInvalidQuality
	}
	// Check if url is valid
	u, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return nil, false, ErrInvalidImageUrl
	}

	// Generate image name
	s := sha256.New()
	s.Write([]byte(imageUrl))
	imageName := fmt.Sprintf("%x_%d_%d.webp", s.Sum(nil), w, q)

	// Check if image is in the cache
	optimizedImage, err := is.ir.GetImage(ctx, imageName)
	if err != nil {
		return nil, false, err
	}
	if optimizedImage != nil {
		return optimizedImage, true, nil
	}

	// Check image size
	response, err := http.Head(u.String())
	if err != nil {
		return nil, false, ErrInvalidImageUrl
	}
	defer response.Body.Close()
	contentLength := response.ContentLength

	if contentLength > int64(envs.MaxImageSize) {
		return nil, false, ErrInvalidImageSize
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "image/") {
		return nil, false, ErrInvalidImageType
	}

	// Download image
	response, err = http.Get(u.String())
	if err != nil {
		return nil, false, ErrInvalidImageUrl
	}
	defer response.Body.Close()

	var imageBuffer bytes.Buffer
	_, err = imageBuffer.ReadFrom(response.Body)
	if err != nil {
		return nil, false, err
	}
	// Resizing and compressing image to webp
	compressedImage, err := is.ic.CompressImage(imageBuffer.Bytes(), q, w)
	err = is.ir.SaveImage(ctx, imageName, compressedImage)
	if err != nil {
		return nil, false, err
	}

	return compressedImage, false, nil
}
