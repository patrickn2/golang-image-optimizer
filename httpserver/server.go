package httpserver

import (
	"log"
	"net/http"

	"github.com/CAFxX/httpcompression"
	"github.com/patrickn2/go-image-optimizer/handler"
)

func Start(h *handler.Handler, port, imagePath string) {
	contentType := httpcompression.ContentTypes([]string{"image/svg+xml"}, false)
	compress, err := httpcompression.DefaultAdapter(contentType)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("GET "+imagePath, compress(http.HandlerFunc(h.OptimizeImage)))

	log.Println("Listening on port", port)
	http.ListenAndServe(":"+port, nil)
}
