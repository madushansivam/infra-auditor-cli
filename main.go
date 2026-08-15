package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	target := flag.String("target", "", "the resource idemtified to audit")
	flag.Parse()

	if *target == "" {
		fmt.Println("Error: --target is required")
		os.Exit(1)

	}
	fmt.Printf("Auditing target: %s\n", *target)
}
