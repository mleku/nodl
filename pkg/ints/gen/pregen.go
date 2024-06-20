package main

import (
	"fmt"
	"os"
)

func main() {
	fh, err := os.Create("pkg/ints/base10k.txt")
	if err != nil {
		panic(err)
	}
	for i := range 10000 {
		fmt.Fprintf(fh, "%04d", i)
	}
}
