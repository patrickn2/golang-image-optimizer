package httpserver

import (
	"log"
	"net/http"
	"strings"
)

func compressMiddleware(h func(w http.ResponseWriter, r *http.Request)) http.Handler {
	next := http.HandlerFunc(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Start", r.RequestURI)
		next.ServeHTTP(w, r)
		acceptEncoding := r.Header.Get("Accept-Encoding")
		var encoding string
		if strings.Contains(acceptEncoding, "zstd") {
			encoding = "zstd"
		} else if strings.Contains(acceptEncoding, "br") {
			encoding = "br"
		} else if strings.Contains(acceptEncoding, "gzip") {
			encoding = "gzip"
		}
		if w.Header().Get("Content-Type") == "image/svg+xml" && encoding != "" {
			w.Header().Set("Content-Encoding", encoding)

		}
		log.Println("End", r.RequestURI)
	})
}
