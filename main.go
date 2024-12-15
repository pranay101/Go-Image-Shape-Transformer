package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

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
		a, err := genImage(file, ext, 33, Primitive.ModeBeziers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Seek(0, 0)
		b, err := genImage(file, ext, 33, Primitive.ModeCombo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Seek(0, 0)
		c, err := genImage(file, ext, 33, Primitive.ModeRotatedRect)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Seek(0, 0)
		d, err := genImage(file, ext, 33, Primitive.ModeTriangle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		html := `<html><body>
                {{range .}}
                    <img styles="width: 240; display: inline-block" src="/{{.}}">
                {{end}}
                </body></html>`
		tpl := template.Must(template.New("").Parse(html))
		tpl.Execute(w, []string{a, b, c, d})
		redirectUrl := fmt.Sprintf("/%s", a)
		http.Redirect(w, r, redirectUrl, http.StatusFound)

	})

	fs := http.FileServer(http.Dir("./img/"))
	mux.Handle("/img/", http.StripPrefix("/img", fs))
	log.Fatal(http.ListenAndServe(":8000", mux))
}

func genImage(file io.Reader, ext string, numShapes int, mode Primitive.Mode) (string, error) {
	out, err := Primitive.Transform(file, ext, numShapes, Primitive.WithNode(mode))
	if err != nil {
		return "", err
	}
	outFile, err := tempfile("", ext)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	io.Copy(outFile, out)
	return outFile.Name(), nil

}
func tempfile(prefix, ext string) (*os.File, error) {
	in, err := os.CreateTemp("./img/", prefix)
	if err != nil {
		return nil, errors.New("main: failed to create temporary file")
	}

	defer in.Close()
	defer os.Remove(in.Name())
	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
