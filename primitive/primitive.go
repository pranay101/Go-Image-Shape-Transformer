package Primitive

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Mode int

const (
	ModeCombo Mode = iota
	ModeTriangle
	ModeRect
	ModeEllipse
	ModeRotateDrect
	ModeBeziers
	ModeRotatedEllipse
	ModePolygon
)

func WithNode(mode Mode) func() []string {
	return func() []string {
		return []string{"-m", fmt.Sprintf("%d", mode)}
	}
}

func Transform(image io.Reader, numShapes int, opts ...func() []string) (io.Reader, error) {
	in, err := tempfile("in_", "jpg")

	if err != nil {
		return nil, err
	}

	defer os.Remove(in.Name())

	out, err := tempfile("out_", "jpg")
	if err != nil {
		return nil, err
	}

	defer os.Remove(in.Name())

	// read image into in file
	_, err = io.Copy(in, image)

	if err != nil {
		return nil, err
	}

	//  run primitive w/ -i in.Name() -o out.Name()
	stdCombo, err := primitive(in.Name(), out.Name(), numShapes, ModeCombo)
	if err != nil {
		return nil, err
	}
	fmt.Println(stdCombo)

	// read out into a reader return reader. delete out
	var b = bytes.NewBuffer(nil)
	_, err = io.Copy(b, out)

	if err != nil {
		return nil, err
	}

	return b, nil

}

func primitive(inputFile string, outputFile string, shapeCount int, mode Mode) (string, error) {
	argStr := fmt.Sprintf("-i %s -o %s -n %d -m %d", inputFile, outputFile, shapeCount, mode)

	cmd := exec.Command("primitive", strings.Fields(argStr)...)
	b, err := cmd.CombinedOutput()

	return string(b), err
}

func tempfile(prefix, ext string) (*os.File, error) {
	in, err := os.CreateTemp("", prefix)
	if err != nil {
		return nil, errors.New("primitive: failed to create temporary file.")
	}

	defer os.Remove(in.Name())
	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
