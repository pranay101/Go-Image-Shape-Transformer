package main

import (
	"io"
	"net/http"
	"path/filepath"

	Primitive "Go-Image-Shape-Transformer/primitive"

	"github.com/labstack/gommon/log"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("image")

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()

		ext := filepath.Ext(header.Filename)[1:]
		_ = ext
		out, err := Primitive.Transform(file, ext, 50)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		switch ext {
		case "jpg":
			fallthrough
		case "jpeg":
			w.Header().Set("Content-Type", "image/jpg")
		case "png":
			w.Header().Set("Content-Type", "image/png")

		default:
			http.Error(w, "Invalid Image Type", http.StatusBadRequest)
		}

		io.Copy(w, out)
	})

	log.Fatal(http.ListenAndServe(":8000", mux))
}
