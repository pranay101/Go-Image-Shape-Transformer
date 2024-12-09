package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("primitive", strings.Fields("-i input.jpg -o output.jpg -n 250 -m 6")...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
