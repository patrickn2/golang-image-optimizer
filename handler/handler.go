package handler

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/patrickn2/go-image-optimizer/service"
)

type Handler struct {
	is *service.ImageService
}

func New(is *service.ImageService) *Handler {
	return &Handler{
		is: is,
	}
}

func (h *Handler) OptimizeImage(w http.ResponseWriter, r *http.Request) {
	imageUrl := r.URL.Query().Get("url")
	width := r.URL.Query().Get("w")
	quality := r.URL.Query().Get("q")
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	cacheControl := r.Header.Get("Cache-Control")

	intQuality, err := strconv.Atoi(quality)
	if err != nil {
		intQuality = 75
	}
	if intQuality < 0 {
		intQuality = 75
	}
	if intQuality > 100 {
		intQuality = 100
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
	// Check if url is valid
	_, err = url.ParseRequestURI(imageUrl)
	if err != nil {
		brokenImage, err := h.is.BrokenImage(r.Context(), intWidth)
		if err != nil {
			log.Printf("Error optimizing image: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(brokenImage)
		return
	}

	request := &service.OptimizeRequest{
		Ctx:             r.Context(),
		ImageUrl:        imageUrl,
		Width:           intWidth,
		Quality:         intQuality,
		IfModifiedSince: ifModifiedSince,
		CacheControl:    cacheControl,
	}

	optimizedResponse, err := h.is.Optimize(request)
	if err != nil {
		brokenImage, err := h.is.BrokenImage(r.Context(), intWidth)
		if err != nil {
			log.Printf("Error optimizing image: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(brokenImage)
		return
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
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=7200, must-revalidate")
	w.Header().Set("Content-Security-Policy", "script-src 'none'; frame-src 'none'; sandbox;")
	w.Header().Set("Content-Length", strconv.Itoa(len(optimizedResponse.ImageData)))
	w.Header().Set("X-Cache", cacheMsg)

	if optimizedResponse.ImageData == nil && optimizedResponse.Cache {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Write(optimizedResponse.ImageData)
}
