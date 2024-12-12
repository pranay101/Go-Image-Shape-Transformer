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
	ModeCircle
	ModeRotatedRect
	ModeBeziers
	ModeRotatedEllipse
	ModePolygon
)

func WithNode(mode Mode) func() []string {
	return func() []string {
		return []string{"-m", fmt.Sprintf("%d", mode)}
	}
}

func Transform(image io.Reader, ext string, numShapes int, opts ...func() []string) (io.Reader, error) {

	var args []string
	for _, opt := range opts {
		args = append(args, opt()...)
	}

	in, err := tempfile("in_", ext)

	if err != nil {
		return nil, errors.New("primitive: failed to create temporary input file")
	}

	defer os.Remove(in.Name())

	out, err := tempfile("out_", ext)
	if err != nil {
		return nil, errors.New("primitive: failed to create temporary output file")
	}

	defer os.Remove(out.Name())

	// read image into in file
	_, err = io.Copy(in, image)

	if err != nil {
		return nil, errors.New("primitive: failed to copy image into temp input file")
	}

	//  run primitive w/ -i in.Name() -o out.Name()
	stdCombo, err := primitive(in.Name(), out.Name(), numShapes, args...)
	if err != nil {
		return nil, err
	}
	fmt.Println(stdCombo)

	// read out into a reader return reader. delete out
	var b = bytes.NewBuffer(nil)
	_, err = io.Copy(b, out)

	if err != nil {
		return nil, errors.New("primitive: Failed to copy output file into byte buffer")
	}

	return b, nil

}

func primitive(inputFile string, outputFile string, shapeCount int, args ...string) (string, error) {
	argStr := fmt.Sprintf("-i %s -o %s -n %d", inputFile, outputFile, shapeCount)
	args = append(strings.Fields(argStr), args...)

	cmd := exec.Command("primitive", args...)
	b, err := cmd.CombinedOutput()

	return string(b), err
}

func tempfile(prefix, ext string) (*os.File, error) {
	in, err := os.CreateTemp("", prefix)
	if err != nil {
		return nil, errors.New("primitive: failed to create temporary file")
	}

	defer os.Remove(in.Name())
	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
