package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	optimized, cache, err := h.is.Optimize(r.Context(), url, width, quality)
	if err != nil {
		if err == service.ErrInvalidImageWidth {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err == service.ErrInvalidImageUrl {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err == service.ErrInvalidQuality {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Error optimizing image: %v\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=7200, must-revalidate")
	w.Header().Set("Content-Security-Policy", "script-src 'none'; frame-src 'none'; sandbox;")
	w.Header().Set("Content-Length", strconv.Itoa(len(optimized)))
	w.Header().Set("X-Cache", fmt.Sprintf("%v", cache))
	w.Write(optimized)
}
