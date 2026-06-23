package main

import (
	"os"
	"strconv"

	"github.com/kngght/asciiart/internal/service/art"
)

func main() {
	width, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	art.MakeASCIIArt(os.Args[1], width)
}
