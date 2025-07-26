package main

import (
	"fmt"
	"net/http"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := Run(); err != http.ErrServerClosed {
		panic(err)
	}
	fmt.Println("Server Shutdown gracefully")
}
