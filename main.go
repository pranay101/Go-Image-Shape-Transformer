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
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		nString := r.FormValue("numShapes")
		if nString == "" {
			renderNumShapeChoices(w, r, f, ext, Primitive.Mode(mode))
			return
		}

		numShapes, err := strconv.Atoi(nString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = numShapes
		http.Redirect(w, r, "/img/"+filepath.Base(f.Name()), http.StatusFound)

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
func renderNumShapeChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string, mode Primitive.Mode) {
	opts := []genOps{
		{N: 10, M: mode},
		{N: 50, M: mode},
		{N: 100, M: mode},
		{N: 150, M: mode},
	}

	_ = r
	imgs, err := genImages(rs, ext, opts...)

	if err != nil {
		http.Error(w, "Something went Wrong", http.StatusInternalServerError)
		return
	}

	html := `<html><body>
                {{range .}}
                <a href="/modify/{{.Name}}?mode={{.Mode}}&numShapes={{.NumShapes}}">
                    <img style="width: 240px;" src="/img/{{.Name}}">
                </a>
                {{end}}
                </body></html>`
	tpl := template.Must(template.New("").Parse(html))

	type dataStruct struct {
		Name      string
		Mode      Primitive.Mode
		NumShapes int
	}

	var data []dataStruct

	for i, img := range imgs {
		data = append(data, dataStruct{
			Name:      filepath.Base(img),
			Mode:      opts[i].M,
			NumShapes: opts[i].N,
		})
	}

	err = tpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func renderModeChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string) {
	opts := []genOps{
		{N: 10, M: Primitive.ModeBeziers},
		{N: 10, M: Primitive.ModeEllipse},
		{N: 10, M: Primitive.ModeRotatedRect},
		{N: 10, M: Primitive.ModeTriangle},
	}

	imgs, err := genImages(rs, ext, opts...)
	_ = r
	if err != nil {
		http.Error(w, "Something went Wrong", http.StatusInternalServerError)
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

	type dataStruct struct {
		Name string
		Mode Primitive.Mode
	}

	var data []dataStruct

	for i, img := range imgs {
		data = append(data, dataStruct{
			Name: filepath.Base(img),
			Mode: opts[i].M,
		})
	}

	err = tpl.Execute(w, data)
	if err != nil {
		panic(err)
	}

}

type genOps struct {
	N int
	M Primitive.Mode
}

func genImages(rs io.ReadSeeker, ext string, opts ...genOps) ([]string, error) {
	var ret []string
	for _, opt := range opts {
		rs.Seek(0, 0)
		f, err := genImage(rs, ext, opt.N, opt.M)
		if err != nil {
			return nil, err
		}

		ret = append(ret, f)
	}

	return ret, nil
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
