package main

import (
	"fmt"
	"os"
)

func main() {
	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "server startup failed: %v\n", err)
		os.Exit(1)
	}
}
