package main

import (
	"io"
	"os"

	Primitive "Go-Image-Shape-Transformer/primitive"
)

func main() {
	f, err := os.Open("./inputImages/input.jpg")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	out, err := Primitive.Transform(f, 50)

	if err != nil {
		panic(err)
	}
	os.Remove("out.jpg")
	outFile, err := os.Create("outPutImages/out-1.jpg")

	if err != nil {
		panic(err)
	}

	io.Copy(outFile, out)
}
