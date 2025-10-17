package main

import (
	"fmt"
	"os"

	"psjungle/internal/psjungle"
)

func main() {
	if err := psjungle.Run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
