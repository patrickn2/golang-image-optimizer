package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/patrickn2/go-image-optimizer/config"
	"github.com/patrickn2/go-image-optimizer/service"
)

type Handler struct {
	is   *service.ImageService
	envs *config.Envs
}

func New(is *service.ImageService, e *config.Envs) *Handler {
	return &Handler{
		is:   is,
		envs: e,
	}
}

func (h *Handler) OptimizeImage(w http.ResponseWriter, r *http.Request) {
	imageUrl := r.URL.Query().Get("url")
	width := r.URL.Query().Get("w")
	height := r.URL.Query().Get("h")
	quality := r.URL.Query().Get("q")
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	cacheControl := r.Header.Get("Cache-Control")
	accept := r.Header.Get("Accept")

	formats := strings.Split(accept, ",")

	intQuality, err := strconv.Atoi(quality)
	if err != nil || intQuality < 0 || intQuality > 100 {
		intQuality = h.envs.DefaultQuality
	}

	intWidth, err := strconv.Atoi(width)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if intWidth < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	intHeight, err := strconv.Atoi(height)
	if err != nil || intHeight < 1 {
		intHeight = 0
	}

	request := &service.OptimizeRequest{
		Ctx:                  r.Context(),
		ImageUrl:             imageUrl,
		Width:                intWidth,
		Height:               intHeight,
		Quality:              intQuality,
		IfModifiedSince:      ifModifiedSince,
		CacheControl:         cacheControl,
		MaxImageSize:         h.envs.MaxImageSize,
		AuthorizedDomains:    h.envs.AuthorizedHostnames,
		ImageDownloadTimeout: h.envs.ImageDownloadTimeout,
		AcceptedFormats:      formats,
	}

	optimizedResponse, err := h.is.Optimize(request)
	if err != nil {
		switch err {
		case service.ErrDomainNotAuthorized:
			log.Printf("%v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		case service.ErrInvalidImageUrl:
			brokenImageRequest := &service.BrokenImageRequest{
				Ctx:             r.Context(),
				BrokenImageData: h.envs.BrokenImageData,
				Quality:         intQuality,
				Width:           intWidth,
				Height:          intHeight,
				AcceptedFormats: formats,
			}
			optimizedResponse, err = h.is.BrokenImage(brokenImageRequest)
			if err != nil {
				log.Printf("Error optimizing image: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			log.Printf("Error optimizing image: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	cacheMsg := "MISS"
	if optimizedResponse.Cache {
		cacheMsg = "HIT"
	}
	ageSeconds := 0
	if optimizedResponse.Modified != nil && optimizedResponse.Cache {
		ageSeconds = int(time.Since(*optimizedResponse.Modified).Seconds())
	}

	lastModified := time.Now().UTC()
	if optimizedResponse.Modified != nil {
		lastModified = *optimizedResponse.Modified
	}
	w.Header().Set("Date", lastModified.Format(http.TimeFormat))
	w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))
	w.Header().Set("Age", strconv.Itoa(ageSeconds))
	w.Header().Set("Content-Type", fmt.Sprintf("%s", optimizedResponse.ImageFormat))
	w.Header().Set("Cache-Control", "public, max-age=7200, must-revalidate")
	w.Header().Set("Content-Security-Policy", "script-src 'none'; frame-src 'none'; sandbox;")
	w.Header().Set("Content-Length", strconv.Itoa(len(optimizedResponse.ImageData)))
	w.Header().Set("X-Cache", cacheMsg)
	w.Header().Set("Vary", "Accept")

	if optimizedResponse.ImageData == nil && optimizedResponse.Cache {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Write(optimizedResponse.ImageData)
	log.Println("---------------------------------------------------")
}
