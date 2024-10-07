package handler

import (
	"log"
	"net/http"
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
	url := r.URL.Query().Get("url")
	width := r.URL.Query().Get("w")
	quality := r.URL.Query().Get("q")
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	cacheControl := r.Header.Get("Cache-Control")

	request := &service.OptimizeRequest{
		Ctx:             r.Context(),
		ImageUrl:        url,
		Width:           width,
		Quality:         quality,
		IfModifiedSince: ifModifiedSince,
		CacheControl:    cacheControl,
	}

	optimizedResponse := h.is.Optimize(request)
	if optimizedResponse.Error != nil {
		if optimizedResponse.Error == service.ErrInvalidImageWidth {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if optimizedResponse.Error == service.ErrInvalidImageUrl {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if optimizedResponse.Error == service.ErrInvalidQuality {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Error optimizing image: %v\n", optimizedResponse.Error)
		w.WriteHeader(http.StatusInternalServerError)
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

	lastModified := time.Now()
	if optimizedResponse.Modified != nil {
		lastModified = *optimizedResponse.Modified
	}
	w.Header().Set("Date", lastModified.Format(http.TimeFormat))
	w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))
	w.Header().Set("Age", strconv.Itoa(ageSeconds))
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=7200, must-revalidate")
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
