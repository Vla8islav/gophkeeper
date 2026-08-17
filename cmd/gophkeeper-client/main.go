package main

import (
	"fmt"
	"os"

	"github.com/Vla8islav/gophkeeper/internal/client"
)

func main() {
	if err := client.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
