package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	Primitive "Go-Image-Shape-Transformer/primitive"

	"github.com/labstack/gommon/log"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	mux.HandleFunc("/modify/", func(w http.ResponseWriter, r *http.Request) {
		// imgPath := r.URL.Path[(len("/modify/")):]
		f, err := os.Open("./img/" + filepath.Base(r.URL.Path))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()
		ext := filepath.Ext(f.Name())
		modeString := r.FormValue("mode")
		if modeString == "" {
			renderModeChoices(w, r, f, ext)
			return
		}

		mode, err := strconv.Atoi(modeString)
		_ = mode
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "image/jpg")
		io.Copy(w, f)

	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("image")

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		ext := filepath.Ext(header.Filename)[1:]
		onDisk, err := tempfile("", ext)
		if err != nil {
			http.Error(w, "Something went Wrong", http.StatusInternalServerError)
			return
		}

		defer onDisk.Close()
		_, err = io.Copy(onDisk, file)

		if err != nil {
			http.Error(w, "Something went Wrong", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/modify/"+onDisk.Name(), http.StatusFound)

	})

	fs := http.FileServer(http.Dir("./img/"))
	mux.Handle("/img/", http.StripPrefix("/img", fs))
	log.Fatal(http.ListenAndServe(":8000", mux))
}

func renderModeChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string) {
	a, err := genImage(rs, ext, 33, Primitive.ModeBeziers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rs.Seek(0, 0)
	b, err := genImage(rs, ext, 33, Primitive.ModeCombo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rs.Seek(0, 0)
	c, err := genImage(rs, ext, 33, Primitive.ModeRotatedRect)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rs.Seek(0, 0)
	d, err := genImage(rs, ext, 33, Primitive.ModeTriangle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := `<html><body>
                {{range .}}
                <a href="/modify/{{.Name}}?mode={{.Mode}}">
                    <img style="width: 240px;" src="/img/{{.Name}}">
                </a>
                {{end}}
                </body></html>`
	tpl := template.Must(template.New("").Parse(html))

	data := []struct {
		Name string
		Mode Primitive.Mode
	}{
		{Name: filepath.Base(a), Mode: Primitive.ModeCircle},
		{Name: filepath.Base(b), Mode: Primitive.ModeRect},
		{Name: filepath.Base(c), Mode: Primitive.ModeEllipse},
		{Name: filepath.Base(d), Mode: Primitive.ModeTriangle},
	}
	err = tpl.Execute(w, data)
    if err != nil {
        panic(err)
    }
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
