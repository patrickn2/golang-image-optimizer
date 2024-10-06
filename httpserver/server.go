package httpserver

import (
	"log"
	"net/http"

	"github.com/patrickn2/go-image-optimizer/handler"
)

func Start(h *handler.Handler, port, imagePath string) {
	srv := http.NewServeMux()

	srv.HandleFunc("GET "+imagePath, h.OptimizeImage)

	log.Println("Listening on port", port)
	http.ListenAndServe(":"+port, srv)
}
